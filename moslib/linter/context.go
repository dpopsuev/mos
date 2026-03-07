package linter

import (
	"os"
	"path/filepath"

	"github.com/dpopsuev/mos/moslib/artifact"
	"github.com/dpopsuev/mos/moslib/dsl"
	"github.com/dpopsuev/mos/moslib/names"
	"github.com/dpopsuev/mos/moslib/schema"
)

// LoadArtifactSchemas is injected by main to load artifact schemas from the
// governance registry, avoiding a circular import between linter and governance.
// Returns schemas keyed by kind. If nil, the linter falls back to built-in defaults.
var LoadArtifactSchemas func(root string) ([]schema.ArtifactSchema, error)

// CustomArtifactSchema is a type alias for backward compatibility.
type CustomArtifactSchema = schema.ArtifactSchema

// ProjectContext holds the loaded .mos/ directory state that validators use.
type ProjectContext struct {
	Root              string
	Config            *dsl.File
	Lexicon        *MergedLexicon
	Keywords          *dsl.KeywordMap
	Layers            []LayerDef
	LayerSet          map[string]bool
	RuleIDs           map[string]string            // id -> file path
	ContractIDs       map[string]string            // id -> file path
	TemplateBlocks    []string
	CustomArtifacts   []schema.ArtifactSchema       // custom artifact type schemas
	ArtifactIDs       map[string]map[string]string  // kind -> id -> file path

	projectRoot       string                        // parent of .mos/
}

// MergedLexicon is the result of overlaying project.mos on default.mos.
type MergedLexicon struct {
	Terms          map[string]string
	ArtifactLabels map[string]string
	LayerNames     map[string]bool
}

// LayerDef is one layer entry from resolution/layers.mos.
type LayerDef struct {
	ID           string
	Level        int64
	InheritsFrom []string
}

