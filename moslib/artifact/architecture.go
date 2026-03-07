package artifact

import (
	"github.com/dpopsuev/mos/moslib/arch"
	"github.com/dpopsuev/mos/moslib/dsl"
	"github.com/dpopsuev/mos/moslib/model"
)

func init() {
	arch.ReadArchitectureFn = ReadArchitecture
	arch.ReadConfigFn = ReadConfig
	arch.ReadSpecIncludesFn = ReadSpecificationIncludes
}

// Re-export architecture types from moslib/arch/ for backward compatibility.
type ArchService = arch.ArchService
type ArchEdge = arch.ArchEdge
type ArchForbidden = arch.ArchForbidden
type ArchModel = arch.ArchModel
type ComponentGroup = arch.ComponentGroup
type SyncOptions = arch.SyncOptions

func ParseArchModel(ab *dsl.ArtifactBlock) ArchModel { return arch.ParseArchModel(ab) }
func RenderMermaid(m ArchModel) string               { return arch.RenderMermaid(m) }
func RenderArchMos(m ArchModel) string               { return arch.RenderArchMos(m) }
func RenderArchMarkdown(m ArchModel) string           { return arch.RenderArchMarkdown(m) }
func ProjectToArchModel(proj *model.Project, opts SyncOptions) ArchModel {
	return arch.ProjectToArchModel(proj, opts)
}
func CheckForbiddenEdges(live, declared ArchModel) []string {
	return arch.CheckForbiddenEdges(live, declared)
}
func LoadComponentGroups(root string) ([]ComponentGroup, error) {
	return arch.LoadComponentGroups(root)
}

func ComputeChurn(root string, days int, modPath string) map[string]int {
	return arch.ComputeChurn(root, days, modPath)
}