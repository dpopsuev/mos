package export

import (
	"strings"

	"github.com/dpopsuev/mos/moslib/artifact"
	"github.com/dpopsuev/mos/moslib/dsl"
	"github.com/dpopsuev/mos/moslib/mesh"
)

// Reader is the GovernReader used by export functions. Defaults to artifact.NewReader().
var Reader artifact.GovernReader = artifact.NewReader()

// relationalFieldsForKind returns field->relation mappings for a given artifact kind.
// specification uses "satisfies" for spec→need; contract and batch use "justifies".
func relationalFieldsForKind(kind string) []struct {
	field    string
	relation string
} {
	base := []struct {
		field    string
		relation string
	}{
		{"precondition", "precondition"},
		{"sprint", "scheduled_in"},
		{"batch", "grouped_in"},
	}
	if kind == "specification" {
		return append([]struct{ field, relation string }{{"satisfies", "satisfies"}}, base...)
	}
	return append([]struct{ field, relation string }{{"justifies", "justifies"}}, base...)
}

// ExportGovernance populates g with nodes and edges from all governance
// artifacts (contracts, specs, needs, rules, sprints, batches, etc.).
func ExportGovernance(root string, g *mesh.Graph) error {
	kinds, err := Reader.ListKinds(root)
	if err != nil {
		return err
	}

	for _, kind := range kinds {
		items, err := Reader.List(root, kind, "")
		if err != nil {
			continue
		}

		for _, item := range items {
			meta := map[string]string{
				"plane":  "governance",
				"status": item.Status,
			}
			g.AddNode(mesh.Node{
				ID:    item.ID,
				Kind:  kind,
				Label: item.Title,
				Meta:  meta,
			})

			ab, err := Reader.Read(item.Path)
			if err != nil {
				continue
			}

			for _, rf := range relationalFieldsForKind(kind) {
				val, _ := dsl.FieldString(ab.Items, rf.field)
				for _, target := range splitCSV(val) {
					g.AddEdge(mesh.Edge{From: item.ID, To: target, Relation: rf.relation})
				}
			}

			extractScopeDeps(ab, item.ID, g)
			extractContractsDeps(ab, item.ID, g)
		}
	}

	return nil
}

// extractScopeDeps reads scope { depends_on = [...] } blocks.
func extractScopeDeps(ab *dsl.ArtifactBlock, fromID string, g *mesh.Graph) {
	for _, item := range ab.Items {
		blk, ok := item.(*dsl.Block)
		if !ok || blk.Name != "scope" {
			continue
		}
		deps := dsl.FieldStringSlice(blk.Items, "depends_on")
		for _, dep := range deps {
			g.AddEdge(mesh.Edge{From: fromID, To: dep, Relation: "depends_on"})
		}
	}
}

// extractContractsDeps reads contracts = "CON-X,CON-Y" on sprints.
func extractContractsDeps(ab *dsl.ArtifactBlock, fromID string, g *mesh.Graph) {
	val, _ := dsl.FieldString(ab.Items, "contracts")
	for _, target := range splitCSV(val) {
		g.AddEdge(mesh.Edge{From: fromID, To: target, Relation: "contains"})
	}
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
