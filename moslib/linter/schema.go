package linter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dpopsuev/mos/moslib/dsl"
	"github.com/dpopsuev/mos/moslib/names"
	"github.com/dpopsuev/mos/moslib/schema"
)

func validateConfig(ctx *ProjectContext) []Diagnostic {
	path := filepath.Join(ctx.Root, "config.mos")
	f, err := parseDSLFile(path, ctx.Keywords)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return []Diagnostic{makeParseDiag(path, err)}
	}

	ab, ok := f.Artifact.(*dsl.ArtifactBlock)
	if !ok {
		return []Diagnostic{{File: path, Severity: SeverityError, Rule: "config-schema", Message: "invalid artifact"}}
	}

	var diags []Diagnostic

	mosBlock := astFindBlock(ab.Items, "mos")
	if mosBlock == nil {
		diags = append(diags, Diagnostic{File: path, Severity: SeverityError, Rule: "config-schema", Message: "missing required block 'mos'"})
	} else {
		if !astHasField(mosBlock.Items, "version") {
			diags = append(diags, Diagnostic{File: path, Severity: SeverityError, Rule: "config-schema", Message: "missing required field 'version' in mos block"})
		}
	}

	backend := astFindBlock(ab.Items, "backend")
	if backend != nil {
		if t, ok := astFieldString(backend.Items, "type"); ok {
			if t != "git" && t != "native" {
				diags = append(diags, Diagnostic{File: path, Severity: SeverityError, Rule: "config-enum", Message: fmt.Sprintf("backend.type = %q: must be \"git\" or \"native\"", t)})
			}
		}
	}

	// names.model is a free-form human-protocol label (e.g. "bdfl",
	// "committee", "consensus", "federation", "meritocracy"). No enum
	// validation -- users define their own governance lexicon.

	diags = append(diags, validateProjectBlocks(path, ab.Items)...)

	return diags
}

func validateProjectBlocks(path string, items []dsl.Node) []Diagnostic {
	var diags []Diagnostic
	prefixes := map[string]string{}
	defaultCount := 0

	for _, item := range items {
		blk, ok := item.(*dsl.Block)
		if !ok || blk.Name != "project" {
			continue
		}
		name := blk.Title
		if name == "" {
			diags = append(diags, Diagnostic{File: path, Severity: SeverityError, Rule: "config-project", Message: "project block requires a name"})
			continue
		}

		prefix, hasPrefix := astFieldString(blk.Items, "prefix")
		if !hasPrefix || prefix == "" {
			diags = append(diags, Diagnostic{File: path, Severity: SeverityError, Rule: "config-project", Message: fmt.Sprintf("project %q: missing required field 'prefix'", name)})
		} else if prev, dup := prefixes[prefix]; dup {
			diags = append(diags, Diagnostic{File: path, Severity: SeverityError, Rule: "config-project", Message: fmt.Sprintf("project %q: duplicate prefix %q (already used by %q)", name, prefix, prev)})
		} else {
			prefixes[prefix] = name
		}

		_, hasSeq := astFieldInt(blk.Items, "sequence")
		if !hasSeq {
			diags = append(diags, Diagnostic{File: path, Severity: SeverityError, Rule: "config-project", Message: fmt.Sprintf("project %q: missing required field 'sequence'", name)})
		}

		if dsl.FieldBool(blk.Items, "default") {
			defaultCount++
		}
	}

	if defaultCount > 1 {
		diags = append(diags, Diagnostic{File: path, Severity: SeverityError, Rule: "config-project", Message: fmt.Sprintf("multiple projects marked as default (%d); at most one allowed", defaultCount)})
	}

	return diags
}

func validateDeclaration(ctx *ProjectContext) []Diagnostic {
	path := filepath.Join(ctx.Root, "declaration.mos")
	f, err := parseDSLFile(path, ctx.Keywords)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return []Diagnostic{makeParseDiag(path, err)}
	}

	ab, ok := f.Artifact.(*dsl.ArtifactBlock)
	if !ok {
		return []Diagnostic{{File: path, Severity: SeverityError, Rule: "declaration-schema", Message: "invalid artifact"}}
	}

	var diags []Diagnostic
	for _, field := range []string{"name", "created"} {
		if !astHasField(ab.Items, field) {
			diags = append(diags, Diagnostic{File: path, Severity: SeverityError, Rule: "declaration-schema", Message: fmt.Sprintf("missing required field '%s'", field)})
		}
	}

	return diags
}

// ValidateRuleFile is the exported entry point for single-rule validation (used by benchmarks).
func ValidateRuleFile(path string, ctx *ProjectContext) []Diagnostic {
	return validateRule(path, ctx)
}

