package artifact

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateContract(t *testing.T) {
	root := setupScaffold(t)

	contractPath, err := CreateContract(root, "CON-2026-042", ContractOpts{
		Title:  "Test Contract",
		Status: "draft",
		Goal:   "Test goal",
	})
	if err != nil {
		t.Fatalf("CreateContract failed: %v", err)
	}

	expected := filepath.Join(root, ".mos", "contracts", "active", "CON-2026-042", "contract.mos")
	if contractPath != expected {
		t.Errorf("contract path = %s, want %s", contractPath, expected)
	}
	assertParses(t, contractPath)
	assertLintClean(t, root)
}

func TestCreateContractWithDeps(t *testing.T) {
	root := setupScaffold(t)

	CreateContract(root, "CON-A", ContractOpts{Title: "A", Status: "active"})

	contractPath, err := CreateContract(root, "CON-B", ContractOpts{
		Title:     "B depends on A",
		Status:    "active",
		Goal:      "Depends on CON-A",
		DependsOn: []string{"CON-A"},
	})
	if err != nil {
		t.Fatalf("CreateContract with deps failed: %v", err)
	}

	data, _ := os.ReadFile(contractPath)
	content := string(data)
	if !strings.Contains(content, "depends_on") {
		t.Error("expected depends_on in contract content")
	}
	if !strings.Contains(content, "CON-A") {
		t.Error("expected CON-A reference in depends_on")
	}
}

func TestCreateContractArchive(t *testing.T) {
	root := setupScaffold(t)

	contractPath, err := CreateContract(root, "CON-OLD", ContractOpts{
		Title:  "Old Contract",
		Status: "complete",
	})
	if err != nil {
		t.Fatalf("CreateContract (archive) failed: %v", err)
	}

	if !strings.Contains(contractPath, filepath.Join("contracts", "archive")) {
		t.Errorf("expected archive path, got %s", contractPath)
	}
}

func TestContractList(t *testing.T) {
	root := setupScaffold(t)

	CreateContract(root, "CON-A", ContractOpts{Title: "Alpha", Status: "active"})
	CreateContract(root, "CON-B", ContractOpts{Title: "Beta", Status: "draft"})
	CreateContract(root, "CON-C", ContractOpts{Title: "Gamma", Status: "complete"})

	all, err := ListContracts(root, ListOpts{})
	if err != nil {
		t.Fatalf("ListContracts failed: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3 contracts, got %d", len(all))
	}

	active, err := ListContracts(root, ListOpts{Status: "active"})
	if err != nil {
		t.Fatalf("ListContracts(active) failed: %v", err)
	}
	if len(active) != 1 {
		t.Fatalf("expected 1 active contract, got %d", len(active))
	}
	if active[0].ID != "CON-A" {
		t.Errorf("expected CON-A, got %s", active[0].ID)
	}
}

func TestContractShow(t *testing.T) {
	root := setupScaffold(t)

	CreateContract(root, "CON-SHOW", ContractOpts{Title: "Showable", Status: "draft", Goal: "Test show"})

	content, err := ShowContract(root, "CON-SHOW")
	if err != nil {
		t.Fatalf("ShowContract failed: %v", err)
	}
	if !strings.Contains(content, "Showable") {
		t.Error("expected title in show output")
	}
	if !strings.Contains(content, "Test show") {
		t.Error("expected goal in show output")
	}
}

