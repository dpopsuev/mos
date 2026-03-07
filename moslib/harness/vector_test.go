package harness

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFilterByVector(t *testing.T) {
	specs := []HarnessSpec{
		{RuleID: "a", Vector: "functional"},
		{RuleID: "b", Vector: "structural"},
		{RuleID: "c", Vector: "performance"},
		{RuleID: "d", Vector: ""},
		{RuleID: "e", Vector: "functional"},
	}

	functional := FilterByVector(specs, "functional")
	if len(functional) != 2 {
		t.Errorf("expected 2 functional specs, got %d", len(functional))
	}

	structural := FilterByVector(specs, "structural")
	if len(structural) != 1 {
		t.Errorf("expected 1 structural spec, got %d", len(structural))
	}

	untagged := FilterByVector(specs, "")
	if len(untagged) != 1 {
		t.Errorf("expected 1 untagged spec, got %d", len(untagged))
	}
}

func TestVectorResultScoring(t *testing.T) {
	root := t.TempDir()
	mosDir := filepath.Join(root, ".mos")
	rulesDir := filepath.Join(mosDir, "rules", "mechanical")
	if err := os.MkdirAll(rulesDir, 0o755); err != nil {
		t.Fatal(err)
	}

	ruleContent := `rule "vec-test-pass" {
  name = "Always Pass"
  type = "mechanical"
  vector = "functional"

  harness {
    command = "true"
    timeout = "10s"
  }
}
`
	if err := os.WriteFile(filepath.Join(rulesDir, "vec-test.mos"), []byte(ruleContent), 0o644); err != nil {
		t.Fatal(err)
	}

	results, err := EvaluateVectors(root, mosDir)
	if err != nil {
		t.Fatalf("EvaluateVectors: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 vector results, got %d", len(results))
	}

	var functional *VectorResult
	for i := range results {
		if results[i].Kind == VectorFunctional {
			functional = &results[i]
			break
		}
	}
	if functional == nil {
		t.Fatal("no functional vector result")
	}
	if !functional.Pass {
		t.Error("expected functional vector to pass")
	}
	if functional.Score != 100 {
		t.Errorf("expected score 100, got %.1f", functional.Score)
	}
	if len(functional.Details) != 1 {
		t.Errorf("expected 1 detail, got %d", len(functional.Details))
	}
}

func TestFormatVectorsText(t *testing.T) {
	results := []VectorResult{
		{Kind: "functional", Score: 100, Pass: true, Details: []VectorDetail{{RuleID: "a", Pass: true, Score: 100}}},
		{Kind: "structural", Score: 50, Pass: false, Details: []VectorDetail{{RuleID: "b", Pass: false}}},
		{Kind: "performance", Score: 0, Pass: true},
	}
	text := FormatVectorsText(results)
	if text == "" {
		t.Error("expected non-empty text")
	}
	if !contains(text, "functional") || !contains(text, "structural") || !contains(text, "performance") {
		t.Error("expected all three vector names in output")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
