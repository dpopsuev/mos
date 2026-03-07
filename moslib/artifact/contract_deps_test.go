package artifact

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestContractGraphSimple(t *testing.T) {
	root := setupScaffold(t)

	CreateContract(root, "CON-A", ContractOpts{Title: "A", Status: "draft"})
	CreateContract(root, "CON-B", ContractOpts{Title: "B", Status: "draft", DependsOn: []string{"CON-A"}})
	CreateContract(root, "CON-C", ContractOpts{Title: "C", Status: "draft", DependsOn: []string{"CON-A", "CON-B"}})

	contracts, err := ContractGraph(root)
	if err != nil {
		t.Fatalf("ContractGraph failed: %v", err)
	}

	byID := make(map[string]ContractInfo)
	for _, c := range contracts {
		byID[c.ID] = c
	}

	if len(byID["CON-A"].DependsOn) != 0 {
		t.Errorf("CON-A should have no deps, got %v", byID["CON-A"].DependsOn)
	}
	if len(byID["CON-B"].DependsOn) != 1 || byID["CON-B"].DependsOn[0] != "CON-A" {
		t.Errorf("CON-B deps = %v, want [CON-A]", byID["CON-B"].DependsOn)
	}
	if len(byID["CON-C"].DependsOn) != 2 {
		t.Errorf("CON-C deps = %v, want [CON-A, CON-B]", byID["CON-C"].DependsOn)
	}
}

func TestContractGraphCycleDetection(t *testing.T) {
	root := setupScaffold(t)

	CreateContract(root, "CYC-A", ContractOpts{Title: "A", Status: "draft"})
	CreateContract(root, "CYC-B", ContractOpts{Title: "B", Status: "draft"})

	if err := LinkContract(root, "CYC-B", "CYC-A"); err != nil {
		t.Fatalf("link B->A should succeed: %v", err)
	}

	err := LinkContract(root, "CYC-A", "CYC-B")
	if err == nil {
		t.Fatal("expected cycle error when linking CYC-A -> CYC-B")
	}
	if !strings.Contains(err.Error(), "cycle") {
		t.Errorf("expected 'cycle' in error, got: %v", err)
	}
}

func TestLinkContract(t *testing.T) {
	root := setupScaffold(t)

	CreateContract(root, "LINK-A", ContractOpts{Title: "A", Status: "draft"})
	CreateContract(root, "LINK-B", ContractOpts{Title: "B", Status: "draft"})

	if err := LinkContract(root, "LINK-B", "LINK-A"); err != nil {
		t.Fatalf("LinkContract failed: %v", err)
	}

	contracts, _ := ContractGraph(root)
	byID := make(map[string]ContractInfo)
	for _, c := range contracts {
		byID[c.ID] = c
	}
	if len(byID["LINK-B"].DependsOn) != 1 || byID["LINK-B"].DependsOn[0] != "LINK-A" {
		t.Errorf("LINK-B deps = %v, want [LINK-A]", byID["LINK-B"].DependsOn)
	}
}

func TestUnlinkContract(t *testing.T) {
	root := setupScaffold(t)

	CreateContract(root, "UNL-A", ContractOpts{Title: "A", Status: "draft"})
	CreateContract(root, "UNL-B", ContractOpts{Title: "B", Status: "draft", DependsOn: []string{"UNL-A"}})

	if err := UnlinkContract(root, "UNL-B", "UNL-A"); err != nil {
		t.Fatalf("UnlinkContract failed: %v", err)
	}

	contracts, _ := ContractGraph(root)
	byID := make(map[string]ContractInfo)
	for _, c := range contracts {
		byID[c.ID] = c
	}
	if len(byID["UNL-B"].DependsOn) != 0 {
		t.Errorf("UNL-B deps should be empty after unlink, got %v", byID["UNL-B"].DependsOn)
	}
}

func TestLinkContractIdempotent(t *testing.T) {
	root := setupScaffold(t)

	CreateContract(root, "IDEM-A", ContractOpts{Title: "A", Status: "draft"})
	CreateContract(root, "IDEM-B", ContractOpts{Title: "B", Status: "draft", DependsOn: []string{"IDEM-A"}})

	if err := LinkContract(root, "IDEM-B", "IDEM-A"); err != nil {
		t.Fatalf("LinkContract idempotent failed: %v", err)
	}

	contracts, _ := ContractGraph(root)
	byID := make(map[string]ContractInfo)
	for _, c := range contracts {
		byID[c.ID] = c
	}
	if len(byID["IDEM-B"].DependsOn) != 1 {
		t.Errorf("expected 1 dep after idempotent link, got %d", len(byID["IDEM-B"].DependsOn))
	}
}

func TestUnlinkContractNotPresent(t *testing.T) {
	root := setupScaffold(t)

	CreateContract(root, "NP-A", ContractOpts{Title: "A", Status: "draft"})

	if err := UnlinkContract(root, "NP-A", "NONEXISTENT"); err != nil {
		t.Fatalf("UnlinkContract not-present should be no-op, got: %v", err)
	}
}

// --- apply artifact tests ---

