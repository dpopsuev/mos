package artifact

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/dpopsuev/mos/moslib/dsl"
	"github.com/dpopsuev/mos/moslib/schema"
)

// GenericInfo represents a discovered instance of a custom artifact type.
type GenericInfo struct {
	ID     string
	Kind   string
	Title  string
	Status string
	Path   string
}

// genericArtifact implements the Artifact interface for custom types.
type genericArtifact struct {
	typeDef ArtifactTypeDef
}

func (g genericArtifact) Kind() string { return g.typeDef.Kind }

func (g genericArtifact) ResolvePath(root string, ab *dsl.ArtifactBlock) (string, error) {
	id := ab.Name
	if id == "" {
		return "", fmt.Errorf("%s artifact must have an ID", g.typeDef.Kind)
	}
	status := artifactField(ab, FieldStatus)
	if status == "" && len(g.typeDef.Lifecycle.ActiveStates) > 0 {
		status = g.typeDef.Lifecycle.ActiveStates[0]
	}
	subDir := ActiveDir
	if g.typeDef.IsArchiveStatus(status) {
		subDir = ArchiveDir
	}
	return filepath.Join(root, MosDir, g.typeDef.Directory, subDir, id, g.typeDef.Kind+".mos"), nil
}

func (g genericArtifact) FindExisting(root string, id string) (string, error) {
	return FindGenericPath(root, g.typeDef, id)
}

func (g genericArtifact) Validate(path, _ string) error {
	_, err := dsl.ReadArtifact(path)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	return nil
}

// FindGenericArtifactPath returns the file path for a generic artifact instance.
func FindGenericArtifactPath(root string, td ArtifactTypeDef, id string) (string, error) {
	return FindGenericPath(root, td, id)
}

func FindGenericPath(root string, td ArtifactTypeDef, id string) (string, error) {
	for _, sub := range []string{ActiveDir, ArchiveDir} {
		p := filepath.Join(root, MosDir, td.Directory, sub, id, td.Kind+".mos")
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	if resolved, err := resolveSlug(root, td.Directory, td.Kind, id); err == nil {
		return resolved, nil
	}
	return "", fmt.Errorf("%s %q not found", td.Kind, id)
}

// resolveSlug scans artifact directories for an artifact whose slug field
// matches the given value. Returns the path to the matching artifact file.
func resolveSlug(root, directory, kind, slug string) (string, error) {
	baseDir := filepath.Join(root, MosDir, directory)
	for _, sub := range []string{ActiveDir, ArchiveDir} {
		subDir := filepath.Join(baseDir, sub)
		entries, err := os.ReadDir(subDir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			p := filepath.Join(subDir, e.Name(), kind+".mos")
			ab, err := dsl.ReadArtifact(p)
			if err != nil {
				continue
			}
			if s, ok := dsl.FieldString(ab.Items, "slug"); ok && s == slug {
				return p, nil
			}
		}
	}
	return "", fmt.Errorf("no %s with slug %q", kind, slug)
}

// GenericCreate creates a new instance of a custom artifact type.
// GenericCreateWithTemplate creates an artifact, optionally merging from a template.
func GenericCreateWithTemplate(root string, td ArtifactTypeDef, id string, fields map[string]string, templateName string) (string, error) {
	path, err := GenericCreate(root, td, id, fields)
	if err != nil {
		return "", err
	}
	if templateName == "" {
		return path, nil
	}
	tmpl, err := LoadTemplate(root, templateName)
	if err != nil {
		return "", fmt.Errorf("template merge: %w", err)
	}
	return path, dsl.WithArtifact(path, func(ab *dsl.ArtifactBlock) error {
		MergeTemplate(ab, tmpl)
		return nil
	})
}

func GenericCreate(root string, td ArtifactTypeDef, id string, fields map[string]string) (string, error) {
	mosDir := filepath.Join(root, MosDir)
	if _, err := os.Stat(mosDir); err != nil {
		return "", fmt.Errorf(".mos/ directory not found; run mos init first")
	}

	status := fields[FieldStatus]
	if status == "" && len(td.Lifecycle.ActiveStates) > 0 {
		status = td.Lifecycle.ActiveStates[0]
	}

	if !td.IsValidStatus(status) {
		return "", fmt.Errorf("invalid status %q for %s; valid states: %s",
			status, td.Kind, strings.Join(td.AllStates(), ", "))
	}

	subDir := ActiveDir
	if td.IsArchiveStatus(status) {
		subDir = ArchiveDir
	}

	targetDir := filepath.Join(mosDir, td.Directory, subDir, id)
	targetPath := filepath.Join(targetDir, td.Kind+".mos")

	if _, err := os.Stat(targetPath); err == nil {
		return "", fmt.Errorf("%s %q already exists", td.Kind, id)
	}

	// Deterministic field ordering: title and status first, then alphabetical
	keys := slices.Sorted(maps.Keys(fields))

	var items []dsl.Node
	for _, priorityKey := range []string{FieldTitle, FieldStatus} {
		if val, ok := fields[priorityKey]; ok {
			items = append(items, &dsl.Field{
				Key:   priorityKey,
				Value: &dsl.StringVal{Text: val},
			})
		}
	}
	for _, k := range keys {
		if k == FieldTitle || k == FieldStatus {
			continue
		}
		items = append(items, &dsl.Field{
			Key:   k,
			Value: &dsl.StringVal{Text: fields[k]},
		})
	}

	// Ensure status is present
	if !dsl.HasField(items, FieldStatus) && status != "" {
		items = append([]dsl.Node{&dsl.Field{
			Key:   FieldStatus,
			Value: &dsl.StringVal{Text: status},
		}}, items...)
	}

	file := &dsl.File{
		Artifact: &dsl.ArtifactBlock{
			Kind:  td.Kind,
			Name:  id,
			Items: items,
		},
	}

	if err := os.MkdirAll(targetDir, DirPerm); err != nil {
		return "", fmt.Errorf("creating directory: %w", err)
	}
	if err := writeArtifact(targetPath, file); err != nil {
		return "", fmt.Errorf("writing artifact: %w", err)
	}

	return targetPath, nil
}

// GenericList lists instances of a custom artifact type.
func GenericList(root string, td ArtifactTypeDef, statusFilter string) ([]GenericInfo, error) {
	mosDir := filepath.Join(root, MosDir)
	var results []GenericInfo

	for _, sub := range []string{ActiveDir, ArchiveDir} {
		dir := filepath.Join(mosDir, td.Directory, sub)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			artPath := filepath.Join(dir, e.Name(), td.Kind+".mos")
			if _, err := os.Stat(artPath); err != nil {
				continue
			}
			info, err := readGenericInfo(td, e.Name(), artPath)
			if err != nil {
				continue
			}
			if statusFilter != "" && info.Status != statusFilter {
				continue
			}
			results = append(results, info)
		}
	}

	return results, nil
}

