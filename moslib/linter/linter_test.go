package linter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dpopsuev/mos/moslib/dsl"
)

// --- fixture helpers ---

func makeDir(t *testing.T, base string, parts ...string) string {
	t.Helper()
	p := filepath.Join(append([]string{base}, parts...)...)
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

const minimalConfig = `config {
  mos { version = 1 }
  backend { type = "git" }
}
`

const validFeatureBlock = `
  feature "Sample rule" {
    scenario "baseline" {
      labels = ["happy_path"]
      given {
        a project
      }
      when {
        I run lint
      }
      then {
        it passes
      }
    }
  }
`

func validCstDir(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	mos := filepath.Join(root, ".mos")

	writeFile(t, filepath.Join(mos, "config.mos"), `
config {
  mos {
    version = 1
  }
  backend {
    type = "git"
  }
}
`)

	writeFile(t, filepath.Join(mos, "declaration.mos"), `
declaration {
  name = "test-project"
  created = 2026-01-01T00:00:00Z
  authors = ["alice"]
}
`)

	writeFile(t, filepath.Join(mos, "lexicon", "default.mos"), `
lexicon {
  terms {
    governance = "the process of governing"
  }
  artifact_labels {
    rule = "Rule"
    contract = "Contract"
  }
}
`)

	makeDir(t, mos, "resolution")
	writeFile(t, filepath.Join(mos, "resolution", "layers.mos"), `
layers {
  layer "repository" {
    level = 1
  }
  layer "organization" {
    level = 2
  }
}
`)

	makeDir(t, mos, "rules", "mechanical")
	writeFile(t, filepath.Join(mos, "rules", "mechanical", "no-binaries.mos"), `
rule "R-001" {
  name = "No binaries"
  type = "mechanical"
  scope = "repository"
  enforcement = "error"
`+validFeatureBlock+`
}
`)

	makeDir(t, mos, "contracts", "active", "CON-001")
	writeFile(t, filepath.Join(mos, "contracts", "active", "CON-001", "contract.mos"), `
contract "CON-001" {
  title = "Test contract"
  status = "active"
  slug = "test-contract"

  bill {
    introduced_by = "alice"
    introduced_at = 2026-01-15T00:00:00Z
    intent = "Testing"
  }
`+validFeatureBlock+`
}
`)

	return root
}

// --- Context loader tests ---

func TestLoadContextValid(t *testing.T) {
	root := validCstDir(t)
	ctx, err := LoadContext(filepath.Join(root, ".mos"))
	if err != nil {
		t.Fatalf("LoadContext: %v", err)
	}

	if ctx.Config == nil {
		t.Fatal("expected Config to be loaded")
	}
	ab := ctx.Config.Artifact.(*dsl.ArtifactBlock)
	mosBlock := astFindBlock(ab.Items, "mos")
	if mosBlock == nil {
		t.Fatal("expected mos block")
	}
	if v, ok := astFieldInt(mosBlock.Items, "version"); !ok || v != 1 {
		t.Errorf("version = %d, want 1", v)
	}

	if len(ctx.Lexicon.Terms) == 0 {
		t.Error("expected lexicon terms to be loaded")
	}
	if _, ok := ctx.Lexicon.Terms["governance"]; !ok {
		t.Error("expected 'governance' term in lexicon")
	}

	if len(ctx.Layers) != 2 {
		t.Errorf("layers = %d, want 2", len(ctx.Layers))
	}
	if !ctx.LayerSet["repository"] {
		t.Error("expected 'repository' in LayerSet")
	}

	if _, ok := ctx.RuleIDs["R-001"]; !ok {
		t.Error("expected rule R-001 to be inventoried")
	}
	if _, ok := ctx.ContractIDs["CON-001"]; !ok {
		t.Error("expected contract CON-001 to be inventoried")
	}
}

func TestLoadContextMinimal(t *testing.T) {
	root := t.TempDir()
	mos := filepath.Join(root, ".mos")
	makeDir(t, mos)
	writeFile(t, filepath.Join(mos, "config.mos"), `config {
  mos {
    version = 1
  }
}
`)

	ctx, err := LoadContext(mos)
	if err != nil {
		t.Fatalf("LoadContext: %v", err)
	}
	if ctx.Config == nil {
		t.Fatal("expected Config")
	}
}

func TestVocabularyMerge(t *testing.T) {
	root := t.TempDir()
	mos := filepath.Join(root, ".mos")

	writeFile(t, filepath.Join(mos, "config.mos"), `config {
  mos {
    version = 1
  }
}
`)

	writeFile(t, filepath.Join(mos, "lexicon", "default.mos"), `
lexicon {
  artifact_labels {
    rule = "Rule"
    contract = "Contract"
  }
}
`)

	writeFile(t, filepath.Join(mos, "lexicon", "project.mos"), `
lexicon {
  artifact_labels {
    rule = "Directive"
  }
}
`)

	ctx, err := LoadContext(mos)
	if err != nil {
		t.Fatalf("LoadContext: %v", err)
	}

	if ctx.Lexicon.ArtifactLabels["rule"] != "Directive" {
		t.Errorf("label = %q, want Directive (project override)", ctx.Lexicon.ArtifactLabels["rule"])
	}
	if ctx.Lexicon.ArtifactLabels["contract"] != "Contract" {
		t.Errorf("label = %q, want Contract (default kept)", ctx.Lexicon.ArtifactLabels["contract"])
	}
}

// --- Schema validator tests ---

func TestValidateConfigMissingVersion(t *testing.T) {
	root := t.TempDir()
	mos := filepath.Join(root, ".mos")
	writeFile(t, filepath.Join(mos, "config.mos"), `config {
  mos {
  }
}
`)
	ctx, _ := LoadContext(mos)
	diags := validateConfig(ctx)
	assertHasDiag(t, diags, "config-schema", SeverityError)
}

func TestValidateConfigInvalidBackend(t *testing.T) {
	root := t.TempDir()
	mos := filepath.Join(root, ".mos")
	writeFile(t, filepath.Join(mos, "config.mos"), `config {
  mos {
    version = 1
  }
  backend {
    type = "svn"
  }
}
`)
	ctx, _ := LoadContext(mos)
	diags := validateConfig(ctx)
	assertHasDiag(t, diags, "config-enum", SeverityError)
}

func TestValidateRuleMissingFeature(t *testing.T) {
	root := t.TempDir()
	mos := filepath.Join(root, ".mos")
	writeFile(t, filepath.Join(mos, "config.mos"), `config {
  mos {
    version = 1
  }
}
`)
	ruleDir := makeDir(t, mos, "rules", "mechanical")
	rulePath := filepath.Join(ruleDir, "bad.mos")
	writeFile(t, rulePath, `rule "R-BAD" {
  name = "No feature"
  type = "mechanical"
  scope = "repository"
  enforcement = "error"
}
`)

	ctx, _ := LoadContext(mos)
	diags := validateRule(rulePath, ctx)
	assertHasDiag(t, diags, "rule-schema", SeverityError)
}

func TestValidateRuleInvalidScope(t *testing.T) {
	root := t.TempDir()
	mos := filepath.Join(root, ".mos")
	writeFile(t, filepath.Join(mos, "config.mos"), `config {
  mos {
    version = 1
  }
}
`)
	writeFile(t, filepath.Join(mos, "resolution", "layers.mos"), `layers {
  layer "repository" {
    level = 1
  }
}
`)
	ruleDir := makeDir(t, mos, "rules", "mechanical")
	writeFile(t, filepath.Join(ruleDir, "bad-scope.mos"), `rule "R-SCOPE" {
  name = "Bad scope"
  type = "mechanical"
  scope = "galaxy"
  enforcement = "error"
`+validFeatureBlock+`
}
`)

	l := &Linter{}
	diags, err := l.Lint(root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	assertHasDiag(t, diags, "layer-ref", SeverityError)
}

func TestValidateRuleWithInclude(t *testing.T) {
	root := t.TempDir()
	mos := filepath.Join(root, ".mos")
	writeFile(t, filepath.Join(mos, "config.mos"), `config {
  mos {
    version = 1
  }
}
`)
	ruleDir := makeDir(t, mos, "rules", "mechanical")
	writeFile(t, filepath.Join(ruleDir, "acceptance.mos"), `feature "included" {
  scenario "s" {
    given { something }
    then { it works }
  }
}
`)
	rulePath := filepath.Join(ruleDir, "with-include.mos")
	writeFile(t, rulePath, `rule "R-INC" {
  name = "Include rule"
  type = "mechanical"
  scope = "repository"
  enforcement = "error"
  spec {
    include "acceptance.mos"
  }
}
`)

	ctx, _ := LoadContext(mos)
	diags := validateRule(rulePath, ctx)
	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics for valid include, got %d: %v", len(diags), diags)
	}
}

func TestValidateRuleWithMissingInclude(t *testing.T) {
	root := t.TempDir()
	mos := filepath.Join(root, ".mos")
	writeFile(t, filepath.Join(mos, "config.mos"), `config {
  mos {
    version = 1
  }
}
`)
	ruleDir := makeDir(t, mos, "rules", "mechanical")
	rulePath := filepath.Join(ruleDir, "missing-include.mos")
	writeFile(t, rulePath, `rule "R-MISS" {
  name = "Missing include"
  type = "mechanical"
  scope = "repository"
  enforcement = "error"
  spec {
    include "nonexistent.mos"
  }
}
`)

	ctx, _ := LoadContext(mos)
	diags := validateRule(rulePath, ctx)
	assertHasDiag(t, diags, "include-resolve", SeverityError)
}

func TestValidateContractMissingBill(t *testing.T) {
	root := t.TempDir()
	mos := filepath.Join(root, ".mos")
	writeFile(t, filepath.Join(mos, "config.mos"), `config {
  mos {
    version = 1
  }
}
`)
	contractDir := makeDir(t, mos, "contracts", "active", "CON-BAD")
	contractPath := filepath.Join(contractDir, "contract.mos")
	writeFile(t, contractPath, `contract "CON-BAD" {
  title = "Missing bill"
  status = "draft"
`+validFeatureBlock+`
}
`)

	ctx, _ := LoadContext(mos)
	diags := validateContract(contractPath, ctx)
	assertHasDiag(t, diags, "contract-schema", SeverityInfo)
}

// --- Cross-ref tests ---

func TestCrossRefRuleDoesNotExist(t *testing.T) {
	root := t.TempDir()
	mos := filepath.Join(root, ".mos")
	writeFile(t, filepath.Join(mos, "config.mos"), `config {
  mos {
    version = 1
  }
}
`)
	contractDir := makeDir(t, mos, "contracts", "active", "CON-XREF")
	writeFile(t, filepath.Join(contractDir, "contract.mos"), `contract "CON-XREF" {
  title = "Cross-ref test"
  status = "active"

  bill {
    introduced_by = "alice"
    introduced_at = 2026-01-15T00:00:00Z
    intent = "Testing cross-refs"
  }

  execution {
    rules_override = ["R-NONEXISTENT"]
  }
`+validFeatureBlock+`
}
`)

	l := &Linter{}
	diags, err := l.Lint(root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	assertHasDiag(t, diags, "crossref-rule", SeverityError)
}

func TestCrossRefRuleExists(t *testing.T) {
	root := validCstDir(t)
	mos := filepath.Join(root, ".mos")

	contractDir := makeDir(t, mos, "contracts", "active", "CON-XREF-OK")
	contractPath := filepath.Join(contractDir, "contract.mos")
	writeFile(t, contractPath, `contract "CON-XREF-OK" {
  title = "Cross-ref ok"
  status = "active"

  bill {
    introduced_by = "alice"
    introduced_at = 2026-01-15T00:00:00Z
    intent = "Testing cross-refs"
  }

  execution {
    rules_override = ["R-001"]
  }
`+validFeatureBlock+`
}
`)

	ctx, _ := LoadContext(mos)
	diags := validateContract(contractPath, ctx)
	for _, d := range diags {
		if d.Rule == "crossref-rule" {
			t.Errorf("unexpected crossref error: %s", d.Message)
		}
	}
}

// --- Template conformance tests ---

func TestTemplateConformanceMissingBlock(t *testing.T) {
	root := t.TempDir()
	mos := filepath.Join(root, ".mos")
	writeFile(t, filepath.Join(mos, "config.mos"), `config {
  mos {
    version = 1
  }
}
`)
	writeFile(t, filepath.Join(mos, "templates", "contract.mos"), `contract "template" {
  bill { }
  execution { }
  coverage { }
}
`)
	contractDir := makeDir(t, mos, "contracts", "active", "CON-TPL")
	writeFile(t, filepath.Join(contractDir, "contract.mos"), `contract "CON-TPL" {
  title = "Template test"
  status = "active"

  bill {
    introduced_by = "alice"
    introduced_at = 2026-01-15T00:00:00Z
    intent = "Testing"
  }
`+validFeatureBlock+`
}
`)

	l := &Linter{}
	diags, err := l.Lint(root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	assertHasDiag(t, diags, "template-conformance", SeverityWarning)
}

// --- Full linter integration ---

func TestLintValidCstDir(t *testing.T) {
	root := validCstDir(t)
	l := &Linter{}
	diags, err := l.Lint(root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if len(diags) != 0 {
		for _, d := range diags {
			t.Errorf("unexpected diagnostic: %s: %s [%s] %s", d.File, d.Severity, d.Rule, d.Message)
		}
	}
}

func TestLintNoCstDir(t *testing.T) {
	root := t.TempDir()
	l := &Linter{}
	_, err := l.Lint(root)
	if err == nil {
		t.Fatal("expected error when no .mos/ directory exists")
	}
}

// --- JSON serialization ---

func TestDiagnosticJSON(t *testing.T) {
	diags := []Diagnostic{
		{File: "config.mos", Severity: SeverityError, Rule: "config-schema", Message: "test"},
		{File: "rule.mos", Severity: SeverityWarning, Rule: "vocab-term", Message: "test2"},
	}
	data, err := json.Marshal(diags)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got []map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got[0]["severity"] != "error" {
		t.Errorf("severity = %v, want error", got[0]["severity"])
	}
	if got[1]["severity"] != "warning" {
		t.Errorf("severity = %v, want warning", got[1]["severity"])
	}
	if got[0]["rule"] != "config-schema" {
		t.Errorf("rule = %v, want config-schema", got[0]["rule"])
	}
}

// --- CON-2026-038: Criterion linkage linter tests ---

func addArtifactTypesToConfig(t *testing.T, mos string) {
	t.Helper()
	configPath := filepath.Join(mos, "config.mos")
	data, _ := os.ReadFile(configPath)
	content := string(data)
	extra := `
  artifact_type "need" {
    directory = "needs"
    fields {
      title { required = true }
      sensation { required = true }
      status { required = true }
      acceptance {}
    }
    lifecycle {
      active_states = ["identified", "validated", "addressed"]
      archive_states = ["retired"]
      expects_downstream {
        via = "satisfies"
        after = "validated"
        severity = "warn"
      }
      urgency_propagation {
        critical = "error"
        high = "warn"
        medium = "info"
        low = "ignore"
      }
    }
  }
  artifact_type "specification" {
    directory = "specifications"
    fields {
      title { required = true }
      enforcement { required = true }
      satisfies {}
      addresses {}
      non_goals {}
    }
    lifecycle {
      active_states = ["active"]
      archive_states = ["retired"]
      expects_downstream {
        via = "implements"
        after = "active"
        severity = "warn"
      }
    }
  }
  artifact_type "architecture" {
    directory = "architectures"
    fields {
      title { required = true }
      implements {}
    }
    lifecycle {
      active_states = ["draft", "active"]
      archive_states = ["superseded"]
      expects_downstream {
        via = "implements"
        after = "active"
        severity = "warn"
      }
    }
  }
  artifact_type "binder" {
    directory = "binders"
    fields {
      title { required = true }
    }
  }
  artifact_type "doc" {
    directory = "docs"
    fields {
      title { required = true }
    }
  }
`
	content = strings.TrimSuffix(strings.TrimSpace(content), "}")
	content += "\n" + extra + "\n}\n"
	writeFile(t, configPath, content)
}

func TestCON038_CriterionExistsDanglingAddress(t *testing.T) {
	root := validCstDir(t)
	mos := filepath.Join(root, ".mos")
	addArtifactTypesToConfig(t, mos)

	needDir := makeDir(t, mos, "needs", "active", "NEED-001")
	writeFile(t, filepath.Join(needDir, "need.mos"), `need "NEED-001" {
  title = "Fast CI"
  sensation = "CI is slow"
  status = "validated"
  acceptance {
    criterion "sub-10-min" {
      description = "P95 under 10min"
      verified_by = "harness"
    }
  }
}`)

	specDir := makeDir(t, mos, "specifications", "active", "SPEC-001")
	writeFile(t, filepath.Join(specDir, "specification.mos"), `specification "SPEC-001" {
  title = "Pipeline optimization"
  enforcement = "warn"
  satisfies = "NEED-001"
  addresses = ["nonexistent"]
  status = "active"
}`)

	l := &Linter{}
	diags, err := l.Lint(root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	assertHasDiag(t, diags, "criterion-exists", SeverityError)
}

func TestCON038_CriterionExistsValidAddress(t *testing.T) {
	root := validCstDir(t)
	mos := filepath.Join(root, ".mos")
	addArtifactTypesToConfig(t, mos)

	needDir := makeDir(t, mos, "needs", "active", "NEED-001")
	writeFile(t, filepath.Join(needDir, "need.mos"), `need "NEED-001" {
  title = "Fast CI"
  sensation = "CI is slow"
  status = "validated"
  acceptance {
    criterion "sub-10-min" {
      description = "P95 under 10min"
      verified_by = "harness"
    }
  }
}`)

	specDir := makeDir(t, mos, "specifications", "active", "SPEC-001")
	writeFile(t, filepath.Join(specDir, "specification.mos"), `specification "SPEC-001" {
  title = "Pipeline optimization"
  enforcement = "warn"
  satisfies = "NEED-001"
  addresses = ["sub-10-min"]
  status = "active"
}`)

	l := &Linter{}
	diags, err := l.Lint(root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	for _, d := range diags {
		if d.Rule == "criterion-exists" {
			t.Errorf("unexpected criterion-exists error: %s", d.Message)
		}
	}
}

func TestCON038_CriterionCoverageUncovered(t *testing.T) {
	root := validCstDir(t)
	mos := filepath.Join(root, ".mos")
	addArtifactTypesToConfig(t, mos)

	needDir := makeDir(t, mos, "needs", "active", "NEED-001")
	writeFile(t, filepath.Join(needDir, "need.mos"), `need "NEED-001" {
  title = "Fast CI"
  sensation = "CI is slow"
  status = "validated"
  acceptance {
    criterion "sub-10-min" {
      description = "P95 under 10min"
      verified_by = "harness"
    }
    criterion "no-regression" {
      description = "Coverage stable"
      verified_by = "harness"
    }
  }
}`)

	specDir := makeDir(t, mos, "specifications", "active", "SPEC-001")
	writeFile(t, filepath.Join(specDir, "specification.mos"), `specification "SPEC-001" {
  title = "Pipeline optimization"
  enforcement = "warn"
  satisfies = "NEED-001"
  addresses = ["sub-10-min"]
  status = "active"
}`)

	l := &Linter{}
	diags, err := l.Lint(root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	assertHasDiag(t, diags, "criterion-coverage", SeverityWarning)
}

func TestCON038_CriterionCoverageFullyCovered(t *testing.T) {
	root := validCstDir(t)
	mos := filepath.Join(root, ".mos")
	addArtifactTypesToConfig(t, mos)

	needDir := makeDir(t, mos, "needs", "active", "NEED-001")
	writeFile(t, filepath.Join(needDir, "need.mos"), `need "NEED-001" {
  title = "Fast CI"
  sensation = "CI is slow"
  status = "validated"
  acceptance {
    criterion "sub-10-min" {
      description = "P95 under 10min"
      verified_by = "harness"
    }
  }
}`)

	specDir := makeDir(t, mos, "specifications", "active", "SPEC-001")
	writeFile(t, filepath.Join(specDir, "specification.mos"), `specification "SPEC-001" {
  title = "Pipeline optimization"
  enforcement = "warn"
  satisfies = "NEED-001"
  addresses = ["sub-10-min"]
  status = "active"
}`)

	l := &Linter{}
	diags, err := l.Lint(root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	for _, d := range diags {
		if d.Rule == "criterion-coverage" {
			t.Errorf("unexpected criterion-coverage diagnostic: %s", d.Message)
		}
	}
}

// --- CON-2026-040: Scope violation detection ---

func TestCON040_SpecTitleOverlappingExcludesFlagged(t *testing.T) {
	root := validCstDir(t)
	mos := filepath.Join(root, ".mos")
	addArtifactTypesToConfig(t, mos)

	needDir := makeDir(t, mos, "needs", "active", "NEED-001")
	writeFile(t, filepath.Join(needDir, "need.mos"), `need "NEED-001" {
  title = "Faster CI"
  sensation = "CI is slow"
  status = "validated"
  scope {
    excludes = ["migration", "rewriting"]
  }
}`)

	specDir := makeDir(t, mos, "specifications", "active", "SPEC-001")
	writeFile(t, filepath.Join(specDir, "specification.mos"), `specification "SPEC-001" {
  title = "Database migration strategy"
  enforcement = "warn"
  satisfies = "NEED-001"
  status = "active"
}`)

	l := &Linter{}
	diags, err := l.Lint(root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	assertHasDiag(t, diags, "scope-violation", SeverityError)
}

func TestCON040_SpecTitleCleanOfExcludes(t *testing.T) {
	root := validCstDir(t)
	mos := filepath.Join(root, ".mos")
	addArtifactTypesToConfig(t, mos)

	needDir := makeDir(t, mos, "needs", "active", "NEED-001")
	writeFile(t, filepath.Join(needDir, "need.mos"), `need "NEED-001" {
  title = "Faster CI"
  sensation = "CI is slow"
  status = "validated"
  scope {
    excludes = ["migration", "rewriting"]
  }
}`)

	specDir := makeDir(t, mos, "specifications", "active", "SPEC-001")
	writeFile(t, filepath.Join(specDir, "specification.mos"), `specification "SPEC-001" {
  title = "Pipeline parallelization"
  enforcement = "warn"
  satisfies = "NEED-001"
  status = "active"
}`)

	l := &Linter{}
	diags, err := l.Lint(root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	for _, d := range diags {
		if d.Rule == "scope-violation" && strings.Contains(d.Message, "SPEC-001") {
			t.Errorf("unexpected scope-violation: %s", d.Message)
		}
	}
}

func TestCON040_ArchForbiddenEdgeViolated(t *testing.T) {
	root := validCstDir(t)
	mos := filepath.Join(root, ".mos")
	addArtifactTypesToConfig(t, mos)

	archDir := makeDir(t, mos, "architectures", "active", "ARCH-001")
	writeFile(t, filepath.Join(archDir, "architecture.mos"), `architecture "ARCH-001" {
  title = "Service architecture"
  resolution = "service"
  status = "active"

  service "CLI" {}
  service "Database" {}

  edge "direct-access" {
    from = "CLI"
    to = "Database"
  }

  forbidden "no-direct-db" {
    from = "CLI"
    to = "Database"
    reason = "must go through API layer"
  }
}`)

	l := &Linter{}
	diags, err := l.Lint(root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	assertHasDiag(t, diags, "scope-violation", SeverityError)
}

func TestCON040_ArchNoForbiddenViolation(t *testing.T) {
	root := validCstDir(t)
	mos := filepath.Join(root, ".mos")
	addArtifactTypesToConfig(t, mos)

	archDir := makeDir(t, mos, "architectures", "active", "ARCH-001")
	writeFile(t, filepath.Join(archDir, "architecture.mos"), `architecture "ARCH-001" {
  title = "Service architecture"
  resolution = "service"
  status = "active"

  service "CLI" {}
  service "API" {}
  service "Database" {}

  edge "cli-to-api" {
    from = "CLI"
    to = "API"
  }

  forbidden "no-direct-db" {
    from = "CLI"
    to = "Database"
    reason = "must go through API layer"
  }
}`)

	l := &Linter{}
	diags, err := l.Lint(root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	for _, d := range diags {
		if d.Rule == "scope-violation" && strings.Contains(d.Message, "ARCH-001") {
			t.Errorf("unexpected scope-violation: %s", d.Message)
		}
	}
}

// --- CON-2026-042: Urgency Propagation ---

func TestCON042_CriticalNeedPromotesOrphanToError(t *testing.T) {
	root := validCstDir(t)
	mos := filepath.Join(root, ".mos")
	addArtifactTypesToConfig(t, mos)

	needDir := makeDir(t, mos, "needs", "active", "NEED-CRIT")
	writeFile(t, filepath.Join(needDir, "need.mos"), `need "NEED-CRIT" {
  title = "Critical orphan need"
  sensation = "Emergency"
  status = "validated"
  urgency = "critical"
}`)

	l := &Linter{}
	diags, err := l.Lint(root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	found := false
	for _, d := range diags {
		if d.Rule == "lifecycle-orphan" && strings.Contains(d.Message, "NEED-CRIT") {
			found = true
			if d.Severity != SeverityError {
				t.Errorf("critical urgency should promote orphan to error, got %s", d.Severity)
			}
		}
	}
	if !found {
		t.Error("expected lifecycle-orphan diagnostic for NEED-CRIT")
	}
}

func TestCON042_LowUrgencyDoesNotPromote(t *testing.T) {
	root := validCstDir(t)
	mos := filepath.Join(root, ".mos")
	addArtifactTypesToConfig(t, mos)

	needDir := makeDir(t, mos, "needs", "active", "NEED-LOW")
	writeFile(t, filepath.Join(needDir, "need.mos"), `need "NEED-LOW" {
  title = "Low urgency orphan"
  sensation = "Minor discomfort"
  status = "validated"
  urgency = "low"
}`)

	l := &Linter{}
	diags, err := l.Lint(root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	for _, d := range diags {
		if d.Rule == "lifecycle-orphan" && strings.Contains(d.Message, "NEED-LOW") {
			if d.Severity != SeverityWarning {
				t.Errorf("low urgency should keep orphan as warning, got %s", d.Severity)
			}
		}
	}
}

func TestCON042_CriticalUncoveredCriterionPromotedToError(t *testing.T) {
	root := validCstDir(t)
	mos := filepath.Join(root, ".mos")
	addArtifactTypesToConfig(t, mos)

	needDir := makeDir(t, mos, "needs", "active", "NEED-CRIT2")
	writeFile(t, filepath.Join(needDir, "need.mos"), `need "NEED-CRIT2" {
  title = "Critical need with criteria"
  sensation = "Emergency"
  status = "validated"
  urgency = "critical"
  acceptance {
    criterion "fast" {
      description = "Must be fast"
      verified_by = "harness"
    }
  }
}`)

	l := &Linter{}
	diags, err := l.Lint(root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	found := false
	for _, d := range diags {
		if d.Rule == "criterion-coverage" && strings.Contains(d.Message, "NEED-CRIT2") {
			found = true
			if d.Severity != SeverityError {
				t.Errorf("critical urgency should promote criterion-coverage to error, got %s", d.Severity)
			}
		}
	}
	if !found {
		t.Error("expected criterion-coverage diagnostic for NEED-CRIT2")
	}
}

func TestCON042_NoUrgencyKeepsDefaultSeverity(t *testing.T) {
	root := validCstDir(t)
	mos := filepath.Join(root, ".mos")
	addArtifactTypesToConfig(t, mos)

	needDir := makeDir(t, mos, "needs", "active", "NEED-NOURG")
	writeFile(t, filepath.Join(needDir, "need.mos"), `need "NEED-NOURG" {
  title = "Need without urgency"
  sensation = "Some pain"
  status = "validated"
  acceptance {
    criterion "fast" {
      description = "Must be fast"
      verified_by = "harness"
    }
  }
}`)

	l := &Linter{}
	diags, err := l.Lint(root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	for _, d := range diags {
		if d.Rule == "criterion-coverage" && strings.Contains(d.Message, "NEED-NOURG") {
			if d.Severity != SeverityWarning {
				t.Errorf("no urgency should keep warning, got %s", d.Severity)
			}
		}
	}
}

// --- CON-2026-039: Lifecycle orphan detection ---

func TestCON039_OrphanNeedFlagged(t *testing.T) {
	root := validCstDir(t)
	mos := filepath.Join(root, ".mos")
	addArtifactTypesToConfig(t, mos)

	needDir := makeDir(t, mos, "needs", "active", "NEED-ORPHAN")
	writeFile(t, filepath.Join(needDir, "need.mos"), `need "NEED-ORPHAN" {
  title = "Orphan need"
  sensation = "Pain"
  status = "validated"
}`)

	l := &Linter{}
	diags, err := l.Lint(root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	assertHasDiag(t, diags, "lifecycle-orphan", SeverityWarning)
}

func TestCON039_NonOrphanNeedClean(t *testing.T) {
	root := validCstDir(t)
	mos := filepath.Join(root, ".mos")
	addArtifactTypesToConfig(t, mos)

	needDir := makeDir(t, mos, "needs", "active", "NEED-001")
	writeFile(t, filepath.Join(needDir, "need.mos"), `need "NEED-001" {
  title = "Referenced need"
  sensation = "Pain"
  status = "validated"
}`)

	specDir := makeDir(t, mos, "specifications", "active", "SPEC-001")
	writeFile(t, filepath.Join(specDir, "specification.mos"), `specification "SPEC-001" {
  title = "Spec for need"
  enforcement = "warn"
  satisfies = "NEED-001"
  status = "active"
}`)

	l := &Linter{}
	diags, err := l.Lint(root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	for _, d := range diags {
		if d.Rule == "lifecycle-orphan" && strings.Contains(d.Message, "NEED-001") {
			t.Errorf("unexpected orphan diagnostic for NEED-001: %s", d.Message)
		}
	}
}

func TestCON039_NeedBelowThresholdNotFlagged(t *testing.T) {
	root := validCstDir(t)
	mos := filepath.Join(root, ".mos")
	addArtifactTypesToConfig(t, mos)

	needDir := makeDir(t, mos, "needs", "active", "NEED-EARLY")
	writeFile(t, filepath.Join(needDir, "need.mos"), `need "NEED-EARLY" {
  title = "Early need"
  sensation = "Pain"
  status = "identified"
}`)

	l := &Linter{}
	diags, err := l.Lint(root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	for _, d := range diags {
		if d.Rule == "lifecycle-orphan" && strings.Contains(d.Message, "NEED-EARLY") {
			t.Errorf("need at identified should not be flagged as orphan: %s", d.Message)
		}
	}
}

func TestIDCollisionDetection(t *testing.T) {
	root := t.TempDir()
	mos := filepath.Join(root, ".mos")
	writeFile(t, filepath.Join(mos, "config.mos"), `config {
  mos {
    version = 1
  }
}
`)
	makeDir(t, mos, "contracts", "active", "CON-DUP-001")
	writeFile(t, filepath.Join(mos, "contracts", "active", "CON-DUP-001", "contract.mos"),
		`contract "CON-DUP-001" { title = "dup" status = "active" }`)

	makeDir(t, mos, "contracts", "archive", "CON-DUP-001")
	writeFile(t, filepath.Join(mos, "contracts", "archive", "CON-DUP-001", "contract.mos"),
		`contract "CON-DUP-001" { title = "dup" status = "complete" }`)

	l := &Linter{}
	diags, err := l.Lint(root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	assertHasDiag(t, diags, "id-collision", SeverityError)
}

func TestIDCollisionNoFalsePositive(t *testing.T) {
	root := t.TempDir()
	mos := filepath.Join(root, ".mos")
	writeFile(t, filepath.Join(mos, "config.mos"), `config {
  mos {
    version = 1
  }
}
`)
	makeDir(t, mos, "contracts", "active", "CON-OK-001")
	writeFile(t, filepath.Join(mos, "contracts", "active", "CON-OK-001", "contract.mos"),
		`contract "CON-OK-001" { title = "ok" status = "active" }`)

	makeDir(t, mos, "contracts", "archive", "CON-OK-002")
	writeFile(t, filepath.Join(mos, "contracts", "archive", "CON-OK-002", "contract.mos"),
		`contract "CON-OK-002" { title = "ok2" status = "complete" }`)

	l := &Linter{}
	diags, err := l.Lint(root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	for _, d := range diags {
		if d.Rule == "id-collision" {
			t.Errorf("unexpected id-collision diagnostic: %s", d.Message)
		}
	}
}

func TestIDFormatValidation(t *testing.T) {
	root := t.TempDir()
	mosDir := makeDir(t, root, ".mos")
	makeDir(t, mosDir, "contracts", "active", "CON-2026-001")
	writeFile(t, filepath.Join(mosDir, "contracts", "active", "CON-2026-001", "contract.mos"),
		`contract "CON-2026-001" { title = "Valid ID" status = "draft" goal = "test" }`)
	makeDir(t, mosDir, "contracts", "active", "bad_id")
	writeFile(t, filepath.Join(mosDir, "contracts", "active", "bad_id", "contract.mos"),
		`contract "bad_id" { title = "Bad ID" status = "draft" goal = "test" }`)

	configContent := `config { project "mos" { prefix = "CON" sequence = 5 } }`
	writeFile(t, filepath.Join(mosDir, "config.mos"), configContent)
	configFile, _ := dsl.Parse(configContent, nil)

	ctx := &ProjectContext{
		Root:        mosDir,
		Config:      configFile,
		ContractIDs: map[string]string{},
	}
	diags := validateIDFormat(ctx)

	foundBad := false
	for _, d := range diags {
		if d.Rule == "id-format" && strings.Contains(d.Message, "bad_id") {
			foundBad = true
		}
		if d.Rule == "id-format" && strings.Contains(d.Message, "CON-2026-001") {
			t.Error("valid ID should not trigger id-format warning")
		}
	}
	if !foundBad {
		t.Error("expected id-format warning for 'bad_id'")
	}
}

func TestSlugFormatIsError(t *testing.T) {
	root := t.TempDir()
	mosDir := makeDir(t, root, ".mos")
	makeDir(t, mosDir, "contracts", "active", "CON-2026-001")
	writeFile(t, filepath.Join(mosDir, "contracts", "active", "CON-2026-001", "contract.mos"),
		`contract "CON-2026-001" { title = "Test" status = "draft" goal = "test" slug = "BAD SLUG" }`)

	ctx := &ProjectContext{
		Root:        mosDir,
		ContractIDs: map[string]string{},
	}
	diags := validateSlugs(ctx)

	assertHasDiag(t, diags, "slug-format", SeverityError)
}

// --- helpers ---

func assertHasDiag(t *testing.T, diags []Diagnostic, rule string, sev Severity) {
	t.Helper()
	for _, d := range diags {
		if d.Rule == rule && d.Severity == sev {
			return
		}
	}
	t.Errorf("expected diagnostic with rule=%q severity=%s, got: %v", rule, sev, diags)
}

// --- Scenario labels lint checks ---

func TestScenarioLabelsHappyPathPresent(t *testing.T) {
	root := t.TempDir()
	mos := filepath.Join(root, ".mos")
	writeFile(t, filepath.Join(mos, "config.mos"), minimalConfig)

	makeDir(t, mos, "contracts", "active", "CON-HP")
	writeFile(t, filepath.Join(mos, "contracts", "active", "CON-HP", "contract.mos"), `
contract "CON-HP" {
  title = "Has happy path"
  status = "draft"

  feature "Login" {
    scenario "success" {
      labels = ["happy_path"]
      given {
        a user
      }
      when {
        they login
      }
      then {
        they see dashboard
      }
    }
  }
}
`)
	ctx, _ := LoadContext(mos)
	contractPath := filepath.Join(mos, "contracts", "active", "CON-HP", "contract.mos")
	diags := validateContract(contractPath, ctx)
	for _, d := range diags {
		if d.Rule == "scenario-labels" {
			t.Errorf("unexpected scenario-labels diagnostic: %s", d.Message)
		}
	}
}

func TestScenarioLabelsMissingHappyPath(t *testing.T) {
	root := t.TempDir()
	mos := filepath.Join(root, ".mos")
	writeFile(t, filepath.Join(mos, "config.mos"), minimalConfig)

	makeDir(t, mos, "contracts", "active", "CON-NHP")
	writeFile(t, filepath.Join(mos, "contracts", "active", "CON-NHP", "contract.mos"), `
contract "CON-NHP" {
  title = "No happy path"
  status = "draft"

  feature "Login" {
    scenario "edge case" {
      given {
        a user
      }
      when {
        they login
      }
      then {
        error shown
      }
    }
  }
}
`)
	ctx, _ := LoadContext(mos)
	contractPath := filepath.Join(mos, "contracts", "active", "CON-NHP", "contract.mos")
	diags := validateContract(contractPath, ctx)
	assertHasDiag(t, diags, "scenario-labels", SeverityInfo)
}

// --- Test matrix spec enforcement ---

func TestSpecEnforcementWithTestMatrix(t *testing.T) {
	root := t.TempDir()
	mos := filepath.Join(root, ".mos")
	writeFile(t, filepath.Join(mos, "config.mos"), minimalConfig)

	makeDir(t, mos, "specifications", "active", "SPEC-TM")
	writeFile(t, filepath.Join(mos, "specifications", "active", "SPEC-TM", "specification.mos"), `
specification "SPEC-TM" {
  title = "With test matrix"
  status = "draft"
  enforcement = "warn"

  test_matrix {
    unit {
      symbol = "TestFoo"
    }
  }
}
`)
	ctx, _ := LoadContext(mos)
	diags := validateSpecEnforcement(mos, ctx)
	for _, d := range diags {
		if d.Rule == "spec-traceability" {
			t.Errorf("unexpected spec-traceability diagnostic when test_matrix present: %s", d.Message)
		}
	}
}

func TestSpecEnforcementWithoutTestMatrixOrSymbol(t *testing.T) {
	root := t.TempDir()
	mos := filepath.Join(root, ".mos")
	writeFile(t, filepath.Join(mos, "config.mos"), minimalConfig)

	makeDir(t, mos, "specifications", "active", "SPEC-BARE")
	writeFile(t, filepath.Join(mos, "specifications", "active", "SPEC-BARE", "specification.mos"), `
specification "SPEC-BARE" {
  title = "No symbol or test matrix"
  status = "draft"
  enforcement = "warn"
}
`)
	ctx, _ := LoadContext(mos)
	diags := validateSpecEnforcement(mos, ctx)
	assertHasDiag(t, diags, "spec-traceability", SeverityWarning)
}

// --- Persona-actor lint checks ---

func TestPersonaActorMatchesDeclared(t *testing.T) {
	root := t.TempDir()
	mos := filepath.Join(root, ".mos")
	writeFile(t, filepath.Join(mos, "config.mos"), minimalConfig)

	makeDir(t, mos, "contracts", "active", "CON-PA1")
	writeFile(t, filepath.Join(mos, "contracts", "active", "CON-PA1", "contract.mos"), `
contract "CON-PA1" {
  title = "Persona match"
  status = "draft"

  personas {
    contributor = "regular user"
    admin = "elevated user"
  }

  feature "Auth" {
    scenario "contributor drafts" {
      actor = "contributor"
      labels = ["happy_path"]
      given {
        a contract
      }
      when {
        contributor creates
      }
      then {
        draft created
      }
    }
  }
}
`)
	ctx, _ := LoadContext(mos)
	contractPath := filepath.Join(mos, "contracts", "active", "CON-PA1", "contract.mos")
	diags := validateContract(contractPath, ctx)
	for _, d := range diags {
		if d.Rule == "persona-actor" && d.Severity == SeverityWarning {
			t.Errorf("unexpected persona-actor warning: %s", d.Message)
		}
	}
}

func TestPersonaActorUndeclared(t *testing.T) {
	root := t.TempDir()
	mos := filepath.Join(root, ".mos")
	writeFile(t, filepath.Join(mos, "config.mos"), minimalConfig)

	makeDir(t, mos, "contracts", "active", "CON-PA2")
	writeFile(t, filepath.Join(mos, "contracts", "active", "CON-PA2", "contract.mos"), `
contract "CON-PA2" {
  title = "Undeclared actor"
  status = "draft"

  personas {
    contributor = "regular user"
  }

  feature "Auth" {
    scenario "attacker breaks in" {
      actor = "attacker"
      labels = ["adversarial"]
      given {
        a system
      }
      when {
        attacker probes
      }
      then {
        blocked
      }
    }
  }
}
`)
	ctx, _ := LoadContext(mos)
	contractPath := filepath.Join(mos, "contracts", "active", "CON-PA2", "contract.mos")
	diags := validateContract(contractPath, ctx)
	assertHasDiag(t, diags, "persona-actor", SeverityWarning)
}

func TestPersonaActorNoPersonasBlock(t *testing.T) {
	root := t.TempDir()
	mos := filepath.Join(root, ".mos")
	writeFile(t, filepath.Join(mos, "config.mos"), minimalConfig)

	makeDir(t, mos, "contracts", "active", "CON-PA3")
	writeFile(t, filepath.Join(mos, "contracts", "active", "CON-PA3", "contract.mos"), `
contract "CON-PA3" {
  title = "No personas block"
  status = "draft"

  feature "Auth" {
    scenario "admin does something" {
      actor = "admin"
      labels = ["happy_path"]
      given {
        a system
      }
      when {
        admin acts
      }
      then {
        result
      }
    }
  }
}
`)
	ctx, _ := LoadContext(mos)
	contractPath := filepath.Join(mos, "contracts", "active", "CON-PA3", "contract.mos")
	diags := validateContract(contractPath, ctx)
	assertHasDiag(t, diags, "persona-actor", SeverityInfo)
}

func TestPersonaDeclaredButNoActorsUsed(t *testing.T) {
	root := t.TempDir()
	mos := filepath.Join(root, ".mos")
	writeFile(t, filepath.Join(mos, "config.mos"), minimalConfig)

	makeDir(t, mos, "contracts", "active", "CON-PA4")
	writeFile(t, filepath.Join(mos, "contracts", "active", "CON-PA4", "contract.mos"), `
contract "CON-PA4" {
  title = "Personas but no actors"
  status = "draft"

  personas {
    contributor = "regular user"
    admin = "elevated user"
  }

  feature "Auth" {
    scenario "something happens" {
      labels = ["happy_path"]
      given {
        a system
      }
      when {
        something
      }
      then {
        result
      }
    }
  }
}
`)
	ctx, _ := LoadContext(mos)
	contractPath := filepath.Join(mos, "contracts", "active", "CON-PA4", "contract.mos")
	diags := validateContract(contractPath, ctx)
	assertHasDiag(t, diags, "persona-actor", SeverityInfo)
}

// --- Spec enforcement (CON-2026-207) ---

func TestResolveSymbolPackageMissing(t *testing.T) {
	root := t.TempDir()
	res := ResolveSymbol(root, "example.com/mymod", "example.com/mymod/nosuchpkg.Foo")
	if res.PackageExists {
		t.Error("expected PackageExists=false for missing package")
	}
}

func TestResolveSymbolSymbolMissing(t *testing.T) {
	root := t.TempDir()
	pkgDir := filepath.Join(root, "mypkg")
	makeDir(t, pkgDir)
	writeFile(t, filepath.Join(pkgDir, "api.go"), `package mypkg

func ExistingFunc() {}
`)
	res := ResolveSymbol(root, "example.com/mymod", "mypkg.MissingFunc")
	if !res.PackageExists {
		t.Error("expected PackageExists=true")
	}
	if res.SymbolExists {
		t.Error("expected SymbolExists=false for missing symbol")
	}
}

func TestResolveSymbolFound(t *testing.T) {
	root := t.TempDir()
	pkgDir := filepath.Join(root, "mypkg")
	makeDir(t, pkgDir)
	writeFile(t, filepath.Join(pkgDir, "api.go"), `package mypkg

func Handler() {}

type Service struct{}

var DefaultTimeout = 30
`)
	for _, sym := range []string{"mypkg.Handler", "mypkg.Service", "mypkg.DefaultTimeout"} {
		res := ResolveSymbol(root, "example.com/mymod", sym)
		if !res.PackageExists || !res.SymbolExists {
			t.Errorf("expected symbol %q to resolve, got pkg=%v sym=%v", sym, res.PackageExists, res.SymbolExists)
		}
	}
}

func TestResolveSymbolWithModulePrefix(t *testing.T) {
	root := t.TempDir()
	pkgDir := filepath.Join(root, "internal", "api")
	makeDir(t, pkgDir)
	writeFile(t, filepath.Join(pkgDir, "handler.go"), `package api

func ServeHTTP() {}
`)
	res := ResolveSymbol(root, "example.com/mymod", "example.com/mymod/internal/api.ServeHTTP")
	if !res.PackageExists || !res.SymbolExists {
		t.Errorf("expected symbol to resolve with full module path, got pkg=%v sym=%v", res.PackageExists, res.SymbolExists)
	}
}

func TestSpecEnforcementSymbolPackageMissing(t *testing.T) {
	root := t.TempDir()
	mos := filepath.Join(root, ".mos")
	writeFile(t, filepath.Join(mos, "config.mos"), minimalConfig)
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/testmod\n\ngo 1.22\n")

	makeDir(t, mos, "specifications", "active", "SPEC-PKG")
	writeFile(t, filepath.Join(mos, "specifications", "active", "SPEC-PKG", "specification.mos"), `
specification "SPEC-PKG" {
  title = "Missing package"
  status = "draft"
  enforcement = "enforced"
  symbol = "nosuchpkg.Handler"
  harness = "go test ./..."
}
`)
	ctx, _ := LoadContext(mos)
	diags := validateSpecEnforcement(mos, ctx)
	found := false
	for _, d := range diags {
		if d.Rule == "spec-enforcement" && strings.Contains(d.Message, "does not exist") {
			found = true
			if d.Severity != SeverityError {
				t.Errorf("expected error severity for enforced spec, got %s", d.Severity)
			}
		}
	}
	if !found {
		t.Errorf("expected spec-enforcement diagnostic for missing package, got: %v", diags)
	}
}

func TestSpecEnforcementSymbolMissing(t *testing.T) {
	root := t.TempDir()
	mos := filepath.Join(root, ".mos")
	writeFile(t, filepath.Join(mos, "config.mos"), minimalConfig)
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/testmod\n\ngo 1.22\n")

	pkgDir := filepath.Join(root, "api")
	makeDir(t, pkgDir)
	writeFile(t, filepath.Join(pkgDir, "server.go"), `package api

func OtherFunc() {}
`)
	makeDir(t, mos, "specifications", "active", "SPEC-SYM")
	writeFile(t, filepath.Join(mos, "specifications", "active", "SPEC-SYM", "specification.mos"), `
specification "SPEC-SYM" {
  title = "Missing symbol"
  status = "draft"
  enforcement = "warn"
  symbol = "api.Handler"
  harness = "go test ./..."
}
`)
	ctx, _ := LoadContext(mos)
	diags := validateSpecEnforcement(mos, ctx)
	found := false
	for _, d := range diags {
		if d.Rule == "spec-enforcement" && strings.Contains(d.Message, "not exported") {
			found = true
			if d.Severity != SeverityWarning {
				t.Errorf("expected warning severity for warn spec, got %s", d.Severity)
			}
		}
	}
	if !found {
		t.Errorf("expected spec-enforcement diagnostic for missing symbol, got: %v", diags)
	}
}

func TestSpecEnforcementSymbolFound(t *testing.T) {
	root := t.TempDir()
	mos := filepath.Join(root, ".mos")
	writeFile(t, filepath.Join(mos, "config.mos"), minimalConfig)
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/testmod\n\ngo 1.22\n")

	pkgDir := filepath.Join(root, "api")
	makeDir(t, pkgDir)
	writeFile(t, filepath.Join(pkgDir, "server.go"), `package api

func Handler() {}
`)
	makeDir(t, mos, "specifications", "active", "SPEC-OK")
	writeFile(t, filepath.Join(mos, "specifications", "active", "SPEC-OK", "specification.mos"), `
specification "SPEC-OK" {
  title = "Valid symbol"
  status = "draft"
  enforcement = "enforced"
  symbol = "api.Handler"
  harness = "go test ./..."
}
`)
	ctx, _ := LoadContext(mos)
	diags := validateSpecEnforcement(mos, ctx)
	for _, d := range diags {
		if d.Rule == "spec-enforcement" {
			t.Errorf("unexpected spec-enforcement diagnostic when symbol exists: %s", d.Message)
		}
	}
}

func directiveConfig() string {
	return `config {
  mos { version = 1 }
  backend { type = "git" }

  artifact_type "directive" {
    directory = "directives"
    fields {
      title {
        required = true
      }
      status {
        required = true
        enum = ["declared" "active" "achieved" "superseded"]
      }
      text {}
    }
    lifecycle {
      active_states  = ["declared" "active"]
      archive_states = ["achieved" "superseded"]
    }
  }
}
`
}

func TestDirectiveAlignmentFires(t *testing.T) {
	root := t.TempDir()
	mos := filepath.Join(root, ".mos")

	writeFile(t, filepath.Join(mos, "config.mos"), directiveConfig())

	makeDir(t, mos, "directives", "active", "DIR-2026-001")
	writeFile(t, filepath.Join(mos, "directives", "active", "DIR-2026-001", "directive.mos"), `directive "DIR-2026-001" {
  title  = "Ship PoC"
  status = "active"
  text   = "Focus on PoC delivery"
}
`)

	makeDir(t, mos, "contracts", "active", "CON-NO-JUST")
	writeFile(t, filepath.Join(mos, "contracts", "active", "CON-NO-JUST", "contract.mos"), `contract "CON-NO-JUST" {
  title  = "No justifies"
  status = "active"
}
`)

	l := &Linter{}
	diags, err := l.Lint(root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	assertHasDiag(t, diags, "directive-alignment", SeverityInfo)
}

func TestDirectiveAlignmentSilentWhenJustifies(t *testing.T) {
	root := t.TempDir()
	mos := filepath.Join(root, ".mos")

	writeFile(t, filepath.Join(mos, "config.mos"), directiveConfig())

	makeDir(t, mos, "directives", "active", "DIR-2026-001")
	writeFile(t, filepath.Join(mos, "directives", "active", "DIR-2026-001", "directive.mos"), `directive "DIR-2026-001" {
  title  = "Ship PoC"
  status = "active"
}
`)

	makeDir(t, mos, "contracts", "active", "CON-HAS-JUST")
	writeFile(t, filepath.Join(mos, "contracts", "active", "CON-HAS-JUST", "contract.mos"), `contract "CON-HAS-JUST" {
  title    = "Has justifies"
  status   = "active"
  justifies = "NEED-2026-001"
}
`)

	l := &Linter{}
	diags, err := l.Lint(root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	for _, d := range diags {
		if d.Rule == "directive-alignment" && strings.Contains(d.Message, "CON-HAS-JUST") {
			t.Errorf("should not fire for contract with justifies: %s", d.Message)
		}
	}
}

// --- CON-2026-234: Structured machine-readable diagnostics ---

func TestExtractArtifactID(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{".mos/contracts/active/CON-2026-100/contract.mos", "CON-2026-100"},
		{".mos/rules/mechanical/RUL-2026-001.mos", "RUL-2026-001"},
		{".mos/sprints/active/SPR-2026-036/sprint.mos", "SPR-2026-036"},
		{".mos/contracts/archive/CON-2026-050/contract.mos", "CON-2026-050"},
		{"config.mos", "config"},
		{".mos/config.mos", "config"},
		{"", ""},
	}
	for _, tt := range tests {
		got := ExtractArtifactID(tt.path)
		if got != tt.want {
			t.Errorf("ExtractArtifactID(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestDiagnosticJSONIncludesNewFields(t *testing.T) {
	diags := []Diagnostic{
		{
			File:            ".mos/contracts/active/CON-2026-100/contract.mos",
			Severity:        SeverityError,
			Rule:            "crossref-rule",
			Message:         "references unknown artifact",
			ArtifactID:      "CON-2026-100",
			SuggestedAction: "Update the reference to point to an existing artifact ID",
		},
	}
	data, err := json.Marshal(diags)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got []map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got[0]["artifact_id"] != "CON-2026-100" {
		t.Errorf("artifact_id = %v, want CON-2026-100", got[0]["artifact_id"])
	}
	if got[0]["suggested_action"] != "Update the reference to point to an existing artifact ID" {
		t.Errorf("suggested_action = %v, want fix text", got[0]["suggested_action"])
	}
}

func TestDiagnosticJSONOmitsEmptyNewFields(t *testing.T) {
	d := Diagnostic{File: "test.mos", Severity: SeverityInfo, Rule: "custom", Message: "msg"}
	data, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	s := string(data)
	if strings.Contains(s, "artifact_id") {
		t.Errorf("empty artifact_id should be omitted, got: %s", s)
	}
	if strings.Contains(s, "suggested_action") {
		t.Errorf("empty suggested_action should be omitted, got: %s", s)
	}
}

func TestEnrichDiagnostics(t *testing.T) {
	diags := []Diagnostic{
		{File: ".mos/contracts/active/CON-2026-100/contract.mos", Rule: "crossref-rule"},
		{File: ".mos/rules/mechanical/RUL-2026-001.mos", Rule: "dsl-parse"},
		{File: "config.mos", Rule: "unknown-rule-no-suggestion"},
	}
	enrichDiagnostics(diags)

	if diags[0].ArtifactID != "CON-2026-100" {
		t.Errorf("diags[0].ArtifactID = %q, want CON-2026-100", diags[0].ArtifactID)
	}
	if diags[0].SuggestedAction != suggestedActions["crossref-rule"] {
		t.Errorf("diags[0].SuggestedAction = %q, want %q", diags[0].SuggestedAction, suggestedActions["crossref-rule"])
	}
	if diags[1].ArtifactID != "RUL-2026-001" {
		t.Errorf("diags[1].ArtifactID = %q, want RUL-2026-001", diags[1].ArtifactID)
	}
	if diags[1].SuggestedAction == "" {
		t.Error("diags[1].SuggestedAction should not be empty for dsl-parse")
	}
	if diags[2].SuggestedAction != "" {
		t.Errorf("diags[2].SuggestedAction = %q, want empty for unknown rule", diags[2].SuggestedAction)
	}
}

// --- CON-2026-235: Incremental diff-aware linting ---

func TestFilterNewOnly(t *testing.T) {
	baseline := []Diagnostic{
		{File: "a.mos", Rule: "r1", Message: "existing issue"},
		{File: "b.mos", Rule: "r2", Message: "another old issue"},
	}
	all := []Diagnostic{
		{File: "a.mos", Rule: "r1", Message: "existing issue"},
		{File: "b.mos", Rule: "r2", Message: "another old issue"},
		{File: "c.mos", Rule: "r3", Message: "brand new issue"},
	}
	result := FilterNewOnly(all, baseline)
	if len(result) != 1 {
		t.Fatalf("FilterNewOnly returned %d diags, want 1", len(result))
	}
	if result[0].File != "c.mos" || result[0].Rule != "r3" {
		t.Errorf("unexpected diagnostic: %+v", result[0])
	}
}

func TestFilterNewOnlyDuplicates(t *testing.T) {
	baseline := []Diagnostic{
		{File: "a.mos", Rule: "r1", Message: "dup"},
	}
	all := []Diagnostic{
		{File: "a.mos", Rule: "r1", Message: "dup"},
		{File: "a.mos", Rule: "r1", Message: "dup"},
	}
	result := FilterNewOnly(all, baseline)
	if len(result) != 1 {
		t.Fatalf("FilterNewOnly returned %d diags, want 1 (second dup is new)", len(result))
	}
}

func TestLintFilesScoping(t *testing.T) {
	root := t.TempDir()
	mos := filepath.Join(root, ".mos")
	writeFile(t, filepath.Join(mos, "config.mos"), `config {
  mos {
    version = 1
  }
}
`)

	contractDir := filepath.Join(mos, "contracts", "active", "CON-2026-900")
	makeDir(t, contractDir)
	writeFile(t, filepath.Join(contractDir, "contract.mos"), `contract "CON-2026-900" {
  title = "Test"
  status = "draft"
  kind = "feature"
}`)

	contractDir2 := filepath.Join(mos, "contracts", "active", "CON-2026-901")
	makeDir(t, contractDir2)
	writeFile(t, filepath.Join(contractDir2, "contract.mos"), `contract "CON-2026-901" {
  title = "Other"
  status = "draft"
  kind = "feature"
}`)

	l := &Linter{}
	scoped, err := l.LintFiles(root, []string{
		filepath.Join(mos, "contracts", "active", "CON-2026-900", "contract.mos"),
	})
	if err != nil {
		t.Fatalf("LintFiles: %v", err)
	}

	for _, d := range scoped {
		if strings.Contains(d.File, "CON-2026-901") {
			t.Errorf("LintFiles should not include diagnostics for unscoped file CON-2026-901, got: %+v", d)
		}
	}
}

func TestDetectChangedFilesNoVCS(t *testing.T) {
	root := t.TempDir()
	mos := filepath.Join(root, ".mos")
	os.MkdirAll(mos, 0755)

	changed, err := DetectChangedFiles(root)
	if err != nil {
		t.Fatalf("DetectChangedFiles: %v", err)
	}
	if changed != nil {
		t.Errorf("expected nil (no VCS), got %v", changed)
	}
}

func addArtifactTypesWithDerivesToConfig(t *testing.T, mos string) {
	t.Helper()
	configPath := filepath.Join(mos, "config.mos")
	data, _ := os.ReadFile(configPath)
	content := string(data)
	extra := `
  artifact_type "need" {
    directory = "needs"
    fields {
      title { required = true }
      sensation { required = true }
      status { required = true }
      acceptance {}
      originating {}
      derives_from { link = true ref_kind = "need" }
    }
    lifecycle {
      active_states = ["identified", "validated", "addressed"]
      archive_states = ["retired"]
      expects_downstream { via = "satisfies" after = "validated" severity = "warn" }
    }
  }
  artifact_type "specification" {
    directory = "specifications"
    fields {
      title { required = true }
      enforcement { required = true }
      satisfies { link = true ref_kind = "need" }
      addresses {}
    }
    lifecycle {
      active_states = ["active"]
      archive_states = ["retired"]
    }
  }
`
	content = strings.TrimSuffix(strings.TrimSpace(content), "}")
	content += "\n" + extra + "\n}\n"
	writeFile(t, configPath, content)
}

func TestSpecOnOriginating_WarnWhenDerivedNeedsExist(t *testing.T) {
	root := validCstDir(t)
	mos := filepath.Join(root, ".mos")
	addArtifactTypesWithDerivesToConfig(t, mos)

	needDir := makeDir(t, mos, "needs", "active", "NEED-001")
	writeFile(t, filepath.Join(needDir, "need.mos"), `need "NEED-001" {
  title = "Originating"
  sensation = "Root need"
  status = "validated"
  originating = "true"
}`)

	need2Dir := makeDir(t, mos, "needs", "active", "NEED-002")
	writeFile(t, filepath.Join(need2Dir, "need.mos"), `need "NEED-002" {
  title = "Derived"
  sensation = "Child"
  status = "identified"
  derives_from = "NEED-001"
}`)

	specDir := makeDir(t, mos, "specifications", "active", "SPEC-001")
	writeFile(t, filepath.Join(specDir, "specification.mos"), `specification "SPEC-001" {
  title = "On originating"
  enforcement = "warn"
  satisfies = "NEED-001"
  status = "active"
}`)

	l := &Linter{}
	diags, err := l.Lint(root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	assertHasDiag(t, diags, "spec-on-originating", SeverityWarning)
}

func TestSpecOnOriginating_NoWarnWithoutDerivedNeeds(t *testing.T) {
	root := validCstDir(t)
	mos := filepath.Join(root, ".mos")
	addArtifactTypesWithDerivesToConfig(t, mos)

	needDir := makeDir(t, mos, "needs", "active", "NEED-001")
	writeFile(t, filepath.Join(needDir, "need.mos"), `need "NEED-001" {
  title = "Originating"
  sensation = "Root need"
  status = "validated"
  originating = "true"
}`)

	specDir := makeDir(t, mos, "specifications", "active", "SPEC-001")
	writeFile(t, filepath.Join(specDir, "specification.mos"), `specification "SPEC-001" {
  title = "On originating"
  enforcement = "warn"
  satisfies = "NEED-001"
  status = "active"
}`)

	l := &Linter{}
	diags, err := l.Lint(root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	for _, d := range diags {
		if d.Rule == "spec-on-originating" {
			t.Errorf("unexpected spec-on-originating warning: %s", d.Message)
		}
	}
}

func TestSpecOnOriginating_NoWarnForDerivedNeedSpec(t *testing.T) {
	root := validCstDir(t)
	mos := filepath.Join(root, ".mos")
	addArtifactTypesWithDerivesToConfig(t, mos)

	needDir := makeDir(t, mos, "needs", "active", "NEED-001")
	writeFile(t, filepath.Join(needDir, "need.mos"), `need "NEED-001" {
  title = "Originating"
  sensation = "Root need"
  status = "validated"
  originating = "true"
}`)

	need2Dir := makeDir(t, mos, "needs", "active", "NEED-002")
	writeFile(t, filepath.Join(need2Dir, "need.mos"), `need "NEED-002" {
  title = "Derived"
  sensation = "Child"
  status = "identified"
  derives_from = "NEED-001"
}`)

	specDir := makeDir(t, mos, "specifications", "active", "SPEC-001")
	writeFile(t, filepath.Join(specDir, "specification.mos"), `specification "SPEC-001" {
  title = "On derived"
  enforcement = "warn"
  satisfies = "NEED-002"
  status = "active"
}`)

	l := &Linter{}
	diags, err := l.Lint(root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	for _, d := range diags {
		if d.Rule == "spec-on-originating" {
			t.Errorf("unexpected spec-on-originating warning: %s", d.Message)
		}
	}
}