func TestLinkContractDanglingRef(t *testing.T) {
	root := setupScaffold(t)
	CreateContract(root, "GAP1-A", ContractOpts{Title: "A", Status: "draft"})

	err := LinkContract(root, "GAP1-A", "NONEXISTENT")
	if err == nil {
		t.Fatal("expected error when linking to nonexistent contract")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}

func TestCreateContractDanglingDep(t *testing.T) {
	root := setupScaffold(t)

	_, err := CreateContract(root, "GAP1-B", ContractOpts{
		Title:     "B",
		Status:    "draft",
		DependsOn: []string{"NOPE"},
	})
	if err == nil {
		t.Fatal("expected error when creating contract with nonexistent dependency")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}

func TestLintContractDanglingDep(t *testing.T) {
	root := setupScaffold(t)
	CreateContract(root, "LINT-A", ContractOpts{Title: "A", Status: "draft"})

	contractPath, _ := FindContractPath(root, "LINT-A")
	data, _ := os.ReadFile(contractPath)
	content := strings.Replace(string(data), "}", `
  scope {
    depends_on = ["DOES-NOT-EXIST"]
  }
}`, 1)
	os.WriteFile(contractPath, []byte(content), 0644)

	if LintAll == nil {
		t.Skip("linter not configured")
	}
	allDiags, err := LintAll(root)
	if err != nil {
		t.Fatalf("LintAll: %v", err)
	}

	found := false
	for _, d := range allDiags {
		if strings.Contains(d.Message, "depends_on") && strings.Contains(d.Message, "DOES-NOT-EXIST") {
			found = true
		}
	}
	if !found {
		t.Error("expected linter diagnostic for dangling depends_on reference")
	}
}

// --- gap 2: safe delete with dependents check ---

func TestDeleteContractWithDependents(t *testing.T) {
	root := setupScaffold(t)
	CreateContract(root, "DEP-A", ContractOpts{Title: "A", Status: "draft"})
	CreateContract(root, "DEP-B", ContractOpts{Title: "B", Status: "draft", DependsOn: []string{"DEP-A"}})

	err := DeleteContract(root, "DEP-A", false)
	if err == nil {
		t.Fatal("expected error deleting contract with dependents")
	}
	if !strings.Contains(err.Error(), "DEP-B") {
		t.Errorf("expected error to mention dependent 'DEP-B', got: %v", err)
	}
}

func TestDeleteContractWithDependentsForce(t *testing.T) {
	root := setupScaffold(t)
	CreateContract(root, "DEPF-A", ContractOpts{Title: "A", Status: "draft"})
	CreateContract(root, "DEPF-B", ContractOpts{Title: "B", Status: "draft", DependsOn: []string{"DEPF-A"}})

	err := DeleteContract(root, "DEPF-A", true)
	if err != nil {
		t.Fatalf("force delete should succeed: %v", err)
	}
	if _, err := FindContractPath(root, "DEPF-A"); err == nil {
		t.Error("expected contract to be deleted")
	}
}

func TestDeleteContractNoDependents(t *testing.T) {
	root := setupScaffold(t)
	CreateContract(root, "LEAF-A", ContractOpts{Title: "A", Status: "draft"})

	err := DeleteContract(root, "LEAF-A", false)
	if err != nil {
		t.Fatalf("deleting leaf contract should succeed: %v", err)
	}
}

// --- gap 4: list output shows depends_on ---

func TestContractListTextShowsDeps(t *testing.T) {
	root := setupScaffold(t)
	CreateContract(root, "LIST-A", ContractOpts{Title: "A", Status: "draft"})
	CreateContract(root, "LIST-B", ContractOpts{Title: "B", Status: "draft", DependsOn: []string{"LIST-A"}})

	contracts, err := ListContracts(root, ListOpts{})
	if err != nil {
		t.Fatalf("ListContracts: %v", err)
	}
	for _, c := range contracts {
		if c.ID == "LIST-B" && len(c.DependsOn) != 1 {
			t.Errorf("expected LIST-B to have 1 dep, got %d", len(c.DependsOn))
		}
	}
}

// --- gap 5: topological sort ---

func TestContractGraphTopologicalOrder(t *testing.T) {
	root := setupScaffold(t)
	CreateContract(root, "TOPO-A", ContractOpts{Title: "A", Status: "draft"})
	CreateContract(root, "TOPO-B", ContractOpts{Title: "B", Status: "draft", DependsOn: []string{"TOPO-A"}})
	CreateContract(root, "TOPO-C", ContractOpts{Title: "C", Status: "draft", DependsOn: []string{"TOPO-B"}})

	all, _ := ContractGraph(root)
	sorted, cycles := TopologicalSort(all)

	if len(cycles) > 0 {
		t.Errorf("unexpected cycles: %v", cycles)
	}
	if len(sorted) != 3 {
		t.Fatalf("expected 3, got %d", len(sorted))
	}

	idxA, idxB, idxC := -1, -1, -1
	for i, c := range sorted {
		switch c.ID {
		case "TOPO-A":
			idxA = i
		case "TOPO-B":
			idxB = i
		case "TOPO-C":
			idxC = i
		}
	}
	if idxA >= idxB || idxB >= idxC {
		t.Errorf("expected A < B < C, got A=%d B=%d C=%d", idxA, idxB, idxC)
	}
}

// --- gap 6: single-node graph ---

func TestContractGraphSingleNode(t *testing.T) {
	root := setupScaffold(t)
	CreateContract(root, "SN-A", ContractOpts{Title: "A", Status: "draft"})
	CreateContract(root, "SN-B", ContractOpts{Title: "B", Status: "draft", DependsOn: []string{"SN-A"}})
	CreateContract(root, "SN-C", ContractOpts{Title: "C", Status: "draft", DependsOn: []string{"SN-B"}})
	CreateContract(root, "SN-D", ContractOpts{Title: "D", Status: "draft"})

	result, err := ContractGraphNode(root, "SN-B")
	if err != nil {
		t.Fatalf("ContractGraphNode: %v", err)
	}

	ids := make(map[string]bool)
	for _, c := range result {
		ids[c.ID] = true
	}
	if !ids["SN-A"] || !ids["SN-B"] || !ids["SN-C"] {
		t.Errorf("expected A, B, C in neighborhood, got %v", ids)
	}
	if ids["SN-D"] {
		t.Error("SN-D should not be in neighborhood of SN-B")
	}
}

func TestContractGraphSingleNodeNotFound(t *testing.T) {
	root := setupScaffold(t)
	_, err := ContractGraphNode(root, "NOPE")
	if err == nil {
		t.Fatal("expected error for nonexistent contract")
	}
}

// --- gap 7: cycle detection ---

func TestContractGraphCycleWarning(t *testing.T) {
	contracts := []ContractInfo{
		{ID: "A", DependsOn: []string{"B"}},
		{ID: "B", DependsOn: []string{"A"}},
	}
	cycles := DetectCycles(contracts)
	if len(cycles) == 0 {
		t.Fatal("expected cycle A->B->A to be detected")
	}
}

func TestLinkContractCreatesCycle(t *testing.T) {
	root := setupScaffold(t)
	CreateContract(root, "CYC2-A", ContractOpts{Title: "A", Status: "draft"})
	CreateContract(root, "CYC2-B", ContractOpts{Title: "B", Status: "draft", DependsOn: []string{"CYC2-A"}})

	err := LinkContract(root, "CYC2-A", "CYC2-B")
	if err == nil {
		t.Fatal("expected cycle error")
	}
	if !strings.Contains(err.Error(), "cycle") {
		t.Errorf("expected 'cycle' in error, got: %v", err)
	}
}

// --- gap 8: contract update with DependsOn ---

func TestContractUpdateAddDependsOn(t *testing.T) {
	root := setupScaffold(t)
	CreateContract(root, "UPD-A", ContractOpts{Title: "A", Status: "draft"})
	CreateContract(root, "UPD-B", ContractOpts{Title: "B", Status: "draft"})

	deps := []string{"UPD-A"}
	err := UpdateContract(root, "UPD-B", ContractUpdateOpts{DependsOn: &deps})
	if err != nil {
		t.Fatalf("UpdateContract with deps: %v", err)
	}

	bPath, _ := FindContractPath(root, "UPD-B")
	info, _ := readContractInfo("UPD-B", bPath)
	if len(info.DependsOn) != 1 || info.DependsOn[0] != "UPD-A" {
		t.Errorf("expected depends_on = [UPD-A], got %v", info.DependsOn)
	}
}

func TestContractUpdateRemoveDependsOn(t *testing.T) {
	root := setupScaffold(t)
	CreateContract(root, "UPDR-A", ContractOpts{Title: "A", Status: "draft"})
	CreateContract(root, "UPDR-B", ContractOpts{Title: "B", Status: "draft", DependsOn: []string{"UPDR-A"}})

	empty := []string{}
	err := UpdateContract(root, "UPDR-B", ContractUpdateOpts{DependsOn: &empty})
	if err != nil {
		t.Fatalf("UpdateContract remove deps: %v", err)
	}

	bPath, _ := FindContractPath(root, "UPDR-B")
	info, _ := readContractInfo("UPDR-B", bPath)
	if len(info.DependsOn) != 0 {
		t.Errorf("expected no deps, got %v", info.DependsOn)
	}
}

func TestContractUpdateDanglingDep(t *testing.T) {
	root := setupScaffold(t)
	CreateContract(root, "UPDE-A", ContractOpts{Title: "A", Status: "draft"})

	deps := []string{"NONEXISTENT"}
	err := UpdateContract(root, "UPDE-A", ContractUpdateOpts{DependsOn: &deps})
	if err == nil {
		t.Fatal("expected error for dangling dependency in update")
	}
}

// --- gap 9: contract rename ---

func TestRenameContract(t *testing.T) {
	root := setupScaffold(t)
	CreateContract(root, "OLD-ID", ContractOpts{Title: "Rename Me", Status: "draft"})

	err := RenameContract(root, "OLD-ID", "NEW-ID")
	if err != nil {
		t.Fatalf("RenameContract: %v", err)
	}

	if _, err := FindContractPath(root, "OLD-ID"); err == nil {
		t.Error("expected old ID to be gone")
	}
	path, err := FindContractPath(root, "NEW-ID")
	if err != nil {
		t.Fatalf("new ID not found: %v", err)
	}
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), `"NEW-ID"`) {
		t.Error("artifact name not updated to NEW-ID")
	}
}