// LoadContext reads a .mos/ directory and builds a ProjectContext.
func LoadContext(mosDir string) (*ProjectContext, error) {
	root := filepath.Dir(mosDir)
	ctx := &ProjectContext{
		Root:        mosDir,
		RuleIDs:     make(map[string]string),
		ContractIDs: make(map[string]string),
		LayerSet:    make(map[string]bool),
		ArtifactIDs: make(map[string]map[string]string),
		projectRoot: root,
	}

	if err := ctx.loadLexicon(); err != nil {
		return nil, err
	}
	if err := ctx.loadConfig(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if err := ctx.loadLayers(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if err := ctx.loadTemplate(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	ctx.inventoryRules()
	ctx.inventoryContracts()
	ctx.loadCustomArtifactSchemas()
	ctx.inventoryCustomArtifacts()

	return ctx, nil
}

func (ctx *ProjectContext) loadConfig() error {
	f, err := artifact.ReadConfig(ctx.projectRoot)
	if err != nil {
		return err
	}
	ctx.Config = f
	return nil
}

func (ctx *ProjectContext) loadLexicon() error {
	merged := &MergedLexicon{
		Terms:          make(map[string]string),
		ArtifactLabels: make(map[string]string),
		LayerNames:     make(map[string]bool),
	}

	for _, name := range []string{"default.mos", "project.mos"} {
		f, err := artifact.ReadLexiconFile(ctx.projectRoot, name)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		extractLexicon(f, merged)
	}

	if defaultFile, err := artifact.ReadLexiconFile(ctx.projectRoot, "default.mos"); err == nil {
		ctx.Keywords = dsl.ExtractKeywords(defaultFile)
	}
	if ctx.Keywords == nil {
		ctx.Keywords = dsl.DefaultKeywords()
	}

	ctx.Lexicon = merged
	return nil
}

func extractLexicon(f *dsl.File, merged *MergedLexicon) {
	ab, ok := f.Artifact.(*dsl.ArtifactBlock)
	if !ok {
		return
	}
	for _, item := range ab.Items {
		blk, ok := item.(*dsl.Block)
		if !ok {
			continue
		}
		switch blk.Name {
		case "terms":
			for _, bi := range blk.Items {
				if field, ok := bi.(*dsl.Field); ok {
					if sv, ok := field.Value.(*dsl.StringVal); ok {
						merged.Terms[field.Key] = sv.Text
					}
				}
			}
		case "artifact_labels":
			for _, bi := range blk.Items {
				if field, ok := bi.(*dsl.Field); ok {
					if sv, ok := field.Value.(*dsl.StringVal); ok {
						merged.ArtifactLabels[field.Key] = sv.Text
					}
				}
			}
		case "resolution_layers":
			for _, bi := range blk.Items {
				if field, ok := bi.(*dsl.Field); ok {
					merged.LayerNames[field.Key] = true
				}
				if nb, ok := bi.(*dsl.Block); ok {
					merged.LayerNames[nb.Name] = true
				}
			}
		}
	}
}

func (ctx *ProjectContext) loadLayers() error {
	f, err := artifact.ReadLayers(ctx.projectRoot)
	if err != nil {
		return err
	}
	ab, ok := f.Artifact.(*dsl.ArtifactBlock)
	if !ok {
		return nil
	}
	for _, item := range ab.Items {
		blk, ok := item.(*dsl.Block)
		if !ok || blk.Name != "layer" {
			continue
		}
		ld := LayerDef{ID: blk.Title}
		if ld.ID == "" {
			ld.ID = blk.Name
		}
		if level, ok := astFieldInt(blk.Items, "level"); ok {
			ld.Level = level
		}
		ld.InheritsFrom = astFieldStringSlice(blk.Items, "inherits_from")
		ctx.Layers = append(ctx.Layers, ld)
		ctx.LayerSet[ld.ID] = true
	}
	return nil
}

func (ctx *ProjectContext) loadTemplate() error {
	f, err := artifact.ReadTemplate(ctx.projectRoot)
	if err != nil {
		return err
	}
	ab, ok := f.Artifact.(*dsl.ArtifactBlock)
	if !ok {
		return nil
	}
	for _, item := range ab.Items {
		if blk, ok := item.(*dsl.Block); ok {
			ctx.TemplateBlocks = append(ctx.TemplateBlocks, blk.Name)
		}
	}
	return nil
}

func (ctx *ProjectContext) inventoryRules() {
	ctx.RuleIDs = artifact.ReadRuleInventory(ctx.projectRoot, ctx.Keywords)
}

func (ctx *ProjectContext) inventoryContracts() {
	base := artifact.ReadContractInventory(ctx.projectRoot, ctx.Keywords)
	for id, path := range base {
		ctx.ContractIDs[id] = path
	}
	for _, path := range base {
		f, err := artifact.ReadDSLFile(path, ctx.Keywords)
		if err != nil {
			continue
		}
		if ab, ok := f.Artifact.(*dsl.ArtifactBlock); ok {
			inventoryNestedContracts(ab.Items, path, ctx)
		}
	}
}

// inventoryNestedContracts recursively registers sub-contracts found
// inside umbrella contract-of-contracts. A titled block is a sub-contract
// if it contains a "status" field (the canonical contract marker),
// distinguishing it from other titled blocks like tracker adapters.
func inventoryNestedContracts(items []dsl.Node, parentPath string, ctx *ProjectContext) {
	for _, item := range items {
		blk, ok := item.(*dsl.Block)
		if !ok || blk.Title == "" {
			continue
		}
		if !astHasField(blk.Items, names.FieldStatus) {
			continue
		}
		ctx.ContractIDs[blk.Title] = parentPath
		inventoryNestedContracts(blk.Items, parentPath, ctx)
	}
}

// loadCustomArtifactSchemas populates CustomArtifacts from the governance registry
// via the injected LoadArtifactSchemas function (single source of truth).
// Falls back to schema.DefaultCoreSchemas when the hook is not wired (e.g. in tests).
func (ctx *ProjectContext) loadCustomArtifactSchemas() {
	var schemas []schema.ArtifactSchema

	if LoadArtifactSchemas != nil {
		loaded, err := LoadArtifactSchemas(ctx.projectRoot)
		if err == nil {
			schemas = loaded
		}
	}

	if len(schemas) == 0 {
		for _, s := range schema.DefaultCoreSchemas() {
			if s != nil {
				schemas = append(schemas, *s)
			}
		}
		schemas = append(schemas, ctx.parseConfigArtifactSchemas()...)
	}

	for _, s := range schemas {
		switch s.Kind {
		case "contract", "rule":
			continue
		}
		ctx.CustomArtifacts = append(ctx.CustomArtifacts, s)
	}
}

// parseConfigArtifactSchemas is the legacy fallback that extracts artifact_type
// blocks from config.mos when the registry hook is not available.
func (ctx *ProjectContext) parseConfigArtifactSchemas() []schema.ArtifactSchema {
	if ctx.Config == nil {
		return nil
	}
	ab, ok := ctx.Config.Artifact.(*dsl.ArtifactBlock)
	if !ok {
		return nil
	}
	var out []schema.ArtifactSchema
	for _, item := range ab.Items {
		blk, ok := item.(*dsl.Block)
		if !ok || blk.Name != "artifact_type" || blk.Title == "" {
			continue
		}
		sch := schema.ArtifactSchema{Kind: blk.Title}
		for _, sub := range blk.Items {
			switch v := sub.(type) {
			case *dsl.Field:
				if v.Key == "directory" {
					if sv, ok := v.Value.(*dsl.StringVal); ok {
						sch.Directory = sv.Text
					}
				}
			case *dsl.Block:
				switch v.Name {
				case "fields":
					sch.Fields = parseFieldSchemasFromBlock(v)
				case "lifecycle":
					sch.ActiveStates = dsl.FieldStringSlice(v.Items, "active_states")
					sch.ArchiveStates = dsl.FieldStringSlice(v.Items, "archive_states")
					sch.ExpectsDownstream = parseExpectsDownstreamFromBlock(v)
					sch.UrgencyPropagation = parseUrgencyPropagationFromBlock(v)
				}
			}
		}
		if sch.Directory == "" {
			sch.Directory = blk.Title + "s"
		}
		out = append(out, sch)
	}
	return out
}

func parseExpectsDownstreamFromBlock(lifecycleBlk *dsl.Block) *schema.ExpectsDownstream {
	for _, item := range lifecycleBlk.Items {
		blk, ok := item.(*dsl.Block)
		if !ok || blk.Name != "expects_downstream" {
			continue
		}
		ed := &schema.ExpectsDownstream{}
		for _, fi := range blk.Items {
			field, ok := fi.(*dsl.Field)
			if !ok {
				continue
			}
			if sv, ok := field.Value.(*dsl.StringVal); ok {
				switch field.Key {
				case "via":
					ed.Via = sv.Text
				case "after":
					ed.After = sv.Text
				case "severity":
					ed.Severity = sv.Text
				}
			}
		}
		if ed.Via != "" {
			return ed
		}
	}
	return nil
}

func parseUrgencyPropagationFromBlock(lifecycleBlk *dsl.Block) map[string]string {
	for _, item := range lifecycleBlk.Items {
		blk, ok := item.(*dsl.Block)
		if !ok || blk.Name != "urgency_propagation" {
			continue
		}
		m := make(map[string]string)
		for _, fi := range blk.Items {
			field, ok := fi.(*dsl.Field)
			if !ok {
				continue
			}
			if sv, ok := field.Value.(*dsl.StringVal); ok {
				m[field.Key] = sv.Text
			}
		}
		if len(m) > 0 {
			return m
		}
	}
	return nil
}

func parseFieldSchemasFromBlock(blk *dsl.Block) []schema.FieldSchema {
	var fields []schema.FieldSchema
	for _, item := range blk.Items {
		sub, ok := item.(*dsl.Block)
		if !ok {
			continue
		}
		fs := schema.FieldSchema{Name: sub.Name}
		fs.Required = dsl.FieldBool(sub.Items, "required")
		fs.Enum = dsl.FieldStringSlice(sub.Items, "enum")
		fs.Link = dsl.FieldBool(sub.Items, "link")
		fs.RefKind, _ = dsl.FieldString(sub.Items, "ref_kind")
		fields = append(fields, fs)
	}
	return fields
}

// inventoryCustomArtifacts discovers instances of custom artifact types.
func (ctx *ProjectContext) inventoryCustomArtifacts() {
	for _, sch := range ctx.CustomArtifacts {
		ids := artifact.ReadArtifactInventory(ctx.projectRoot, sch.Directory, sch.Kind)
		if len(ids) > 0 {
			ctx.ArtifactIDs[sch.Kind] = ids
		}
	}
}
