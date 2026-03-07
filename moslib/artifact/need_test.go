package artifact

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dpopsuev/mos/moslib/dsl"
)

func TestCON033_NeedTypeInDefaultConfig(t *testing.T) {
	root := setupScaffold(t)
	reg := loadTestRegistry(t, root)

	need, ok := reg.Types["need"]
	if !ok {
		t.Fatal("need type not in registry from config.mos")
	}
	if need.Directory != "needs" {
		t.Errorf("need.Directory = %q, want needs", need.Directory)
	}
	if !need.Ledger {
		t.Error("need should have ledger enabled")
	}

	hasSensation := false
	hasUrgency := false
	for _, f := range need.Fields {
		if f.Name == "sensation" && f.Required {
			hasSensation = true
		}
		if f.Name == "urgency" && len(f.Enum) == 4 {
			hasUrgency = true
		}
	}
	if !hasSensation {
		t.Error("need missing required sensation field")
	}
	if !hasUrgency {
		t.Error("need missing urgency enum field")
	}
}

func TestCON033_CreateNeedWithRequiredFields(t *testing.T) {
	root := setupScaffold(t)
	reg := loadTestRegistry(t, root)
	td := reg.Types["need"]

	path, err := GenericCreate(root, td, "NEED-2026-001", map[string]string{
		"title":     "Faster CI feedback",
		"sensation": "CI pipelines take 45min; developers context-switch and lose flow",
		"urgency":   "high",
		"status":    "identified",
	})
	if err != nil {
		t.Fatalf("GenericCreate need: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("need file not found: %v", err)
	}
	assertParses(t, path)
}

func TestCON033_ListNeedsFilteredByStatus(t *testing.T) {
	root := setupScaffold(t)
	reg := loadTestRegistry(t, root)
	td := reg.Types["need"]

	GenericCreate(root, td, "NEED-001", map[string]string{
		"title": "Need A", "sensation": "Pain A", "urgency": "high", "status": "identified",
	})
	GenericCreate(root, td, "NEED-002", map[string]string{
		"title": "Need B", "sensation": "Pain B", "urgency": "low", "status": "validated",
	})

	items, err := GenericList(root, td, "identified")
	if err != nil {
		t.Fatalf("GenericList: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 identified need, got %d", len(items))
	}
	if items[0].ID != "NEED-001" {
		t.Errorf("expected NEED-001, got %s", items[0].ID)
	}
}

func TestCON033_NeedLifecycleTransitions(t *testing.T) {
	root := setupScaffold(t)
	reg := loadTestRegistry(t, root)
	td := reg.Types["need"]

	GenericCreate(root, td, "NEED-LC-001", map[string]string{
		"title": "Lifecycle test", "sensation": "Pain", "status": "identified",
	})

	if err := GenericUpdateStatus(root, td, "NEED-LC-001", "validated"); err != nil {
		t.Fatalf("transition to validated: %v", err)
	}
	if err := GenericUpdateStatus(root, td, "NEED-LC-001", "addressed"); err != nil {
		t.Fatalf("transition to addressed: %v", err)
	}
	if err := GenericUpdateStatus(root, td, "NEED-LC-001", "retired"); err != nil {
		t.Fatalf("transition to retired: %v", err)
	}

	items, _ := GenericList(root, td, "retired")
	if len(items) != 1 {
		t.Error("expected retired need in archive")
	}
}

func TestCON033_NeedScopeExcludesMachineReadable(t *testing.T) {
	root := setupScaffold(t)
	reg := loadTestRegistry(t, root)
	td := reg.Types["need"]

	GenericCreate(root, td, "NEED-SCOPE-001", map[string]string{
		"title":     "Scoped need",
		"sensation": "Build times are too long",
		"status":    "identified",
	})

	needPath, _ := FindGenericPath(root, td, "NEED-SCOPE-001")
	data, _ := os.ReadFile(needPath)
	f, _ := dsl.Parse(string(data), nil)
	ab := f.Artifact.(*dsl.ArtifactBlock)

	scopeBlock := &dsl.Block{
		Name: "scope",
		Items: []dsl.Node{
			&dsl.Field{Key: "includes", Value: &dsl.ListVal{Items: []dsl.Value{
				&dsl.StringVal{Text: "pipeline parallelization"},
				&dsl.StringVal{Text: "caching strategy"},
			}}},
			&dsl.Field{Key: "excludes", Value: &dsl.ListVal{Items: []dsl.Value{
				&dsl.StringVal{Text: "rewriting test suite"},
				&dsl.StringVal{Text: "migrating CI provider"},
			}}},
		},
	}
	ab.Items = append(ab.Items, scopeBlock)
	writeArtifact(needPath, f)

	data2, _ := os.ReadFile(needPath)
	f2, err := dsl.Parse(string(data2), nil)
	if err != nil {
		t.Fatalf("re-parse failed: %v", err)
	}
	ab2 := f2.Artifact.(*dsl.ArtifactBlock)

	foundScope := false
	for _, item := range ab2.Items {
		blk, ok := item.(*dsl.Block)
		if !ok || blk.Name != "scope" {
			continue
		}
		foundScope = true
		for _, fi := range blk.Items {
			field, ok := fi.(*dsl.Field)
			if !ok {
				continue
			}
			if field.Key == "excludes" {
				lv, ok := field.Value.(*dsl.ListVal)
				if !ok || len(lv.Items) != 2 {
					t.Errorf("excludes should have 2 items, got %v", field.Value)
				}
			}
		}
	}
	if !foundScope {
		t.Error("scope block not found after write")
	}
}

func TestCON033_NeedDirectoriesCreatedByInit(t *testing.T) {
	root := setupScaffold(t)
	for _, sub := range []string{"active", "archive"} {
		dir := filepath.Join(root, ".mos", "needs", sub)
		if _, err := os.Stat(dir); err != nil {
			t.Errorf("needs/%s directory not created by init: %v", sub, err)
		}
	}
}

// --- CON-2026-034: Architecture -- The Abstraction Primitive (Native DSL) ---