func validateRule(path string, ctx *ProjectContext) []Diagnostic {
	f, err := parseDSLFile(path, ctx.Keywords)
	if err != nil {
		return []Diagnostic{makeParseDiag(path, err)}
	}

	ab, ok := f.Artifact.(*dsl.ArtifactBlock)
	if !ok {
		return []Diagnostic{{File: path, Severity: SeverityError, Rule: "rule-schema", Message: "invalid artifact"}}
	}

	var diags []Diagnostic

	if ab.Name == "" {
		diags = append(diags, Diagnostic{File: path, Severity: SeverityError, Rule: "rule-schema", Message: "rule artifact must have a name (id)"})
	}

	for _, field := range []string{"name", "type", "scope", "enforcement"} {
		if !astHasField(ab.Items, field) {
			diags = append(diags, Diagnostic{File: path, Severity: SeverityError, Rule: "rule-schema", Message: fmt.Sprintf("missing required field '%s'", field)})
		}
	}

	if t, ok := astFieldString(ab.Items, "type"); ok {
		if t != "mechanical" && t != "interpretive" {
			diags = append(diags, Diagnostic{File: path, Severity: SeverityError, Rule: "rule-enum", Message: fmt.Sprintf("type = %q: must be \"mechanical\" or \"interpretive\"", t)})
		}
	}

	if e, ok := astFieldString(ab.Items, "enforcement"); ok {
		if e != "error" && e != "warning" && e != "info" {
			diags = append(diags, Diagnostic{File: path, Severity: SeverityError, Rule: "rule-enum", Message: fmt.Sprintf("enforcement = %q: must be \"error\", \"warning\", or \"info\"", e)})
		}
	}

	ruleType, _ := astFieldString(ab.Items, "type")
	features := astFindFeatures(ab.Items)
	includes := astFindIncludes(ab.Items)
	hasHarness := astFindBlock(ab.Items, "harness") != nil
	if ruleType != "interpretive" && len(features) == 0 && len(includes) == 0 && !hasHarness {
		diags = append(diags, Diagnostic{File: path, Severity: SeverityError, Rule: "rule-schema", Message: "rule must have at least one feature block, spec with includes, or harness block"})
	}

	for _, inc := range includes {
		target := inc.Path
		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(path), target)
		}
		if _, err := os.Stat(target); err != nil {
			diags = append(diags, Diagnostic{File: path, Line: inc.Line, Severity: SeverityError, Rule: "include-resolve", Message: fmt.Sprintf("include %q: file not found", inc.Path)})
		}
	}

	return diags
}

// ValidateContractFile is the exported entry point for single-contract validation.
func ValidateContractFile(path string, ctx *ProjectContext) []Diagnostic {
	return validateContract(path, ctx)
}

func validateContract(path string, ctx *ProjectContext) []Diagnostic {
	f, err := parseDSLFile(path, ctx.Keywords)
	if err != nil {
		return []Diagnostic{makeParseDiag(path, err)}
	}

	ab, ok := f.Artifact.(*dsl.ArtifactBlock)
	if !ok {
		return []Diagnostic{{File: path, Severity: SeverityError, Rule: "contract-schema", Message: "invalid artifact"}}
	}

	var diags []Diagnostic

	if ab.Name == "" {
		diags = append(diags, Diagnostic{File: path, Severity: SeverityError, Rule: "contract-schema", Message: "contract artifact must have a name (id)"})
	}

	for _, field := range []string{names.FieldTitle, names.FieldStatus} {
		if !astHasField(ab.Items, field) {
			diags = append(diags, Diagnostic{File: path, Severity: SeverityError, Rule: "contract-schema", Message: fmt.Sprintf("missing required field '%s'", field)})
		}
	}

	bill := astFindBlock(ab.Items, "bill")
	if bill == nil {
		diags = append(diags, Diagnostic{File: path, Severity: SeverityInfo, Rule: "contract-schema", Message: "no 'bill' block present (expected for governance lifecycle)"})
	} else {
		for _, field := range []string{"introduced_by", "introduced_at", "intent"} {
			if !astHasField(bill.Items, field) {
				diags = append(diags, Diagnostic{File: path, Severity: SeverityError, Rule: "contract-schema", Message: fmt.Sprintf("missing required field '%s' in bill block", field)})
			}
		}
	}

	for _, inc := range astFindIncludes(ab.Items) {
		target := inc.Path
		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(path), target)
		}
		if _, err := os.Stat(target); err != nil {
			diags = append(diags, Diagnostic{File: path, Line: inc.Line, Severity: SeverityError, Rule: "include-resolve", Message: fmt.Sprintf("include %q: file not found", inc.Path)})
		}
	}

	scope := astFindBlock(ab.Items, "scope")
	if scope != nil {
		for _, item := range scope.Items {
			field, ok := item.(*dsl.Field)
			if !ok || field.Key != "depends_on" {
				continue
			}
			lv, ok := field.Value.(*dsl.ListVal)
			if !ok {
				continue
			}
			for _, v := range lv.Items {
				sv, ok := v.(*dsl.StringVal)
				if !ok {
					continue
				}
				if _, exists := ctx.ContractIDs[sv.Text]; !exists {
					diags = append(diags, Diagnostic{
						File:     path,
						Severity: SeverityError,
						Rule:     "depends_on-resolve",
						Message:  fmt.Sprintf("depends_on reference %q: contract not found", sv.Text),
					})
				}
			}
		}
	}

	diags = append(diags, checkHappyPathLabels(path, ab)...)
	diags = append(diags, checkPersonaActors(path, ab)...)

	return diags
}

