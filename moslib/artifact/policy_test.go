package artifact

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMatchesArtifact_SingleField(t *testing.T) {
	p := Policy{
		Predicates: []Predicate{
			{Fields: map[string]string{"kind": "contract"}},
		},
	}

	if !MatchesArtifact(p, map[string]string{"kind": "contract", "status": "active"}) {
		t.Error("expected match for kind=contract")
	}

	if MatchesArtifact(p, map[string]string{"kind": "specification"}) {
		t.Error("expected no match for kind=specification")
	}
}

func TestMatchesArtifact_MultiFieldAND(t *testing.T) {
	p := Policy{
		Predicates: []Predicate{
			{Fields: map[string]string{"kind": "contract", "status": "active"}},
		},
	}

	if !MatchesArtifact(p, map[string]string{"kind": "contract", "status": "active", "title": "foo"}) {
		t.Error("expected match when all predicate fields match")
	}

	if MatchesArtifact(p, map[string]string{"kind": "contract", "status": "draft"}) {
		t.Error("expected no match when status differs")
	}
}

func TestMatchesArtifact_MultiPredicateOR(t *testing.T) {
	p := Policy{
		Predicates: []Predicate{
			{Fields: map[string]string{"kind": "contract"}},
			{Fields: map[string]string{"kind": "specification"}},
		},
	}

	if !MatchesArtifact(p, map[string]string{"kind": "contract"}) {
		t.Error("expected match via first predicate")
	}
	if !MatchesArtifact(p, map[string]string{"kind": "specification"}) {
		t.Error("expected match via second predicate")
	}
	if MatchesArtifact(p, map[string]string{"kind": "rule"}) {
		t.Error("expected no match for kind=rule")
	}
}

func TestMatchesArtifact_NoPredicates(t *testing.T) {
	p := Policy{Predicates: nil}
	if MatchesArtifact(p, map[string]string{"kind": "contract"}) {
		t.Error("expected no match when policy has no predicates")
	}
}

func TestMatchingPolicies(t *testing.T) {
	policies := []Policy{
		{RuleID: "p1", Predicates: []Predicate{{Fields: map[string]string{"kind": "contract"}}}},
		{RuleID: "p2", Predicates: []Predicate{{Fields: map[string]string{"kind": "specification"}}}},
		{RuleID: "p3", Predicates: []Predicate{{Fields: map[string]string{"kind": "contract", "status": "active"}}}},
	}

	matched := MatchingPolicies(policies, map[string]string{"kind": "contract", "status": "active"})
	if len(matched) != 2 {
		t.Fatalf("expected 2 matching policies, got %d", len(matched))
	}
	ids := map[string]bool{}
	for _, m := range matched {
		ids[m.RuleID] = true
	}
	if !ids["p1"] || !ids["p3"] {
		t.Errorf("expected p1 and p3 to match, got %v", ids)
	}
}

