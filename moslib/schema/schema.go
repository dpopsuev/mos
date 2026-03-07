package schema

// TransitionDef describes a gated transition between two enum values.
type TransitionDef struct {
	From       string
	To         string
	VerifiedBy string // e.g. "harness"
}

// FieldSchema describes a single field in an artifact type schema.
// Merges FieldDef (governance) and CustomFieldSchema (linter) into a unified type.
// Default, Ordered, Transitions are governance-specific; linter ignores them.
type FieldSchema struct {
	Name        string
	Required    bool
	Enum        []string
	Default     string
	Ordered     bool
	Transitions []TransitionDef
	Link        bool   // true if this field references another artifact
	RefKind     string // expected artifact kind for the link target (e.g. "specification")
}

// ExpectsDownstream declares that instances of this type expect to be
// referenced by downstream artifacts via a specific link field.
type ExpectsDownstream struct {
	Via      string // link field name (e.g. "satisfies" for need, "justifies" for spec)
	After    string // state threshold -- only flag orphans past this state
	Severity string // "warn" or "error"
}

// ArtifactSchema describes a custom artifact type for discovery and validation.
type ArtifactSchema struct {
	Kind               string
	Directory          string
	Fields             []FieldSchema
	ExpectsDownstream  *ExpectsDownstream
	ActiveStates       []string
	ArchiveStates      []string
	UrgencyPropagation map[string]string
}

// LinkFieldNames returns the names of all fields with Link == true.
func (s *ArtifactSchema) LinkFieldNames() []string {
	var names []string
	for _, f := range s.Fields {
		if f.Link {
			names = append(names, f.Name)
		}
	}
	return names
}

// DefaultCoreSchemas returns built-in ArtifactSchemas for core types (contract,
// specification) that have link fields. Used when config.mos does not define
// artifact_type for these kinds (e.g. linter skips them for dedicated validation).
func DefaultCoreSchemas() map[string]*ArtifactSchema {
	return map[string]*ArtifactSchema{
		"contract": {
			Kind:          "contract",
			Directory:     "contracts",
			ActiveStates:  []string{"draft", "active"},
			ArchiveStates: []string{"complete", "cancelled", "abandoned", "closed", "duplicate"},
			Fields: []FieldSchema{
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
			Fields: []FieldSchema{
				{Name: "satisfies", Link: true, RefKind: "need"},
				{Name: "addresses", Link: true},
			},
		},
	}
}
