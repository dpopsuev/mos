package artifact

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dpopsuev/mos/moslib/dsl"
	"github.com/dpopsuev/mos/moslib/schema"
)

// MigrationDiff describes a single instance that needs migration.
type MigrationDiff struct {
	Kind     string
	ID       string
	Path     string
	Missing  []FieldDefault // fields that need to be backfilled
	UpToDate bool
}

// FieldDefault is a field name + default value pair to be inserted.
type FieldDefault struct {
	Name    string
	Default string
}

// ComputeMigration walks all artifact instances for every type in the registry
// and checks for missing fields that have default values defined.
func ComputeMigration(root string, reg *Registry) ([]MigrationDiff, error) {
	var diffs []MigrationDiff

	for _, td := range reg.Types {
		fieldsWithDefaults := fieldsHavingDefaults(td.Fields)
		if len(fieldsWithDefaults) == 0 {
			continue
		}

		instances, err := discoverInstances(root, td)
		if err != nil {
			continue
		}

		for _, inst := range instances {
			existing, err := existingFieldNames(inst.Path)
			if err != nil {
				continue
			}

			var missing []FieldDefault
			for _, fd := range fieldsWithDefaults {
				if _, ok := existing[fd.Name]; !ok {
					missing = append(missing, FieldDefault{Name: fd.Name, Default: fd.Default})
				}
			}

			diff := MigrationDiff{
				Kind:     td.Kind,
				ID:       inst.ID,
				Path:     inst.Path,
				Missing:  missing,
				UpToDate: len(missing) == 0,
			}
			diffs = append(diffs, diff)
		}
	}
	return diffs, nil
}

// ApplyMigration writes the missing default fields into each artifact instance.
func ApplyMigration(diffs []MigrationDiff) (int, error) {
	applied := 0
	for _, diff := range diffs {
		if diff.UpToDate {
			continue
		}
		if err := backfillFields(diff.Path, diff.Missing); err != nil {
			return applied, fmt.Errorf("migrating %s (%s): %w", diff.ID, diff.Path, err)
		}
		applied++
	}
	return applied, nil
}

// FormatMigrationPlan returns a human-readable summary of pending migrations.
func FormatMigrationPlan(diffs []MigrationDiff) string {
	var b strings.Builder
	pending := 0
	upToDate := 0
	for _, d := range diffs {
		if d.UpToDate {
			upToDate++
			continue
		}
		pending++
		fmt.Fprintf(&b, "%s %s:\n", d.Kind, d.ID)
		for _, m := range d.Missing {
			fmt.Fprintf(&b, "  + %s = %q\n", m.Name, m.Default)
		}
	}
	if pending == 0 {
		return "All instances are up to date.\n"
	}
	fmt.Fprintf(&b, "\n%d instance(s) need migration, %d up to date.\n", pending, upToDate)
	return b.String()
}

type instanceRef struct {
	ID   string
	Path string
}

func fieldsHavingDefaults(fields []schema.FieldSchema) []schema.FieldSchema {
	var result []schema.FieldSchema
	for _, f := range fields {
		if f.Default != "" {
			result = append(result, f)
		}
	}
	return result
}

func discoverInstances(root string, td ArtifactTypeDef) ([]instanceRef, error) {
	var refs []instanceRef
	base := filepath.Join(root, MosDir, td.Directory)
	for _, sub := range []string{ActiveDir, ArchiveDir} {
		dir := filepath.Join(base, sub)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			p := filepath.Join(dir, e.Name(), td.Kind+".mos")
			if _, err := os.Stat(p); err == nil {
				refs = append(refs, instanceRef{ID: e.Name(), Path: p})
			}
		}
	}
	return refs, nil
}

func existingFieldNames(path string) (map[string]bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	f, err := dsl.Parse(string(data), nil)
	if err != nil {
		return nil, err
	}
	ab, ok := f.Artifact.(*dsl.ArtifactBlock)
	if !ok {
		return nil, fmt.Errorf("not an artifact block")
	}
	names := make(map[string]bool)
	for _, item := range ab.Items {
		if field, ok := item.(*dsl.Field); ok {
			names[field.Key] = true
		}
	}
	return names, nil
}

func backfillFields(path string, fields []FieldDefault) error {
	return dsl.WithArtifact(path, func(ab *dsl.ArtifactBlock) error {
		for _, fd := range fields {
			dsl.SetField(&ab.Items, fd.Name, &dsl.StringVal{Text: fd.Default})
		}
		return nil
	})
}