func TestContractShowNotFound(t *testing.T) {
	root := setupScaffold(t)

	_, err := ShowContract(root, "NONEXISTENT")
	if err == nil {
		t.Fatal("expected error for nonexistent contract, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestContractStatusUpdate(t *testing.T) {
	root := setupScaffold(t)

	CreateContract(root, "CON-STATUS", ContractOpts{Title: "Status Test", Status: "draft"})

	activePath := filepath.Join(root, ".mos", "contracts", "active", "CON-STATUS", "contract.mos")
	if _, err := os.Stat(activePath); err != nil {
		t.Fatalf("expected contract in active/, got error: %v", err)
	}

	if err := UpdateContractStatus(root, "CON-STATUS", "complete"); err != nil {
		t.Fatalf("UpdateContractStatus failed: %v", err)
	}

	archivePath := filepath.Join(root, ".mos", "contracts", "archive", "CON-STATUS", "contract.mos")
	if _, err := os.Stat(archivePath); err != nil {
		t.Fatalf("expected contract in archive/ after status update, got error: %v", err)
	}

	if _, err := os.Stat(activePath); !os.IsNotExist(err) {
		t.Error("expected contract to be removed from active/")
	}

	data, _ := os.ReadFile(archivePath)
	content := string(data)
	if !strings.Contains(content, `"complete"`) {
		t.Error("expected status 'complete' in updated contract")
	}

	assertLintClean(t, root)
}

func TestContractStatusSameDir(t *testing.T) {
	root := setupScaffold(t)

	CreateContract(root, "CON-SAME", ContractOpts{Title: "Same Dir", Status: "draft"})

	if err := UpdateContractStatus(root, "CON-SAME", "active"); err != nil {
		t.Fatalf("UpdateContractStatus (same dir) failed: %v", err)
	}

	activePath := filepath.Join(root, ".mos", "contracts", "active", "CON-SAME", "contract.mos")
	if _, err := os.Stat(activePath); err != nil {
		t.Fatalf("expected contract to remain in active/: %v", err)
	}

	data, _ := os.ReadFile(activePath)
	if !strings.Contains(string(data), `"active"`) {
		t.Error("expected status 'active' in updated contract")
	}
}

func TestCreateContractWithSpecFile(t *testing.T) {
	root := setupScaffold(t)

	specContent := `feature "Widget lifecycle" {
  scenario "Create widget" {
    given {
      an empty database
    }
    when {
      the user creates a widget named "foo"
    }
    then {
      the widget exists in the database
    }
  }
}
`
	specPath := filepath.Join(root, "spec.mos")
	if err := os.WriteFile(specPath, []byte(specContent), 0644); err != nil {
		t.Fatalf("writing spec file: %v", err)
	}

	contractPath, err := CreateContract(root, "CON-SPEC", ContractOpts{
		Title:    "Spec Contract",
		Status:   "draft",
		Goal:     "Test spec file composition",
		SpecFile: specPath,
	})
	if err != nil {
		t.Fatalf("CreateContract with spec file failed: %v", err)
	}

	assertParses(t, contractPath)

	data, _ := os.ReadFile(contractPath)
	content := string(data)
	if !strings.Contains(content, "Widget lifecycle") {
		t.Error("expected feature title in contract content")
	}
	if !strings.Contains(content, "Create widget") {
		t.Error("expected scenario title in contract content")
	}
	if !strings.Contains(content, "an empty database") {
		t.Error("expected given step content in contract")
	}
	assertLintClean(t, root)
}

func TestCreateContractWithCoverageFile(t *testing.T) {
	root := setupScaffold(t)

	coverageContent := `coverage {
  unit {
    applies = true
  }
  integration {
    applies = true
  }
  performance {
    applies = false
    rationale = "No SLA defined"
  }
}
`
	coveragePath := filepath.Join(root, "coverage.mos")
	if err := os.WriteFile(coveragePath, []byte(coverageContent), 0644); err != nil {
		t.Fatalf("writing coverage file: %v", err)
	}

	contractPath, err := CreateContract(root, "CON-COV", ContractOpts{
		Title:        "Coverage Contract",
		Status:       "draft",
		CoverageFile: coveragePath,
	})
	if err != nil {
		t.Fatalf("CreateContract with coverage file failed: %v", err)
	}

	assertParses(t, contractPath)

	data, _ := os.ReadFile(contractPath)
	content := string(data)
	if !strings.Contains(content, "coverage") {
		t.Error("expected coverage block in contract content")
	}
	if !strings.Contains(content, "unit") {
		t.Error("expected unit pillar in coverage block")
	}
	if !strings.Contains(content, "No SLA defined") {
		t.Error("expected rationale in coverage block")
	}
	assertLintClean(t, root)
}

func TestCreateContractWithSpecAndCoverage(t *testing.T) {
	root := setupScaffold(t)

	CreateContract(root, "CON-SPEC", ContractOpts{Title: "Spec dep", Status: "draft"})
	CreateContract(root, "CON-COV", ContractOpts{Title: "Cov dep", Status: "draft"})

	specContent := `feature "Full contract" {
  scenario "Everything works" {
    given {
      a valid system
    }
    when {
      all checks pass
    }
    then {
      success is reported
    }
  }
}
`
	coverageContent := `coverage {
  unit {
    applies = true
  }
}
`
	specPath := filepath.Join(root, "spec.mos")
	coveragePath := filepath.Join(root, "coverage.mos")
	os.WriteFile(specPath, []byte(specContent), 0644)
	os.WriteFile(coveragePath, []byte(coverageContent), 0644)

	contractPath, err := CreateContract(root, "CON-BOTH", ContractOpts{
		Title:        "Both Contract",
		Status:       "draft",
		Goal:         "Test both composition paths",
		DependsOn:    []string{"CON-SPEC", "CON-COV"},
		SpecFile:     specPath,
		CoverageFile: coveragePath,
	})
	if err != nil {
		t.Fatalf("CreateContract with both failed: %v", err)
	}

	assertParses(t, contractPath)

	data, _ := os.ReadFile(contractPath)
	content := string(data)
	if !strings.Contains(content, "Full contract") {
		t.Error("expected feature in contract content")
	}
	if !strings.Contains(content, "coverage") {
		t.Error("expected coverage in contract content")
	}
	if !strings.Contains(content, "depends_on") {
		t.Error("expected depends_on in contract content")
	}
	assertLintClean(t, root)
}

func TestFullPipeline(t *testing.T) {
	root := t.TempDir()
	writeGoMod(t, root, "github.com/test/pipeline")

	if err := Init(root, InitOpts{Model: "bdfl", Scope: "cabinet"}); err != nil {
		t.Fatalf("Init: %v", err)
	}

	if _, err := CreateRule(root, "build-pass", RuleOpts{
		Name:        "Build Must Pass",
		Type:        "mechanical",
		Enforcement: "error",
		HarnessCmd:  "go build ./...",
	}); err != nil {
		t.Fatalf("CreateRule 1: %v", err)
	}

	if _, err := CreateRule(root, "test-pass", RuleOpts{
		Name:        "Tests Must Pass",
		Type:        "mechanical",
		Enforcement: "error",
		HarnessCmd:  "go test ./...",
	}); err != nil {
		t.Fatalf("CreateRule 2: %v", err)
	}

	if _, err := CreateContract(root, "CON-001", ContractOpts{
		Title:  "Phase 1",
		Status: "active",
		Goal:   "Build the foundation",
	}); err != nil {
		t.Fatalf("CreateContract: %v", err)
	}

	assertLintClean(t, root)
}

// --- contract update tests ---

func TestContractUpdateTitle(t *testing.T) {
	root := setupScaffold(t)
	CreateContract(root, "CON-UPD", ContractOpts{Title: "Original", Status: "draft", Goal: "Original goal"})

	newTitle := "Updated Title"
	if err := UpdateContract(root, "CON-UPD", ContractUpdateOpts{Title: &newTitle}); err != nil {
		t.Fatalf("UpdateContract failed: %v", err)
	}

	content, _ := ShowContract(root, "CON-UPD")
	if !strings.Contains(content, "Updated Title") {
		t.Error("expected updated title in contract")
	}
	if !strings.Contains(content, "Original goal") {
		t.Error("expected original goal to be preserved")
	}
	assertLintClean(t, root)
}

func TestContractUpdateGoal(t *testing.T) {
	root := setupScaffold(t)
	CreateContract(root, "CON-UPD", ContractOpts{Title: "Test", Status: "draft", Goal: "Old goal"})

	newGoal := "New goal"
	if err := UpdateContract(root, "CON-UPD", ContractUpdateOpts{Goal: &newGoal}); err != nil {
		t.Fatalf("UpdateContract failed: %v", err)
	}

	content, _ := ShowContract(root, "CON-UPD")
	if !strings.Contains(content, "New goal") {
		t.Error("expected updated goal")
	}
}

func TestContractUpdateAddsGoal(t *testing.T) {
	root := setupScaffold(t)
	CreateContract(root, "CON-NOGOAL", ContractOpts{Title: "No Goal", Status: "draft"})

	newGoal := "Added goal"
	if err := UpdateContract(root, "CON-NOGOAL", ContractUpdateOpts{Goal: &newGoal}); err != nil {
		t.Fatalf("UpdateContract failed: %v", err)
	}

	content, _ := ShowContract(root, "CON-NOGOAL")
	if !strings.Contains(content, "Added goal") {
		t.Error("expected newly added goal in contract")
	}
}

func TestContractUpdateStatus(t *testing.T) {
	root := setupScaffold(t)
	CreateContract(root, "CON-MV", ContractOpts{Title: "Moving", Status: "draft"})

	activePath := filepath.Join(root, ".mos", "contracts", "active", "CON-MV", "contract.mos")
	if _, err := os.Stat(activePath); err != nil {
		t.Fatalf("expected contract in active/: %v", err)
	}

	newStatus := "complete"
	if err := UpdateContract(root, "CON-MV", ContractUpdateOpts{Status: &newStatus}); err != nil {
		t.Fatalf("UpdateContract status failed: %v", err)
	}

	archivePath := filepath.Join(root, ".mos", "contracts", "archive", "CON-MV", "contract.mos")
	if _, err := os.Stat(archivePath); err != nil {
		t.Fatalf("expected contract in archive/ after update: %v", err)
	}
	if _, err := os.Stat(activePath); !os.IsNotExist(err) {
		t.Error("expected contract removed from active/")
	}
	assertLintClean(t, root)
}

func TestContractUpdateInvalidStatus(t *testing.T) {
	root := setupScaffold(t)
	CreateContract(root, "CON-BAD", ContractOpts{Title: "Bad", Status: "draft"})

	badStatus := "invalid"
	err := UpdateContract(root, "CON-BAD", ContractUpdateOpts{Status: &badStatus})
	if err == nil {
		t.Fatal("expected error for invalid status, got nil")
	}
}

func TestContractUpdateNilFieldsNoOp(t *testing.T) {
	root := setupScaffold(t)
	CreateContract(root, "CON-NOOP", ContractOpts{Title: "Original", Status: "draft", Goal: "Stay"})

	if err := UpdateContract(root, "CON-NOOP", ContractUpdateOpts{}); err != nil {
		t.Fatalf("UpdateContract no-op failed: %v", err)
	}

	content, _ := ShowContract(root, "CON-NOOP")
	if !strings.Contains(content, "Original") {
		t.Error("expected title preserved after no-op update")
	}
	if !strings.Contains(content, "Stay") {
		t.Error("expected goal preserved after no-op update")
	}
}

func TestContractUpdateSpecFile(t *testing.T) {
	root := setupScaffold(t)

	specV1 := `feature "Version 1" {
  scenario "Old behavior" {
    given {
      a system exists
    }
    when {
      user acts on it
    }
    then {
      old result is produced
    }
  }
}
`
	specV2 := `feature "Version 2" {
  scenario "New behavior" {
    given {
      a system exists
    }
    when {
      user acts on it
    }
    then {
      new result is produced
    }
  }
}
`
	specPathV1 := filepath.Join(root, "spec-v1.mos")
	specPathV2 := filepath.Join(root, "spec-v2.mos")
	os.WriteFile(specPathV1, []byte(specV1), 0644)
	os.WriteFile(specPathV2, []byte(specV2), 0644)

	CreateContract(root, "CON-SPEC-UPD", ContractOpts{
		Title:    "Spec Update",
		Status:   "draft",
		SpecFile: specPathV1,
	})

	content, _ := ShowContract(root, "CON-SPEC-UPD")
	if !strings.Contains(content, "Version 1") {
		t.Fatal("expected v1 spec in initial contract")
	}

	specFileArg := specPathV2
	if err := UpdateContract(root, "CON-SPEC-UPD", ContractUpdateOpts{SpecFile: &specFileArg}); err != nil {
		t.Fatalf("UpdateContract spec-file failed: %v", err)
	}

	content, _ = ShowContract(root, "CON-SPEC-UPD")
	if strings.Contains(content, "Version 1") {
		t.Error("expected v1 spec to be replaced")
	}
	if !strings.Contains(content, "Version 2") {
		t.Error("expected v2 spec in updated contract")
	}
	assertLintClean(t, root)
}

func TestContractUpdateNotFound(t *testing.T) {
	root := setupScaffold(t)

	newTitle := "Ghost"
	err := UpdateContract(root, "NONEXISTENT", ContractUpdateOpts{Title: &newTitle})
	if err == nil {
		t.Fatal("expected error for nonexistent contract, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

// --- contract delete tests ---

func TestContractDelete(t *testing.T) {
	root := setupScaffold(t)
	CreateContract(root, "CON-DEL", ContractOpts{Title: "Deletable", Status: "draft"})

	if err := DeleteContract(root, "CON-DEL", false); err != nil {
		t.Fatalf("DeleteContract failed: %v", err)
	}

	contracts, _ := ListContracts(root, ListOpts{})
	for _, c := range contracts {
		if c.ID == "CON-DEL" {
			t.Error("expected CON-DEL to be deleted")
		}
	}

	contractDir := filepath.Join(root, ".mos", "contracts", "active", "CON-DEL")
	if _, err := os.Stat(contractDir); !os.IsNotExist(err) {
		t.Error("expected contract directory to be removed")
	}
}

func TestContractDeleteArchived(t *testing.T) {
	root := setupScaffold(t)
	CreateContract(root, "CON-DEL-ARC", ContractOpts{Title: "Archived", Status: "complete"})

	if err := DeleteContract(root, "CON-DEL-ARC", false); err != nil {
		t.Fatalf("DeleteContract (archive) failed: %v", err)
	}

	contractDir := filepath.Join(root, ".mos", "contracts", "archive", "CON-DEL-ARC")
	if _, err := os.Stat(contractDir); !os.IsNotExist(err) {
		t.Error("expected archived contract directory to be removed")
	}
}

func TestContractDeleteNotFound(t *testing.T) {
	root := setupScaffold(t)

	err := DeleteContract(root, "NONEXISTENT", false)
	if err == nil {
		t.Fatal("expected error deleting nonexistent contract, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

// --- rule update tests ---