// checkHappyPathLabels warns when a feature has scenarios but none labeled happy_path.
func checkHappyPathLabels(path string, ab *dsl.ArtifactBlock) []Diagnostic {
	var diags []Diagnostic
	for _, item := range ab.Items {
		fb, ok := item.(*dsl.FeatureBlock)
		if !ok {
			continue
		}
		hasScenarios := false
		hasHappyPath := false
		for _, group := range fb.Groups {
			switch g := group.(type) {
			case *dsl.Scenario:
				hasScenarios = true
				for _, l := range g.Labels() {
					if l == "happy_path" {
						hasHappyPath = true
					}
				}
			case *dsl.Group:
				for _, sc := range g.Scenarios {
					hasScenarios = true
					for _, l := range sc.Labels() {
						if l == "happy_path" {
							hasHappyPath = true
						}
					}
				}
			}
		}
		if hasScenarios && !hasHappyPath {
			featureName := fb.Name
			if featureName == "" {
				featureName = "(unnamed)"
			}
			diags = append(diags, Diagnostic{
				File:     path,
				Line:     fb.Line,
				Severity: SeverityInfo,
				Rule:     "scenario-labels",
				Message:  fmt.Sprintf("feature %q has no scenario labeled happy_path", featureName),
			})
		}
	}
	return diags
}

// checkPersonaActors cross-references actor field values in scenarios against
// a declared personas block on the artifact.
func checkPersonaActors(path string, ab *dsl.ArtifactBlock) []Diagnostic {
	var diags []Diagnostic

	declared := extractPersonas(ab)
	actors := collectScenarioActors(ab)

	if len(declared) == 0 && len(actors) > 0 {
		diags = append(diags, Diagnostic{
			File:     path,
			Severity: SeverityInfo,
			Rule:     "persona-actor",
			Message:  "scenarios use actor field but artifact has no personas declaration",
		})
		return diags
	}

	if len(declared) == 0 {
		return diags
	}

	for _, a := range actors {
		if _, ok := declared[a.actor]; !ok {
			diags = append(diags, Diagnostic{
				File:     path,
				Line:     a.line,
				Severity: SeverityWarning,
				Rule:     "persona-actor",
				Message:  fmt.Sprintf("actor %q not declared in personas block", a.actor),
			})
		}
	}

	for _, item := range ab.Items {
		fb, ok := item.(*dsl.FeatureBlock)
		if !ok {
			continue
		}
		if !featureHasActors(fb) && featureHasScenarios(fb) {
			featureName := fb.Name
			if featureName == "" {
				featureName = "(unnamed)"
			}
			diags = append(diags, Diagnostic{
				File:     path,
				Line:     fb.Line,
				Severity: SeverityInfo,
				Rule:     "persona-actor",
				Message:  fmt.Sprintf("feature %q has no actor coverage despite personas being declared", featureName),
			})
		}
	}

	return diags
}

func extractPersonas(ab *dsl.ArtifactBlock) map[string]bool {
	result := make(map[string]bool)
	for _, item := range ab.Items {
		blk, ok := item.(*dsl.Block)
		if !ok || blk.Name != "personas" {
			continue
		}
		for _, fi := range blk.Items {
			field, ok := fi.(*dsl.Field)
			if !ok {
				continue
			}
			result[field.Key] = true
		}
	}
	return result
}

type actorRef struct {
	actor string
	line  int
}

