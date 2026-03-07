package artifact

import (
	"os"
	"path/filepath"
	"testing"
)

func setupQueryWorkspace(t *testing.T) string {
	t.Helper()
	root := setupScaffold(t)
	seedProjectConfig(t, root)

	mos := filepath.Join(root, ".mos")
	active := filepath.Join(mos, "contracts", "active")

	c1 := filepath.Join(active, "CON-2026-001")
	os.MkdirAll(c1, 0755)
	os.WriteFile(filepath.Join(c1, "contract.mos"), []byte(`contract "CON-2026-001" {
  title = "Alpha"
  status = "active"
  sprint = "SPR-2026-001"
  justifies = "NEED-2026-001"
  labels = ["infra", "urgent"]
}`), 0644)

	c2 := filepath.Join(active, "CON-2026-002")
	os.MkdirAll(c2, 0755)
	os.WriteFile(filepath.Join(c2, "contract.mos"), []byte(`contract "CON-2026-002" {
  title = "Beta"
  status = "draft"
  sprint = "SPR-2026-001"
  labels = ["ui"]
}`), 0644)

	c3 := filepath.Join(active, "CON-2026-003")
	os.MkdirAll(c3, 0755)
	os.WriteFile(filepath.Join(c3, "contract.mos"), []byte(`contract "CON-2026-003" {
  title = "Gamma"
  status = "active"
  labels = ["infra"]
}`), 0644)

	return root
}

func TestQueryAll(t *testing.T) {
	root := setupQueryWorkspace(t)

	results, err := QueryArtifacts(root, QueryOpts{})
	if err != nil {
		t.Fatalf("QueryArtifacts: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
}

func TestQueryByStatus(t *testing.T) {
	root := setupQueryWorkspace(t)

	results, err := QueryArtifacts(root, QueryOpts{Status: "active"})
	if err != nil {
		t.Fatalf("QueryArtifacts: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 active results, got %d", len(results))
	}
}

func TestQueryByLabels(t *testing.T) {
	root := setupQueryWorkspace(t)

	results, err := QueryArtifacts(root, QueryOpts{Labels: []string{"ui"}})
	if err != nil {
		t.Fatalf("QueryArtifacts: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result with label 'ui', got %d", len(results))
	}
	if len(results) > 0 && results[0].ID != "CON-2026-002" {
		t.Errorf("expected CON-2026-002, got %s", results[0].ID)
	}
}

func TestQueryReferences(t *testing.T) {
	root := setupQueryWorkspace(t)

	results, err := QueryArtifacts(root, QueryOpts{References: "NEED-2026-001"})
	if err != nil {
		t.Fatalf("QueryArtifacts: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result referencing NEED-2026-001, got %d", len(results))
	}
}

func TestQueryReferencesBySprint(t *testing.T) {
	root := setupQueryWorkspace(t)

	results, err := QueryArtifacts(root, QueryOpts{References: "SPR-2026-001"})
	if err != nil {
		t.Fatalf("QueryArtifacts: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results referencing SPR-2026-001, got %d", len(results))
	}
}

func TestQueryCount(t *testing.T) {
	root := setupQueryWorkspace(t)

	results, err := QueryArtifacts(root, QueryOpts{})
	if err != nil {
		t.Fatalf("QueryArtifacts: %v", err)
	}
	output := FormatQueryResults(results, QueryOpts{Count: true})
	if output != "3\n" {
		t.Errorf("expected '3\\n', got %q", output)
	}
}

func TestQueryGroupByStatus(t *testing.T) {
	root := setupQueryWorkspace(t)

	results, err := QueryArtifacts(root, QueryOpts{})
	if err != nil {
		t.Fatalf("QueryArtifacts: %v", err)
	}
	groups := GroupResults(results, "status")
	if groups["active"] != 2 {
		t.Errorf("expected 2 active, got %d", groups["active"])
	}
	if groups["draft"] != 1 {
		t.Errorf("expected 1 draft, got %d", groups["draft"])
	}
}
