package artifact

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCON088_GenericAddCoverage(t *testing.T) {
	root := setupScaffold(t)
	td := ArtifactTypeDef{Kind: "contract", Directory: "contracts"}
	GenericCreate(root, td, "CON-COV-001", map[string]string{
		"title": "Coverage test", "status": "draft", "goal": "test",
	})

	fields := map[string]string{"unit": "yes", "integration": "no", "e2e": "yes"}
	if err := GenericAddCoverage(root, td, "CON-COV-001", fields); err != nil {
		t.Fatalf("GenericAddCoverage: %v", err)
	}

	path := filepath.Join(root, ".mos", "contracts", ActiveDir, "CON-COV-001", "contract.mos")
	data, _ := os.ReadFile(path)
	content := string(data)
	if !strings.Contains(content, "coverage") {
		t.Error("expected coverage block")
	}
	if !strings.Contains(content, `unit = "yes"`) {
		t.Error("expected unit = yes")
	}
	if !strings.Contains(content, `e2e = "yes"`) {
		t.Error("expected e2e = yes")
	}

	fields2 := map[string]string{"unit": "no"}
	if err := GenericAddCoverage(root, td, "CON-COV-001", fields2); err != nil {
		t.Fatalf("GenericAddCoverage replace: %v", err)
	}
	data2, _ := os.ReadFile(path)
	if strings.Contains(string(data2), `unit = "yes"`) {
		t.Error("expected coverage to be replaced, not appended")
	}
}

func TestCON088_GenericAddBill(t *testing.T) {
	root := setupScaffold(t)
	td := ArtifactTypeDef{Kind: "contract", Directory: "contracts"}
	GenericCreate(root, td, "CON-BILL-001", map[string]string{
		"title": "Bill test", "status": "draft", "goal": "test",
	})

	if err := GenericAddBill(root, td, "CON-BILL-001", "alice", "Amend rule X"); err != nil {
		t.Fatalf("GenericAddBill: %v", err)
	}

	path := filepath.Join(root, ".mos", "contracts", ActiveDir, "CON-BILL-001", "contract.mos")
	data, _ := os.ReadFile(path)
	content := string(data)
	if !strings.Contains(content, "bill") {
		t.Error("expected bill block")
	}
	if !strings.Contains(content, `introduced_by = "alice"`) {
		t.Error("expected introduced_by = alice")
	}
	if !strings.Contains(content, `intent = "Amend rule X"`) {
		t.Error("expected intent field")
	}

	if err := GenericAddBill(root, td, "CON-BILL-001", "bob", "Override"); err != nil {
		t.Fatalf("GenericAddBill replace: %v", err)
	}
	data2, _ := os.ReadFile(path)
	if strings.Contains(string(data2), "alice") {
		t.Error("expected bill to be replaced, not appended")
	}
}

func TestCON088_GenericAddSpec(t *testing.T) {
	root := setupScaffold(t)
	td := ArtifactTypeDef{Kind: "contract", Directory: "contracts"}
	GenericCreate(root, td, "CON-SPEC-001", map[string]string{
		"title": "Spec test", "status": "draft", "goal": "test",
	})

	if err := GenericAddSpec(root, td, "CON-SPEC-001", []string{"api.yaml", "models.yaml"}); err != nil {
		t.Fatalf("GenericAddSpec: %v", err)
	}

	path := filepath.Join(root, ".mos", "contracts", ActiveDir, "CON-SPEC-001", "contract.mos")
	data, _ := os.ReadFile(path)
	content := string(data)
	if !strings.Contains(content, "spec") {
		t.Error("expected spec block")
	}
	if !strings.Contains(content, `include "api.yaml"`) {
		t.Error("expected include api.yaml")
	}
	if !strings.Contains(content, `include "models.yaml"`) {
		t.Error("expected include models.yaml")
	}

	if err := GenericAddSpec(root, td, "CON-SPEC-001", []string{"replaced.yaml"}); err != nil {
		t.Fatalf("GenericAddSpec replace: %v", err)
	}
	data2, _ := os.ReadFile(path)
	if strings.Contains(string(data2), "api.yaml") {
		t.Error("expected spec to be replaced, not appended")
	}
}

func TestCON088_GenericRemoveScenario(t *testing.T) {
	root := setupScaffold(t)
	td := ArtifactTypeDef{Kind: "contract", Directory: "contracts"}
	GenericCreate(root, td, "CON-RS-001", map[string]string{
		"title": "Remove scenario test", "status": "draft", "goal": "test",
	})

	GenericAddFeature(root, td, "CON-RS-001", "Feature A", "")
	GenericAddScenario(root, td, "CON-RS-001", "Feature A", "Keep this", "given", "when", "then")
	GenericAddScenario(root, td, "CON-RS-001", "Feature A", "Remove this", "g2", "w2", "t2")

	path := filepath.Join(root, ".mos", "contracts", ActiveDir, "CON-RS-001", "contract.mos")
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "Remove this") {
		t.Fatal("expected both scenarios before removal")
	}

	if err := GenericRemoveScenario(root, td, "CON-RS-001", "Feature A", "Remove this"); err != nil {
		t.Fatalf("GenericRemoveScenario: %v", err)
	}

	data2, _ := os.ReadFile(path)
	content := string(data2)
	if strings.Contains(content, "Remove this") {
		t.Error("expected scenario to be removed")
	}
	if !strings.Contains(content, "Keep this") {
		t.Error("expected other scenario to remain")
	}

	err := GenericRemoveScenario(root, td, "CON-RS-001", "Feature A", "Nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent scenario")
	}

	err = GenericRemoveScenario(root, td, "CON-RS-001", "Nonexistent Feature", "Keep this")
	if err == nil {
		t.Error("expected error for nonexistent feature")
	}
}

// --- CON-2026-125: Declarative link fields ---
