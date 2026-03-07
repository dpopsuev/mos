package artifact

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dpopsuev/mos/moslib/dsl"
)

// Relocation records a single artifact directory move.
type Relocation struct {
	ID   string
	Kind string
	From string
	To   string
}

// RelocateMisplacedArtifacts scans all artifact types and moves directories
// whose status does not match the active/archive placement.
func RelocateMisplacedArtifacts(root string) ([]Relocation, error) {
	reg, err := LoadRegistry(root)
	if err != nil {
		return nil, fmt.Errorf("relocate: %w", err)
	}

	mosDir := filepath.Join(root, MosDir)
	var relocations []Relocation

	for _, td := range reg.Types {
		if len(td.Lifecycle.ActiveStates) == 0 && len(td.Lifecycle.ArchiveStates) == 0 {
			continue
		}

		for _, pair := range []struct {
			srcSub, dstSub string
			shouldMove     func(string) bool
		}{
			{ActiveDir, ArchiveDir, td.IsArchiveStatus},
			{ArchiveDir, ActiveDir, func(s string) bool { return s != "" && !td.IsArchiveStatus(s) }},
		} {
			srcDir := filepath.Join(mosDir, td.Directory, pair.srcSub)
			entries, err := os.ReadDir(srcDir)
			if err != nil {
				continue
			}
			for _, e := range entries {
				if !e.IsDir() {
					continue
				}
				artPath := filepath.Join(srcDir, e.Name(), td.Kind+".mos")
				ab, err := dsl.ReadArtifact(artPath)
				if err != nil {
					continue
				}
				status, _ := dsl.FieldString(ab.Items, FieldStatus)
				if !pair.shouldMove(status) {
					continue
				}

				oldDir := filepath.Join(srcDir, e.Name())
				dstDir := filepath.Join(mosDir, td.Directory, pair.dstSub)
				newDir := filepath.Join(dstDir, e.Name())

				if err := os.MkdirAll(dstDir, DirPerm); err != nil {
					return relocations, fmt.Errorf("relocate %s: %w", e.Name(), err)
				}
				if err := os.Rename(oldDir, newDir); err != nil {
					return relocations, fmt.Errorf("relocate %s: %w", e.Name(), err)
				}
				relocations = append(relocations, Relocation{
					ID:   e.Name(),
					Kind: td.Kind,
					From: filepath.Join(td.Directory, pair.srcSub),
					To:   filepath.Join(td.Directory, pair.dstSub),
				})
			}
		}
	}
	return relocations, nil
}
