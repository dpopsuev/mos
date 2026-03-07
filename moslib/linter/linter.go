package linter

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dpopsuev/mos/moslib/dsl"
	"github.com/dpopsuev/mos/moslib/schema"
)

// Diagnostic represents a single lint finding.
type Diagnostic struct {
	File            string   `json:"file"`
	Line            int      `json:"line,omitempty"`
	Severity        Severity `json:"severity"`
	Rule            string   `json:"rule"`
	Message         string   `json:"message"`
	ArtifactID      string   `json:"artifact_id,omitempty"`
	SuggestedAction string   `json:"suggested_action,omitempty"`
	Expected        []string `json:"expected,omitempty"`
	Got             string   `json:"got,omitempty"`
}

// Severity classifies a lint finding.
type Severity int

const (
	SeverityError Severity = iota
	SeverityWarning
	SeverityInfo
)

func (s Severity) String() string {
	switch s {
	case SeverityError:
		return "error"
	case SeverityWarning:
		return "warning"
	case SeverityInfo:
		return "info"
	default:
		return "unknown"
	}
}

func (s Severity) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", s.String())), nil
}

// Linter validates .mos/ artifacts against the format specification.
type Linter struct{}

// Lint validates all artifacts under root. If root itself is a .mos/ directory
// it is used directly; otherwise .mos/ is expected as a child.
func (l *Linter) Lint(root string) ([]Diagnostic, error) {
	mosDir := root
	if filepath.Base(root) != ".mos" {
		mosDir = filepath.Join(root, ".mos")
	}

	if _, err := os.Stat(mosDir); err != nil {
		return nil, fmt.Errorf("cannot find .mos directory: %w", err)
	}

	ctx, err := LoadContext(mosDir)
	if err != nil {
		return nil, fmt.Errorf("loading context: %w", err)
	}

	var diags []Diagnostic

	diags = append(diags, validateConfig(ctx)...)
	diags = append(diags, validateDeclaration(ctx)...)

	for _, sub := range []string{"mechanical", "interpretive"} {
		dir := filepath.Join(mosDir, "rules", sub)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".mos") {
				continue
			}
			diags = append(diags, validateRule(filepath.Join(dir, e.Name()), ctx)...)
		}
	}

	for _, sub := range []string{"active", "archive"} {
		dir := filepath.Join(mosDir, "contracts", sub)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			contractPath := filepath.Join(dir, e.Name(), "contract.mos")
			if _, err := os.Stat(contractPath); err == nil {
				diags = append(diags, validateContract(contractPath, ctx)...)
			}
		}
	}

	diags = append(diags, validateArtifactTypeBlocks(ctx)...)

	for _, schema := range ctx.CustomArtifacts {
		for _, sub := range []string{"active", "archive"} {
			dir := filepath.Join(mosDir, schema.Directory, sub)
			entries, err := os.ReadDir(dir)
			if err != nil {
				continue
			}
			for _, e := range entries {
				if !e.IsDir() {
					continue
				}
				artPath := filepath.Join(dir, e.Name(), schema.Kind+".mos")
				if _, err := os.Stat(artPath); err == nil {
					diags = append(diags, ValidateGenericFile(artPath, schema, ctx)...)
				}
			}
		}
	}

	diags = append(diags, validateSpecEnforcement(mosDir, ctx)...)

	diags = append(diags, validateCrossRefs(ctx)...)

	diags = append(diags, validateSlugs(ctx)...)

	diags = append(diags, validateIDFormat(ctx)...)

	diags = append(diags, validateDirectoryPlacement(ctx)...)

	diags = append(diags, RunStructuralChecks(root, mosDir)...)

	diags = append(diags, validateSprintOrdering(ctx)...)

	diags = append(diags, validateTrajectory(root, mosDir)...)

	enrichDiagnostics(diags)
	return diags, nil
}

// LintFiles validates only the specified files but loads full context for
// cross-reference checks. Diagnostics outside the file set are filtered out.
func (l *Linter) LintFiles(root string, files []string) ([]Diagnostic, error) {
	allDiags, err := l.Lint(root)
	if err != nil {
		return nil, err
	}
	fileSet := make(map[string]bool, len(files))
	for _, f := range files {
		fileSet[filepath.Clean(f)] = true
	}
	var filtered []Diagnostic
	for _, d := range allDiags {
		if fileSet[filepath.Clean(d.File)] {
			filtered = append(filtered, d)
		}
	}
	return filtered, nil
}

// FilterNewOnly returns diagnostics from all that are not present in baseline.
// Matching is by File + Rule + Message.
func FilterNewOnly(all, baseline []Diagnostic) []Diagnostic {
	type key struct{ file, rule, msg string }
	baseSet := make(map[key]int, len(baseline))
	for _, d := range baseline {
		baseSet[key{d.File, d.Rule, d.Message}]++
	}
	var result []Diagnostic
	for _, d := range all {
		k := key{d.File, d.Rule, d.Message}
		if baseSet[k] > 0 {
			baseSet[k]--
		} else {
			result = append(result, d)
		}
	}
	return result
}

