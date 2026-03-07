package artifact

import "github.com/dpopsuev/mos/moslib/registry"

// Re-export registry types for backward compatibility.
type HookDef = registry.HookDef
type TransitionGate = registry.TransitionGate
type LifecycleDef = registry.LifecycleDef
type ArtifactTypeDef = registry.ArtifactTypeDef
type Registry = registry.Registry
type FieldOpts = registry.FieldOpts

var CoreKinds = registry.CoreKinds

func DefaultRegistry() *Registry           { return registry.DefaultRegistry() }
func LoadRegistry(root string) (*Registry, error) { return registry.LoadRegistry(root) }

func AddArtifactType(root, kind, directory string) error { return registry.AddArtifactType(root, kind, directory) }
func RemoveArtifactType(root, kind string) error         { return registry.RemoveArtifactType(root, kind) }
func RemoveProject(root, name string) error              { return registry.RemoveProject(root, name) }
func AddFieldToType(root, kind string, field FieldOpts) error { return registry.AddFieldToType(root, kind, field) }
func SetTypeLifecycle(root, kind string, activeStates, archiveStates []string) error {
	return registry.SetTypeLifecycle(root, kind, activeStates, archiveStates)
}
func SetTypeDirectory(root, kind, directory string) error { return registry.SetTypeDirectory(root, kind, directory) }
func SetFieldEnum(root, kind, fieldName string, enum []string) error {
	return registry.SetFieldEnum(root, kind, fieldName, enum)
}
