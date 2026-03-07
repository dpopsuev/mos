package export

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/dpopsuev/mos/moslib/mesh"
	"github.com/dpopsuev/mos/moslib/model"
)

func mkart(t *testing.T, root, kind, dir, id, content string) {
	t.Helper()
	d := filepath.Join(root, ".mos", dir, "active", id)
	if err := os.MkdirAll(d, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(d, kind+".mos"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func setupTestProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	mkart(t, root, "specification", "specifications", "SPEC-2026-001", `specification "SPEC-2026-001" {
  title = "Auth Spec"
  status = "active"
  spec {
    include "moslib/auth"
  }
}
`)
	mkart(t, root, "contract", "contracts", "CON-2026-001", `contract "CON-2026-001" {
  title = "Auth Contract"
  status = "active"
  justifies = "SPEC-2026-001"
  sprint = "SPR-2026-001"
}
`)
	mkart(t, root, "sprint", "sprints", "SPR-2026-001", `sprint "SPR-2026-001" {
  title = "Sprint One"
  status = "active"
  contracts = "CON-2026-001"
}
`)
	mkart(t, root, "need", "needs", "NEED-2026-001", `need "NEED-2026-001" {
  title = "Core Need"
  status = "active"
}
`)
	mkart(t, root, "architecture", "architectures", "ARCH-test", `architecture "ARCH-test" {
  title = "Test Architecture"
  status = "active"
  resolution = "package"

  component "API" {
    package = "moslib/api"
    trust_zone = "external"
  }

  component "Core" {
    package = "moslib/core"
    trust_zone = "internal"
  }

  edge "api-to-core" {
    from = "API"
    to = "Core"
    protocol = "function_call"
  }

  forbidden "core-to-api" {
    from = "Core"
    to = "API"
    reason = "violates layering"
  }
}
`)

	return root
}

func TestExportGovernance(t *testing.T) {
	root := setupTestProject(t)
	g := &mesh.Graph{}

	if err := ExportGovernance(root, g); err != nil {
		t.Fatalf("ExportGovernance: %v", err)
	}

	if len(g.Nodes) == 0 {
		t.Fatal("expected governance nodes")
	}

	kinds := map[string]int{}
	for _, n := range g.Nodes {
		kinds[n.Kind]++
		if n.Meta["plane"] != "governance" {
			t.Errorf("node %s has plane=%q, want governance", n.ID, n.Meta["plane"])
		}
	}

	for _, expected := range []string{"contract", "sprint", "specification", "need"} {
		if kinds[expected] == 0 {
			t.Errorf("expected at least one %s node", expected)
		}
	}

	hasEdge := func(from, to, rel string) bool {
		for _, e := range g.Edges {
			if e.From == from && e.To == to && e.Relation == rel {
				return true
			}
		}
		return false
	}

	if !hasEdge("CON-2026-001", "SPEC-2026-001", "justifies") {
		t.Error("expected justifies edge from CON-2026-001 to SPEC-2026-001")
	}
	if !hasEdge("CON-2026-001", "SPR-2026-001", "scheduled_in") {
		t.Error("expected scheduled_in edge from CON-2026-001 to SPR-2026-001")
	}
	if !hasEdge("SPR-2026-001", "CON-2026-001", "contains") {
		t.Error("expected contains edge from SPR-2026-001 to CON-2026-001")
	}
}

func TestExportSurvey(t *testing.T) {
	g := &mesh.Graph{}

	proj := &model.Project{
		Path: "example.com/test",
		Namespaces: []*model.Namespace{
			{
				Name:       "auth",
				ImportPath: "example.com/test/auth",
				Symbols: []*model.Symbol{
					{Name: "Login", Kind: model.SymbolFunction, Exported: true},
					{Name: "helper", Kind: model.SymbolFunction, Exported: false},
				},
			},
			{
				Name:       "db",
				ImportPath: "example.com/test/db",
			},
		},
		DependencyGraph: &model.DependencyGraph{
			Edges: []model.DependencyEdge{
				{From: "example.com/test/auth", To: "example.com/test/db"},
				{From: "example.com/test/auth", To: "fmt", External: true},
			},
		},
	}

	ExportProjectToGraph(proj, g)

	nsNodes := g.NodesOfKind("namespace")
	if len(nsNodes) != 2 {
		t.Fatalf("expected 2 namespace nodes, got %d", len(nsNodes))
	}

	symNodes := g.NodesOfKind("symbol")
	if len(symNodes) != 1 {
		t.Fatalf("expected 1 exported symbol node, got %d", len(symNodes))
	}
	if symNodes[0].ID != "example.com/test/auth.Login" {
		t.Errorf("unexpected symbol ID: %s", symNodes[0].ID)
	}

	for _, n := range g.Nodes {
		if n.Meta["plane"] != "code" {
			t.Errorf("node %s has plane=%q, want code", n.ID, n.Meta["plane"])
		}
	}

	hasEdge := func(from, to, rel string) bool {
		for _, e := range g.Edges {
			if e.From == from && e.To == to && e.Relation == rel {
				return true
			}
		}
		return false
	}

	if !hasEdge("example.com/test/auth", "example.com/test/db", "imports") {
		t.Error("expected imports edge auth->db")
	}
	if !hasEdge("example.com/test/auth", "fmt", "imports_external") {
		t.Error("expected imports_external edge auth->fmt")
	}
	if !hasEdge("example.com/test/auth", "example.com/test/auth.Login", "declares") {
		t.Error("expected declares edge auth->Login")
	}
}

func TestExportArchitecture(t *testing.T) {
	root := setupTestProject(t)
	g := &mesh.Graph{}

	if err := ExportArchitecture(root, g); err != nil {
		t.Fatalf("ExportArchitecture: %v", err)
	}

	components := g.NodesOfKind("component")
	if len(components) != 2 {
		t.Fatalf("expected 2 component nodes, got %d", len(components))
	}

	for _, n := range components {
		if n.Meta["plane"] != "architecture" {
			t.Errorf("component %s has plane=%q, want architecture", n.ID, n.Meta["plane"])
		}
	}

	hasEdge := func(from, to, rel string) bool {
		for _, e := range g.Edges {
			if e.From == from && e.To == to && e.Relation == rel {
				return true
			}
		}
		return false
	}

	if !hasEdge("API", "Core", "depends_on") {
		t.Error("expected depends_on edge API->Core")
	}
	if !hasEdge("Core", "API", "forbidden") {
		t.Error("expected forbidden edge Core->API")
	}
}

func TestExportMesh(t *testing.T) {
	root := setupTestProject(t)
	g := &mesh.Graph{}

	if err := ExportMesh(root, g); err != nil {
		t.Fatalf("ExportMesh: %v", err)
	}

	hasEdge := func(from, to, rel string) bool {
		for _, e := range g.Edges {
			if e.From == from && e.To == to && e.Relation == rel {
				return true
			}
		}
		return false
	}

	if !hasEdge("moslib/auth", "SPEC-2026-001", "governed_by") {
		t.Error("expected governed_by edge from moslib/auth to SPEC-2026-001")
	}
	if !hasEdge("CON-2026-001", "SPEC-2026-001", "justifies") {
		t.Error("expected justifies edge from CON-2026-001 to SPEC-2026-001")
	}
}

func TestExportUnifiedGraph(t *testing.T) {
	root := setupTestProject(t)

	g, errs := ExportUnifiedGraph(root)
	for _, e := range errs {
		if e != nil {
			t.Logf("non-fatal export error: %v", e)
		}
	}

	if len(g.Nodes) == 0 {
		t.Fatal("expected nodes in unified graph")
	}

	planes := map[string]int{}
	for _, n := range g.Nodes {
		planes[n.Meta["plane"]]++
	}

	if planes["governance"] == 0 {
		t.Error("expected governance plane nodes")
	}
}

func TestDiffGraphs(t *testing.T) {
	old := &mesh.Graph{}
	old.AddNode(mesh.Node{ID: "A", Kind: "test", Meta: map[string]string{"v": "1"}})
	old.AddNode(mesh.Node{ID: "B", Kind: "test", Meta: map[string]string{"v": "1"}})
	old.AddNode(mesh.Node{ID: "C", Kind: "test"})
	old.AddEdge(mesh.Edge{From: "A", To: "B", Relation: "links"})
	old.AddEdge(mesh.Edge{From: "B", To: "C", Relation: "links"})

	newG := &mesh.Graph{}
	newG.AddNode(mesh.Node{ID: "A", Kind: "test", Meta: map[string]string{"v": "2"}})
	newG.AddNode(mesh.Node{ID: "B", Kind: "test", Meta: map[string]string{"v": "1"}})
	newG.AddNode(mesh.Node{ID: "D", Kind: "test"})
	newG.AddEdge(mesh.Edge{From: "A", To: "B", Relation: "links"})
	newG.AddEdge(mesh.Edge{From: "A", To: "D", Relation: "links"})

	delta := DiffGraphs(old, newG)

	if len(delta.AddedNodes) != 1 || delta.AddedNodes[0].ID != "D" {
		t.Errorf("expected 1 added node (D), got %v", delta.AddedNodes)
	}
	if len(delta.RemovedNodes) != 1 || delta.RemovedNodes[0].ID != "C" {
		t.Errorf("expected 1 removed node (C), got %v", delta.RemovedNodes)
	}
	if len(delta.ModifiedNodes) != 1 || delta.ModifiedNodes[0].ID != "A" {
		t.Errorf("expected 1 modified node (A), got %v", delta.ModifiedNodes)
	}
	if len(delta.AddedEdges) != 1 {
		t.Errorf("expected 1 added edge, got %d", len(delta.AddedEdges))
	}
	if len(delta.RemovedEdges) != 1 {
		t.Errorf("expected 1 removed edge, got %d", len(delta.RemovedEdges))
	}
}

func TestGraphJSONSerialization(t *testing.T) {
	g := &mesh.Graph{}
	g.AddNode(mesh.Node{ID: "CON-1", Kind: "contract", Label: "Test", Meta: map[string]string{"plane": "governance"}})
	g.AddNode(mesh.Node{ID: "pkg/foo", Kind: "namespace", Label: "foo", Meta: map[string]string{"plane": "code"}})
	g.AddEdge(mesh.Edge{From: "pkg/foo", To: "CON-1", Relation: "governed_by"})

	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	var decoded mesh.Graph
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	if len(decoded.Nodes) != 2 {
		t.Errorf("expected 2 nodes after roundtrip, got %d", len(decoded.Nodes))
	}
	if len(decoded.Edges) != 1 {
		t.Errorf("expected 1 edge after roundtrip, got %d", len(decoded.Edges))
	}

	if decoded.Nodes[0].Meta["plane"] != "governance" {
		t.Errorf("expected plane=governance, got %q", decoded.Nodes[0].Meta["plane"])
	}
}

func TestDiffGraphsJSONSerialization(t *testing.T) {
	delta := &GraphDelta{
		AddedNodes:   []mesh.Node{{ID: "X", Kind: "test"}},
		RemovedNodes: []mesh.Node{{ID: "Y", Kind: "test"}},
		ModifiedNodes: []NodeDiff{{
			ID:      "Z",
			OldMeta: map[string]string{"v": "1"},
			NewMeta: map[string]string{"v": "2"},
		}},
		AddedEdges:   []mesh.Edge{{From: "X", To: "Z", Relation: "new"}},
		RemovedEdges: []mesh.Edge{{From: "Y", To: "Z", Relation: "old"}},
	}

	data, err := json.Marshal(delta)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	var decoded GraphDelta
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	if len(decoded.AddedNodes) != 1 {
		t.Errorf("expected 1 added node, got %d", len(decoded.AddedNodes))
	}
	if len(decoded.RemovedNodes) != 1 {
		t.Errorf("expected 1 removed node, got %d", len(decoded.RemovedNodes))
	}
	if len(decoded.ModifiedNodes) != 1 {
		t.Errorf("expected 1 modified node, got %d", len(decoded.ModifiedNodes))
	}
}
