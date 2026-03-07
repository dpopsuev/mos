package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dpopsuev/mos/moslib/dsl"
	"github.com/dpopsuev/mos/moslib/names"
)

// ProjectDef represents a project namespace from config.mos.
type ProjectDef struct {
	Name     string
	Prefix   string
	Sequence int
	Default  bool
}

// LoadProjects reads project blocks from .mos/config.mos.
func LoadProjects(root string) ([]ProjectDef, error) {
	configPath := filepath.Join(root, names.MosDir, names.ConfigFile)
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("reading config.mos: %w", err)
	}
	f, err := dsl.Parse(string(data), nil)
	if err != nil {
		return nil, fmt.Errorf("parsing config.mos: %w", err)
	}
	ab, ok := f.Artifact.(*dsl.ArtifactBlock)
	if !ok {
		return nil, fmt.Errorf("config.mos: invalid artifact structure")
	}

	var projects []ProjectDef
	for _, item := range ab.Items {
		blk, ok := item.(*dsl.Block)
		if !ok || blk.Name != "project" {
			continue
		}
		prefix, _ := dsl.FieldString(blk.Items, "prefix")
		seq, _ := dsl.FieldInt(blk.Items, "sequence")
		p := ProjectDef{
			Name:     blk.Title,
			Prefix:   prefix,
			Sequence: int(seq),
			Default:  dsl.FieldBool(blk.Items, "default"),
		}
		projects = append(projects, p)
	}
	return projects, nil
}

// FindDefaultProject returns the project with default = true, or nil.
func FindDefaultProject(projects []ProjectDef) *ProjectDef {
	for i := range projects {
		if projects[i].Default {
			return &projects[i]
		}
	}
	return nil
}

// FindProjectByPrefix returns the project whose prefix matches (case-insensitive).
// This enables --kind bug → project with prefix "BUG".
func FindProjectByPrefix(projects []ProjectDef, prefix string) *ProjectDef {
	upper := strings.ToUpper(prefix)
	for i := range projects {
		if strings.ToUpper(projects[i].Prefix) == upper {
			return &projects[i]
		}
	}
	return nil
}

// NextID generates the next ID for a project and atomically bumps the sequence.
// Format: PREFIX-YYYY-NNN (zero-padded to 3 digits).
func NextID(root, projectName string) (string, error) {
	configPath := filepath.Join(root, names.MosDir, names.ConfigFile)
	lockPath := configPath + ".lock"

	lock, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	for i := 0; err != nil && i < 100; i++ {
		time.Sleep(10 * time.Millisecond)
		lock, err = os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	}
	if err != nil {
		return "", fmt.Errorf("acquiring lock on config.mos: %w", err)
	}
	defer func() {
		lock.Close()
		os.Remove(lockPath)
	}()

	var id string
	if err := dsl.WithArtifact(configPath, func(ab *dsl.ArtifactBlock) error {
		var targetBlk *dsl.Block
		for _, item := range ab.Items {
			if blk, ok := item.(*dsl.Block); ok && blk.Name == "project" && blk.Title == projectName {
				targetBlk = blk
				break
			}
		}
		if targetBlk == nil {
			return fmt.Errorf("project %q not found in config.mos", projectName)
		}

		prefix, _ := dsl.FieldString(targetBlk.Items, "prefix")
		seq, seqOk := dsl.FieldInt(targetBlk.Items, "sequence")
		if prefix == "" {
			return fmt.Errorf("project %q has no prefix", projectName)
		}
		if !seqOk {
			return fmt.Errorf("project %q has no sequence field", projectName)
		}

		mosDir := filepath.Join(root, names.MosDir)
		artifactDirs := collectArtifactDirs(ab, mosDir)

		newSeq := int(seq) + 1
		year := time.Now().UTC().Format("2006")
		id = fmt.Sprintf("%s-%s-%03d", prefix, year, newSeq)

		for idExistsOnDisk(mosDir, id, artifactDirs) {
			newSeq++
			id = fmt.Sprintf("%s-%s-%03d", prefix, year, newSeq)
		}

		dsl.SetField(&targetBlk.Items, "sequence", &dsl.IntegerVal{Raw: fmt.Sprintf("%d", newSeq), Val: int64(newSeq)})
		return nil
	}); err != nil {
		return "", err
	}
	return id, nil
}

func collectArtifactDirs(ab *dsl.ArtifactBlock, mosDir string) []string {
	dirs := []string{filepath.Join(mosDir, "contracts")}
	for _, item := range ab.Items {
		blk, ok := item.(*dsl.Block)
		if !ok || blk.Name != "artifact_type" {
			continue
		}
		dir := blk.Title + "s"
		if d, ok := dsl.FieldString(blk.Items, "directory"); ok && d != "" {
			dir = d
		}
		dirs = append(dirs, filepath.Join(mosDir, dir))
	}
	return dirs
}

func idExistsOnDisk(mosDir, id string, artifactDirs []string) bool {
	for _, dir := range artifactDirs {
		for _, sub := range []string{names.ActiveDir, names.ArchiveDir} {
			candidate := filepath.Join(dir, sub, id)
			if _, err := os.Stat(candidate); err == nil {
				return true
			}
		}
	}
	return false
}

// NextIDForType generates the next ID for an artifact type by scanning
// existing IDs in the type's directory. Does not require a project block.
// Format: PREFIX-YYYY-NNN (zero-padded to 3 digits).
func NextIDForType(root, prefix, directory string) (string, error) {
	mosDir := filepath.Join(root, names.MosDir)
	baseDir := filepath.Join(mosDir, directory)
	year := time.Now().UTC().Format("2006")
	upperPrefix := strings.ToUpper(prefix)

	maxSeq := 0
	for _, sub := range []string{names.ActiveDir, names.ArchiveDir} {
		dir := filepath.Join(baseDir, sub)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			parts := strings.SplitN(e.Name(), "-", 3)
			if len(parts) != 3 || strings.ToUpper(parts[0]) != upperPrefix {
				continue
			}
			n, err := strconv.Atoi(parts[2])
			if err != nil {
				continue
			}
			if n > maxSeq {
				maxSeq = n
			}
		}
	}

	newSeq := maxSeq + 1
	id := fmt.Sprintf("%s-%s-%03d", upperPrefix, year, newSeq)
	return id, nil
}

// AddProject adds a new project block to config.mos.
func AddProject(root, name, prefix string) error {
	configPath := filepath.Join(root, names.MosDir, names.ConfigFile)
	return dsl.WithArtifact(configPath, func(ab *dsl.ArtifactBlock) error {
		for _, item := range ab.Items {
			blk, ok := item.(*dsl.Block)
			if ok && blk.Name == "project" && blk.Title == name {
				return fmt.Errorf("project %q already exists in config.mos", name)
			}
		}
		newBlk := &dsl.Block{
			Name:  "project",
			Title: name,
			Items: []dsl.Node{
				&dsl.Field{Key: "prefix", Value: &dsl.StringVal{Text: prefix}},
				&dsl.Field{Key: "sequence", Value: &dsl.IntegerVal{Raw: "0", Val: 0}},
			},
		}
		ab.Items = append(ab.Items, newBlk)
		return nil
	})
}
