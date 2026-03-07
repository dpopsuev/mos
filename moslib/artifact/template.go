package artifact

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dpopsuev/mos/moslib/dsl"
)

const templatesDir = "templates"

// LoadTemplate reads a template from .mos/templates/<name>.mos and returns
// its parsed artifact block. Templates are regular .mos files whose blocks
// and fields serve as scaffolding for new artifacts.
func LoadTemplate(root, name string) (*dsl.ArtifactBlock, error) {
	path := filepath.Join(root, MosDir, templatesDir, name+".mos")
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("template %q not found: %w", name, err)
	}
	ab, err := dsl.ReadArtifact(path)
	if err != nil {
		return nil, fmt.Errorf("reading template %q: %w", name, err)
	}
	return ab, nil
}

// MergeTemplate copies blocks and fields from tmpl into target.
// Fields already present in target are not overwritten.
// Blocks (sections, features, etc.) are appended from the template.
func MergeTemplate(target, tmpl *dsl.ArtifactBlock) {
	existingFields := make(map[string]bool)
	for _, item := range target.Items {
		if f, ok := item.(*dsl.Field); ok {
			existingFields[f.Key] = true
		}
	}

	for _, item := range tmpl.Items {
		switch v := item.(type) {
		case *dsl.Field:
			if !existingFields[v.Key] {
				target.Items = append(target.Items, v)
			}
		case *dsl.Block, *dsl.FeatureBlock, *dsl.SpecBlock:
			target.Items = append(target.Items, item)
		}
	}
}
