package main

import (
	"strings"
	"testing"

	"github.com/dpopsuev/mos/moslib/model"
)

func testStyles() *Styles {
	return BuildStyles(DarkTheme())
}

func TestGraphEmptySectionsHidden(t *testing.T) {
	s := testStyles()

	t.Run("no imports and no importers renders single placeholder", func(t *testing.T) {
		mod := &model.Project{
			Path:            "example.com/m",
			Namespaces:      []*model.Namespace{model.NewNamespace("leaf", "example.com/m/leaf")},
			DependencyGraph: model.NewDependencyGraph(),
		}
		out := renderGraph(mod, 0, 0, 80, 40, s)
		if strings.Contains(out, "Imports") {
			t.Error("expected no 'Imports' header when there are no outgoing edges")
		}
		if strings.Contains(out, "Imported by") {
			t.Error("expected no 'Imported by' header when there are no incoming edges")
		}
		if !strings.Contains(out, "no imports") {
			t.Error("expected '(no imports)' placeholder for package with no edges")
		}
	})

	t.Run("only outgoing imports hides imported-by section", func(t *testing.T) {
		mod := &model.Project{
			Path: "example.com/m",
			Namespaces: []*model.Namespace{
				model.NewNamespace("a", "example.com/m/a"),
				model.NewNamespace("b", "example.com/m/b"),
			},
			DependencyGraph: model.NewDependencyGraph(),
		}
		mod.DependencyGraph.AddEdge("example.com/m/a", "example.com/m/b", false)

		out := renderGraph(mod, 0, 0, 80, 40, s)
		if !strings.Contains(out, "Imports") {
			t.Error("expected 'Imports' header for package with outgoing edges")
		}
		if strings.Contains(out, "Imported by") {
			t.Error("expected no 'Imported by' header when there are no incoming edges")
		}
	})

	t.Run("only incoming imports hides imports section", func(t *testing.T) {
		mod := &model.Project{
			Path: "example.com/m",
			Namespaces: []*model.Namespace{
				model.NewNamespace("a", "example.com/m/a"),
				model.NewNamespace("b", "example.com/m/b"),
			},
			DependencyGraph: model.NewDependencyGraph(),
		}
		mod.DependencyGraph.AddEdge("example.com/m/b", "example.com/m/a", false)

		out := renderGraph(mod, 0, 0, 80, 40, s)
		if strings.Contains(out, "Imports") && !strings.Contains(out, "Imported by") {
			t.Error("expected no 'Imports' header when there are no outgoing edges")
		}
		if !strings.Contains(out, "Imported by") {
			t.Error("expected 'Imported by' header for package with incoming edges")
		}
	})

	t.Run("both sections rendered when both exist", func(t *testing.T) {
		mod := &model.Project{
			Path: "example.com/m",
			Namespaces: []*model.Namespace{
				model.NewNamespace("a", "example.com/m/a"),
				model.NewNamespace("b", "example.com/m/b"),
			},
			DependencyGraph: model.NewDependencyGraph(),
		}
		mod.DependencyGraph.AddEdge("example.com/m/a", "example.com/m/b", false)
		mod.DependencyGraph.AddEdge("example.com/m/b", "example.com/m/a", false)

		out := renderGraph(mod, 0, 0, 80, 40, s)
		if !strings.Contains(out, "Imports") {
			t.Error("expected 'Imports' header")
		}
		if !strings.Contains(out, "Imported by") {
			t.Error("expected 'Imported by' header")
		}
	})
}

func TestGraphLineCountMatchesRenderedLines(t *testing.T) {
	s := testStyles()

	mod := &model.Project{
		Path: "example.com/m",
		Namespaces: []*model.Namespace{
			model.NewNamespace("a", "example.com/m/a"),
			model.NewNamespace("b", "example.com/m/b"),
			model.NewNamespace("c", "example.com/m/c"),
		},
		DependencyGraph: model.NewDependencyGraph(),
	}
	mod.DependencyGraph.AddEdge("example.com/m/a", "example.com/m/b", false)
	mod.DependencyGraph.AddEdge("example.com/m/c", "example.com/m/a", false)

	count := graphLineCount(mod, 0)
	rendered := renderGraph(mod, 0, 0, 80, 100, s)
	renderedLines := strings.Count(rendered, "\n") + 1

	if count != renderedLines {
		t.Errorf("graphLineCount=%d but rendered %d lines\nrendered:\n%s", count, renderedLines, rendered)
	}
}

func TestGraphSymbolLevel(t *testing.T) {
	s := testStyles()

	t.Run("symbol with dependencies shows symbol-level graph", func(t *testing.T) {
		mod := &model.Project{Path: "example.com/m"}
		sym := &model.Symbol{
			Name:         "Hello",
			Kind:         model.SymbolFunction,
			Exported:     true,
			Dependencies: []string{"fmt", "example.com/m/util"},
		}
		out := renderGraphForSymbol(mod, sym, 0, 80, 40, s)
		if out == "" {
			t.Fatal("expected non-empty output for symbol with dependencies")
		}
		if !strings.Contains(out, "Hello") {
			t.Error("expected symbol name in header")
		}
		if !strings.Contains(out, "fmt") {
			t.Error("expected 'fmt' in dependencies")
		}
		if !strings.Contains(out, "util") {
			t.Error("expected 'util' in dependencies")
		}
	})

	t.Run("symbol without dependencies returns empty", func(t *testing.T) {
		mod := &model.Project{Path: "example.com/m"}
		sym := &model.Symbol{
			Name:     "NoImport",
			Kind:     model.SymbolFunction,
			Exported: true,
		}
		out := renderGraphForSymbol(mod, sym, 0, 80, 40, s)
		if out != "" {
			t.Errorf("expected empty output for symbol without dependencies, got: %s", out)
		}
	})

	t.Run("nil symbol returns empty", func(t *testing.T) {
		mod := &model.Project{Path: "example.com/m"}
		out := renderGraphForSymbol(mod, nil, 0, 80, 40, s)
		if out != "" {
			t.Errorf("expected empty output for nil symbol, got: %s", out)
		}
	})
}

func TestGraphNoDataStates(t *testing.T) {
	s := testStyles()

	t.Run("negative index", func(t *testing.T) {
		mod := &model.Project{Path: "m"}
		out := renderGraph(mod, -1, 0, 80, 40, s)
		if !strings.Contains(out, "select a package") {
			t.Error("expected 'select a package' for invalid index")
		}
	})

	t.Run("nil dependency graph", func(t *testing.T) {
		mod := &model.Project{
			Path:       "m",
			Namespaces: []*model.Namespace{model.NewNamespace("a", "m/a")},
		}
		out := renderGraph(mod, 0, 0, 80, 40, s)
		if !strings.Contains(out, "no import data") {
			t.Error("expected 'no import data' for nil graph")
		}
	})
}
