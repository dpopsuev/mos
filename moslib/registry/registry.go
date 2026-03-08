package registry

import (
	"fmt"
	"slices"
	"strings"

	"github.com/dpopsuev/mos/moslib/schema"
)

// TransitionDef is a type alias for schema.TransitionDef (backward compatibility).
type TransitionDef = schema.TransitionDef

// HookDef describes a declarative lifecycle hook that fires when child elements
// reach a threshold on a watched field.
type HookDef struct {
	Trigger    string // "on_any" or "on_all"
	WatchField string // field to watch on children (e.g. "status")
	Threshold  string // value to match (e.g. "implemented")
	SetField   string // field to set on parent (e.g. "status")
	SetValue   string // value to set (e.g. "active")
}

// TransitionGate describes a gated lifecycle transition that requires
// a precondition to be met before the transition is allowed.
type TransitionGate struct {
	From string
	To   string
	Gate string // e.g. "criteria_coverage"
}

// LifecycleDef describes the state machine for an artifact type.
type LifecycleDef struct {
	ActiveStates       []string
	ArchiveStates      []string
	Hooks              []HookDef
	ExpectsDownstream  *schema.ExpectsDownstream
	Gates              []TransitionGate
	UrgencyPropagation map[string]string // urgency level -> diagnostic severity
}

// ArtifactTypeDef describes an artifact type (Custom Artifact Definition).
type ArtifactTypeDef struct {
	Kind           string
	Directory      string
	Prefix         string   // primary ID prefix (e.g. "CON", "SPEC")
	Prefixes       []string // additional alias prefixes (e.g. "BUG", "TASK" for contract)
	Core           bool     // true for types with dedicated CLI commands (contract, spec, rule, binder)
	Fields         []schema.FieldSchema
	ScenarioFields []schema.FieldSchema
	Lifecycle      LifecycleDef
	Version        string
	Ledger         bool
}

// ToSchema converts an ArtifactTypeDef into a schema.ArtifactSchema
// so linter and other consumers can use a single schema representation.
func (td ArtifactTypeDef) ToSchema() schema.ArtifactSchema {
	as := schema.ArtifactSchema{
		Kind:      td.Kind,
		Directory: td.Directory,
		Fields:    td.Fields,
	}
	if td.Lifecycle.ExpectsDownstream != nil {
		as.ExpectsDownstream = td.Lifecycle.ExpectsDownstream
	}
	as.ActiveStates = td.Lifecycle.ActiveStates
	as.ArchiveStates = td.Lifecycle.ArchiveStates
	as.UrgencyPropagation = td.Lifecycle.UrgencyPropagation
	return as
}

// Registry holds all known artifact type definitions.
type Registry struct {
	Types      map[string]ArtifactTypeDef
	PrefixKind map[string]string // uppercase prefix → kind (e.g. "CON" → "contract")
}

// ArtifactSchemas returns all artifact type definitions as schema.ArtifactSchema values,
// keyed by kind. This is the single source of truth for schema information.
func (r *Registry) ArtifactSchemas() map[string]*schema.ArtifactSchema {
	out := make(map[string]*schema.ArtifactSchema, len(r.Types))
	for kind, td := range r.Types {
		s := td.ToSchema()
		out[kind] = &s
	}
	return out
}

