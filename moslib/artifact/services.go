package artifact

import (
	"github.com/dpopsuev/mos/moslib/dsl"
	"github.com/dpopsuev/mos/moslib/registry"
)

// Services aggregates the capabilities that library packages (arch, clone)
// need from the artifact layer. Callers register these once in main.go init()
// instead of threading individual function parameters.
type Services struct {
	Create   func(root string, td registry.ArtifactTypeDef, id string, fields map[string]string) (string, error)
	FindPath func(root string, td registry.ArtifactTypeDef, id string) (string, error)
	FieldStr func(items []dsl.Node, key string) string
}

// DefaultServices returns a Services instance backed by the default artifact implementations.
func DefaultServices() *Services {
	return &Services{
		Create:   GenericCreate,
		FindPath: FindGenericArtifactPath,
		FieldStr: FieldStr,
	}
}

// Svc is the global service locator. Set during init() in cmd/mos/main.go.
var Svc = DefaultServices()
