package artifact

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplyArtifactContractCreate(t *testing.T) {
	root := setupScaffold(t)

	content := []byte(`contract "CON-APPLY-001" {
  title = "Applied Contract"
  status = "draft"
  goal = "Test apply create"
}
`)
	path, err := ApplyArtifact(root, content)
	if err != nil {
		t.Fatalf("ApplyArtifact create failed: %v", err)
	}

	expected := filepath.Join(root, ".mos", "contracts", "active", "CON-APPLY-001", "contract.mos")
	if path != expected {
		t.Errorf("path = %s, want %s", path, expected)
	}
	assertParses(t, path)
	assertLintClean(t, root)
}

func TestApplyArtifactContractUpdate(t *testing.T) {
	root := setupScaffold(t)

	CreateContract(root, "CON-APPLY-UPD", ContractOpts{Title: "Original", Status: "draft", Goal: "Old"})

	content := []byte(`contract "CON-APPLY-UPD" {
  title = "Updated via Apply"
  status = "draft"
  goal = "New goal via apply"
}
`)
	path, err := ApplyArtifact(root, content)
	if err != nil {
		t.Fatalf("ApplyArtifact update failed: %v", err)
	}

	data, _ := os.ReadFile(path)
	s := string(data)
	if !strings.Contains(s, "Updated via Apply") {
		t.Error("expected updated title")
	}
	if !strings.Contains(s, "New goal via apply") {
		t.Error("expected updated goal")
	}
	assertLintClean(t, root)
}

func TestApplyArtifactContractStatusMove(t *testing.T) {
	root := setupScaffold(t)

	CreateContract(root, "CON-APPLY-MV", ContractOpts{Title: "Mover", Status: "draft"})

	activePath := filepath.Join(root, ".mos", "contracts", "active", "CON-APPLY-MV", "contract.mos")
	if _, err := os.Stat(activePath); err != nil {
		t.Fatalf("expected contract in active/: %v", err)
	}

	content := []byte(`contract "CON-APPLY-MV" {
  title = "Mover"
  status = "complete"
}
`)
	path, err := ApplyArtifact(root, content)
	if err != nil {
		t.Fatalf("ApplyArtifact status move failed: %v", err)
	}

	archivePath := filepath.Join(root, ".mos", "contracts", "archive", "CON-APPLY-MV", "contract.mos")
	if path != archivePath {
		t.Errorf("expected archive path, got %s", path)
	}
	if _, err := os.Stat(activePath); !os.IsNotExist(err) {
		t.Error("expected old active/ directory removed")
	}
	assertLintClean(t, root)
}

func TestApplyArtifactRuleCreate(t *testing.T) {
	root := setupScaffold(t)

	content := []byte(`rule "applied-rule" {
  name = "Applied Rule"
  type = "mechanical"
  scope = "project"
  enforcement = "error"

  harness {
    command = "echo ok"
  }
}
`)
	path, err := ApplyArtifact(root, content)
	if err != nil {
		t.Fatalf("ApplyArtifact rule create failed: %v", err)
	}

	expected := filepath.Join(root, ".mos", "rules", "mechanical", "applied-rule.mos")
	if path != expected {
		t.Errorf("path = %s, want %s", path, expected)
	}
	assertParses(t, path)
	assertLintClean(t, root)
}

func TestApplyArtifactRuleTypeMove(t *testing.T) {
	root := setupScaffold(t)

	CreateRule(root, "type-mover", RuleOpts{Name: "Type Mover", Type: "mechanical", Enforcement: "error", HarnessCmd: "echo ok"})

	mechPath := filepath.Join(root, ".mos", "rules", "mechanical", "type-mover.mos")
	if _, err := os.Stat(mechPath); err != nil {
		t.Fatalf("expected rule in mechanical/: %v", err)
	}

	content := []byte(`rule "type-mover" {
  name = "Type Mover"
  type = "interpretive"
  scope = "project"
  enforcement = "warning"

  harness {
    command = "echo ok"
  }
}
`)
	path, err := ApplyArtifact(root, content)
	if err != nil {
		t.Fatalf("ApplyArtifact rule type move failed: %v", err)
	}

	interpPath := filepath.Join(root, ".mos", "rules", "interpretive", "type-mover.mos")
	if path != interpPath {
		t.Errorf("expected interpretive path, got %s", path)
	}
	if _, err := os.Stat(mechPath); !os.IsNotExist(err) {
		t.Error("expected old mechanical/ file removed")
	}
	assertLintClean(t, root)
}