func TestRenameContractUpdatesDependents(t *testing.T) {
	root := setupScaffold(t)
	CreateContract(root, "REN-A", ContractOpts{Title: "A", Status: "draft"})
	CreateContract(root, "REN-B", ContractOpts{Title: "B", Status: "draft", DependsOn: []string{"REN-A"}})

	err := RenameContract(root, "REN-A", "REN-Z")
	if err != nil {
		t.Fatalf("RenameContract: %v", err)
	}

	bPath, _ := FindContractPath(root, "REN-B")
	info, _ := readContractInfo("REN-B", bPath)
	if len(info.DependsOn) != 1 || info.DependsOn[0] != "REN-Z" {
		t.Errorf("expected REN-B depends_on = [REN-Z], got %v", info.DependsOn)
	}
}

func TestRenameContractNotFound(t *testing.T) {
	root := setupScaffold(t)
	err := RenameContract(root, "NOPE", "NEW")
	if err == nil {
		t.Fatal("expected error for nonexistent contract")
	}
}

func TestRenameContractConflict(t *testing.T) {
	root := setupScaffold(t)
	CreateContract(root, "CONF-A", ContractOpts{Title: "A", Status: "draft"})
	CreateContract(root, "CONF-B", ContractOpts{Title: "B", Status: "draft"})

	err := RenameContract(root, "CONF-A", "CONF-B")
	if err == nil {
		t.Fatal("expected error for conflicting rename target")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' in error, got: %v", err)
	}
}

