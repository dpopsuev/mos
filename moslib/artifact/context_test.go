package artifact

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCON027_CreateRuleWithAppliesTo(t *testing.T) {
	root := setupScaffold(t)

	_, err := CreateRule(root, "R-BUG", RuleOpts{
		Name:       "Bug workflow",
		Type:       "mechanical",
		AppliesTo:  []string{"bug", "feature"},
		HarnessCmd: "echo ok",
	})
	if err != nil {
		t.Fatalf("CreateRule: %v", err)
	}

	info, err := readRuleInfo(filepath.Join(root, ".mos", "rules", "mechanical", "R-BUG.mos"))
	if err != nil {
		t.Fatalf("readRuleInfo: %v", err)
	}
	if len(info.AppliesTo) != 2 || info.AppliesTo[0] != "bug" || info.AppliesTo[1] != "feature" {
		t.Errorf("expected applies_to [bug, feature], got %v", info.AppliesTo)
	}
}

func TestCON027_ContractContextResolvesMatchingRules(t *testing.T) {
	root := setupScaffold(t)

	CreateRule(root, "R-BUG", RuleOpts{
		Name:       "Bug workflow",
		Type:       "mechanical",
		AppliesTo:  []string{"bug"},
		HarnessCmd: "echo ok",
	})

	CreateContract(root, "BUG-TEST-001", ContractOpts{
		Title:  "Test bug",
		Status: "active",
		Kind:   "bug",
		Goal:   "Fix the thing",
	})

	ctx, err := ContractContext(root, "BUG-TEST-001")
	if err != nil {
		t.Fatalf("ContractContext: %v", err)
	}

	if ctx.Title != "Test bug" {
		t.Errorf("expected title 'Test bug', got %q", ctx.Title)
	}
	if ctx.Kind != "bug" {
		t.Errorf("expected kind 'bug', got %q", ctx.Kind)
	}
	if len(ctx.Rules) != 1 {
		t.Fatalf("expected 1 matching rule, got %d", len(ctx.Rules))
	}
	if ctx.Rules[0].ID != "R-BUG" {
		t.Errorf("expected rule R-BUG, got %q", ctx.Rules[0].ID)
	}
}

func TestCON027_RulesWithoutAppliesToExcluded(t *testing.T) {
	root := setupScaffold(t)

	CreateRule(root, "R-GENERAL", RuleOpts{
		Name:       "General rule",
		Type:       "mechanical",
		HarnessCmd: "echo ok",
	})

	CreateContract(root, "CON-TEST-001", ContractOpts{
		Title:  "Test feature",
		Status: "active",
		Kind:   "feature",
	})

	ctx, err := ContractContext(root, "CON-TEST-001")
	if err != nil {
		t.Fatalf("ContractContext: %v", err)
	}

	if len(ctx.Rules) != 0 {
		t.Errorf("expected 0 rules for feature contract with no matching rules, got %d", len(ctx.Rules))
	}
}

func TestCON027_MultipleMatchingRulesConcatenated(t *testing.T) {
	root := setupScaffold(t)

	CreateRule(root, "R-BUG-A", RuleOpts{
		Name:       "Bug rule A",
		Type:       "mechanical",
		AppliesTo:  []string{"bug"},
		HarnessCmd: "echo ok",
	})
	CreateRule(root, "R-BUG-B", RuleOpts{
		Name:       "Bug rule B",
		Type:       "interpretive",
		AppliesTo:  []string{"bug", "feature"},
		HarnessCmd: "echo ok",
	})

	CreateContract(root, "BUG-MULTI-001", ContractOpts{
		Title:  "Multi rule bug",
		Status: "active",
		Kind:   "bug",
	})

	ctx, err := ContractContext(root, "BUG-MULTI-001")
	if err != nil {
		t.Fatalf("ContractContext: %v", err)
	}

	if len(ctx.Rules) != 2 {
		t.Fatalf("expected 2 matching rules, got %d", len(ctx.Rules))
	}
}

func TestCON027_ContextIncludesContractMetadata(t *testing.T) {
	root := setupScaffold(t)

	CreateRule(root, "R-META", RuleOpts{
		Name:       "Meta rule",
		Type:       "mechanical",
		AppliesTo:  []string{"feature"},
		HarnessCmd: "echo ok",
	})

	CreateContract(root, "CON-META-001", ContractOpts{
		Title:  "Metadata test",
		Status: "active",
		Kind:   "feature",
		Goal:   "Test metadata in context output",
	})

	ctx, err := ContractContext(root, "CON-META-001")
	if err != nil {
		t.Fatalf("ContractContext: %v", err)
	}

	output := FormatContext(ctx)
	if !strings.Contains(output, "CON-META-001") {
		t.Error("output missing contract ID")
	}
	if !strings.Contains(output, "Metadata test") {
		t.Error("output missing title")
	}
	if !strings.Contains(output, "feature") {
		t.Error("output missing kind")
	}
	if !strings.Contains(output, "Test metadata in context output") {
		t.Error("output missing goal")
	}
	if !strings.Contains(output, "R-META") {
		t.Error("output missing matching rule")
	}
}

// --- CON-2026-028: Schema Migration ---
