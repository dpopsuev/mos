package export

import (
	"github.com/dpopsuev/mos/moslib/mesh"
)

// ExportUnifiedGraph builds a complete mesh.Graph covering all four data
// planes: governance, code (survey), architecture, and mesh (cross-plane edges).
// Errors from individual planes are collected but do not abort the export;
// the returned graph contains whatever was successfully loaded.
func ExportUnifiedGraph(root string) (*mesh.Graph, []error) {
	g := &mesh.Graph{}
	var errs []error

	if err := ExportGovernance(root, g); err != nil {
		errs = append(errs, err)
	}
	if err := ExportSurvey(root, g); err != nil {
		errs = append(errs, err)
	}
	if err := ExportArchitecture(root, g); err != nil {
		errs = append(errs, err)
	}
	if err := ExportMesh(root, g); err != nil {
		errs = append(errs, err)
	}

	return g, errs
}