func TestApplyArtifactLexicon(t *testing.T) {
	root := setupScaffold(t)

	content := []byte(`lexicon "default" {
  terms {
    pillar = "A test category"
    component = "A product component"
  }
}
`)
	path, err := ApplyArtifact(root, content)
	if err != nil {
		t.Fatalf("ApplyArtifact lexicon failed: %v", err)
	}

	expected := filepath.Join(root, ".mos", "lexicon", "default.mos")
	if path != expected {
		t.Errorf("path = %s, want %s", path, expected)
	}

	terms, err := ListTerms(root)
	if err != nil {
		t.Fatalf("ListTerms after apply: %v", err)
	}
	if len(terms) != 2 {
		t.Errorf("expected 2 terms, got %d", len(terms))
	}
}

func TestApplyArtifactInvalidDSL(t *testing.T) {
	root := setupScaffold(t)

	content := []byte(`this is not valid DSL at all {{{`)
	_, err := ApplyArtifact(root, content)
	if err == nil {
		t.Fatal("expected error for invalid DSL, got nil")
	}
}

func TestApplyArtifactUnknownKind(t *testing.T) {
	root := setupScaffold(t)

	content := []byte(`widget "foo" {
  name = "bar"
}
`)
	_, err := ApplyArtifact(root, content)
	if err == nil {
		t.Fatal("expected error for unknown kind, got nil")
	}
	if !strings.Contains(err.Error(), "unknown artifact kind") {
		t.Errorf("expected 'unknown artifact kind' error, got: %v", err)
	}
}

func TestEditArtifact(t *testing.T) {
	root := setupScaffold(t)

	CreateContract(root, "CON-EDIT", ContractOpts{Title: "Before Edit", Status: "draft", Goal: "Original"})

	t.Setenv("EDITOR", "sed -i s/Before\\ Edit/After\\ Edit/")

	if err := EditArtifact(root, "contract", "CON-EDIT"); err != nil {
		t.Fatalf("EditArtifact failed: %v", err)
	}

	content, err := ShowContract(root, "CON-EDIT")
	if err != nil {
		t.Fatalf("ShowContract after edit: %v", err)
	}
	if !strings.Contains(content, "After Edit") {
		t.Error("expected edited title in contract")
	}
	if !strings.Contains(content, "Original") {
		t.Error("expected goal preserved after edit")
	}
}

func TestEditArtifactNoChange(t *testing.T) {
	root := setupScaffold(t)

	CreateContract(root, "CON-NOEDIT", ContractOpts{Title: "Unchanged", Status: "draft"})

	t.Setenv("EDITOR", "true")

	if err := EditArtifact(root, "contract", "CON-NOEDIT"); err != nil {
		t.Fatalf("EditArtifact no-change failed: %v", err)
	}

	content, _ := ShowContract(root, "CON-NOEDIT")
	if !strings.Contains(content, "Unchanged") {
		t.Error("expected title unchanged")
	}
}

func TestEditArtifactLexicon(t *testing.T) {
	root := setupScaffold(t)

	AddTerm(root, "pillar", "A test category")

	t.Setenv("EDITOR", "true")

	if err := EditArtifact(root, "lexicon", ""); err != nil {
		t.Fatalf("EditArtifact lexicon failed: %v", err)
	}

	terms, _ := ListTerms(root)
	if len(terms) != 1 {
		t.Errorf("expected 1 term after no-op edit, got %d", len(terms))
	}
}

// --- gap 1: dangling dependency validation ---

func TestApplyArtifactPreservesCreatedAt(t *testing.T) {
	root := setupScaffold(t)
	CreateContract(root, "TS-E", ContractOpts{Title: "E", Status: "draft"})

	ePath1, _ := FindContractPath(root, "TS-E")
	info1, _ := readContractInfo("TS-E", ePath1)

	content := fmt.Sprintf(`contract "TS-E" {
  title = "Updated E"
  status = "draft"
  created_at = %s
  updated_at = %s
}
`, info1.CreatedAt, info1.CreatedAt)
	ApplyArtifact(root, []byte(content))

	ePath2, _ := FindContractPath(root, "TS-E")
	info2, _ := readContractInfo("TS-E", ePath2)
	if info2.CreatedAt != info1.CreatedAt {
		t.Errorf("created_at should be preserved: was %s, now %s", info1.CreatedAt, info2.CreatedAt)
	}
}

// --- primitive 12: checkpoints / progress ---
