package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/dpopsuev/mos/moslib/dsl"
	"github.com/dpopsuev/mos/moslib/names"
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

// DefaultRegistry returns a registry seeded with link field metadata for
// built-in types. Custom artifact types are loaded from config.mos via LoadRegistry.
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

// LoadRegistry reads artifact_type blocks (Custom Artifact Definitions) from config.mos.
func LoadRegistry(root string) (*Registry, error) {
	reg := DefaultRegistry()

	configPath := filepath.Join(root, names.MosDir, names.ConfigFile)
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return reg, nil
		}
		return nil, fmt.Errorf("reading config.mos: %w", err)
	}

	f, err := dsl.Parse(string(data), nil)
	if err != nil {
		return nil, fmt.Errorf("parsing config.mos: %w", err)
	}

	ab, ok := f.Artifact.(*dsl.ArtifactBlock)
	if !ok {
		return reg, nil
	}

	dsl.WalkBlocks(ab.Items, func(blk *dsl.Block) bool {
		if blk.Name != "artifact_type" {
			return true
		}
		kind := blk.Title
		if kind == "" {
			return false
		}

		td := parseArtifactTypeDef(kind, blk)

		if defaults, ok := reg.Types[kind]; ok {
			td.Fields = mergeFieldDefs(defaults.Fields, td.Fields)
			td.Core = td.Core || defaults.Core
		}

		if CoreKinds[kind] {
			td.Core = true
		}

		reg.Types[kind] = td
		if td.Prefix != "" {
			reg.PrefixKind[td.Prefix] = kind
		}
		for _, alias := range td.Prefixes {
			reg.PrefixKind[alias] = kind
		}
		return false
	})

	projects, _ := LoadProjects(root)
	dirToKind := make(map[string]string)
	for kind, td := range reg.Types {
		if td.Directory != "" {
			dirToKind[strings.ToLower(td.Directory)] = kind
		}
	}
	for _, p := range projects {
		upper := strings.ToUpper(p.Prefix)
		if kind, ok := dirToKind[strings.ToLower(p.Name)]; ok {
			reg.PrefixKind[upper] = kind
			if td := reg.Types[kind]; td.Prefix == "" {
				td.Prefix = p.Prefix
				reg.Types[kind] = td
			}
		} else {
			if kind, ok := dirToKind["contracts"]; ok {
				reg.PrefixKind[upper] = kind
			}
		}
	}

	return reg, nil
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

// mergeFieldDefs merges default field definitions (e.g. built-in link metadata)
// with config-loaded field definitions. Config-loaded fields take precedence;
// default fields not present in config are appended.
func mergeFieldDefs(defaults, loaded []schema.FieldSchema) []schema.FieldSchema {
	configNames := make(map[string]bool, len(loaded))
	for _, f := range loaded {
		configNames[f.Name] = true
	}
	for _, df := range defaults {
		if !configNames[df.Name] {
			loaded = append(loaded, df)
		}
	}
	return loaded
}

func parseArtifactTypeDef(kind string, blk *dsl.Block) ArtifactTypeDef {
	td := ArtifactTypeDef{Kind: kind}

	td.Directory, _ = dsl.FieldString(blk.Items, "directory")
	td.Prefix, _ = dsl.FieldString(blk.Items, "prefix")
	td.Prefixes = dsl.FieldStringSlice(blk.Items, "prefixes")
	td.Core = dsl.FieldBool(blk.Items, "core")
	td.Version, _ = dsl.FieldString(blk.Items, "version")
	td.Ledger = dsl.FieldBool(blk.Items, "ledger")

	if fb := dsl.FindBlock(blk.Items, "fields"); fb != nil {
		td.Fields = parseFieldDefs(fb)
	}
	if sb := dsl.FindBlock(blk.Items, "scenario_fields"); sb != nil {
		td.ScenarioFields = parseFieldDefs(sb)
	}
	if lb := dsl.FindBlock(blk.Items, "lifecycle"); lb != nil {
		td.Lifecycle = parseLifecycleDef(lb)
	}

	if td.Directory == "" {
		td.Directory = kind + "s"
	}

	return td
}

func parseFieldDefs(blk *dsl.Block) []schema.FieldSchema {
	var fields []schema.FieldSchema
	dsl.WalkBlocks(blk.Items, func(sub *dsl.Block) bool {
		fd := schema.FieldSchema{Name: sub.Name}
		fd.Required = dsl.FieldBool(sub.Items, "required")
		fd.Enum = dsl.FieldStringSlice(sub.Items, "enum")
		fd.Default, _ = dsl.FieldString(sub.Items, "default")
		fd.Ordered = dsl.FieldBool(sub.Items, "ordered")
		fd.Link = dsl.FieldBool(sub.Items, "link")
		fd.RefKind, _ = dsl.FieldString(sub.Items, "ref_kind")
		if tb := dsl.FindBlock(sub.Items, "transitions"); tb != nil {
			fd.Transitions = parseTransitionDefs(tb)
		}
		fields = append(fields, fd)
		return false
	})
	return fields
}

