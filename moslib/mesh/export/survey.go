package export

import (
	"github.com/dpopsuev/mos/moslib/mesh"
	"github.com/dpopsuev/mos/moslib/model"
	"github.com/dpopsuev/mos/moslib/survey"
)

// ExportSurvey populates g with nodes and edges from source code structure.
// Each namespace becomes a node with plane=code; dependency graph edges become
// "imports" edges. Exported symbols are added as child nodes.
func ExportSurvey(root string, g *mesh.Graph) error {
	sc := &survey.AutoScanner{}
	proj, err := sc.Scan(root)
	if err != nil {
		return err
	}
	exportProject(proj, g)
	return nil
}

// ExportProjectToGraph populates g from a pre-scanned project, useful for
// testing without a real source tree.
func ExportProjectToGraph(proj *model.Project, g *mesh.Graph) {
	exportProject(proj, g)
}

func exportProject(proj *model.Project, g *mesh.Graph) {
	for _, ns := range proj.Namespaces {
		g.AddNode(mesh.Node{
			ID:    ns.ImportPath,
			Kind:  "namespace",
			Label: ns.Name,
			Meta:  map[string]string{"plane": "code"},
		})

		for _, sym := range ns.Symbols {
			if !sym.Exported {
				continue
			}
			symID := ns.ImportPath + "." + sym.Name
			g.AddNode(mesh.Node{
				ID:    symID,
				Kind:  "symbol",
				Label: sym.Name,
				Meta: map[string]string{
					"plane":       "code",
					"symbol_kind": sym.Kind.String(),
				},
			})
			g.AddEdge(mesh.Edge{From: ns.ImportPath, To: symID, Relation: "declares"})
		}
	}

	if proj.DependencyGraph == nil {
		return
	}
	for _, e := range proj.DependencyGraph.Edges {
		rel := "imports"
		if e.External {
			rel = "imports_external"
		}
		g.AddEdge(mesh.Edge{From: e.From, To: e.To, Relation: rel})
	}
}
