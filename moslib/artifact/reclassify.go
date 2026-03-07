package artifact

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dpopsuev/mos/moslib/dsl"
)

// ReclassifyResult describes the outcome of a reclassify operation.
type ReclassifyResult struct {
	OldID   string
	NewID   string
	OldKind string
	NewKind string
	NewPath string
}

// Reclassify moves an artifact from one kind to another, preserving all content.
// It generates a new ID with the target kind's prefix, writes the artifact
// under the target directory, and leaves a tombstone at the old location.
func Reclassify(root, id, toKind string) (*ReclassifyResult, error) {
	reg, err := LoadRegistry(root)
	if err != nil {
		return nil, fmt.Errorf("reclassify: %w", err)
	}

	fromKind, err := reg.ResolveKindFromID(id)
	if err != nil {
		return nil, fmt.Errorf("reclassify: cannot resolve source kind: %w", err)
	}
	if fromKind == toKind {
		return nil, fmt.Errorf("reclassify: artifact %q is already a %s", id, toKind)
	}

	srcTD, ok := reg.Types[fromKind]
	if !ok {
		return nil, fmt.Errorf("reclassify: source kind %q not in registry", fromKind)
	}
	dstTD, ok := reg.Types[toKind]
	if !ok {
		return nil, fmt.Errorf("reclassify: target kind %q not in registry", toKind)
	}
	if dstTD.Prefix == "" {
		return nil, fmt.Errorf("reclassify: target kind %q has no prefix for ID generation", toKind)
	}

	srcPath, err := FindGenericPath(root, srcTD, id)
	if err != nil {
		return nil, fmt.Errorf("reclassify: %w", err)
	}

	ab, err := dsl.ReadArtifact(srcPath)
	if err != nil {
		return nil, fmt.Errorf("reclassify: reading source: %w", err)
	}

	newID, err := NextIDForType(root, dstTD.Prefix, dstTD.Directory)
	if err != nil {
		return nil, fmt.Errorf("reclassify: generating new ID: %w", err)
	}

	ab.Kind = toKind
	ab.Name = newID

	dstDir := filepath.Join(root, MosDir, dstTD.Directory, ActiveDir, newID)
	dstPath := filepath.Join(dstDir, dstTD.Kind+".mos")

	if err := os.MkdirAll(dstDir, DirPerm); err != nil {
		return nil, fmt.Errorf("reclassify: creating target dir: %w", err)
	}

	file := &dsl.File{Artifact: ab}
	if err := writeArtifact(dstPath, file); err != nil {
		os.RemoveAll(dstDir)
		return nil, fmt.Errorf("reclassify: writing target: %w", err)
	}

	srcDir := filepath.Dir(srcPath)
	if err := writeTombstone(srcDir, srcTD.Kind, id, newID, toKind); err != nil {
		return nil, fmt.Errorf("reclassify: writing tombstone: %w", err)
	}

	return &ReclassifyResult{
		OldID:   id,
		NewID:   newID,
		OldKind: fromKind,
		NewKind: toKind,
		NewPath: dstPath,
	}, nil
}

func writeTombstone(dir, srcKind, oldID, newID, newKind string) error {
	for _, entry := range mustReadDir(dir) {
		name := entry.Name()
		if name == srcKind+".mos" {
			continue
		}
		os.Remove(filepath.Join(dir, name))
	}

	tombstone := &dsl.File{
		Artifact: &dsl.ArtifactBlock{
			Kind: srcKind,
			Name: oldID,
			Items: []dsl.Node{
				&dsl.Field{Key: "status", Value: &dsl.StringVal{Text: "reclassified"}},
				&dsl.Field{Key: "reclassified_to", Value: &dsl.StringVal{Text: newID}},
				&dsl.Field{Key: "reclassified_kind", Value: &dsl.StringVal{Text: newKind}},
			},
		},
	}

	path := filepath.Join(dir, srcKind+".mos")
	content := dsl.Format(tombstone, nil)
	return atomicWriteFile(path, []byte(content), FilePerm)
}

func mustReadDir(dir string) []os.DirEntry {
	entries, _ := os.ReadDir(dir)
	return entries
}
