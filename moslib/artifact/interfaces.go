package artifact

import (
	"github.com/dpopsuev/mos/moslib/dsl"
	"github.com/dpopsuev/mos/moslib/schema"
)

// GovernReader is Interface A: the narrow read-only API surface for Insight
// sub-tools (mesh/export, topology, arch). Consumers that only need to
// discover, read, and list governance artifacts use this interface.
type GovernReader interface {
	FindPath(root, kind, id string) (string, error)
	Read(path string) (*dsl.ArtifactBlock, error)
	Query(root string, opts QueryOpts) ([]QueryResult, error)
	List(root, kind, statusFilter string) ([]GenericInfo, error)
	ListKinds(root string) ([]string, error)
}

// GovernEnforcer is Interface B: the narrow API surface for Guard sub-tools
// (guardcmd, audit). Consumers that enforce governance constraints use this
// interface, which extends GovernReader with schema loading.
type GovernEnforcer interface {
	GovernReader
	LoadSchemas(root string) ([]schema.ArtifactSchema, error)
}

// reader implements GovernReader backed by artifact package functions.
type reader struct{}

func (r *reader) FindPath(root, kind, id string) (string, error) {
	reg, err := LoadRegistry(root)
	if err != nil {
		return "", err
	}
	td, ok := reg.Types[kind]
	if !ok {
		return FindContractPath(root, id)
	}
	return FindGenericPath(root, td, id)
}

func (r *reader) Read(path string) (*dsl.ArtifactBlock, error) {
	return dsl.ReadArtifact(path)
}

func (r *reader) Query(root string, opts QueryOpts) ([]QueryResult, error) {
	return QueryArtifacts(root, opts)
}

func (r *reader) List(root, kind, statusFilter string) ([]GenericInfo, error) {
	reg, err := LoadRegistry(root)
	if err != nil {
		return nil, err
	}
	td, ok := reg.Types[kind]
	if !ok {
		return nil, nil
	}
	return GenericList(root, td, statusFilter)
}

func (r *reader) ListKinds(root string) ([]string, error) {
	reg, err := LoadRegistry(root)
	if err != nil {
		return nil, err
	}
	kinds := make([]string, 0, len(reg.Types))
	for k := range reg.Types {
		kinds = append(kinds, k)
	}
	return kinds, nil
}

// enforcer implements GovernEnforcer.
type enforcer struct {
	reader
}

func (e *enforcer) LoadSchemas(root string) ([]schema.ArtifactSchema, error) {
	reg, err := LoadRegistry(root)
	if err != nil {
		return nil, err
	}
	var schemas []schema.ArtifactSchema
	for _, td := range reg.Types {
		schemas = append(schemas, schema.ArtifactSchema{
			Kind:      td.Kind,
			Directory: td.Directory,
		})
	}
	for _, s := range schema.DefaultCoreSchemas() {
		if s != nil {
			schemas = append(schemas, *s)
		}
	}
	return schemas, nil
}

// NewReader returns a GovernReader backed by artifact package functions.
func NewReader() GovernReader {
	return &reader{}
}

// NewEnforcer returns a GovernEnforcer backed by artifact package functions.
func NewEnforcer() GovernEnforcer {
	return &enforcer{}
}
