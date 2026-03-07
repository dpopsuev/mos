package mesh

import "testing"

func TestAddNodeDedup(t *testing.T) {
	var g Graph
	g.AddNode(Node{ID: "a", Kind: "package", Label: "pkg/a"})
	g.AddNode(Node{ID: "a", Kind: "package", Label: "pkg/a"})
	g.AddNode(Node{ID: "b", Kind: "spec", Label: "SPEC-001"})
	if len(g.Nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(g.Nodes))
	}
}

func TestAddEdgeDedup(t *testing.T) {
	var g Graph
	g.AddEdge(Edge{From: "a", To: "b", Relation: "includes"})
	g.AddEdge(Edge{From: "a", To: "b", Relation: "includes"})
	g.AddEdge(Edge{From: "a", To: "b", Relation: "imports"})
	if len(g.Edges) != 2 {
		t.Fatalf("expected 2 edges, got %d", len(g.Edges))
	}
}

func TestNodesOfKind(t *testing.T) {
	var g Graph
	g.AddNode(Node{ID: "a", Kind: "package"})
	g.AddNode(Node{ID: "b", Kind: "spec"})
	g.AddNode(Node{ID: "c", Kind: "package"})

	pkgs := g.NodesOfKind("package")
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 package nodes, got %d", len(pkgs))
	}
}

func TestEdgesFromTo(t *testing.T) {
	var g Graph
	g.AddEdge(Edge{From: "a", To: "b", Relation: "includes"})
	g.AddEdge(Edge{From: "a", To: "c", Relation: "imports"})
	g.AddEdge(Edge{From: "b", To: "c", Relation: "justifies"})

	from := g.EdgesFrom("a")
	if len(from) != 2 {
		t.Fatalf("expected 2 edges from a, got %d", len(from))
	}
	to := g.EdgesTo("c")
	if len(to) != 2 {
		t.Fatalf("expected 2 edges to c, got %d", len(to))
	}
}
