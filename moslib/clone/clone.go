package clone

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dpopsuev/mos/moslib/dsl"
	"github.com/dpopsuev/mos/moslib/names"
	"github.com/dpopsuev/mos/moslib/registry"
)

// Opts controls clone behavior.
type Opts struct {
	From   string
	Kinds  []string
	State  string
	Group  string
}

// Result describes a single cloned artifact.
type Result struct {
	OldID string
	NewID string
	Title string
	Kind  string
}

// GenericCreateFn abstracts artifact creation.
type GenericCreateFn func(root string, td registry.ArtifactTypeDef, id string, fields map[string]string) (string, error)

// FindPathFn locates an artifact's filesystem path.
type FindPathFn func(root string, td registry.ArtifactTypeDef, id string) (string, error)

// FieldStrFn reads a string field from a DSL node list.
type FieldStrFn func(items []dsl.Node, key string) string

// Run clones artifacts from a source project into the destination project.
func Run(dstRoot string, opts Opts, create GenericCreateFn, findPath FindPathFn, fieldStr FieldStrFn) ([]Result, map[string]string, error) {
	srcRoot := opts.From
	srcMos := filepath.Join(srcRoot, names.MosDir)
	if _, err := os.Stat(srcMos); err != nil {
		return nil, nil, fmt.Errorf("source project %s has no .mos/ directory", srcRoot)
	}

	dstReg, err := registry.LoadRegistry(dstRoot)
	if err != nil {
		return nil, nil, fmt.Errorf("load destination registry: %w", err)
	}
	srcReg, err := registry.LoadRegistry(srcRoot)
	if err != nil {
		return nil, nil, fmt.Errorf("load source registry: %w", err)
	}

	kinds := opts.Kinds
	if len(kinds) == 0 {
		for k := range srcReg.Types {
			kinds = append(kinds, k)
		}
	}

	idMap := make(map[string]string)
	var results []Result

	for _, kind := range kinds {
		srcTD, ok := srcReg.Types[kind]
		if !ok {
			continue
		}
		dstTD, ok := dstReg.Types[kind]
		if !ok {
			continue
		}

		srcDir := filepath.Join(srcMos, srcTD.Directory, names.ActiveDir)
		entries, err := os.ReadDir(srcDir)
		if err != nil {
			continue
		}

		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			artFile := filepath.Join(srcDir, e.Name(), kind+".mos")
			data, err := os.ReadFile(artFile)
			if err != nil {
				continue
			}
			f, err := dsl.Parse(string(data), nil)
			if err != nil {
				continue
			}
			ab, ok := f.Artifact.(*dsl.ArtifactBlock)
			if !ok {
				continue
			}

			if opts.State != "" {
				state := fieldStr(ab.Items, "state")
				if state != opts.State {
					continue
				}
			}
			if opts.Group != "" {
				group := fieldStr(ab.Items, "group")
				if group != opts.Group {
					continue
				}
			}

			fields := ExtractFields(ab)

			genID, err := registry.NextIDForType(dstRoot, dstTD.Prefix, dstTD.Directory)
			if err != nil {
				continue
			}
			newID, err := create(dstRoot, dstTD, genID, fields)
			if err != nil {
				continue
			}

			idMap[e.Name()] = newID

			newPath, err := findPath(dstRoot, dstTD, newID)
			if err == nil {
				CopySpecBlocks(artFile, newPath)
			}

			results = append(results, Result{
				OldID: e.Name(),
				NewID: newID,
				Title: fields["title"],
				Kind:  kind,
			})
		}
	}

	if len(idMap) > 0 {
		RemapReferences(dstRoot, dstReg, idMap, findPath)
	}

	return results, idMap, nil
}

// ExtractFields reads top-level scalar fields from an artifact block.
func ExtractFields(ab *dsl.ArtifactBlock) map[string]string {
	fields := make(map[string]string)
	for _, item := range ab.Items {
		f, ok := item.(*dsl.Field)
		if !ok {
			continue
		}
		switch v := f.Value.(type) {
		case *dsl.StringVal:
			fields[f.Key] = v.Text
		case *dsl.BoolVal:
			if v.Val {
				fields[f.Key] = "true"
			} else {
				fields[f.Key] = "false"
			}
		}
	}
	if _, ok := fields["status"]; !ok {
		fields["status"] = "draft"
	}
	return fields
}

// CopySpecBlocks copies spec and section blocks from a source artifact file
// into a destination artifact.
func CopySpecBlocks(srcPath, dstPath string) {
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return
	}
	f, err := dsl.Parse(string(data), nil)
	if err != nil {
		return
	}
	ab, ok := f.Artifact.(*dsl.ArtifactBlock)
	if !ok {
		return
	}

	_ = dsl.WithArtifact(dstPath, func(dstAB *dsl.ArtifactBlock) error {
		for _, item := range ab.Items {
			switch v := item.(type) {
			case *dsl.SpecBlock:
				dstAB.Items = append(dstAB.Items, v)
			case *dsl.Block:
				if v.Name == "section" {
					dstAB.Items = append(dstAB.Items, v)
				}
			}
		}
		return nil
	})
}

// RemapReferences updates cross-references in cloned artifacts using the ID map.
func RemapReferences(root string, reg *registry.Registry, idMap map[string]string, findPath FindPathFn) int {
	linkFields := []string{"justifies", "satisfies", "implements", "documents", "addresses", "depends_on", "parent", "sprint", "batch", "derives_from"}
	remapped := 0

	for _, newID := range idMap {
		kind, err := reg.ResolveKindFromID(newID)
		if err != nil {
			continue
		}
		td, ok := reg.Types[kind]
		if !ok {
			continue
		}
		artPath, err := findPath(root, td, newID)
		if err != nil {
			continue
		}

		_ = dsl.WithArtifact(artPath, func(ab *dsl.ArtifactBlock) error {
			for _, item := range ab.Items {
				f, ok := item.(*dsl.Field)
				if !ok {
					continue
				}
				isLink := false
				for _, lf := range linkFields {
					if f.Key == lf {
						isLink = true
						break
					}
				}
				if !isLink {
					continue
				}
				sv, ok := f.Value.(*dsl.StringVal)
				if !ok {
					continue
				}
				if newRef, exists := idMap[sv.Text]; exists {
					sv.Text = newRef
					remapped++
				}
			}
			return nil
		})
	}
	return remapped
}
