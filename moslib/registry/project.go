package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dpopsuev/mos/moslib/names"
)

// ProjectDef represents a project namespace from config.mos.
type ProjectDef struct {
	Name     string
	Prefix   string
	Sequence int
	Default  bool
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