func collectScenarioActors(ab *dsl.ArtifactBlock) []actorRef {
	var refs []actorRef
	for _, item := range ab.Items {
		fb, ok := item.(*dsl.FeatureBlock)
		if !ok {
			continue
		}
		for _, group := range fb.Groups {
			switch g := group.(type) {
			case *dsl.Scenario:
				if a := g.Actor(); a != "" {
					refs = append(refs, actorRef{actor: a, line: g.Line})
				}
			case *dsl.Group:
				for _, sc := range g.Scenarios {
					if a := sc.Actor(); a != "" {
						refs = append(refs, actorRef{actor: a, line: sc.Line})
					}
				}
			}
		}
	}
	return refs
}

func featureHasActors(fb *dsl.FeatureBlock) bool {
	for _, group := range fb.Groups {
		switch g := group.(type) {
		case *dsl.Scenario:
			if g.Actor() != "" {
				return true
			}
		case *dsl.Group:
			for _, sc := range g.Scenarios {
				if sc.Actor() != "" {
					return true
				}
			}
		}
	}
	return false
}

func featureHasScenarios(fb *dsl.FeatureBlock) bool {
	for _, group := range fb.Groups {
		switch group.(type) {
		case *dsl.Scenario:
			return true
		case *dsl.Group:
			return true
		}
	}
	return false
}

// ValidateGenericFile validates a custom artifact instance against its schema.
func ValidateGenericFile(path string, sch schema.ArtifactSchema, ctx *ProjectContext) []Diagnostic {
	var diags []Diagnostic

	data, err := os.ReadFile(path)
	if err != nil {
		return append(diags, Diagnostic{
			File:     path,
			Severity: SeverityError,
			Rule:     "schema/parse",
			Message:  fmt.Sprintf("cannot read file: %v", err),
		})
	}

	f, err := dsl.Parse(string(data), ctx.Keywords)
	if err != nil {
		return append(diags, Diagnostic{
			File:     path,
			Severity: SeverityError,
			Rule:     "schema/parse",
			Message:  fmt.Sprintf("parse error: %v", err),
		})
	}

	ab, ok := f.Artifact.(*dsl.ArtifactBlock)
	if !ok {
		return append(diags, Diagnostic{
			File:     path,
			Severity: SeverityError,
			Rule:     "schema/structure",
			Message:  "file does not contain a valid artifact",
		})
	}

	if ab.Kind != sch.Kind {
		diags = append(diags, Diagnostic{
			File:     path,
			Severity: SeverityError,
			Rule:     "schema/kind",
			Message:  fmt.Sprintf("expected artifact kind %q, got %q", sch.Kind, ab.Kind),
		})
	}

	fieldValues := make(map[string]string)
	for _, item := range ab.Items {
		field, ok := item.(*dsl.Field)
		if !ok {
			continue
		}
		if sv, ok := field.Value.(*dsl.StringVal); ok {
			fieldValues[field.Key] = sv.Text
		} else {
			fieldValues[field.Key] = ""
		}
	}

	schemaFields := make(map[string]bool)
	for _, fd := range sch.Fields {
		schemaFields[fd.Name] = true
		if fd.Required {
			if _, present := fieldValues[fd.Name]; !present {
				diags = append(diags, Diagnostic{
					File:     path,
					Severity: SeverityError,
					Rule:     "schema/required-field",
					Message:  fmt.Sprintf("required field %q is missing", fd.Name),
				})
			}
		}
		if len(fd.Enum) > 0 {
			if val, present := fieldValues[fd.Name]; present {
				valid := false
				for _, e := range fd.Enum {
					if val == e {
						valid = true
						break
					}
				}
				if !valid {
					diags = append(diags, Diagnostic{
						File:     path,
						Severity: SeverityError,
						Rule:     "schema/enum",
						Message:  fmt.Sprintf("field %q has invalid value %q; valid values: %s", fd.Name, val, strings.Join(fd.Enum, ", ")),
					})
				}
			}
		}
	}

	for key := range fieldValues {
		if !schemaFields[key] {
			diags = append(diags, Diagnostic{
				File:     path,
				Severity: SeverityWarning,
				Rule:     "schema/unknown-field",
				Message:  fmt.Sprintf("field %q is not defined in the %s schema", key, sch.Kind),
			})
		}
	}

	diags = append(diags, checkHappyPathLabels(path, ab)...)
	diags = append(diags, checkPersonaActors(path, ab)...)

	return diags
}