func parseTransitionDefs(blk *dsl.Block) []schema.TransitionDef {
	var defs []schema.TransitionDef
	dsl.WalkBlocks(blk.Items, func(sub *dsl.Block) bool {
		td := schema.TransitionDef{From: sub.Name}
		td.To, _ = dsl.FieldString(sub.Items, "to")
		td.VerifiedBy, _ = dsl.FieldString(sub.Items, "verified_by")
		defs = append(defs, td)
		return false
	})
	return defs
}

func parseLifecycleDef(blk *dsl.Block) LifecycleDef {
	var ld LifecycleDef
	ld.ActiveStates = dsl.FieldStringSlice(blk.Items, "active_states")
	ld.ArchiveStates = dsl.FieldStringSlice(blk.Items, "archive_states")
	if hb := dsl.FindBlock(blk.Items, "hooks"); hb != nil {
		ld.Hooks = parseHookDefs(hb)
	}
	if eb := dsl.FindBlock(blk.Items, "expects_downstream"); eb != nil {
		ld.ExpectsDownstream = parseExpectsDownstream(eb)
	}
	dsl.WalkBlocks(blk.Items, func(b *dsl.Block) bool {
		if b.Name == "transition" {
			if g := parseTransitionGate(b); g != nil {
				ld.Gates = append(ld.Gates, *g)
			}
		}
		return false
	})
	if ub := dsl.FindBlock(blk.Items, "urgency_propagation"); ub != nil {
		ld.UrgencyPropagation = parseUrgencyPropagation(ub)
	}
	return ld
}

func parseExpectsDownstream(blk *dsl.Block) *schema.ExpectsDownstream {
	ed := &schema.ExpectsDownstream{}
	ed.Via, _ = dsl.FieldString(blk.Items, "via")
	ed.After, _ = dsl.FieldString(blk.Items, "after")
	ed.Severity, _ = dsl.FieldString(blk.Items, "severity")
	if ed.Via == "" {
		return nil
	}
	return ed
}

func parseUrgencyPropagation(blk *dsl.Block) map[string]string {
	raw := dsl.ToMap(&dsl.ArtifactBlock{Items: blk.Items})
	m := make(map[string]string)
	for k, v := range raw {
		if k == "kind" || k == "id" {
			continue
		}
		if s, ok := v.(string); ok {
			m[k] = s
		}
	}
	if len(m) == 0 {
		return nil
	}
	return m
}

func parseTransitionGate(blk *dsl.Block) *TransitionGate {
	g := &TransitionGate{From: blk.Title}
	g.To, _ = dsl.FieldString(blk.Items, "to")
	g.Gate, _ = dsl.FieldString(blk.Items, "gate")
	if g.From == "" || g.To == "" || g.Gate == "" {
		return nil
	}
	return g
}