// --- primitive 10: seal / lock ---

func TestCreateContractTimestamps(t *testing.T) {
	root := setupScaffold(t)
	CreateContract(root, "TS-A", ContractOpts{Title: "A", Status: "draft"})

	aPath, _ := FindContractPath(root, "TS-A")
	info, _ := readContractInfo("TS-A", aPath)
	if info.CreatedAt == "" {
		t.Error("expected created_at to be set")
	}
	if info.UpdatedAt == "" {
		t.Error("expected updated_at to be set")
	}
	if info.CreatedAt != info.UpdatedAt {
		t.Errorf("on creation, created_at (%s) should equal updated_at (%s)", info.CreatedAt, info.UpdatedAt)
	}
}

func TestUpdateContractTimestampUpdated(t *testing.T) {
	root := setupScaffold(t)
	CreateContract(root, "TS-B", ContractOpts{Title: "B", Status: "draft"})

	bPath1, _ := FindContractPath(root, "TS-B")
	info1, _ := readContractInfo("TS-B", bPath1)

	time.Sleep(1100 * time.Millisecond)

	title := "Updated B"
	UpdateContract(root, "TS-B", ContractUpdateOpts{Title: &title})

	bPath2, _ := FindContractPath(root, "TS-B")
	info2, _ := readContractInfo("TS-B", bPath2)
	if info2.CreatedAt != info1.CreatedAt {
		t.Errorf("created_at should not change: was %s, now %s", info1.CreatedAt, info2.CreatedAt)
	}
	if info2.UpdatedAt == info1.UpdatedAt {
		t.Error("expected updated_at to change after update")
	}
}

func TestLinkContractTimestampUpdated(t *testing.T) {
	root := setupScaffold(t)
	CreateContract(root, "TS-C", ContractOpts{Title: "C", Status: "draft"})
	CreateContract(root, "TS-D", ContractOpts{Title: "D", Status: "draft"})

	dPath1, _ := FindContractPath(root, "TS-D")
	info1, _ := readContractInfo("TS-D", dPath1)

	time.Sleep(1100 * time.Millisecond)

	LinkContract(root, "TS-D", "TS-C")

	dPath2, _ := FindContractPath(root, "TS-D")
	info2, _ := readContractInfo("TS-D", dPath2)
	if info2.UpdatedAt == info1.UpdatedAt {
		t.Error("expected updated_at to change after link")
	}
}

func TestContractProgressFromScenarios(t *testing.T) {
	root := setupScaffold(t)
	content := `contract "PROG-A" {
  title = "Progress Test"
  status = "active"
  created_at = 2026-03-01T00:00:00Z
  updated_at = 2026-03-01T00:00:00Z

  feature "Auth" {
    scenario "Login" {
      status = "done"
      given {
        a user
      }
      when {
        they login
      }
      then {
        access granted
      }
    }
    scenario "Signup" {
      status = "done"
      given {
        a visitor
      }
      when {
        they signup
      }
      then {
        account created
      }
    }
    scenario "Reset" {
      status = "pending"
      given {
        a user
      }
      when {
        they reset password
      }
      then {
        email sent
      }
    }
  }
}
`
	ApplyArtifact(root, []byte(content))

	done, total, _, err := ContractProgress(root, "PROG-A")
	if err != nil {
		t.Fatalf("ContractProgress: %v", err)
	}
	if done != 2 || total != 3 {
		t.Errorf("progress = %d/%d, want 2/3", done, total)
	}
}

func TestContractProgressFromSteps(t *testing.T) {
	root := setupScaffold(t)
	content := `contract "PROG-B" {
  title = "Steps Test"
  status = "active"
  created_at = 2026-03-01T00:00:00Z
  updated_at = 2026-03-01T00:00:00Z

  steps {
    step "Design" { status = "done" }
    step "Implement" { status = "in_progress" }
    step "Test" { status = "pending" }
  }
}
`
	ApplyArtifact(root, []byte(content))

	done, total, current, err := ContractProgress(root, "PROG-B")
	if err != nil {
		t.Fatalf("ContractProgress: %v", err)
	}
	if done != 1 || total != 3 {
		t.Errorf("progress = %d/%d, want 1/3", done, total)
	}
	if current != "Implement" {
		t.Errorf("current = %q, want Implement", current)
	}
}

func TestContractProgressMixed(t *testing.T) {
	root := setupScaffold(t)
	content := `contract "PROG-C" {
  title = "Mixed Test"
  status = "active"
  created_at = 2026-03-01T00:00:00Z
  updated_at = 2026-03-01T00:00:00Z

  steps {
    step "Phase1" { status = "done" }
  }

  feature "Core" {
    scenario "Thing" {
      status = "done"
      given {
        something
      }
      when {
        action
      }
      then {
        result
      }
    }
  }
}
`
	ApplyArtifact(root, []byte(content))

	done, total, _, err := ContractProgress(root, "PROG-C")
	if err != nil {
		t.Fatalf("ContractProgress: %v", err)
	}
	if done != 2 || total != 2 {
		t.Errorf("progress = %d/%d, want 2/2", done, total)
	}
}

