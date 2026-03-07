package artifact

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateRule(t *testing.T) {
	root := setupScaffold(t)

	rulePath, err := CreateRule(root, "go-build-pass", RuleOpts{
		Name:           "Go Build Must Pass",
		Type:           "mechanical",
		Enforcement:    "error",
		HarnessCmd:     "go build ./...",
		HarnessTimeout: "2m",
	})
	if err != nil {
		t.Fatalf("CreateRule failed: %v", err)
	}

	expected := filepath.Join(root, ".mos", "rules", "mechanical", "go-build-pass.mos")
	if rulePath != expected {
		t.Errorf("rule path = %s, want %s", rulePath, expected)
	}
	assertParses(t, rulePath)
	assertLintClean(t, root)
}

func TestCreateRuleInterpretive(t *testing.T) {
	root := setupScaffold(t)

	rulePath, err := CreateRule(root, "bdd-specs", RuleOpts{
		Name:        "BDD Specs Required",
		Type:        "interpretive",
		Enforcement: "warning",
	})
	if err != nil {
		t.Fatalf("CreateRule failed: %v", err)
	}

	expected := filepath.Join(root, ".mos", "rules", "interpretive", "bdd-specs.mos")
	if rulePath != expected {
		t.Errorf("rule path = %s, want %s", rulePath, expected)
	}
	assertParses(t, rulePath)
}

func TestCreateRuleWithHarnessRequires(t *testing.T) {
	root := setupScaffold(t)

	rulePath, err := CreateRule(root, "ptp-e2e", RuleOpts{
		Name:           "PTP End-to-End",
		Type:           "mechanical",
		Enforcement:    "error",
		HarnessCmd:     "ginkgo run ./tests/cnf/core/network/ptp/",
		HarnessTimeout: "30m",
		HarnessRequires: map[string]string{
			"platform":   "ocp",
			"go_version": ">=1.25",
		},
	})
	if err != nil {
		t.Fatalf("CreateRule with harness requires failed: %v", err)
	}

	assertParses(t, rulePath)

	data, _ := os.ReadFile(rulePath)
	content := string(data)
	if !strings.Contains(content, "requires") {
		t.Error("expected requires block in rule content")
	}
	if !strings.Contains(content, "platform") {
		t.Error("expected platform field in requires block")
	}
	if !strings.Contains(content, "go_version") {
		t.Error("expected go_version field in requires block")
	}
	assertLintClean(t, root)
}