func readGenericInfo(td ArtifactTypeDef, id, path string) (GenericInfo, error) {
	ab, err := dsl.ReadArtifact(path)
	if err != nil {
		return GenericInfo{}, err
	}
	title, _ := dsl.FieldString(ab.Items, FieldTitle)
	status, _ := dsl.FieldString(ab.Items, FieldStatus)
	return GenericInfo{
		ID:     id,
		Kind:   td.Kind,
		Title:  title,
		Status: status,
		Path:   path,
	}, nil
}

// GenericShow returns formatted content of a custom artifact instance.
func GenericShow(root string, td ArtifactTypeDef, id string) (string, error) {
	path, err := FindGenericPath(root, td, id)
	if err != nil {
		return "", fmt.Errorf("GenericShow: %w", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("GenericShow: %w", err)
	}
	return string(data), nil
}

// GenericDelete removes a custom artifact instance.
func GenericDelete(root string, td ArtifactTypeDef, id string) error {
	path, err := FindGenericPath(root, td, id)
	if err != nil {
		return fmt.Errorf("GenericDelete: %w", err)
	}
	return os.RemoveAll(filepath.Dir(path))
}

// GenericUpdateStatus updates the status of a custom artifact instance.
func GenericUpdateStatus(root string, td ArtifactTypeDef, id, newStatus string) error {
	if !td.IsValidStatus(newStatus) {
		return fmt.Errorf("invalid status %q for %s; valid states: %s",
			newStatus, td.Kind, strings.Join(td.AllStates(), ", "))
	}

	path, err := FindGenericPath(root, td, id)
	if err != nil {
		return fmt.Errorf("GenericUpdateStatus: %w", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("GenericUpdateStatus: %w", err)
	}

	f, err := dsl.Parse(string(data), nil) // writes to different path on lifecycle move: cannot migrate to WithArtifact
	if err != nil {
		return fmt.Errorf("GenericUpdateStatus: %w", err)
	}

	ab, ok := f.Artifact.(*dsl.ArtifactBlock)
	if !ok {
		return fmt.Errorf("invalid artifact structure")
	}

	currentStatus, _ := dsl.FieldString(ab.Items, FieldStatus)
	if err := evaluateTransitionGates(root, td, id, ab, currentStatus, newStatus); err != nil {
		return fmt.Errorf("GenericUpdateStatus: %w", err)
	}

	dsl.SetField(&ab.Items, FieldStatus, &dsl.StringVal{Text: newStatus})

	newSubDir := ActiveDir
	if td.IsArchiveStatus(newStatus) {
		newSubDir = ArchiveDir
	}

	mosDir := filepath.Join(root, MosDir)
	newDir := filepath.Join(mosDir, td.Directory, newSubDir, id)
	newPath := filepath.Join(newDir, td.Kind+".mos")

	if err := os.MkdirAll(newDir, DirPerm); err != nil {
		return fmt.Errorf("GenericUpdateStatus: %w", err)
	}
	if err := writeArtifact(newPath, f); err != nil {
		return fmt.Errorf("GenericUpdateStatus: %w", err)
	}

	oldDir := filepath.Dir(path)
	if oldDir != newDir {
		os.RemoveAll(oldDir)
	}

	return nil
}

// GenericUpdate updates fields on an existing custom artifact instance.
// It updates existing fields in-place and inserts new ones at the end.
// If the new status requires a lifecycle move (active↔archive), the directory is relocated.
// validateLinkFields checks that link field values reference valid artifact
// kinds (when RefKind is set) and that the target artifacts exist.
func validateLinkFields(root string, td ArtifactTypeDef, fields map[string]string) error {
	reg, err := LoadRegistry(root)
	if err != nil {
		return nil
	}

	fieldDefs := make(map[string]schema.FieldSchema)
	for _, fd := range td.Fields {
		fieldDefs[fd.Name] = fd
	}

	for fieldName, value := range fields {
		if value == "" {
			continue
		}
		fd, ok := fieldDefs[fieldName]
		if !ok || !fd.Link {
			continue
		}
		if fd.RefKind == "" {
			continue
		}
		targetTD, ok := reg.Types[fd.RefKind]
		if !ok {
			continue
		}
		if _, err := FindGenericPath(root, targetTD, value); err != nil {
			if fd.RefKind == "contract" {
				if _, cerr := FindContractPath(root, value); cerr == nil {
					continue
				}
			}
			return fmt.Errorf("field %q must reference a %s, but %q was not found as a %s",
				fieldName, fd.RefKind, value, fd.RefKind)
		}
	}
	return nil
}

func GenericUpdate(root string, td ArtifactTypeDef, id string, fields map[string]string) error {
	path, err := FindGenericPath(root, td, id)
	if err != nil {
		return fmt.Errorf("GenericUpdate: %w", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("GenericUpdate: %w", err)
	}

	f, err := dsl.Parse(string(data), nil) // writes to different path on lifecycle move: cannot migrate to WithArtifact
	if err != nil {
		return fmt.Errorf("GenericUpdate: %w", err)
	}

	ab, ok := f.Artifact.(*dsl.ArtifactBlock)
	if !ok {
		return fmt.Errorf("invalid artifact structure")
	}

	if newStatus, ok := fields[FieldStatus]; ok {
		if !td.IsValidStatus(newStatus) {
			return fmt.Errorf("invalid status %q for %s; valid states: %s",
				newStatus, td.Kind, strings.Join(td.AllStates(), ", "))
		}
	}

	if err := validateLinkFields(root, td, fields); err != nil {
		return fmt.Errorf("GenericUpdate: %w", err)
	}

	for k, v := range fields {
		dsl.SetField(&ab.Items, k, &dsl.StringVal{Text: v})
	}

	newSubDir := ActiveDir
	status, _ := dsl.FieldString(ab.Items, FieldStatus)
	if td.IsArchiveStatus(status) {
		newSubDir = ArchiveDir
	}

	mosDir := filepath.Join(root, MosDir)
	newDir := filepath.Join(mosDir, td.Directory, newSubDir, id)
	newPath := filepath.Join(newDir, td.Kind+".mos")

	if err := os.MkdirAll(newDir, DirPerm); err != nil {
		return fmt.Errorf("GenericUpdate: %w", err)
	}
	if err := writeArtifact(newPath, f); err != nil {
		return fmt.Errorf("GenericUpdate: %w", err)
	}

	oldDir := filepath.Dir(path)
	if oldDir != newDir {
		os.RemoveAll(oldDir)
	}

	return nil
}