func TestContractProgressEmpty(t *testing.T) {
	root := setupScaffold(t)
	CreateContract(root, "PROG-D", ContractOpts{Title: "Empty", Status: "draft"})

	done, total, _, err := ContractProgress(root, "PROG-D")
	if err != nil {
		t.Fatalf("ContractProgress: %v", err)
	}
	if done != 0 || total != 0 {
		t.Errorf("progress = %d/%d, want 0/0", done, total)
	}
}

func TestContractProgressAllDone(t *testing.T) {
	root := setupScaffold(t)
	content := `contract "PROG-E" {
  title = "All Done Test"
  status = "complete"
  created_at = 2026-03-01T00:00:00Z
  updated_at = 2026-03-01T00:00:00Z

  steps {
    step "A" { status = "done" }
    step "B" { status = "done" }
  }
}
`
	ApplyArtifact(root, []byte(content))

	done, total, _, err := ContractProgress(root, "PROG-E")
	if err != nil {
		t.Fatalf("ContractProgress: %v", err)
	}
	if done != total || total != 2 {
		t.Errorf("progress = %d/%d, want 2/2", done, total)
	}
}

// --- primitive 13: show modes ---

func TestContractShowShort(t *testing.T) {
	root := setupScaffold(t)
	content := `contract "SHOW-A" {
  title = "Show Test"
  status = "active"
  created_at = 2026-03-01T00:00:00Z
  updated_at = 2026-03-01T00:00:00Z

  steps {
    step "Design" { status = "done" }
    step "Build" { status = "in_progress" }
    step "Ship" { status = "pending" }
  }
}
`
	ApplyArtifact(root, []byte(content))

	short, err := ShowContractShort(root, "SHOW-A")
	if err != nil {
		t.Fatalf("ShowContractShort: %v", err)
	}
	if !strings.Contains(short, "SHOW-A") {
		t.Error("short form should contain contract ID")
	}
	if !strings.Contains(short, "Show Test") {
		t.Error("short form should contain title")
	}
	if !strings.Contains(short, "1/3") {
		t.Error("short form should contain progress")
	}
	if !strings.Contains(short, "Build") {
		t.Error("short form should contain current step")
	}
}

func TestContractShowShortNoCurrent(t *testing.T) {
	root := setupScaffold(t)
	content := `contract "SHOW-B" {
  title = "All Done"
  status = "complete"
  created_at = 2026-03-01T00:00:00Z
  updated_at = 2026-03-01T00:00:00Z

  steps {
    step "A" { status = "done" }
    step "B" { status = "done" }
  }
}
`
	ApplyArtifact(root, []byte(content))

	short, err := ShowContractShort(root, "SHOW-B")
	if err != nil {
		t.Fatalf("ShowContractShort: %v", err)
	}
	if !strings.Contains(short, "all complete") {
		t.Error("expected 'all complete' when all steps are done")
	}
}

func TestContractShowVerbose(t *testing.T) {
	root := setupScaffold(t)
	CreateContract(root, "SHOW-C", ContractOpts{Title: "Verbose Test", Status: "draft"})

	verbose, err := ShowContractVerbose(root, "SHOW-C")
	if err != nil {
		t.Fatalf("ShowContractVerbose: %v", err)
	}
	if !strings.Contains(verbose, "contract") {
		t.Error("verbose should contain full DSL content")
	}
}

func TestContractShowPrevCurrentNext(t *testing.T) {
	root := setupScaffold(t)
	content := `contract "SHOW-D" {
  title = "Nav Test"
  status = "active"
  created_at = 2026-03-01T00:00:00Z
  updated_at = 2026-03-01T00:00:00Z

  steps {
    step "First" { status = "done" }
    step "Second" { status = "in_progress" }
    step "Third" { status = "pending" }
  }
}
`
	ApplyArtifact(root, []byte(content))

	s, err := ContractSummary(root, "SHOW-D")
	if err != nil {
		t.Fatalf("ContractSummary: %v", err)
	}
	if s.Prev != "First" {
		t.Errorf("prev = %q, want First", s.Prev)
	}
	if s.Current != "Second" {
		t.Errorf("current = %q, want Second", s.Current)
	}
	if s.Next != "Third" {
		t.Errorf("next = %q, want Third", s.Next)
	}
}

// --- race condition tests ---

func TestListContractsByProject(t *testing.T) {
	root := setupScaffold(t)
	seedProjectConfig(t, root)

	CreateContract(root, "", ContractOpts{Title: "Bug one", Project: "bugs"})
	CreateContract(root, "", ContractOpts{Title: "Bug two", Project: "bugs"})
	CreateContract(root, "", ContractOpts{Title: "Feature one", Project: "features"})
	CreateContract(root, "", ContractOpts{Title: "Contract one", Project: "contracts"})

	bugs, err := ListContracts(root, ListOpts{Project: "bugs"})
	if err != nil {
		t.Fatalf("ListContracts project=bugs: %v", err)
	}
	if len(bugs) != 2 {
		t.Errorf("expected 2 bug contracts, got %d", len(bugs))
	}
	for _, c := range bugs {
		if !strings.HasPrefix(c.ID, "BUG-") {
			t.Errorf("expected BUG- prefix, got %q", c.ID)
		}
	}

	feats, _ := ListContracts(root, ListOpts{Project: "features"})
	if len(feats) != 1 {
		t.Errorf("expected 1 feature contract, got %d", len(feats))
	}
}