func TestLoadPolicies_WithAndWithoutWhen(t *testing.T) {
	root := t.TempDir()
	mosDir := filepath.Join(root, ".mos")
	mechDir := filepath.Join(mosDir, "rules", "mechanical")
	if err := os.MkdirAll(mechDir, 0755); err != nil {
		t.Fatal(err)
	}

	ruleWithWhen := `rule "has-when" {
  name = "With Predicate"
  type = "mechanical"
  scope = "project"
  enforcement = "warning"

  when {
    kind = "contract"
  }

  harness {
    command = "echo ok"
    timeout = "10s"
  }
}
`
	ruleNoWhen := `rule "no-when" {
  name = "No Predicate"
  type = "mechanical"
  scope = "project"
  enforcement = "error"

  harness {
    command = "echo ok"
    timeout = "10s"
  }
}
`
	if err := os.WriteFile(filepath.Join(mechDir, "has-when.mos"), []byte(ruleWithWhen), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mechDir, "no-when.mos"), []byte(ruleNoWhen), 0644); err != nil {
		t.Fatal(err)
	}

	policies, err := LoadPolicies(root)
	if err != nil {
		t.Fatal(err)
	}

	if len(policies) != 1 {
		t.Fatalf("expected 1 policy (only the one with when), got %d", len(policies))
	}
	if policies[0].RuleID != "has-when" {
		t.Errorf("expected policy rule ID 'has-when', got %q", policies[0].RuleID)
	}
	if policies[0].Enforcement != "warning" {
		t.Errorf("expected enforcement 'warning', got %q", policies[0].Enforcement)
	}
	if len(policies[0].Predicates) != 1 {
		t.Fatalf("expected 1 predicate, got %d", len(policies[0].Predicates))
	}
	if policies[0].Predicates[0].Fields["kind"] != "contract" {
		t.Errorf("expected predicate kind=contract, got %v", policies[0].Predicates[0].Fields)
	}
	if policies[0].Harness == nil || policies[0].Harness.Command != "echo ok" {
		t.Errorf("expected harness command 'echo ok', got %v", policies[0].Harness)
	}
}

func TestAddRuleWhen(t *testing.T) {
	root := t.TempDir()
	mosDir := filepath.Join(root, ".mos")
	mechDir := filepath.Join(mosDir, "rules", "mechanical")
	if err := os.MkdirAll(mechDir, 0755); err != nil {
		t.Fatal(err)
	}

	initial := `rule "test-rule" {
  name = "Test"
  type = "mechanical"
  scope = "project"
  enforcement = "error"
}
`
	if err := os.WriteFile(filepath.Join(mechDir, "test-rule.mos"), []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	err := AddRuleWhen(root, "test-rule", map[string]string{"kind": "contract", "status": "active"})
	if err != nil {
		t.Fatal(err)
	}

	policies, err := LoadPolicies(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(policies) != 1 {
		t.Fatalf("expected 1 policy after adding when, got %d", len(policies))
	}
	if len(policies[0].Predicates) != 1 {
		t.Fatalf("expected 1 predicate, got %d", len(policies[0].Predicates))
	}
	pred := policies[0].Predicates[0]
	if pred.Fields["kind"] != "contract" || pred.Fields["status"] != "active" {
		t.Errorf("unexpected predicate fields: %v", pred.Fields)
	}
}

func TestAddRuleWhen_MultipleWhenBlocks(t *testing.T) {
	root := t.TempDir()
	mosDir := filepath.Join(root, ".mos")
	mechDir := filepath.Join(mosDir, "rules", "mechanical")
	if err := os.MkdirAll(mechDir, 0755); err != nil {
		t.Fatal(err)
	}

	initial := `rule "multi-when" {
  name = "Multi"
  type = "mechanical"
  scope = "project"
  enforcement = "info"
}
`
	if err := os.WriteFile(filepath.Join(mechDir, "multi-when.mos"), []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	if err := AddRuleWhen(root, "multi-when", map[string]string{"kind": "contract"}); err != nil {
		t.Fatal(err)
	}
	if err := AddRuleWhen(root, "multi-when", map[string]string{"kind": "specification"}); err != nil {
		t.Fatal(err)
	}

	policies, err := LoadPolicies(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(policies) != 1 {
		t.Fatalf("expected 1 policy, got %d", len(policies))
	}
	if len(policies[0].Predicates) != 2 {
		t.Fatalf("expected 2 predicates (OR), got %d", len(policies[0].Predicates))
	}

	if !MatchesArtifact(policies[0], map[string]string{"kind": "contract"}) {
		t.Error("expected match via first when block")
	}
	if !MatchesArtifact(policies[0], map[string]string{"kind": "specification"}) {
		t.Error("expected match via second when block")
	}
	if MatchesArtifact(policies[0], map[string]string{"kind": "rule"}) {
		t.Error("expected no match for kind=rule")
	}
}
