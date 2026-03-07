package export

import (
	"strings"

	"github.com/dpopsuev/mos/moslib/artifact"
	"github.com/dpopsuev/mos/moslib/dsl"
	"github.com/dpopsuev/mos/moslib/mesh"
)

// ExportMesh populates g with cross-plane "governed_by" edges linking code
// packages (from the survey plane) to governance artifacts (specs, contracts)
// via include directives and justifies fields.
func ExportMesh(root string, g *mesh.Graph) error {
	specs, err := Reader.List(root, "specification", "")
	if err != nil || specs == nil {
		return err
	}

	for _, si := range specs {
		ab, err := Reader.Read(si.Path)
		if err != nil {
			continue
		}
		for _, item := range ab.Items {
			sb, ok := item.(*dsl.SpecBlock)
			if !ok {
				continue
			}
			for _, inc := range sb.Includes {
				g.AddEdge(mesh.Edge{From: inc.Path, To: si.ID, Relation: "governed_by"})
			}
		}
	}

	contracts, err := Reader.List(root, artifact.KindContract, "")
	if err != nil || contracts == nil {
		return err
	}
	for _, ci := range contracts {
		ab, err := Reader.Read(ci.Path)
		if err != nil {
			continue
		}
		justifies, _ := dsl.FieldString(ab.Items, "justifies")
		for _, target := range splitCSV(justifies) {
			if strings.HasPrefix(target, "SPEC-") || strings.HasPrefix(target, "NEED-") {
				g.AddEdge(mesh.Edge{From: ci.ID, To: target, Relation: "justifies"})
			}
		}
	}

	return nil
}