// validateSpecEnforcement checks specifications based on their enforcement level.
// disabled=skip, warn=warning diagnostics, enforced=error diagnostics.
func validateSpecEnforcement(mosDir string, ctx *ProjectContext) []Diagnostic {
	var diags []Diagnostic

	for _, sub := range []string{names.ActiveDir, names.ArchiveDir} {
		dir := filepath.Join(mosDir, "specifications", sub)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			specPath := filepath.Join(dir, e.Name(), "specification.mos")
			data, err := os.ReadFile(specPath)
			if err != nil {
				continue
			}
			f, err := dsl.Parse(string(data), ctx.Keywords)
			if err != nil {
				continue
			}
			ab, ok := f.Artifact.(*dsl.ArtifactBlock)
			if !ok {
				continue
			}

			fields := map[string]string{}
			for _, item := range ab.Items {
				if field, ok := item.(*dsl.Field); ok {
					if sv, ok := field.Value.(*dsl.StringVal); ok {
						fields[field.Key] = sv.Text
					}
				}
			}

			enforcement := fields["enforcement"]
			if enforcement == "disabled" || enforcement == "" {
				continue
			}

			severity := SeverityWarning
			if enforcement == "enforced" {
				severity = SeverityError
			}

			hasTestMatrix := specHasTestMatrix(ab.Items)

			if fields["symbol"] == "" && !hasTestMatrix {
				diags = append(diags, Diagnostic{
					File:     specPath,
					Severity: severity,
					Rule:     "spec-traceability",
					Message:  fmt.Sprintf("specification %q has no implementation symbol binding", ab.Name),
				})
			}
			if fields["harness"] == "" && !hasTestMatrix {
				diags = append(diags, Diagnostic{
					File:     specPath,
					Severity: severity,
					Rule:     "spec-traceability",
					Message:  fmt.Sprintf("specification %q has no harness binding", ab.Name),
				})
			}

			if symbol := fields["symbol"]; symbol != "" {
				projectRoot := filepath.Dir(mosDir)
				modulePath, err := ReadGoModulePath(filepath.Join(projectRoot, "go.mod"))
				if err == nil {
					res := ResolveSymbol(projectRoot, modulePath, symbol)
					if !res.PackageExists {
						dot := strings.LastIndex(symbol, ".")
						pkg := symbol
						if dot >= 0 {
							pkg = symbol[:dot]
						}
						diags = append(diags, Diagnostic{
							File:     specPath,
							Severity: severity,
							Rule:     "spec-enforcement",
							Message:  fmt.Sprintf("specification %q references symbol %q but package %q does not exist", ab.Name, symbol, pkg),
						})
					} else if !res.SymbolExists {
						dot := strings.LastIndex(symbol, ".")
						sym := symbol
						if dot >= 0 {
							sym = symbol[dot+1:]
						}
						diags = append(diags, Diagnostic{
							File:     specPath,
							Severity: severity,
							Rule:     "spec-enforcement",
							Message:  fmt.Sprintf("specification %q references symbol %q but %q is not exported by the package", ab.Name, symbol, sym),
						})
					}
				}
			}
		}
	}

	return diags
}

// specHasTestMatrix returns true if the spec's AST items contain a test_matrix
// block with at least one layer that has a symbol binding.
func specHasTestMatrix(items []dsl.Node) bool {
	for _, item := range items {
		blk, ok := item.(*dsl.Block)
		if !ok || blk.Name != "test_matrix" {
			continue
		}
		for _, layer := range blk.Items {
			layerBlk, ok := layer.(*dsl.Block)
			if !ok {
				continue
			}
			for _, fi := range layerBlk.Items {
				field, ok := fi.(*dsl.Field)
				if !ok {
					continue
				}
				if field.Key == "symbol" {
					if sv, ok := field.Value.(*dsl.StringVal); ok && sv.Text != "" {
						return true
					}
				}
			}
		}
	}
	return false
}

// validateArtifactTypeBlocks validates artifact_type definitions in config.mos.
func validateArtifactTypeBlocks(ctx *ProjectContext) []Diagnostic {
	var diags []Diagnostic

	if ctx.Config == nil {
		return diags
	}

	ab, ok := ctx.Config.Artifact.(*dsl.ArtifactBlock)
	if !ok {
		return diags
	}

	for _, item := range ab.Items {
		blk, ok := item.(*dsl.Block)
		if !ok || blk.Name != "artifact_type" {
			continue
		}

		configPath := filepath.Join(ctx.Root, "config.mos")

		if blk.Title == "" {
			diags = append(diags, Diagnostic{
				File:     configPath,
				Line:     blk.Line,
				Severity: SeverityError,
				Rule:     "artifact-type/name",
				Message:  "artifact_type block must have a name",
			})
			continue
		}
	}

	return diags
}
