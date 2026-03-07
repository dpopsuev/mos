package artifact

import (
	"os"
	"path/filepath"
	"testing"
)

// Compile-time interface compliance checks.
var _ GovernReader = (*reader)(nil)
var _ GovernEnforcer = (*enforcer)(nil)
var _ GovernReader = (GovernEnforcer)(nil)

func TestNewReader_ImplementsGovernReader(t *testing.T) {
	var _ GovernReader = NewReader()
}

func TestNewEnforcer_ImplementsGovernEnforcer(t *testing.T) {
	var _ GovernEnforcer = NewEnforcer()
}

func setupInterfaceRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	mosDir := filepath.Join(root, ".mos")
	for _, dir := range []string{
		"contracts/active/CON-001",
		"sprints/active/SPR-001",
	} {
		os.MkdirAll(filepath.Join(mosDir, dir), 0755)
	}
	os.WriteFile(filepath.Join(mosDir, "config.mos"), []byte(`config {
  mos { version = 1 }
  backend { type = "git" }
  governance { model = "bdfl" scope = "project" }
}`), 0644)
	os.WriteFile(filepath.Join(mosDir, "contracts", "active", "CON-001", "contract.mos"),
		[]byte(`contract "CON-001" { status = "draft" }`), 0644)
	os.WriteFile(filepath.Join(mosDir, "sprints", "active", "SPR-001", "sprint.mos"),
		[]byte(`sprint "SPR-001" { status = "planned" }`), 0644)
	return root
}

func TestGovernReader_Read(t *testing.T) {
	root := setupInterfaceRoot(t)
	r := NewReader()
	path := filepath.Join(root, ".mos", "contracts", "active", "CON-001", "contract.mos")
	ab, err := r.Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if ab.Name != "CON-001" {
		t.Errorf("expected name CON-001, got %s", ab.Name)
	}
}

func TestGovernReader_List(t *testing.T) {
	root := setupInterfaceRoot(t)
	r := NewReader()
	items, err := r.List(root, "contract", "")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) == 0 {
		t.Error("expected at least one contract")
	}
}

func TestGovernReader_ListKinds(t *testing.T) {
	root := setupInterfaceRoot(t)
	r := NewReader()
	kinds, err := r.ListKinds(root)
	if err != nil {
		t.Fatalf("ListKinds: %v", err)
	}
	if len(kinds) == 0 {
		t.Error("expected at least one kind")
	}
}

func TestGovernReader_Query(t *testing.T) {
	root := setupInterfaceRoot(t)
	r := NewReader()
	results, err := r.Query(root, QueryOpts{Kind: "contract"})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	found := false
	for _, qr := range results {
		if qr.ID == "CON-001" {
			found = true
		}
	}
	if !found {
		t.Error("expected CON-001 in query results")
	}
}

func TestGovernEnforcer_LoadSchemas(t *testing.T) {
	root := setupInterfaceRoot(t)
	e := NewEnforcer()
	schemas, err := e.LoadSchemas(root)
	if err != nil {
		t.Fatalf("LoadSchemas: %v", err)
	}
	if len(schemas) == 0 {
		t.Error("expected at least one schema")
	}
}
