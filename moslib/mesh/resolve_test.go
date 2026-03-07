package mesh

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolve(t *testing.T) {
	root := t.TempDir()
	mosDir := filepath.Join(root, ".mos")

	mkart := func(kind, dir, id, content string) {
		d := filepath.Join(mosDir, dir, "active", id)
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(d, kind+".mos"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	mkart("specification", "specifications", "SPEC-2026-001", `specification "SPEC-2026-001" {
  title = "Test Spec"
  status = "active"
  spec {
    include "moslib/governance"
  }
}
`)
	mkart("contract", "contracts", "CON-2026-001", `contract "CON-2026-001" {
  title = "Test Contract"
  status = "active"
  justifies = "SPEC-2026-001"
  sprint = "SPR-2026-001"
}
`)
	mkart("sprint", "sprints", "SPR-2026-001", `sprint "SPR-2026-001" {
  title = "Test Sprint"
  status = "active"
  contracts = "CON-2026-001"
}
`)

	g, err := Resolve(root, "moslib/governance")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	if len(g.Nodes) == 0 {
		t.Fatal("expected nodes in graph")
	}

	pkgs := g.NodesOfKind("package")
	if len(pkgs) == 0 {
		t.Fatal("expected at least one package node")
	}
	found := false
	for _, n := range pkgs {
		if n.ID == "moslib/governance" {
			found = true
		}
	}
	if !found {
		t.Error("expected moslib/governance package node")
	}

	specs := g.NodesOfKind("spec")
	if len(specs) != 1 || specs[0].ID != "SPEC-2026-001" {
		t.Errorf("expected 1 spec node SPEC-2026-001, got %v", specs)
	}

	contracts := g.NodesOfKind("contract")
	if len(contracts) != 1 || contracts[0].ID != "CON-2026-001" {
		t.Errorf("expected 1 contract node CON-2026-001, got %v", contracts)
	}

	sprints := g.NodesOfKind("sprint")
	if len(sprints) != 1 || sprints[0].ID != "SPR-2026-001" {
		t.Errorf("expected 1 sprint node SPR-2026-001, got %v", sprints)
	}

	specEdges := g.EdgesTo("moslib/governance")
	includeFound := false
	for _, e := range specEdges {
		if e.Relation == "includes" && e.From == "SPEC-2026-001" {
			includeFound = true
		}
	}
	if !includeFound {
		t.Error("expected includes edge from SPEC-2026-001 to moslib/governance")
	}
}
