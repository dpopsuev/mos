package export

import (
	"github.com/dpopsuev/mos/moslib/arch"
	"github.com/dpopsuev/mos/moslib/mesh"
)

// ExportArchitecture populates g with nodes and edges from architecture
// artifacts. Each service/component becomes a node with plane=architecture;
// edges and forbidden constraints become graph edges.
func ExportArchitecture(root string, g *mesh.Graph) error {
	archs, err := Reader.List(root, "architecture", "")
	if err != nil || archs == nil {
		return err
	}

	for _, ai := range archs {
		ab, err := Reader.Read(ai.Path)
		if err != nil {
			continue
		}
		model := arch.ParseArchModel(ab)
		exportArchModel(model, g)
	}
	return nil
}

func exportArchModel(m arch.ArchModel, g *mesh.Graph) {
	for _, svc := range m.Services {
		meta := map[string]string{"plane": "architecture"}
		if svc.TrustZone != "" {
			meta["trust_zone"] = svc.TrustZone
		}
		if svc.Package != "" {
			meta["package"] = svc.Package
		}
		g.AddNode(mesh.Node{
			ID:    svc.Name,
			Kind:  "component",
			Label: svc.Name,
			Meta:  meta,
		})
	}

	for _, e := range m.Edges {
		g.AddEdge(mesh.Edge{From: e.From, To: e.To, Relation: "depends_on"})
	}

	for _, f := range m.Forbidden {
		if f.From != "" && f.To != "" {
			g.AddEdge(mesh.Edge{From: f.From, To: f.To, Relation: "forbidden"})
		}
	}
}