func TestRuleList(t *testing.T) {
	root := setupScaffold(t)

	if _, err := CreateRule(root, "build-pass", RuleOpts{Name: "Build Pass", Type: "mechanical", Enforcement: "error", HarnessCmd: "go build ./..."}); err != nil {
		t.Fatalf("CreateRule build-pass: %v", err)
	}
	if _, err := CreateRule(root, "test-pass", RuleOpts{Name: "Test Pass", Type: "mechanical", Enforcement: "error", HarnessCmd: "go test ./..."}); err != nil {
		t.Fatalf("CreateRule test-pass: %v", err)
	}
	if _, err := CreateRule(root, "bdd-specs", RuleOpts{Name: "BDD Specs", Type: "interpretive", Enforcement: "warning", HarnessCmd: "echo ok"}); err != nil {
		t.Fatalf("CreateRule bdd-specs: %v", err)
	}

	all, err := ListRules(root, "")
	if err != nil {
		t.Fatalf("ListRules failed: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(all))
	}

	mechanical, err := ListRules(root, "mechanical")
	if err != nil {
		t.Fatalf("ListRules(mechanical) failed: %v", err)
	}
	if len(mechanical) != 2 {
		t.Fatalf("expected 2 mechanical rules, got %d", len(mechanical))
	}

	interpretive, err := ListRules(root, "interpretive")
	if err != nil {
		t.Fatalf("ListRules(interpretive) failed: %v", err)
	}
	if len(interpretive) != 1 {
		t.Fatalf("expected 1 interpretive rule, got %d", len(interpretive))
	}
}

func TestRuleShow(t *testing.T) {
	root := setupScaffold(t)

	CreateRule(root, "build-pass", RuleOpts{
		Name:        "Build Must Pass",
		Type:        "mechanical",
		Enforcement: "error",
		HarnessCmd:  "go build ./...",
	})

	content, err := ShowRule(root, "build-pass")
	if err != nil {
		t.Fatalf("ShowRule failed: %v", err)
	}
	if !strings.Contains(content, "Build Must Pass") {
		t.Error("expected rule name in show output")
	}
	if !strings.Contains(content, "go build") {
		t.Error("expected harness command in show output")
	}
}

func TestRuleShowNotFound(t *testing.T) {
	root := setupScaffold(t)

	_, err := ShowRule(root, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent rule, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestCreateRuleInvalidType(t *testing.T) {
	root := setupScaffold(t)

	_, err := CreateRule(root, "bad-rule", RuleOpts{Type: "invalid"})
	if err == nil {
		t.Fatal("expected error for invalid type, got nil")
	}
}

func TestRuleUpdateName(t *testing.T) {
	root := setupScaffold(t)
	CreateRule(root, "my-rule", RuleOpts{Name: "Original Name", Type: "mechanical", Enforcement: "error", HarnessCmd: "echo ok"})

	newName := "Updated Name"
	if err := UpdateRule(root, "my-rule", RuleUpdateOpts{Name: &newName}); err != nil {
		t.Fatalf("UpdateRule failed: %v", err)
	}

	content, _ := ShowRule(root, "my-rule")
	if !strings.Contains(content, "Updated Name") {
		t.Error("expected updated name in rule")
	}
	assertLintClean(t, root)
}

func TestRuleUpdateEnforcement(t *testing.T) {
	root := setupScaffold(t)
	CreateRule(root, "warn-rule", RuleOpts{Name: "Warn", Type: "mechanical", Enforcement: "error", HarnessCmd: "echo ok"})

	newEnf := "warning"
	if err := UpdateRule(root, "warn-rule", RuleUpdateOpts{Enforcement: &newEnf}); err != nil {
		t.Fatalf("UpdateRule failed: %v", err)
	}

	content, _ := ShowRule(root, "warn-rule")
	if !strings.Contains(content, "warning") {
		t.Error("expected enforcement updated to warning")
	}
}

func TestRuleUpdateTypeMovesFile(t *testing.T) {
	root := setupScaffold(t)
	CreateRule(root, "movable-rule", RuleOpts{Name: "Movable", Type: "mechanical", Enforcement: "error", HarnessCmd: "echo ok"})

	mechPath := filepath.Join(root, ".mos", "rules", "mechanical", "movable-rule.mos")
	if _, err := os.Stat(mechPath); err != nil {
		t.Fatalf("expected rule in mechanical/: %v", err)
	}

	newType := "interpretive"
	if err := UpdateRule(root, "movable-rule", RuleUpdateOpts{Type: &newType}); err != nil {
		t.Fatalf("UpdateRule type change failed: %v", err)
	}

	interpPath := filepath.Join(root, ".mos", "rules", "interpretive", "movable-rule.mos")
	if _, err := os.Stat(interpPath); err != nil {
		t.Fatalf("expected rule in interpretive/ after type change: %v", err)
	}
	if _, err := os.Stat(mechPath); !os.IsNotExist(err) {
		t.Error("expected rule removed from mechanical/")
	}

	content, _ := ShowRule(root, "movable-rule")
	if !strings.Contains(content, "interpretive") {
		t.Error("expected type field updated to interpretive")
	}
	assertLintClean(t, root)
}

func TestRuleUpdateInvalidType(t *testing.T) {
	root := setupScaffold(t)
	CreateRule(root, "bad-type", RuleOpts{Name: "Bad", Type: "mechanical", Enforcement: "error", HarnessCmd: "echo ok"})

	badType := "invalid"
	err := UpdateRule(root, "bad-type", RuleUpdateOpts{Type: &badType})
	if err == nil {
		t.Fatal("expected error for invalid type, got nil")
	}
}

func TestRuleUpdateHarnessRequires(t *testing.T) {
	root := setupScaffold(t)
	CreateRule(root, "req-rule", RuleOpts{
		Name:            "Requires",
		Type:            "mechanical",
		Enforcement:     "error",
		HarnessCmd:      "make test",
		HarnessRequires: map[string]string{"platform": "linux"},
	})

	newReqs := map[string]string{"platform": "ocp", "go_version": ">=1.25"}
	if err := UpdateRule(root, "req-rule", RuleUpdateOpts{HarnessRequires: newReqs}); err != nil {
		t.Fatalf("UpdateRule harness-requires failed: %v", err)
	}

	content, _ := ShowRule(root, "req-rule")
	if !strings.Contains(content, "ocp") {
		t.Error("expected updated platform in requires")
	}
	if !strings.Contains(content, "go_version") {
		t.Error("expected go_version added to requires")
	}
	if strings.Contains(content, `"linux"`) {
		t.Error("expected old platform value replaced")
	}
	assertLintClean(t, root)
}

func TestRuleUpdateNilFieldsNoOp(t *testing.T) {
	root := setupScaffold(t)
	CreateRule(root, "noop-rule", RuleOpts{Name: "Stays", Type: "mechanical", Enforcement: "error", HarnessCmd: "echo ok"})

	if err := UpdateRule(root, "noop-rule", RuleUpdateOpts{}); err != nil {
		t.Fatalf("UpdateRule no-op failed: %v", err)
	}

	content, _ := ShowRule(root, "noop-rule")
	if !strings.Contains(content, "Stays") {
		t.Error("expected name preserved after no-op update")
	}
}

func TestRuleUpdateNotFound(t *testing.T) {
	root := setupScaffold(t)

	newName := "Ghost"
	err := UpdateRule(root, "nonexistent", RuleUpdateOpts{Name: &newName})
	if err == nil {
		t.Fatal("expected error for nonexistent rule, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

// --- rule delete tests ---

func TestRuleDelete(t *testing.T) {
	root := setupScaffold(t)
	CreateRule(root, "del-rule", RuleOpts{Name: "Deletable", Type: "mechanical", Enforcement: "error", HarnessCmd: "echo ok"})

	if err := DeleteRule(root, "del-rule"); err != nil {
		t.Fatalf("DeleteRule failed: %v", err)
	}

	rulePath := filepath.Join(root, ".mos", "rules", "mechanical", "del-rule.mos")
	if _, err := os.Stat(rulePath); !os.IsNotExist(err) {
		t.Error("expected rule file to be removed")
	}

	rules, _ := ListRules(root, "")
	for _, r := range rules {
		if r.ID == "del-rule" {
			t.Error("expected del-rule to not appear in list")
		}
	}
}

func TestRuleDeleteInterpretive(t *testing.T) {
	root := setupScaffold(t)
	CreateRule(root, "del-interp", RuleOpts{Name: "Interp", Type: "interpretive", Enforcement: "warning", HarnessCmd: "echo ok"})

	if err := DeleteRule(root, "del-interp"); err != nil {
		t.Fatalf("DeleteRule (interpretive) failed: %v", err)
	}

	rulePath := filepath.Join(root, ".mos", "rules", "interpretive", "del-interp.mos")
	if _, err := os.Stat(rulePath); !os.IsNotExist(err) {
		t.Error("expected interpretive rule file to be removed")
	}
}

func TestRuleDeleteNotFound(t *testing.T) {
	root := setupScaffold(t)

	err := DeleteRule(root, "nonexistent")
	if err == nil {
		t.Fatal("expected error deleting nonexistent rule, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

// --- contract graph tests ---