// Feature 2: Contract Kind

func TestCreateContractWithKind(t *testing.T) {
	root := setupScaffold(t)

	CreateContract(root, "KIND-A", ContractOpts{Title: "Kind Test", Kind: "bug"})

	contracts, _ := ListContracts(root, ListOpts{})
	for _, c := range contracts {
		if c.ID == "KIND-A" {
			if c.Kind != "bug" {
				t.Errorf("Kind = %q, want bug", c.Kind)
			}
			return
		}
	}
	t.Error("contract KIND-A not found")
}

func TestListContractsByKind(t *testing.T) {
	root := setupScaffold(t)

	CreateContract(root, "K-BUG", ContractOpts{Title: "A bug", Kind: "bug"})
	CreateContract(root, "K-FEAT", ContractOpts{Title: "A feature", Kind: "feature"})
	CreateContract(root, "K-TASK", ContractOpts{Title: "A task", Kind: "task"})

	bugs, _ := ListContracts(root, ListOpts{Kind: "bug"})
	if len(bugs) != 1 {
		t.Errorf("expected 1 bug, got %d", len(bugs))
	}
	if len(bugs) > 0 && bugs[0].ID != "K-BUG" {
		t.Errorf("expected K-BUG, got %s", bugs[0].ID)
	}
}

// Feature 3: Contract Labels

func TestCreateContractWithLabels(t *testing.T) {
	root := setupScaffold(t)

	CreateContract(root, "LBL-A", ContractOpts{Title: "Label Test", Labels: []string{"harness", "testability", "dx"}})

	contracts, _ := ListContracts(root, ListOpts{})
	for _, c := range contracts {
		if c.ID == "LBL-A" {
			if len(c.Labels) != 3 {
				t.Errorf("Labels count = %d, want 3", len(c.Labels))
			}
			return
		}
	}
	t.Error("contract LBL-A not found")
}

func TestListContractsByLabel(t *testing.T) {
	root := setupScaffold(t)

	CreateContract(root, "L-ONE", ContractOpts{Title: "One", Labels: []string{"harness", "dx"}})
	CreateContract(root, "L-TWO", ContractOpts{Title: "Two", Labels: []string{"security"}})
	CreateContract(root, "L-THREE", ContractOpts{Title: "Three", Labels: []string{"harness", "security"}})

	harness, _ := ListContracts(root, ListOpts{Label: "harness"})
	if len(harness) != 2 {
		t.Errorf("expected 2 harness-labeled contracts, got %d", len(harness))
	}
}

// Feature 4: Contract Priority

func TestCreateContractWithPriority(t *testing.T) {
	root := setupScaffold(t)

	CreateContract(root, "PRI-A", ContractOpts{Title: "Priority Test", Priority: "p1"})

	contracts, _ := ListContracts(root, ListOpts{})
	for _, c := range contracts {
		if c.ID == "PRI-A" {
			if c.Priority != "p1" {
				t.Errorf("Priority = %q, want p1", c.Priority)
			}
			return
		}
	}
	t.Error("contract PRI-A not found")
}

func TestListContractsByPriority(t *testing.T) {
	root := setupScaffold(t)

	CreateContract(root, "P-ONE", ContractOpts{Title: "P1", Priority: "p1"})
	CreateContract(root, "P-TWO", ContractOpts{Title: "P2", Priority: "p2"})
	CreateContract(root, "P-THREE", ContractOpts{Title: "P1 too", Priority: "p1"})

	p1s, _ := ListContracts(root, ListOpts{Priority: "p1"})
	if len(p1s) != 2 {
		t.Errorf("expected 2 p1 contracts, got %d", len(p1s))
	}
}

// Feature 5: Sequence Atomicity

func TestAssignParentToContract(t *testing.T) {
	root := setupScaffold(t)

	CreateContract(root, "CON-INIT", ContractOpts{Title: "Initiative"})
	CreateContract(root, "EPIC-001", ContractOpts{Title: "Epic One", Parent: "CON-INIT"})

	contracts, _ := ListContracts(root, ListOpts{})
	for _, c := range contracts {
		if c.ID == "EPIC-001" {
			if c.Parent != "CON-INIT" {
				t.Errorf("Parent = %q, want CON-INIT", c.Parent)
			}
			return
		}
	}
	t.Error("EPIC-001 not found")
}

func TestRejectCircularParent(t *testing.T) {
	root := setupScaffold(t)

	CreateContract(root, "A", ContractOpts{Title: "A", Parent: "B"})
	CreateContract(root, "B", ContractOpts{Title: "B"})

	parentB := "A"
	err := UpdateContract(root, "B", ContractUpdateOpts{Parent: &parentB})
	if err == nil {
		t.Error("expected cycle detection error, got nil")
	}
}

func TestReparentContract(t *testing.T) {
	root := setupScaffold(t)

	CreateContract(root, "EPIC-001", ContractOpts{Title: "Epic 1"})
	CreateContract(root, "EPIC-002", ContractOpts{Title: "Epic 2"})
	CreateContract(root, "STORY-001", ContractOpts{Title: "Story", Parent: "EPIC-001"})

	newParent := "EPIC-002"
	if err := UpdateContract(root, "STORY-001", ContractUpdateOpts{Parent: &newParent}); err != nil {
		t.Fatalf("reparent: %v", err)
	}

	contracts, _ := ListContracts(root, ListOpts{})
	for _, c := range contracts {
		if c.ID == "STORY-001" {
			if c.Parent != "EPIC-002" {
				t.Errorf("Parent = %q, want EPIC-002", c.Parent)
			}
			return
		}
	}
	t.Error("STORY-001 not found")
}