func parseHookDefs(blk *dsl.Block) []HookDef {
	var hooks []HookDef
	dsl.WalkBlocks(blk.Items, func(sub *dsl.Block) bool {
		hd := HookDef{Trigger: sub.Name}
		hd.WatchField, _ = dsl.FieldString(sub.Items, "watch_field")
		hd.Threshold, _ = dsl.FieldString(sub.Items, "threshold")
		hd.SetField, _ = dsl.FieldString(sub.Items, "set_field")
		hd.SetValue, _ = dsl.FieldString(sub.Items, "set_value")
		hooks = append(hooks, hd)
		return false
	})
	return hooks
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

// AllStates returns all valid lifecycle states.
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

func pluralize(s string) string {
	if strings.HasSuffix(s, "s") || strings.HasSuffix(s, "x") || strings.HasSuffix(s, "z") ||
		strings.HasSuffix(s, "sh") || strings.HasSuffix(s, "ch") {
		return s + "es"
	}
	if strings.HasSuffix(s, "y") && len(s) > 1 {
		c := s[len(s)-2]
		if c != 'a' && c != 'e' && c != 'i' && c != 'o' && c != 'u' {
			return s[:len(s)-1] + "ies"
		}
	}
	return s + "s"
}

// AddArtifactType adds a new artifact_type block to config.mos with default fields and lifecycle.
func AddArtifactType(root, kind, directory string) error {
	configPath := filepath.Join(root, names.MosDir, names.ConfigFile)
	return dsl.WithArtifact(configPath, func(ab *dsl.ArtifactBlock) error {
		exists := false
		dsl.WalkBlocks(ab.Items, func(blk *dsl.Block) bool {
			if blk.Name == "artifact_type" && blk.Title == kind {
				exists = true
				return false
			}
			return true
		})
		if exists {
			return fmt.Errorf("artifact_type %q already exists in config.mos", kind)
		}

		if directory == "" {
			directory = pluralize(kind)
		}

		fieldsBlk := &dsl.Block{
			Name: "fields",
			Items: []dsl.Node{
				&dsl.Block{Name: names.FieldTitle, Items: []dsl.Node{
					&dsl.Field{Key: "required", Value: &dsl.BoolVal{Val: true}},
				}},
				&dsl.Block{Name: names.FieldStatus, Items: []dsl.Node{
					&dsl.Field{Key: "required", Value: &dsl.BoolVal{Val: true}},
					&dsl.Field{Key: "enum", Value: &dsl.ListVal{Items: []dsl.Value{
						&dsl.StringVal{Text: names.StatusDraft},
						&dsl.StringVal{Text: names.StatusActive},
						&dsl.StringVal{Text: "superseded"},
						&dsl.StringVal{Text: "retired"},
					}}},
				}},
			},
		}

		lifecycleBlk := &dsl.Block{
			Name: "lifecycle",
			Items: []dsl.Node{
				&dsl.Field{Key: "active_states", Value: &dsl.ListVal{Items: []dsl.Value{
					&dsl.StringVal{Text: names.StatusDraft},
					&dsl.StringVal{Text: names.StatusActive},
				}}},
				&dsl.Field{Key: "archive_states", Value: &dsl.ListVal{Items: []dsl.Value{
					&dsl.StringVal{Text: "superseded"},
					&dsl.StringVal{Text: "retired"},
				}}},
			},
		}

		newBlk := &dsl.Block{
			Name:  "artifact_type",
			Title: kind,
			Items: []dsl.Node{
				&dsl.Field{Key: "directory", Value: &dsl.StringVal{Text: directory}},
				fieldsBlk,
				lifecycleBlk,
			},
		}

		ab.Items = append(ab.Items, newBlk)
		return nil
	})
}

// RemoveArtifactType removes an artifact_type block from config.mos.
func RemoveArtifactType(root, kind string) error {
	configPath := filepath.Join(root, names.MosDir, names.ConfigFile)
	return dsl.WithArtifact(configPath, func(ab *dsl.ArtifactBlock) error {
		if !dsl.RemoveNamedBlock(&ab.Items, "artifact_type", kind) {
			return fmt.Errorf("artifact_type %q not found in config.mos", kind)
		}
		return nil
	})
}

// RemoveProject removes a project block from config.mos.
func RemoveProject(root, name string) error {
	configPath := filepath.Join(root, names.MosDir, names.ConfigFile)
	return dsl.WithArtifact(configPath, func(ab *dsl.ArtifactBlock) error {
		if !dsl.RemoveNamedBlock(&ab.Items, "project", name) {
			return fmt.Errorf("project %q not found in config.mos", name)
		}
		return nil
	})
}

// FieldOpts describes a field to add to an artifact_type CAD.
type FieldOpts struct {
	Name     string
	Required bool
	Enum     []string
}

// AddFieldToType appends a field to an existing artifact_type's fields block.
func AddFieldToType(root, kind string, field FieldOpts) error {
	configPath := filepath.Join(root, names.MosDir, names.ConfigFile)
	return dsl.WithArtifact(configPath, func(ab *dsl.ArtifactBlock) error {
		var typeBlk *dsl.Block
		dsl.WalkBlocks(ab.Items, func(blk *dsl.Block) bool {
			if blk.Name == "artifact_type" && blk.Title == kind {
				typeBlk = blk
				return false
			}
			return true
		})
		if typeBlk == nil {
			return fmt.Errorf("artifact_type %q not found in config.mos", kind)
		}

		fieldsBlk := dsl.FindBlock(typeBlk.Items, "fields")
		if fieldsBlk == nil {
			return fmt.Errorf("artifact_type %q has no fields block", kind)
		}

		if dsl.FindBlock(fieldsBlk.Items, field.Name) != nil {
			return fmt.Errorf("field %q already exists in artifact_type %q", field.Name, kind)
		}

		newField := &dsl.Block{Name: field.Name}
		if field.Required {
			newField.Items = append(newField.Items,
				&dsl.Field{Key: "required", Value: &dsl.BoolVal{Val: true}})
		}
		if len(field.Enum) > 0 {
			enumItems := make([]dsl.Value, len(field.Enum))
			for i, v := range field.Enum {
				enumItems[i] = &dsl.StringVal{Text: v}
			}
			newField.Items = append(newField.Items,
				&dsl.Field{Key: "enum", Value: &dsl.ListVal{Items: enumItems}})
		}

		statusBlk := dsl.FindBlock(fieldsBlk.Items, names.FieldStatus)
		statusIdx := -1
		if statusBlk != nil {
			statusIdx = slices.IndexFunc(fieldsBlk.Items, func(n dsl.Node) bool {
				b, ok := n.(*dsl.Block)
				return ok && b == statusBlk
			})
		}
		if statusIdx >= 0 {
			items := make([]dsl.Node, 0, len(fieldsBlk.Items)+1)
			items = append(items, fieldsBlk.Items[:statusIdx]...)
			items = append(items, newField)
			items = append(items, fieldsBlk.Items[statusIdx:]...)
			fieldsBlk.Items = items
		} else {
			fieldsBlk.Items = append(fieldsBlk.Items, newField)
		}
		return nil
	})
}

// SetTypeLifecycle replaces the lifecycle states for an artifact_type.
func SetTypeLifecycle(root, kind string, activeStates, archiveStates []string) error {
	configPath := filepath.Join(root, names.MosDir, names.ConfigFile)
	makeList := func(vals []string) *dsl.ListVal {
		items := make([]dsl.Value, len(vals))
		for i, v := range vals {
			items[i] = &dsl.StringVal{Text: v}
		}
		return &dsl.ListVal{Items: items}
	}
	return dsl.WithArtifact(configPath, func(ab *dsl.ArtifactBlock) error {
		var typeBlk *dsl.Block
		dsl.WalkBlocks(ab.Items, func(blk *dsl.Block) bool {
			if blk.Name == "artifact_type" && blk.Title == kind {
				typeBlk = blk
				return false
			}
			return true
		})
		if typeBlk == nil {
			return fmt.Errorf("artifact_type %q not found in config.mos", kind)
		}

		lifecycleBlk := dsl.FindBlock(typeBlk.Items, "lifecycle")
		if lifecycleBlk == nil {
			return fmt.Errorf("artifact_type %q has no lifecycle block", kind)
		}

		if f := dsl.FindField(lifecycleBlk.Items, "active_states"); f != nil {
			f.Value = makeList(activeStates)
		}
		if f := dsl.FindField(lifecycleBlk.Items, "archive_states"); f != nil {
			f.Value = makeList(archiveStates)
		}
		return nil
	})
}

// SetTypeDirectory updates the directory field for an artifact_type.
func SetTypeDirectory(root, kind, directory string) error {
	configPath := filepath.Join(root, names.MosDir, names.ConfigFile)
	return dsl.WithArtifact(configPath, func(ab *dsl.ArtifactBlock) error {
		var typeBlk *dsl.Block
		dsl.WalkBlocks(ab.Items, func(blk *dsl.Block) bool {
			if blk.Name == "artifact_type" && blk.Title == kind {
				typeBlk = blk
				return false
			}
			return true
		})
		if typeBlk == nil {
			return fmt.Errorf("artifact_type %q not found in config.mos", kind)
		}

		f := dsl.FindField(typeBlk.Items, "directory")
		if f == nil {
			return fmt.Errorf("artifact_type %q has no directory field", kind)
		}
		f.Value = &dsl.StringVal{Text: directory}
		return nil
	})
}

// SetFieldEnum updates the enum for an existing field in an artifact_type.
func SetFieldEnum(root, kind, fieldName string, enum []string) error {
	configPath := filepath.Join(root, names.MosDir, names.ConfigFile)
	enumItems := make([]dsl.Value, len(enum))
	for i, v := range enum {
		enumItems[i] = &dsl.StringVal{Text: v}
	}
	return dsl.WithArtifact(configPath, func(ab *dsl.ArtifactBlock) error {
		var typeBlk *dsl.Block
		dsl.WalkBlocks(ab.Items, func(blk *dsl.Block) bool {
			if blk.Name == "artifact_type" && blk.Title == kind {
				typeBlk = blk
				return false
			}
			return true
		})
		if typeBlk == nil {
			return fmt.Errorf("artifact_type %q not found in config.mos", kind)
		}

		fieldsBlk := dsl.FindBlock(typeBlk.Items, "fields")
		if fieldsBlk == nil {
			return fmt.Errorf("artifact_type %q has no fields block", kind)
		}

		fieldBlk := dsl.FindBlock(fieldsBlk.Items, fieldName)
		if fieldBlk == nil {
			return fmt.Errorf("field %q not found in artifact_type %q", fieldName, kind)
		}

		if f := dsl.FindField(fieldBlk.Items, "enum"); f != nil {
			f.Value = &dsl.ListVal{Items: enumItems}
		} else {
			fieldBlk.Items = append(fieldBlk.Items,
				&dsl.Field{Key: "enum", Value: &dsl.ListVal{Items: enumItems}})
		}
		return nil
	})
}