// ExtractArtifactID derives an artifact ID from a .mos file path.
// Paths follow the convention .mos/<kind>/<active|archive>/<ID>/<file>.mos.
// Falls back to the filename stem if the path does not match.
func ExtractArtifactID(filePath string) string {
	parts := strings.Split(filepath.ToSlash(filePath), "/")
	for i, p := range parts {
		if p == ".mos" && i+3 < len(parts) {
			sub := parts[i+2]
			if sub == "active" || sub == "archive" {
				if i+3 < len(parts) {
					return parts[i+3]
				}
			}
		}
	}
	base := filepath.Base(filePath)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

var suggestedActions = map[string]string{
	"dsl-parse":            "Fix the DSL syntax error in the artifact file",
	"config-schema":        "Add the missing required block or field to config.mos",
	"config-enum":          "Set the field to one of the allowed enum values",
	"crossref-rule":        "Update the reference to point to an existing artifact ID",
	"layer-ref":            "Use a layer name defined in resolution/layers.mos",
	"slug-missing":         "Add a slug field to the artifact block",
	"slug-format":          "Change the slug to lowercase-hyphenated format",
	"slug-unique":          "Choose a slug that is not already used by another artifact",
	"id-format":            "Use the ID format PREFIX-YYYY-NNN matching the artifact kind",
	"id-prefix":            "Change the ID prefix to match the expected kind prefix",
	"id-collision":         "Rename one of the duplicate IDs to be unique",
	"directory-placement":  "Move the artifact to the correct active/ or archive/ directory",
	"sprint-order":         "Close sprints in sequential numeric order",
	"lifecycle-chain":      "Ensure the status transition follows the defined lifecycle",
	"lifecycle-orphan":     "Add a lifecycle block to the artifact_type in config.mos",
	"terminal-in-active":   "Move the artifact with terminal status to archive/",
	"stale-sprint-ref":     "Update the sprint reference to a current active sprint",
	"scope-violation":      "Remove or correct the out-of-scope reference",
	"blame-ref-missing":    "Add a blame block referencing the originating commit or author",
	"trajectory-stall":     "Investigate stalled quality axis and take corrective action",
	"trajectory-regression": "Address the quality regression before it compounds",
}

func enrichDiagnostics(diags []Diagnostic) {
	for i := range diags {
		if diags[i].ArtifactID == "" {
			diags[i].ArtifactID = ExtractArtifactID(diags[i].File)
		}
		if diags[i].SuggestedAction == "" {
			diags[i].SuggestedAction = suggestedActions[diags[i].Rule]
		}
	}
}

func makeParseDiag(path string, err error) Diagnostic {
	d := Diagnostic{File: path, Severity: SeverityError, Rule: "dsl-parse", Message: err.Error()}
	var pe *dsl.ParseError
	if errors.As(err, &pe) {
		d.Line = pe.Line
		d.Expected = pe.Expected
		d.Got = pe.Got
	}
	return d
}

func validateDirectoryPlacement(ctx *ProjectContext) []Diagnostic {
	var diags []Diagnostic

	allSchemas := collectAllArtifactSchemas(ctx)
	for _, sch := range allSchemas {
		if len(sch.ActiveStates) == 0 && len(sch.ArchiveStates) == 0 {
			continue
		}
		archiveSet := make(map[string]bool, len(sch.ArchiveStates))
		for _, s := range sch.ArchiveStates {
			archiveSet[s] = true
		}
		for _, pair := range []struct {
			sub, label, expected string
			isMisplaced          func(string) bool
		}{
			{"active", "active", "archive", func(s string) bool { return archiveSet[s] }},
			{"archive", "archive", "active", func(s string) bool { return !archiveSet[s] && s != "" }},
		} {
			dir := filepath.Join(ctx.Root, sch.Directory, pair.sub)
			entries, err := os.ReadDir(dir)
			if err != nil {
				continue
			}
			for _, e := range entries {
				if !e.IsDir() {
					continue
				}
				artPath := filepath.Join(dir, e.Name(), sch.Kind+".mos")
				f, err := parseDSLFile(artPath, ctx.Keywords)
				if err != nil {
					continue
				}
				ab, ok := f.Artifact.(*dsl.ArtifactBlock)
				if !ok {
					continue
				}
				status, _ := dsl.FieldString(ab.Items, "status")
				if pair.isMisplaced(status) {
					diags = append(diags, Diagnostic{
						File:     artPath,
						Severity: SeverityWarning,
						Rule:     "directory-placement",
						Message:  fmt.Sprintf("artifact %s has status %q but is in %s/ — expected %s/", e.Name(), status, pair.label, pair.expected),
					})
				}
			}
		}
	}
	return diags
}

// collectAllArtifactSchemas returns lifecycle-aware schemas for every
// artifact_type in config.mos (including contract and rule, which the
// CustomArtifacts list normally skips).
func collectAllArtifactSchemas(ctx *ProjectContext) []schema.ArtifactSchema {
	schemas := append([]schema.ArtifactSchema{}, ctx.CustomArtifacts...)
	if ctx.Config == nil {
		return schemas
	}
	ab, ok := ctx.Config.Artifact.(*dsl.ArtifactBlock)
	if !ok {
		return schemas
	}
	known := make(map[string]bool, len(schemas))
	for _, s := range schemas {
		known[s.Kind] = true
	}
	for _, item := range ab.Items {
		blk, ok := item.(*dsl.Block)
		if !ok || blk.Name != "artifact_type" || blk.Title == "" {
			continue
		}
		if known[blk.Title] {
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
				if v.Name == "lifecycle" {
					sch.ActiveStates = dsl.FieldStringSlice(v.Items, "active_states")
					sch.ArchiveStates = dsl.FieldStringSlice(v.Items, "archive_states")
				}
			}
		}
		if sch.Directory == "" {
			sch.Directory = blk.Title + "s"
		}
		if len(sch.ActiveStates) > 0 || len(sch.ArchiveStates) > 0 {
			schemas = append(schemas, sch)
		}
	}
	return schemas
}