// Feature: Subgraph Rendering

func TestGraphClustersChildrenUnderParent(t *testing.T) {
	root := setupScaffold(t)

	CreateContract(root, "INIT", ContractOpts{Title: "Initiative"})
	CreateContract(root, "EPIC-A", ContractOpts{Title: "Epic A", Parent: "INIT"})
	CreateContract(root, "EPIC-B", ContractOpts{Title: "Epic B", Parent: "INIT"})
	CreateContract(root, "STORY-A1", ContractOpts{Title: "Story A1", Parent: "EPIC-A"})
	CreateContract(root, "STORY-A2", ContractOpts{Title: "Story A2", Parent: "EPIC-A"})
	CreateContract(root, "STORY-B1", ContractOpts{Title: "Story B1", Parent: "EPIC-B"})

	children, err := FindChildren(root, "INIT")
	if err != nil {
		t.Fatalf("FindChildren: %v", err)
	}
	if len(children) != 2 {
		t.Errorf("expected 2 direct children of INIT, got %d", len(children))
	}

	epicAChildren, _ := FindChildren(root, "EPIC-A")
	if len(epicAChildren) != 2 {
		t.Errorf("expected 2 children of EPIC-A, got %d", len(epicAChildren))
	}
}

func TestSingleNodeGraphShowsChildren(t *testing.T) {
	root := setupScaffold(t)

	CreateContract(root, "EPIC-001", ContractOpts{Title: "Epic"})
	CreateContract(root, "S1", ContractOpts{Title: "Story 1", Parent: "EPIC-001"})
	CreateContract(root, "S2", ContractOpts{Title: "Story 2", Parent: "EPIC-001"})
	CreateContract(root, "S3", ContractOpts{Title: "Story 3", Parent: "EPIC-001"})

	children, _ := FindChildren(root, "EPIC-001")
	if len(children) != 3 {
		t.Errorf("expected 3 children, got %d", len(children))
	}

	s, err := ContractSummary(root, "EPIC-001")
	if err != nil {
		t.Fatalf("ContractSummary: %v", err)
	}
	if len(s.Children) != 3 {
		t.Errorf("Summary.Children = %d, want 3", len(s.Children))
	}
}

// Feature: Rollup Progress

func TestRollupProgressAggregatesChildren(t *testing.T) {
	root := setupScaffold(t)

	specComplete := filepath.Join(root, "spec-complete.mos")
	os.WriteFile(specComplete, []byte(`feature "F" {
  scenario "S1" { status = "done" }
}
`), 0644)

	specActive := filepath.Join(root, "spec-active.mos")
	os.WriteFile(specActive, []byte(`feature "F" {
  scenario "S1" { status = "done" }
  scenario "S2" { status = "pending" }
}
`), 0644)

	specDraft := filepath.Join(root, "spec-draft.mos")
	os.WriteFile(specDraft, []byte(`feature "F" {
  scenario "S1" { status = "pending" }
  scenario "S2" { status = "pending" }
}
`), 0644)

	CreateContract(root, "EPIC-RP", ContractOpts{Title: "Rollup Epic"})
	CreateContract(root, "S-DONE", ContractOpts{Title: "Done story", Parent: "EPIC-RP", Status: "complete", SpecFile: specComplete})
	CreateContract(root, "S-ACTIVE", ContractOpts{Title: "Active story", Parent: "EPIC-RP", SpecFile: specActive})
	CreateContract(root, "S-DRAFT", ContractOpts{Title: "Draft story", Parent: "EPIC-RP", SpecFile: specDraft})

	s, err := ContractSummary(root, "EPIC-RP")
	if err != nil {
		t.Fatalf("ContractSummary: %v", err)
	}
	if s.RollupProgress == "" {
		t.Error("expected RollupProgress to be set")
	}
	// S-DONE: 1/1, S-ACTIVE: 1/2, S-DRAFT: 0/2 => total 2/5
	if s.RollupProgress != "2/5" {
		t.Errorf("RollupProgress = %q, want 2/5", s.RollupProgress)
	}
}

// Feature: Scoped Listing

func TestListChildrenOfParent(t *testing.T) {
	root := setupScaffold(t)

	CreateContract(root, "INIT-SL", ContractOpts{Title: "Init"})
	CreateContract(root, "E1", ContractOpts{Title: "E1", Parent: "INIT-SL"})
	CreateContract(root, "E2", ContractOpts{Title: "E2", Parent: "INIT-SL"})
	CreateContract(root, "S1", ContractOpts{Title: "S1", Parent: "E1"})
	CreateContract(root, "LONE", ContractOpts{Title: "Lone"})

	children, _ := ListContracts(root, ListOpts{Parent: "INIT-SL"})
	if len(children) != 2 {
		t.Errorf("expected 2 direct children, got %d", len(children))
	}
	for _, c := range children {
		if c.ID != "E1" && c.ID != "E2" {
			t.Errorf("unexpected child: %s", c.ID)
		}
	}
}

