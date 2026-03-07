package artifact

import "github.com/dpopsuev/mos/moslib/registry"

// Re-export project types for backward compatibility.
type ProjectDef = registry.ProjectDef

func LoadProjects(root string) ([]ProjectDef, error) { return registry.LoadProjects(root) }
func FindDefaultProject(projects []ProjectDef) *ProjectDef { return registry.FindDefaultProject(projects) }
func FindProjectByPrefix(projects []ProjectDef, prefix string) *ProjectDef {
	return registry.FindProjectByPrefix(projects, prefix)
}
func NextID(root, projectName string) (string, error) { return registry.NextID(root, projectName) }
func NextIDForType(root, prefix, directory string) (string, error) {
	return registry.NextIDForType(root, prefix, directory)
}
func AddProject(root, name, prefix string) error { return registry.AddProject(root, name, prefix) }