// DefaultRegistry returns a registry seeded with built-in artifact type definitions.
func DefaultRegistry() *Registry {
	reg := &Registry{
		Types: map[string]ArtifactTypeDef{
			"contract": {
				Kind:      "contract",
				Directory: "contracts",
				Prefix:    "CON",
				Prefixes:  []string{"BUG", "TASK", "FEAT"},
				Core:      true,
				Fields: []schema.FieldSchema{
					{Name: "justifies", Link: true, RefKind: "specification"},
					{Name: "implements", Link: true},
					{Name: "documents", Link: true},
					{Name: "sprint", Link: true, RefKind: "sprint"},
					{Name: "batch", Link: true, RefKind: "batch"},
					{Name: "parent", Link: true, RefKind: "contract"},
					{Name: "depends_on", Link: true, RefKind: "contract"},
				},
			},
			"specification": {
				Kind:      "specification",
				Directory: "specifications",
				Prefix:    "SPEC",
				Core:      true,
				Fields: []schema.FieldSchema{
					{Name: "satisfies", Link: true, RefKind: "need"},
					{Name: "addresses", Link: true},
				},
			},
			"rule": {
				Kind:      "rule",
				Directory: "rules",
				Prefix:    "RULE",
				Core:      true,
			},
			"binder": {
				Kind:      "binder",
				Directory: "binders",
				Prefix:    "BND",
				Core:      true,
			},
			"need": {
				Kind:      "need",
				Directory: "needs",
				Prefix:    "NEED",
			},
			"sprint": {
				Kind:      "sprint",
				Directory: "sprints",
				Prefix:    "SPR",
			},
			"batch": {
				Kind:      "batch",
				Directory: "batches",
				Prefix:    "BAT",
			},
			"architecture": {
				Kind:      "architecture",
				Directory: "architectures",
				Prefix:    "ARCH",
				Fields: []schema.FieldSchema{
					{Name: "resolution"},
				},
			},
			"doc": {
				Kind:      "doc",
				Directory: "docs",
				Prefix:    "DOC",
			},
		},
		PrefixKind: make(map[string]string),
	}
	for kind, td := range reg.Types {
		if td.Prefix != "" {
			reg.PrefixKind[td.Prefix] = kind
		}
		for _, alias := range td.Prefixes {
			reg.PrefixKind[alias] = kind
		}
	}
	return reg
}

// ResolveKindFromID extracts the prefix from an ID (everything before the first
// "-") and returns the matching artifact kind from the registry.
func (r *Registry) ResolveKindFromID(id string) (string, error) {
	prefix := id
	if idx := strings.Index(id, "-"); idx > 0 {
		prefix = id[:idx]
	}
	upper := strings.ToUpper(prefix)
	if kind, ok := r.PrefixKind[upper]; ok {
		return kind, nil
	}
	return "", fmt.Errorf("no artifact type found for ID prefix %q", prefix)
}

// CoreKinds are the artifact types that have dedicated CLI commands.
var CoreKinds = map[string]bool{
	"contract":      true,
	"rule":          true,
	"specification": true,
	"binder":        true,
}

// IsValidStatus checks if a status is valid for the given artifact type.
func (td *ArtifactTypeDef) IsValidStatus(status string) bool {
	if len(td.Lifecycle.ActiveStates) == 0 && len(td.Lifecycle.ArchiveStates) == 0 {
		return true
	}
	return slices.Contains(td.Lifecycle.ActiveStates, status) || slices.Contains(td.Lifecycle.ArchiveStates, status)
}

// IsArchiveStatus checks if a status maps to the archive directory.
func (td *ArtifactTypeDef) IsArchiveStatus(status string) bool {
	return slices.Contains(td.Lifecycle.ArchiveStates, status)
}

// AllStates returns all valid lifecycle states for an artifact type.
func (td *ArtifactTypeDef) AllStates() []string {
	var states []string
	states = append(states, td.Lifecycle.ActiveStates...)
	states = append(states, td.Lifecycle.ArchiveStates...)
	return states
}

// LinkFields returns the names of all fields with Link == true.
func (td *ArtifactTypeDef) LinkFields() []string {
	var names []string
	for _, f := range td.Fields {
		if f.Link {
			names = append(names, f.Name)
		}
	}
	return names
}

// TraceFields returns link fields used for traceability chains (derivation
// hierarchy), excluding self-referential links and organizational links
// like sprint and batch.
func (td *ArtifactTypeDef) TraceFields() []string {
	var names []string
	for _, f := range td.Fields {
		if !f.Link {
			continue
		}
		if f.RefKind == td.Kind || f.RefKind == "sprint" || f.RefKind == "batch" {
			continue
		}
		names = append(names, f.Name)
	}
	return names
}

// AllLinkFields returns a deduplicated set of link field names across all types.
func (reg *Registry) AllLinkFields() []string {
	seen := make(map[string]bool)
	var names []string
	for _, td := range reg.Types {
		for _, f := range td.Fields {
			if f.Link && !seen[f.Name] {
				seen[f.Name] = true
				names = append(names, f.Name)
			}
		}
	}
	slices.Sort(names)
	return names
}