func TestListRootContracts(t *testing.T) {
	root := setupScaffold(t)

	CreateContract(root, "ROOT-A", ContractOpts{Title: "Root A"})
	CreateContract(root, "ROOT-B", ContractOpts{Title: "Root B"})
	CreateContract(root, "CHILD-A", ContractOpts{Title: "Child A", Parent: "ROOT-A"})

	roots, _ := ListContracts(root, ListOpts{Roots: true})
	if len(roots) != 2 {
		t.Errorf("expected 2 roots, got %d", len(roots))
	}
	for _, c := range roots {
		if c.Parent != "" {
			t.Errorf("root %s has parent %q", c.ID, c.Parent)
		}
	}
}

func TestRecursiveListing(t *testing.T) {
	root := setupScaffold(t)

	CreateContract(root, "TOP", ContractOpts{Title: "Top"})
	CreateContract(root, "MID", ContractOpts{Title: "Mid", Parent: "TOP"})
	CreateContract(root, "LEAF", ContractOpts{Title: "Leaf", Parent: "MID"})
	CreateContract(root, "OTHER", ContractOpts{Title: "Other"})

	descendants, _ := ListContracts(root, ListOpts{Parent: "TOP", Recursive: true})
	if len(descendants) != 2 {
		t.Errorf("expected 2 descendants, got %d", len(descendants))
	}

	ids := map[string]bool{}
	for _, c := range descendants {
		ids[c.ID] = true
	}
	if !ids["MID"] || !ids["LEAF"] {
		t.Errorf("expected MID and LEAF, got %v", ids)
	}
}

// Feature: Branch Scoping

func TestAssignBranchesToContract(t *testing.T) {
	root := setupScaffold(t)

	CreateContract(root, "BUG-ROOT", ContractOpts{Title: "Root bug"})
	CreateContract(root, "BUG-BP-12", ContractOpts{
		Title:    "Backport v1.2",
		Parent:   "BUG-ROOT",
		Branches: []string{"release/v1.2"},
	})

	contracts, _ := ListContracts(root, ListOpts{})
	for _, c := range contracts {
		if c.ID == "BUG-BP-12" {
			if len(c.Branches) != 1 || c.Branches[0] != "release/v1.2" {
				t.Errorf("Branches = %v, want [release/v1.2]", c.Branches)
			}
			return
		}
	}
	t.Error("BUG-BP-12 not found")
}

func TestFilterContractsByBranch(t *testing.T) {
	root := setupScaffold(t)

	CreateContract(root, "BP-10", ContractOpts{Title: "BP v1.0", Branches: []string{"release/v1.0"}})
	CreateContract(root, "BP-11", ContractOpts{Title: "BP v1.1", Branches: []string{"release/v1.1"}})
	CreateContract(root, "BP-12", ContractOpts{Title: "BP v1.2", Branches: []string{"release/v1.2"}})
	CreateContract(root, "NOSCOPE", ContractOpts{Title: "No scope"})

	results, _ := ListContracts(root, ListOpts{Branch: "release/v1.2"})
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if len(results) > 0 && results[0].ID != "BP-12" {
		t.Errorf("expected BP-12, got %s", results[0].ID)
	}
}

// --- CON-2026-023: Custom Artifact Definitions (CAD) ---

// Feature 1: Schema Registry

func TestBackportTreeWithPerBranchChildren(t *testing.T) {
	root := setupScaffold(t)

	specDone := filepath.Join(root, "spec-done.mos")
	os.WriteFile(specDone, []byte(`feature "F" { scenario "S" { status = "done" } }
`), 0644)
	specPending := filepath.Join(root, "spec-pending.mos")
	os.WriteFile(specPending, []byte(`feature "F" { scenario "S" { status = "pending" } }
`), 0644)

	CreateContract(root, "BUG-ROOT-BT", ContractOpts{Title: "Root bug fix"})
	CreateContract(root, "BUG-V10", ContractOpts{Title: "v1.0 backport", Parent: "BUG-ROOT-BT", Branches: []string{"release/v1.0"}, Status: "complete", SpecFile: specDone})
	CreateContract(root, "BUG-V11", ContractOpts{Title: "v1.1 backport", Parent: "BUG-ROOT-BT", Branches: []string{"release/v1.1"}, SpecFile: specPending})
	CreateContract(root, "BUG-V12", ContractOpts{Title: "v1.2 backport", Parent: "BUG-ROOT-BT", Branches: []string{"release/v1.2"}, SpecFile: specPending})

	s, err := ContractSummary(root, "BUG-ROOT-BT")
	if err != nil {
		t.Fatalf("ContractSummary: %v", err)
	}
	if len(s.Children) != 3 {
		t.Errorf("expected 3 children, got %d", len(s.Children))
	}
	// BUG-V10: 1/1, BUG-V11: 0/1, BUG-V12: 0/1 => 1/3
	if s.RollupProgress != "1/3" {
		t.Errorf("RollupProgress = %q, want 1/3", s.RollupProgress)
	}
}

// ===== CON-2026-024: Specification as a First-Class Artifact =====

// Feature: Specification Artifact

func TestCreateContractWithKindAutoID(t *testing.T) {
	root := setupScaffold(t)

	path, err := CreateContract(root, "", ContractOpts{
		Title: "A bug report",
		Kind:  "bug",
	})
	if err != nil {
		t.Fatalf("CreateContract with kind=bug: %v", err)
	}
	if !strings.Contains(filepath.Base(filepath.Dir(path)), "BUG-") {
		t.Errorf("expected BUG- prefix in path, got %s", path)
	}
}
