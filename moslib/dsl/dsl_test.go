package dsl

import (
	"strings"
	"testing"
)

// --- Parse tests for each artifact type ---

func TestParseRule(t *testing.T) {
	src := `
rule "test-rule" {
  name = "Test Rule"
  type = "mechanical"
  scope = "organization"
  enforcement = "error"
  priority = 100
  deprecated = false
  tags = ["api", "stability"]

  feature "Test Rule" {
    scenario "Basic check" {
      given {
        a project
      }
      when {
        the rule is evaluated
      }
      then {
        it passes
      }
    }
  }
}
`
	f, err := Parse(src, nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	ab, ok := f.Artifact.(*ArtifactBlock)
	if !ok {
		t.Fatal("expected ArtifactBlock")
	}
	if ab.Kind != "rule" {
		t.Errorf("Kind = %q, want rule", ab.Kind)
	}
	if ab.Name != "test-rule" {
		t.Errorf("Name = %q, want test-rule", ab.Name)
	}

	var fields []*Field
	var features []*FeatureBlock
	for _, item := range ab.Items {
		switch v := item.(type) {
		case *Field:
			fields = append(fields, v)
		case *FeatureBlock:
			features = append(features, v)
		}
	}

	if len(fields) != 7 {
		t.Errorf("got %d fields, want 7", len(fields))
	}
	if len(features) != 1 {
		t.Fatalf("got %d feature blocks, want 1", len(features))
	}

	feat := features[0]
	if feat.Name != "Test Rule" {
		t.Errorf("Feature name = %q, want Test Rule", feat.Name)
	}
	if len(feat.Groups) != 1 {
		t.Errorf("got %d scenario groups, want 1", len(feat.Groups))
	}
}

func TestParseContract(t *testing.T) {
	src := `
contract "CON-2026-001" {
  title = "Test Contract"
  status = "active"

  bill {
    introduced_by = "generator"
    introduced_at = "2026-01-01"
    intent = "Testing"
  }

  feature "Contract acceptance" {
    scenario "acceptance" {
      given {
        the contract is active
      }
      when {
        deliverables are reviewed
      }
      then {
        the contract passes
      }
    }
  }
}
`
	f, err := Parse(src, nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	ab := f.Artifact.(*ArtifactBlock)
	if ab.Kind != "contract" {
		t.Errorf("Kind = %q, want contract", ab.Kind)
	}
	if ab.Name != "CON-2026-001" {
		t.Errorf("Name = %q, want CON-2026-001", ab.Name)
	}
}

func TestParseConfig(t *testing.T) {
	src := `
config {
  mos {
    version = 1
  }

  backend {
    type = "git"
  }

  governance {
    model = "committee"
  }
}
`
	f, err := Parse(src, nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	ab := f.Artifact.(*ArtifactBlock)
	if ab.Kind != "config" {
		t.Errorf("Kind = %q, want config", ab.Kind)
	}
	if ab.Name != "" {
		t.Errorf("Name = %q, want empty", ab.Name)
	}

	blockCount := 0
	for _, item := range ab.Items {
		if _, ok := item.(*Block); ok {
			blockCount++
		}
	}
	if blockCount != 3 {
		t.Errorf("got %d nested blocks, want 3", blockCount)
	}
}

func TestParseDeclaration(t *testing.T) {
	src := `
declaration {
  name = "test-project"
  created = "2026-01-01"
  authors = ["alice", "bob"]
}
`
	f, err := Parse(src, nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	ab := f.Artifact.(*ArtifactBlock)
	if ab.Kind != "declaration" {
		t.Errorf("Kind = %q, want declaration", ab.Kind)
	}
}

func TestParseLexicon(t *testing.T) {
	src := `
lexicon {
  terms {
    rule = "A governance directive"
    contract = "A time-boxed work agreement"
  }

  artifact_labels {
    rule = "Rule"
    contract = "Contract"
  }
}
`
	f, err := Parse(src, nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	ab := f.Artifact.(*ArtifactBlock)
	if ab.Kind != "lexicon" {
		t.Errorf("Kind = %q, want lexicon", ab.Kind)
	}
}

func TestParseLayers(t *testing.T) {
	src := `
layers {
  layer "project" {
    level = 1
  }
  layer "organization" {
    level = 2
    inherits_from = "project"
  }
}
`
	f, err := Parse(src, nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	ab := f.Artifact.(*ArtifactBlock)
	if ab.Kind != "layers" {
		t.Errorf("Kind = %q, want layers", ab.Kind)
	}
	if len(ab.Items) != 2 {
		t.Fatalf("got %d layer entries, want 2", len(ab.Items))
	}

	blk0 := ab.Items[0].(*Block)
	if blk0.Name != "layer" {
		t.Errorf("block 0 name = %q, want layer", blk0.Name)
	}
	if blk0.Title != "project" {
		t.Errorf("block 0 title = %q, want project", blk0.Title)
	}
	blk1 := ab.Items[1].(*Block)
	if blk1.Name != "layer" {
		t.Errorf("block 1 name = %q, want layer", blk1.Name)
	}
	if blk1.Title != "organization" {
		t.Errorf("block 1 title = %q, want organization", blk1.Title)
	}
}

// --- Value type tests ---

func TestParseDateTime(t *testing.T) {
	src := `
config {
  created = 2024-06-15T00:00:00Z
  updated = 2026-02-01T14:30:00+02:00
}
`
	f, err := Parse(src, nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	ab := f.Artifact.(*ArtifactBlock)
	for _, item := range ab.Items {
		fld, ok := item.(*Field)
		if !ok {
			continue
		}
		dt, ok := fld.Value.(*DateTimeVal)
		if !ok {
			t.Errorf("field %q: expected DateTimeVal, got %T", fld.Key, fld.Value)
			continue
		}
		if fld.Key == "created" && dt.Raw != "2024-06-15T00:00:00Z" {
			t.Errorf("created = %q", dt.Raw)
		}
	}
}

func TestParseInlineTable(t *testing.T) {
	src := `
config {
  env = { FOO = "bar", BAZ = "qux" }
}
`
	f, err := Parse(src, nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	ab := f.Artifact.(*ArtifactBlock)
	fld := ab.Items[0].(*Field)
	tbl, ok := fld.Value.(*InlineTableVal)
	if !ok {
		t.Fatalf("expected InlineTableVal, got %T", fld.Value)
	}
	if len(tbl.Fields) != 2 {
		t.Errorf("got %d fields, want 2", len(tbl.Fields))
	}
}

func TestParseTripleQuotedString(t *testing.T) {
	src := `
config {
  desc = """
    Line one.
    Line two.
    """
}
`
	f, err := Parse(src, nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	ab := f.Artifact.(*ArtifactBlock)
	fld := ab.Items[0].(*Field)
	sv, ok := fld.Value.(*StringVal)
	if !ok {
		t.Fatalf("expected StringVal, got %T", fld.Value)
	}
	if !strings.Contains(sv.Text, "Line one") {
		t.Errorf("triple-quoted string missing content: %q", sv.Text)
	}
}

// --- Spec block tests ---

func TestParseSpecInclude(t *testing.T) {
	src := `
rule "test" {
  spec {
    include "spec.feature"
  }
}
`
	f, err := Parse(src, nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	ab := f.Artifact.(*ArtifactBlock)
	for _, item := range ab.Items {
		sb, ok := item.(*SpecBlock)
		if !ok {
			continue
		}
		if len(sb.Includes) != 1 {
			t.Fatalf("expected 1 include, got %d", len(sb.Includes))
		}
		if sb.Includes[0].Path != "spec.feature" {
			t.Errorf("include path = %q, want spec.feature", sb.Includes[0].Path)
		}
		return
	}
	t.Fatal("no spec block found")
}

func TestParseFeatureWithBackground(t *testing.T) {
	src := `
rule "test" {
  feature "My Feature" {
    background {
      given {
        a setup step
      }
    }

    scenario "First" {
      given {
        a condition
      }
      when {
        an action
      }
      then {
        a result
      }
    }

    scenario "Second" {
      given {
        another condition
      }
      then {
        another result
      }
    }
  }
}
`
	f, err := Parse(src, nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	ab := f.Artifact.(*ArtifactBlock)
	var fb *FeatureBlock
	for _, item := range ab.Items {
		if f, ok := item.(*FeatureBlock); ok {
			fb = f
			break
		}
	}
	if fb == nil {
		t.Fatal("no feature block found")
	}

	if fb.Name != "My Feature" {
		t.Errorf("feature name = %q", fb.Name)
	}
	if fb.Background == nil {
		t.Fatal("expected background")
	}
	if fb.Background.Given == nil || len(fb.Background.Given.Lines) != 1 {
		t.Errorf("background given lines = %v", fb.Background.Given)
	}
	if len(fb.Groups) != 2 {
		t.Errorf("scenario count = %d, want 2", len(fb.Groups))
	}
}

func TestParseFeatureWithGroups(t *testing.T) {
	src := `
rule "test" {
  feature "Grouped Feature" {
    group "First group" {
      scenario "A" {
        given {
          step a
        }
      }
    }

    group "Second group" {
      scenario "B" {
        given {
          step b
        }
        when {
          action b
        }
        then {
          result b
        }
      }
    }
  }
}
`
	f, err := Parse(src, nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	ab := f.Artifact.(*ArtifactBlock)
	var fb *FeatureBlock
	for _, item := range ab.Items {
		if feat, ok := item.(*FeatureBlock); ok {
			fb = feat
			break
		}
	}
	if fb == nil {
		t.Fatal("no feature block found")
	}

	if len(fb.Groups) != 2 {
		t.Errorf("got %d groups, want 2", len(fb.Groups))
	}

	g0, ok := fb.Groups[0].(*Group)
	if !ok {
		t.Fatal("first group should be Group")
	}
	if g0.Name != "First group" {
		t.Errorf("group 0 name = %q", g0.Name)
	}
	if len(g0.Scenarios) != 1 {
		t.Errorf("group 0 scenarios = %d, want 1", len(g0.Scenarios))
	}
}

func TestParseScenarioSUTTestCase(t *testing.T) {
	src := `
rule "test" {
  feature "Traceability" {
    scenario "Linked scenario" {
      sut = "cmd/lint"
      test = "lint_test.go"
      case = "TestLint/Binary file detected"
      given {
        a binary file in staging
      }
      when {
        the linter runs
      }
      then {
        a warning is emitted
      }
    }
  }
}
`
	f, err := Parse(src, nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	ab := f.Artifact.(*ArtifactBlock)
	var fb *FeatureBlock
	for _, item := range ab.Items {
		if feat, ok := item.(*FeatureBlock); ok {
			fb = feat
			break
		}
	}
	if fb == nil {
		t.Fatal("no feature block")
	}
	if len(fb.Groups) != 1 {
		t.Fatalf("expected 1 scenario, got %d groups", len(fb.Groups))
	}

	sc, ok := fb.Groups[0].(*Scenario)
	if !ok {
		t.Fatal("expected Scenario")
	}
	if sc.SUT() != "cmd/lint" {
		t.Errorf("SUT() = %q, want cmd/lint", sc.SUT())
	}
	if sc.Test() != "lint_test.go" {
		t.Errorf("Test() = %q, want lint_test.go", sc.Test())
	}
	if sc.Case() != "TestLint/Binary file detected" {
		t.Errorf("Case() = %q, want TestLint/Binary file detected", sc.Case())
	}
	if len(sc.Fields) != 3 {
		t.Errorf("Fields count = %d, want 3", len(sc.Fields))
	}
	if sc.Given == nil || len(sc.Given.Lines) != 1 {
		t.Errorf("Given = %v", sc.Given)
	}
	if sc.When == nil || len(sc.When.Lines) != 1 {
		t.Errorf("When = %v", sc.When)
	}
	if sc.Then == nil || len(sc.Then.Lines) != 1 {
		t.Errorf("Then = %v", sc.Then)
	}
}

func TestParseMultiLineStepBlock(t *testing.T) {
	src := `
rule "test" {
  feature "Multi-line steps" {
    scenario "Many conditions" {
      given {
        a user is logged in
        the user has admin role
        the system is in maintenance mode
      }
      then {
        access is denied
        an audit log entry is created
      }
    }
  }
}
`
	f, err := Parse(src, nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	ab := f.Artifact.(*ArtifactBlock)
	fb := ab.Items[0].(*FeatureBlock)
	sc := fb.Groups[0].(*Scenario)

	if sc.Given == nil || len(sc.Given.Lines) != 3 {
		t.Errorf("Given lines = %d, want 3", len(sc.Given.Lines))
	}
	if sc.Then == nil || len(sc.Then.Lines) != 2 {
		t.Errorf("Then lines = %d, want 2", len(sc.Then.Lines))
	}
}

func TestParseInlineFeatureWithoutSpec(t *testing.T) {
	src := `
rule "test" {
  name = "Inline"

  feature "Direct feature" {
    scenario "Simple" {
      given {
        something
      }
      then {
        result
      }
    }
  }
}
`
	f, err := Parse(src, nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	ab := f.Artifact.(*ArtifactBlock)
	var hasSpec, hasFeature bool
	for _, item := range ab.Items {
		switch item.(type) {
		case *SpecBlock:
			hasSpec = true
		case *FeatureBlock:
			hasFeature = true
		}
	}

	if hasSpec {
		t.Error("should not have SpecBlock for inline feature")
	}
	if !hasFeature {
		t.Error("should have FeatureBlock as direct child")
	}
}

func TestParseMixedSpecContent(t *testing.T) {
	src := `
rule "test" {
  spec {
    include "a.mos"

    feature "Inline" {
      scenario "S1" {
        given {
          something
        }
        then {
          result
        }
      }
    }
  }
}
`
	f, err := Parse(src, nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	ab := f.Artifact.(*ArtifactBlock)
	var sb *SpecBlock
	for _, item := range ab.Items {
		if s, ok := item.(*SpecBlock); ok {
			sb = s
			break
		}
	}
	if sb == nil {
		t.Fatal("no spec block found")
	}
	if len(sb.Includes) != 1 {
		t.Errorf("expected 1 include, got %d", len(sb.Includes))
	}
	if len(sb.Features) != 1 {
		t.Errorf("expected 1 feature, got %d", len(sb.Features))
	}
}

func TestParseBackgroundGivenBlock(t *testing.T) {
	src := `
rule "test" {
  feature "BG" {
    background {
      given {
        setup step one
        setup step two
      }
    }

    scenario "Test" {
      given {
        specific condition
      }
      then {
        expected outcome
      }
    }
  }
}
`
	f, err := Parse(src, nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	ab := f.Artifact.(*ArtifactBlock)
	fb := ab.Items[0].(*FeatureBlock)
	if fb.Background == nil {
		t.Fatal("expected background")
	}
	if fb.Background.Given == nil {
		t.Fatal("expected given in background")
	}
	if len(fb.Background.Given.Lines) != 2 {
		t.Errorf("background given lines = %d, want 2", len(fb.Background.Given.Lines))
	}
}

// --- Part A: Open scenario fields ---

func TestParseScenarioCustomFields(t *testing.T) {
	src := `
rule "test" {
  feature "Custom fields" {
    scenario "Rich metadata" {
      sut = "cmd/lint"
      test = "lint_test.go"
      case = "TestLint/Check"
      priority = "high"
      owner = "alice"
      given {
        a project
      }
      then {
        it passes
      }
    }
  }
}
`
	f, err := Parse(src, nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	ab := f.Artifact.(*ArtifactBlock)
	fb := ab.Items[0].(*FeatureBlock)
	sc := fb.Groups[0].(*Scenario)

	if len(sc.Fields) != 5 {
		t.Fatalf("Fields count = %d, want 5", len(sc.Fields))
	}
	if sc.SUT() != "cmd/lint" {
		t.Errorf("SUT() = %q", sc.SUT())
	}
	if sc.Test() != "lint_test.go" {
		t.Errorf("Test() = %q", sc.Test())
	}
	if sc.Case() != "TestLint/Check" {
		t.Errorf("Case() = %q", sc.Case())
	}
	if sc.fieldVal("priority") != "high" {
		t.Errorf("priority = %q", sc.fieldVal("priority"))
	}
	if sc.fieldVal("owner") != "alice" {
		t.Errorf("owner = %q", sc.fieldVal("owner"))
	}
}

func TestParseKeywordAsScenarioField(t *testing.T) {
	src := `
rule "test" {
  feature "Keyword fields" {
    scenario "Date tracking" {
      when = "2026-03-01"
      given = "precondition-met"
      then = "expected-outcome"
      sut = "cmd/lint"
      given {
        a project exists
      }
      when {
        lint runs
      }
      then {
        it passes
      }
    }
  }
}
`
	f, err := Parse(src, nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	ab := f.Artifact.(*ArtifactBlock)
	fb := ab.Items[0].(*FeatureBlock)
	sc := fb.Groups[0].(*Scenario)

	if len(sc.Fields) != 4 {
		t.Fatalf("Fields count = %d, want 4", len(sc.Fields))
	}
	if sc.fieldVal("when") != "2026-03-01" {
		t.Errorf("when field = %q, want 2026-03-01", sc.fieldVal("when"))
	}
	if sc.fieldVal("given") != "precondition-met" {
		t.Errorf("given field = %q", sc.fieldVal("given"))
	}
	if sc.fieldVal("then") != "expected-outcome" {
		t.Errorf("then field = %q", sc.fieldVal("then"))
	}
	if sc.SUT() != "cmd/lint" {
		t.Errorf("SUT() = %q", sc.SUT())
	}
	if sc.Given == nil || len(sc.Given.Lines) == 0 {
		t.Error("Given step block missing")
	}
	if sc.When == nil || len(sc.When.Lines) == 0 {
		t.Error("When step block missing")
	}
	if sc.Then == nil || len(sc.Then.Lines) == 0 {
		t.Error("Then step block missing")
	}

	out := Format(f, nil)
	if !strings.Contains(out, `when = "2026-03-01"`) {
		t.Errorf("Format missing keyword field 'when':\n%s", out)
	}
	rt, err := Parse(out, nil)
	if err != nil {
		t.Fatalf("Round-trip Parse: %v", err)
	}
	rtsc := rt.Artifact.(*ArtifactBlock).Items[0].(*FeatureBlock).Groups[0].(*Scenario)
	if rtsc.fieldVal("when") != "2026-03-01" {
		t.Errorf("Round-trip when = %q", rtsc.fieldVal("when"))
	}
}

// --- Part B: Open artifact type parsing ---

func TestParseCustomArtifactType(t *testing.T) {
	src := `guideline "G-001" {
  name = "Style guide"
}
`
	f, err := Parse(src, nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	ab := f.Artifact.(*ArtifactBlock)
	if ab.Kind != "guideline" {
		t.Errorf("Kind = %q, want guideline", ab.Kind)
	}
	if ab.Name != "G-001" {
		t.Errorf("Name = %q, want G-001", ab.Name)
	}
}

func TestParseCustomUnnamedArtifact(t *testing.T) {
	src := `settings {
  theme = "dark"
}
`
	f, err := Parse(src, nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	ab := f.Artifact.(*ArtifactBlock)
	if ab.Kind != "settings" {
		t.Errorf("Kind = %q, want settings", ab.Kind)
	}
	if ab.Name != "" {
		t.Errorf("Name = %q, want empty", ab.Name)
	}
}

// --- Part C: Named nested blocks ---

func TestParseNamedNestedBlock(t *testing.T) {
	src := `config {
  zone "us-east-1" {
    region = "US"
  }
}
`
	f, err := Parse(src, nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	ab := f.Artifact.(*ArtifactBlock)
	if len(ab.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(ab.Items))
	}

	blk := ab.Items[0].(*Block)
	if blk.Name != "zone" {
		t.Errorf("Name = %q, want zone", blk.Name)
	}
	if blk.Title != "us-east-1" {
		t.Errorf("Title = %q, want us-east-1", blk.Title)
	}
	if len(blk.Items) != 1 {
		t.Fatalf("expected 1 field in zone block, got %d", len(blk.Items))
	}
	fld := blk.Items[0].(*Field)
	if fld.Key != "region" {
		t.Errorf("field key = %q, want region", fld.Key)
	}
}

func TestParseLayersRoundTrip(t *testing.T) {
	src := `layers {
  layer "project" {
    level = 1
  }
  layer "org" {
    level = 2
  }
}
`
	f, err := Parse(src, nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	out1 := Format(f, nil)

	f2, err := Parse(out1, nil)
	if err != nil {
		t.Fatalf("Parse formatted: %v", err)
	}
	out2 := Format(f2, nil)

	if out1 != out2 {
		t.Errorf("layers round-trip not idempotent:\n--- first ---\n%s\n--- second ---\n%s", out1, out2)
	}

	ab := f.Artifact.(*ArtifactBlock)
	for _, item := range ab.Items {
		blk, ok := item.(*Block)
		if !ok {
			t.Errorf("expected *Block, got %T", item)
			continue
		}
		if blk.Title == "" {
			t.Error("expected non-empty Title for layer block")
		}
	}
}

// --- Part D: KeywordMap ---

func TestParseSpanishKeywords(t *testing.T) {
	kw := DefaultKeywords()
	kw.ToMachine["regla"] = "rule"
	kw.ToHuman["rule"] = "regla"
	kw.ToMachine["característica"] = "feature"
	kw.ToHuman["feature"] = "característica"
	kw.ToMachine["escenario"] = "scenario"
	kw.ToHuman["scenario"] = "escenario"
	kw.ToMachine["dado"] = "given"
	kw.ToHuman["given"] = "dado"
	kw.ToMachine["cuando"] = "when"
	kw.ToHuman["when"] = "cuando"
	kw.ToMachine["entonces"] = "then"
	kw.ToHuman["then"] = "entonces"

	src := `regla "R-001" {
  name = "Test"

  característica "Test" {
    escenario "S" {
      dado {
        algo
      }
      entonces {
        resultado
      }
    }
  }
}
`
	f, err := Parse(src, kw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	ab := f.Artifact.(*ArtifactBlock)
	if ab.Kind != "rule" {
		t.Errorf("Kind = %q, want rule (machine name)", ab.Kind)
	}
	if ab.Name != "R-001" {
		t.Errorf("Name = %q, want R-001", ab.Name)
	}

	var fb *FeatureBlock
	for _, item := range ab.Items {
		if feat, ok := item.(*FeatureBlock); ok {
			fb = feat
			break
		}
	}
	if fb == nil {
		t.Fatal("no feature block")
	}
	if len(fb.Groups) != 1 {
		t.Fatalf("expected 1 scenario, got %d", len(fb.Groups))
	}
	sc := fb.Groups[0].(*Scenario)
	if sc.Given == nil {
		t.Error("expected Given step block")
	}
	if sc.Then == nil {
		t.Error("expected Then step block")
	}
}

func TestFormatSpanishKeywords(t *testing.T) {
	kw := DefaultKeywords()
	kw.ToMachine["regla"] = "rule"
	kw.ToHuman["rule"] = "regla"
	kw.ToMachine["característica"] = "feature"
	kw.ToHuman["feature"] = "característica"
	kw.ToMachine["escenario"] = "scenario"
	kw.ToHuman["scenario"] = "escenario"
	kw.ToMachine["dado"] = "given"
	kw.ToHuman["given"] = "dado"
	kw.ToMachine["cuando"] = "when"
	kw.ToHuman["when"] = "cuando"
	kw.ToMachine["entonces"] = "then"
	kw.ToHuman["then"] = "entonces"

	f := &File{
		Artifact: &ArtifactBlock{
			Kind: "rule",
			Name: "R-001",
			Items: []Node{
				&FeatureBlock{
					Name: "Test",
					Groups: []ScenarioContainer{
						&Scenario{
							Name:  "S",
							Given: &StepBlock{Lines: []string{"algo"}},
							Then:  &StepBlock{Lines: []string{"resultado"}},
						},
					},
				},
			},
		},
	}

	out := Format(f, kw)
	if !strings.Contains(out, "regla") {
		t.Errorf("expected 'regla' in output:\n%s", out)
	}
	if !strings.Contains(out, "característica") {
		t.Errorf("expected 'característica' in output:\n%s", out)
	}
	if !strings.Contains(out, "escenario") {
		t.Errorf("expected 'escenario' in output:\n%s", out)
	}
	if !strings.Contains(out, "dado") {
		t.Errorf("expected 'dado' in output:\n%s", out)
	}
	if !strings.Contains(out, "entonces") {
		t.Errorf("expected 'entonces' in output:\n%s", out)
	}
	for _, line := range strings.Split(out, "\n") {
		trimmed := strings.TrimSpace(line)
		for _, eng := range []string{"rule ", "feature ", "scenario ", "given {", "then {"} {
			if strings.HasPrefix(trimmed, eng) {
				t.Errorf("English keyword %q should not start a line, found: %q", eng, trimmed)
			}
		}
	}
}

// --- Part E: Lexicon keyword extraction ---

func TestExtractKeywordsFromLexicon(t *testing.T) {
	src := `lexicon {
  keywords {
    feature = "característica"
    rule = "regla"
    scenario = "escenario"
    given = "dado"
    when = "cuando"
    then = "entonces"
  }

  terms {
    rule = "A governance directive"
  }
}
`
	lexiconFile, err := Parse(src, nil)
	if err != nil {
		t.Fatalf("Parse lexicon: %v", err)
	}

	kw := ExtractKeywords(lexiconFile)

	if kw.ToMachine["característica"] != "feature" {
		t.Errorf("ToMachine[característica] = %q, want feature", kw.ToMachine["característica"])
	}
	if kw.ToHuman["feature"] != "característica" {
		t.Errorf("ToHuman[feature] = %q, want característica", kw.ToHuman["feature"])
	}
	if kw.ToMachine["regla"] != "rule" {
		t.Errorf("ToMachine[regla] = %q, want rule", kw.ToMachine["regla"])
	}
	if kw.ToHuman["rule"] != "regla" {
		t.Errorf("ToHuman[rule] = %q, want regla", kw.ToHuman["rule"])
	}

	// Unmapped keywords fall back to English
	if kw.ToMachine["spec"] != "spec" {
		t.Errorf("ToMachine[spec] = %q, want spec (English default)", kw.ToMachine["spec"])
	}
	if kw.ToHuman["spec"] != "spec" {
		t.Errorf("ToHuman[spec] = %q, want spec (English default)", kw.ToHuman["spec"])
	}
}

func TestExtractKeywordsNilLexicon(t *testing.T) {
	kw := ExtractKeywords(nil)
	if kw.ToMachine["feature"] != "feature" {
		t.Errorf("expected English default for nil lexicon")
	}
}

// --- Full pipeline: lexicon -> parse -> format round-trip ---

func TestFullLocalizationPipeline(t *testing.T) {
	lexiconSrc := `lexicon {
  keywords {
    feature = "característica"
    scenario = "escenario"
    given = "dado"
    when = "cuando"
    then = "entonces"
    rule = "regla"
  }
}
`
	lexicon, err := Parse(lexiconSrc, nil)
	if err != nil {
		t.Fatalf("Parse lexicon: %v", err)
	}
	kw := ExtractKeywords(lexicon)

	ruleSrc := `regla "R-001" {
  name = "Test"

  característica "Feature A" {
    escenario "Scenario 1" {
      sut = "pkg/foo"
      dado {
        precondition
      }
      cuando {
        action
      }
      entonces {
        postcondition
      }
    }
  }
}
`
	f1, err := Parse(ruleSrc, kw)
	if err != nil {
		t.Fatalf("Parse Spanish rule: %v", err)
	}

	ab := f1.Artifact.(*ArtifactBlock)
	if ab.Kind != "rule" {
		t.Errorf("Kind = %q, want rule", ab.Kind)
	}

	out1 := Format(f1, kw)

	f2, err := Parse(out1, kw)
	if err != nil {
		t.Fatalf("Re-parse formatted: %v", err)
	}
	out2 := Format(f2, kw)

	if out1 != out2 {
		t.Errorf("localization round-trip not idempotent:\n--- first ---\n%s\n--- second ---\n%s", out1, out2)
	}

	if !strings.Contains(out1, "regla") {
		t.Errorf("missing Spanish keyword 'regla' in output:\n%s", out1)
	}
	if !strings.Contains(out1, "característica") {
		t.Errorf("missing Spanish keyword 'característica' in output:\n%s", out1)
	}
}

// --- Formatter tests ---

func TestFormatRoundTrip(t *testing.T) {
	src := `rule "test-rule" {
  name = "Test Rule"
  type = "mechanical"
  priority = 42
  tags = ["a", "b"]

  feature "Test" {
    scenario "Basic" {
      given {
        a project
      }
      when {
        the rule runs
      }
      then {
        it passes
      }
    }
  }
}
`
	f, err := Parse(src, nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	formatted := Format(f, nil)

	f2, err := Parse(formatted, nil)
	if err != nil {
		t.Fatalf("Parse formatted: %v", err)
	}

	formatted2 := Format(f2, nil)

	if formatted != formatted2 {
		t.Errorf("round-trip not idempotent:\n--- first ---\n%s\n--- second ---\n%s", formatted, formatted2)
	}
}

func TestFormatCanonicality(t *testing.T) {
	src1 := `rule    "test"   {
  name="Test"
  priority   =   42
  feature    "Test"   {
    scenario   "Basic"   {
      given   {
        a project
      }
      when {
        the rule runs
      }
      then   {
        it passes
      }
    }
  }
}
`
	src2 := `
rule "test" {
    name = "Test"
    priority = 42

    feature "Test" {
        scenario "Basic" {
            given {
                a project
            }
            when {
                the rule runs
            }
            then {
                it passes
            }
        }
    }
}
`
	f1, err := Parse(src1, nil)
	if err != nil {
		t.Fatalf("Parse src1: %v", err)
	}
	f2, err := Parse(src2, nil)
	if err != nil {
		t.Fatalf("Parse src2: %v", err)
	}

	out1 := Format(f1, nil)
	out2 := Format(f2, nil)

	if out1 != out2 {
		t.Errorf("canonical output differs:\n--- src1 ---\n%s\n--- src2 ---\n%s", out1, out2)
	}
}

func TestFormatConfig(t *testing.T) {
	src := `
config {
  mos {
    version = 1
  }
  backend {
    type = "git"
  }
}
`
	f, err := Parse(src, nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	out := Format(f, nil)
	if !strings.Contains(out, "config {") {
		t.Errorf("missing config block in output:\n%s", out)
	}
	if !strings.Contains(out, `version = 1`) {
		t.Errorf("missing version field in output:\n%s", out)
	}
}

func TestFormatLayers(t *testing.T) {
	src := `
layers {
  layer "project" {
    level = 1
  }
  layer "org" {
    level = 2
  }
}
`
	f, err := Parse(src, nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	out := Format(f, nil)
	if !strings.Contains(out, `layer "project"`) {
		t.Errorf("missing project layer:\n%s", out)
	}
	if !strings.Contains(out, `layer "org"`) {
		t.Errorf("missing org layer:\n%s", out)
	}
}

func TestFormatSpecInclude(t *testing.T) {
	src := `
rule "test" {
  spec {
    include "spec.feature"
  }
}
`
	f, err := Parse(src, nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	out := Format(f, nil)
	if !strings.Contains(out, `include "spec.feature"`) {
		t.Errorf("missing include in output:\n%s", out)
	}
}

func TestFormatScenarioSUTTestCase(t *testing.T) {
	src := `
rule "test" {
  feature "F" {
    scenario "S" {
      sut = "cmd/lint"
      test = "lint_test.go"
      case = "TestLint/Binary"
      given {
        something
      }
      then {
        result
      }
    }
  }
}
`
	f, err := Parse(src, nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	out := Format(f, nil)
	if !strings.Contains(out, `sut = "cmd/lint"`) {
		t.Errorf("missing sut in output:\n%s", out)
	}
	if !strings.Contains(out, `test = "lint_test.go"`) {
		t.Errorf("missing test in output:\n%s", out)
	}
	if !strings.Contains(out, `case = "TestLint/Binary"`) {
		t.Errorf("missing case in output:\n%s", out)
	}
}

// --- Error tests ---

func TestParseErrorBadSyntaxAfterKeyword(t *testing.T) {
	_, err := Parse(`unknown 123`, nil)
	if err == nil {
		t.Fatal("expected error for bad token after artifact keyword")
	}
}

func TestParseErrorUnterminatedString(t *testing.T) {
	_, err := Parse(`config { name = "unterminated }`, nil)
	if err == nil {
		t.Fatal("expected error for unterminated string")
	}
}

func TestParseErrorUnterminatedBlock(t *testing.T) {
	_, err := Parse(`config { name = "test" `, nil)
	if err == nil {
		t.Fatal("expected error for unterminated block")
	}
}

func TestParseErrorMissingValue(t *testing.T) {
	_, err := Parse(`config { name = }`, nil)
	if err == nil {
		t.Fatal("expected error for missing value")
	}
}

// --- Lexer tests ---

func TestLexerDateTime(t *testing.T) {
	lex := NewLexer(`2024-06-15T00:00:00Z`, nil)
	tok, err := lex.Next()
	if err != nil {
		t.Fatalf("Lex: %v", err)
	}
	if tok.Type != TokenDateTime {
		t.Errorf("type = %v, want DateTime", tok.Type)
	}
	if tok.Value != "2024-06-15T00:00:00Z" {
		t.Errorf("value = %q", tok.Value)
	}
}

func TestLexerBool(t *testing.T) {
	lex := NewLexer(`true false`, nil)
	t1, _ := lex.Next()
	t2, _ := lex.Next()
	if t1.Type != TokenBool || t1.Value != "true" {
		t.Errorf("t1 = %v %q", t1.Type, t1.Value)
	}
	if t2.Type != TokenBool || t2.Value != "false" {
		t.Errorf("t2 = %v %q", t2.Type, t2.Value)
	}
}

func TestLexerComment(t *testing.T) {
	lex := NewLexer("# a comment\nfoo", nil)
	t1, _ := lex.Next()
	if t1.Type != TokenComment {
		t.Errorf("t1.Type = %v, want Comment", t1.Type)
	}
}

func TestLexerV3Keywords(t *testing.T) {
	keywords := map[string]TokenType{
		"feature":    TokenFeature,
		"background": TokenBackground,
		"scenario":   TokenScenario,
		"given":      TokenGiven,
		"when":       TokenWhen,
		"then":       TokenThen,
		"group":      TokenGroup,
		"spec":       TokenSpec,
		"include":    TokenInclude,
	}
	for word, expected := range keywords {
		lex := NewLexer(word, nil)
		tok, err := lex.Next()
		if err != nil {
			t.Errorf("Lex %q: %v", word, err)
			continue
		}
		if tok.Type != expected {
			t.Errorf("Lex %q: type = %v, want %v", word, tok.Type, expected)
		}
	}
}

// --- Project orchestration ---

func TestLoadKeywordsEmpty(t *testing.T) {
	kw, err := LoadKeywords("")
	if err != nil {
		t.Fatalf("LoadKeywords: %v", err)
	}
	if kw.machineKeyword("feature") != "feature" {
		t.Error("expected identity mapping for empty lexicon")
	}
}

func TestLoadKeywordsFromLexicon(t *testing.T) {
	lexiconSrc := `lexicon {
  keywords {
    feature = "característica"
    scenario = "escenario"
  }
}
`
	kw, err := LoadKeywords(lexiconSrc)
	if err != nil {
		t.Fatalf("LoadKeywords: %v", err)
	}
	if kw.machineKeyword("característica") != "feature" {
		t.Errorf("machineKeyword(característica) = %q", kw.machineKeyword("característica"))
	}
	if kw.humanKeyword("scenario") != "escenario" {
		t.Errorf("humanKeyword(scenario) = %q", kw.humanKeyword("scenario"))
	}
}

func TestParseFilesOrchestration(t *testing.T) {
	lexiconSrc := `lexicon {
  keywords {
    feature = "característica"
    scenario = "escenario"
    given = "dado"
    when = "cuando"
    then = "entonces"
  }
}
`
	sources := map[string]string{
		"rule.mos": `rule "R1" {
  característica "F1" {
    escenario "S1" {
      sut = "cmd/test"
      dado {
        something
      }
      entonces {
        it works
      }
    }
  }
}
`,
	}

	files, kw, err := ParseFiles(lexiconSrc, sources)
	if err != nil {
		t.Fatalf("ParseFiles: %v", err)
	}

	if kw.machineKeyword("característica") != "feature" {
		t.Error("keyword map not applied")
	}

	f, ok := files["rule.mos"]
	if !ok {
		t.Fatal("rule.mos not in results")
	}

	ab := f.Artifact.(*ArtifactBlock)
	if ab.Kind != "rule" {
		t.Errorf("Kind = %q", ab.Kind)
	}
	fb := ab.Items[0].(*FeatureBlock)
	sc := fb.Groups[0].(*Scenario)
	if sc.SUT() != "cmd/test" {
		t.Errorf("SUT = %q", sc.SUT())
	}
	if sc.Given == nil {
		t.Error("Given step block missing")
	}
	if sc.Then == nil {
		t.Error("Then step block missing")
	}
}

func TestLexerSpanishKeywordTranslation(t *testing.T) {
	kw := DefaultKeywords()
	kw.ToMachine["característica"] = "feature"

	lex := NewLexer("característica", kw)
	tok, err := lex.Next()
	if err != nil {
		t.Fatalf("Lex: %v", err)
	}
	if tok.Type != TokenFeature {
		t.Errorf("type = %v, want TokenFeature", tok.Type)
	}
	if tok.Value != "característica" {
		t.Errorf("value = %q, want característica (preserves human keyword)", tok.Value)
	}
}

// --- Nested contracts (contract-of-contracts) ---

func TestParseNestedContracts(t *testing.T) {
	src := `
contract "CON-IPC-PARENT" {
  title = "Interplanetary IPC Framework"
  status = "active"
  depends_on = []

  contract "CON-IPC-TERRA" {
    title = "Terra-side message queue"
    status = "active"
    depends_on = []

    feature "Terra MQ" {
      scenario "message delivery" {
        given {
          a Terra process sends a message
        }
        when {
          the local MQ processes it
        }
        then {
          the message is queued for relay
        }
      }
    }
  }

  contract "CON-IPC-MARS" {
    title = "Mars-side message queue"
    status = "active"
    depends_on = ["CON-IPC-TERRA"]

    feature "Mars MQ" {
      scenario "message receipt" {
        given {
          a relayed message arrives at Mars
        }
        when {
          the Mars MQ dequeues it
        }
        then {
          the destination process receives it
        }
      }
    }
  }

  contract "CON-IPC-RELAY" {
    title = "Interplanetary relay protocol"
    status = "active"
    depends_on = ["CON-IPC-TERRA", "CON-IPC-MARS"]

    feature "Relay" {
      scenario "light-delay tolerant transfer" {
        given {
          messages are queued on Terra
        }
        when {
          the relay window opens
        }
        then {
          messages arrive on Mars within 2x light-delay
        }
      }
    }
  }
}
`

	f, err := Parse(src, nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	ab := f.Artifact.(*ArtifactBlock)
	if ab.Kind != "contract" {
		t.Errorf("parent Kind = %q, want contract", ab.Kind)
	}
	if ab.Name != "CON-IPC-PARENT" {
		t.Errorf("parent Name = %q", ab.Name)
	}

	var subContracts []*Block
	for _, item := range ab.Items {
		if blk, ok := item.(*Block); ok && blk.Name == "contract" {
			subContracts = append(subContracts, blk)
		}
	}

	if len(subContracts) != 3 {
		t.Fatalf("expected 3 sub-contracts, got %d", len(subContracts))
	}

	if subContracts[0].Title != "CON-IPC-TERRA" {
		t.Errorf("sub[0] title = %q, want CON-IPC-TERRA", subContracts[0].Title)
	}
	if subContracts[1].Title != "CON-IPC-MARS" {
		t.Errorf("sub[1] title = %q, want CON-IPC-MARS", subContracts[1].Title)
	}
	if subContracts[2].Title != "CON-IPC-RELAY" {
		t.Errorf("sub[2] title = %q, want CON-IPC-RELAY", subContracts[2].Title)
	}

	// Verify dependency fields survive parsing
	for _, sc := range subContracts {
		found := false
		for _, item := range sc.Items {
			if fld, ok := item.(*Field); ok && fld.Key == "depends_on" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("sub-contract %q missing depends_on field", sc.Title)
		}
	}

	// Verify RELAY has non-linear (diamond) dependency: depends on both TERRA and MARS
	relay := subContracts[2]
	for _, item := range relay.Items {
		if fld, ok := item.(*Field); ok && fld.Key == "depends_on" {
			lv, ok := fld.Value.(*ListVal)
			if !ok {
				t.Fatalf("depends_on not a list")
			}
			if len(lv.Items) != 2 {
				t.Errorf("RELAY depends_on has %d items, want 2 (non-linear)", len(lv.Items))
			}
			break
		}
	}

	// Verify feature blocks inside nested contracts are accessible
	for _, sc := range subContracts {
		var hasFeature bool
		for _, item := range sc.Items {
			if _, ok := item.(*FeatureBlock); ok {
				hasFeature = true
				break
			}
		}
		if !hasFeature {
			t.Errorf("sub-contract %q has no feature block", sc.Title)
		}
	}
}

func TestParseNestedContractsMartianEnglish(t *testing.T) {
	kw := DefaultKeywords()
	for machine, human := range map[string]string{
		"contract": "charter",
		"feature":  "capability",
		"scenario": "protocol",
		"given":    "assuming",
		"when":     "upon",
		"then":     "verify",
	} {
		kw.ToMachine[human] = machine
		kw.ToHuman[machine] = human
	}

	src := `
charter "CON-DUNIX-CONSENSUS" {
  title = "Cross-planetary consensus protocol"
  status = "active"
  ordering = "non-linear"

  charter "CON-QUORUM-TERRA" {
    title = "Terra quorum module"
    status = "active"
    depends_on = []

    capability "Terra quorum" {
      protocol "quorum reached" {
        assuming {
          3 of 5 Terra nodes are online
        }
        upon {
          a proposal is submitted
        }
        verify {
          consensus is reached within 1 second
        }
      }
    }
  }

  charter "CON-QUORUM-MARS" {
    title = "Mars quorum module"
    status = "active"
    depends_on = []

    capability "Mars quorum" {
      protocol "quorum reached" {
        assuming {
          3 of 5 Mars nodes are online
        }
        upon {
          a proposal is submitted
        }
        verify {
          consensus is reached within 1 second
        }
      }
    }
  }

  charter "CON-SPLIT-BRAIN" {
    title = "Split-brain resolution"
    status = "active"
    depends_on = ["CON-QUORUM-TERRA", "CON-QUORUM-MARS"]
    ordering = "after-all"

    capability "Split-brain recovery" {
      protocol "divergent state reconciliation" {
        assuming {
          Terra and Mars made independent decisions during partition
        }
        upon {
          communication is restored
        }
        verify {
          states are merged using CvRDT semantics
          no committed transactions are lost
        }
      }
    }
  }
}
`

	f, err := Parse(src, kw)
	if err != nil {
		t.Fatalf("Parse with Martian keywords: %v", err)
	}

	ab := f.Artifact.(*ArtifactBlock)
	if ab.Kind != "contract" {
		t.Errorf("parent Kind = %q, want 'contract' (machine keyword)", ab.Kind)
	}

	var subCharters []*Block
	for _, item := range ab.Items {
		if blk, ok := item.(*Block); ok && blk.Name == "charter" {
			subCharters = append(subCharters, blk)
		}
	}

	if len(subCharters) != 3 {
		t.Fatalf("expected 3 nested charters, got %d", len(subCharters))
	}

	// The nested block Name preserves the human keyword ("charter"), not machine.
	// This is correct: nested blocks are generic Blocks, not ArtifactBlocks.
	// The Name is the raw identifier from source.
	if subCharters[0].Name != "charter" {
		t.Errorf("sub[0] Name = %q, want 'charter' (human keyword preserved in Block.Name)", subCharters[0].Name)
	}

	// Split-brain sub-charter has non-linear diamond dependency
	splitBrain := subCharters[2]
	if splitBrain.Title != "CON-SPLIT-BRAIN" {
		t.Errorf("sub[2] Title = %q", splitBrain.Title)
	}

	for _, item := range splitBrain.Items {
		if fld, ok := item.(*Field); ok && fld.Key == "depends_on" {
			lv := fld.Value.(*ListVal)
			if len(lv.Items) != 2 {
				t.Errorf("split-brain depends_on: got %d items, want 2", len(lv.Items))
			}
			break
		}
	}

	// Feature blocks inside nested charters parse as FeatureBlocks
	for _, sc := range subCharters {
		var features int
		for _, item := range sc.Items {
			if _, ok := item.(*FeatureBlock); ok {
				features++
			}
		}
		if features == 0 {
			t.Errorf("charter %q has no capability/feature blocks", sc.Title)
		}
	}
}

func TestParseLinearContractChain(t *testing.T) {
	src := `
contract "CON-PIPELINE" {
  title = "Sequential delivery pipeline"
  status = "active"
  ordering = "linear"

  contract "CON-STEP-1" {
    title = "Design phase"
    status = "complete"
    depends_on = []
    sequence = 1
  }

  contract "CON-STEP-2" {
    title = "Implementation phase"
    status = "active"
    depends_on = ["CON-STEP-1"]
    sequence = 2
  }

  contract "CON-STEP-3" {
    title = "Verification phase"
    status = "draft"
    depends_on = ["CON-STEP-2"]
    sequence = 3
  }

  contract "CON-STEP-4" {
    title = "Deployment phase"
    status = "draft"
    depends_on = ["CON-STEP-3"]
    sequence = 4
  }
}
`

	f, err := Parse(src, nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	ab := f.Artifact.(*ArtifactBlock)
	var steps []*Block
	for _, item := range ab.Items {
		if blk, ok := item.(*Block); ok && blk.Name == "contract" {
			steps = append(steps, blk)
		}
	}

	if len(steps) != 4 {
		t.Fatalf("expected 4 linear steps, got %d", len(steps))
	}

	// Verify linear chain: each step depends on the previous
	for i, step := range steps {
		for _, item := range step.Items {
			fld, ok := item.(*Field)
			if !ok || fld.Key != "depends_on" {
				continue
			}
			lv := fld.Value.(*ListVal)
			if i == 0 && len(lv.Items) != 0 {
				t.Errorf("step 0 should have no deps, got %d", len(lv.Items))
			}
			if i > 0 && len(lv.Items) != 1 {
				t.Errorf("step %d should have 1 dep, got %d", i, len(lv.Items))
			}
		}
	}
}

func TestParseRuleWithWhenBlock(t *testing.T) {
	src := `rule "policy-rule" {
  name = "Test Policy"
  type = "mechanical"
  enforcement = "warning"

  when {
    kind = "contract"
    status = "active"
  }

  harness {
    command = "echo ok"
  }
}
`
	f, err := Parse(src, nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	ab, ok := f.Artifact.(*ArtifactBlock)
	if !ok {
		t.Fatal("expected ArtifactBlock")
	}

	var whenBlocks []*Block
	var harnessBlocks []*Block
	for _, item := range ab.Items {
		if blk, ok := item.(*Block); ok {
			switch blk.Name {
			case "when":
				whenBlocks = append(whenBlocks, blk)
			case "harness":
				harnessBlocks = append(harnessBlocks, blk)
			}
		}
	}

	if len(whenBlocks) != 1 {
		t.Fatalf("expected 1 when block, got %d", len(whenBlocks))
	}
	if len(harnessBlocks) != 1 {
		t.Fatalf("expected 1 harness block, got %d", len(harnessBlocks))
	}

	when := whenBlocks[0]
	kind, ok := FieldString(when.Items, "kind")
	if !ok || kind != "contract" {
		t.Errorf("expected when.kind = 'contract', got %q (ok=%v)", kind, ok)
	}
	status, ok := FieldString(when.Items, "status")
	if !ok || status != "active" {
		t.Errorf("expected when.status = 'active', got %q (ok=%v)", status, ok)
	}

	formatted := Format(f, nil)
	f2, err := Parse(formatted, nil)
	if err != nil {
		t.Fatalf("re-parse after format: %v", err)
	}
	ab2 := f2.Artifact.(*ArtifactBlock)
	var whenCount int
	for _, item := range ab2.Items {
		if blk, ok := item.(*Block); ok && blk.Name == "when" {
			whenCount++
		}
	}
	if whenCount != 1 {
		t.Errorf("round-trip: expected 1 when block, got %d", whenCount)
	}
}
