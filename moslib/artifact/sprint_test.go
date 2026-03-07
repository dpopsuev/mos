package artifact

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func setupSprintTest(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	mos := filepath.Join(root, ".mos")

	cfg := `config {
  mos { version = 1 }
  backend { type = "git" }
  project "sprints" { prefix = "SPR" sequence = 0 }
  artifact_type "sprint" {
    directory = "sprints"
    fields {
      title    { required = true }
      status   { required = true enum = ["planned" "in-progress" "complete"] }
      goal     {}
      contracts {}
    }
    lifecycle {
      active_states  = ["planned" "in-progress"]
      archive_states = ["complete"]
    }
  }
}
`
	os.MkdirAll(filepath.Join(mos, "sprints", "active"), 0o755)
	os.MkdirAll(filepath.Join(mos, "sprints", "archive"), 0o755)
	os.MkdirAll(filepath.Join(mos, "contracts", "active"), 0o755)
	os.MkdirAll(filepath.Join(mos, "contracts", "archive"), 0o755)
	os.WriteFile(filepath.Join(mos, "config.mos"), []byte(cfg), 0o644)

	sprintDir := filepath.Join(mos, "sprints", "active", "SPR-2026-100")
	os.MkdirAll(sprintDir, 0o755)
	os.WriteFile(filepath.Join(sprintDir, "sprint.mos"), []byte(`sprint "SPR-2026-100" {
  title = "Test Sprint"
  status = "in-progress"
  contracts = "CON-2026-001,CON-2026-002,CON-2026-003"
}
`), 0o644)

	for _, cid := range []string{"CON-2026-001", "CON-2026-002", "CON-2026-003"} {
		cdir := filepath.Join(mos, "contracts", "active", cid)
		os.MkdirAll(cdir, 0o755)
		status := "active"
		if cid == "CON-2026-003" {
			status = StatusComplete
		}
		os.WriteFile(filepath.Join(cdir, "contract.mos"), []byte(`contract "`+cid+`" {
  title = "Test Contract"
  status = "`+status+`"
  sprint = "SPR-2026-100"
}
`), 0o644)
	}

	return root
}

func TestSprintClose(t *testing.T) {
	root := setupSprintTest(t)

	result, err := SprintClose(root, "SPR-2026-100")
	if err != nil {
		t.Fatalf("SprintClose: %v", err)
	}
	if result.Closed != 2 {
		t.Errorf("closed = %d, want 2", result.Closed)
	}
	if result.AlreadyDone != 1 {
		t.Errorf("already_done = %d, want 1", result.AlreadyDone)
	}
	if len(result.Contracts) != 3 {
		t.Errorf("contracts = %d, want 3", len(result.Contracts))
	}

	for _, cid := range []string{"CON-2026-001", "CON-2026-002", "CON-2026-003"} {
		status, err := GetContractStatus(root, cid)
		if err != nil {
			t.Fatalf("GetContractStatus(%s): %v", cid, err)
		}
		if status != StatusComplete {
			t.Errorf("%s status = %q, want %q", cid, status, StatusComplete)
		}
	}
}

func TestPlanSprint(t *testing.T) {
	root := setupSprintTest(t)
	mos := filepath.Join(root, ".mos")

	// Create unassigned draft contracts
	for _, cid := range []string{"CON-2026-010", "CON-2026-011", "CON-2026-012"} {
		cdir := filepath.Join(mos, "contracts", "active", cid)
		os.MkdirAll(cdir, 0o755)
		os.WriteFile(filepath.Join(cdir, "contract.mos"), []byte(`contract "`+cid+`" {
  title = "Unassigned"
  status = "draft"
}
`), 0o644)
	}

	proposal, err := PlanSprint(root, 8)
	if err != nil {
		t.Fatalf("PlanSprint: %v", err)
	}
	if len(proposal.ContractIDs) != 3 {
		t.Errorf("expected 3 contracts, got %d: %v", len(proposal.ContractIDs), proposal.ContractIDs)
	}
}

func TestPlanSprint_MaxCap(t *testing.T) {
	root := setupSprintTest(t)
	mos := filepath.Join(root, ".mos")

	for i := 0; i < 5; i++ {
		cid := fmt.Sprintf("CON-2026-0%d0", i)
		cdir := filepath.Join(mos, "contracts", "active", cid)
		os.MkdirAll(cdir, 0o755)
		os.WriteFile(filepath.Join(cdir, "contract.mos"), []byte(`contract "`+cid+`" {
  title = "Backlog item"
  status = "draft"
}
`), 0o644)
	}

	proposal, err := PlanSprint(root, 3)
	if err != nil {
		t.Fatalf("PlanSprint: %v", err)
	}
	if len(proposal.ContractIDs) != 3 {
		t.Errorf("expected max 3 contracts, got %d", len(proposal.ContractIDs))
	}
}

func TestValidateKindChange_RejectsMismatch(t *testing.T) {
	root := t.TempDir()
	mos := filepath.Join(root, ".mos")
	os.MkdirAll(filepath.Join(mos, "contracts", "active"), 0o755)

	cfg := `config {
  mos { version = 1 }
  backend { type = "git" }
  project "contracts" { prefix = "CON" sequence = 0 default = true }
  project "bugs" { prefix = "BUG" sequence = 1 }
}
`
	os.WriteFile(filepath.Join(mos, "config.mos"), []byte(cfg), 0o644)

	bugDir := filepath.Join(mos, "contracts", "active", "BUG-2026-001")
	os.MkdirAll(bugDir, 0o755)
	os.WriteFile(filepath.Join(bugDir, "contract.mos"), []byte(`contract "BUG-2026-001" {
  title = "A bug"
  status = "draft"
  kind = "bug"
}
`), 0o644)

	err := SetArtifactField(root, "BUG-2026-001", "kind", "feature")
	if err == nil {
		t.Fatal("expected error when setting kind=feature on BUG- prefix artifact")
	}

	err = SetArtifactField(root, "BUG-2026-001", "kind", "bug")
	if err != nil {
		t.Fatalf("setting kind=bug on BUG- should succeed: %v", err)
	}
}

func TestValidateKindChange_AllowsDefaultPrefix(t *testing.T) {
	root := t.TempDir()
	mos := filepath.Join(root, ".mos")
	os.MkdirAll(filepath.Join(mos, "contracts", "active"), 0o755)

	cfg := `config {
  mos { version = 1 }
  backend { type = "git" }
  project "contracts" { prefix = "CON" sequence = 0 default = true }
  project "bugs" { prefix = "BUG" sequence = 1 }
}
`
	os.WriteFile(filepath.Join(mos, "config.mos"), []byte(cfg), 0o644)

	conDir := filepath.Join(mos, "contracts", "active", "CON-2026-001")
	os.MkdirAll(conDir, 0o755)
	os.WriteFile(filepath.Join(conDir, "contract.mos"), []byte(`contract "CON-2026-001" {
  title = "A contract"
  status = "draft"
  kind = "feature"
}
`), 0o644)

	err := SetArtifactField(root, "CON-2026-001", "kind", "bug")
	if err != nil {
		t.Fatalf("setting kind=bug on CON- (default prefix) should succeed: %v", err)
	}
}

func TestSprintCloseDryRun(t *testing.T) {
	root := setupSprintTest(t)

	result, err := SprintCloseDryRun(root, "SPR-2026-100")
	if err != nil {
		t.Fatalf("SprintCloseDryRun: %v", err)
	}
	if result.Closed != 2 {
		t.Errorf("would close = %d, want 2", result.Closed)
	}
	if result.AlreadyDone != 1 {
		t.Errorf("already_done = %d, want 1", result.AlreadyDone)
	}

	// Verify nothing was actually changed
	status, _ := GetContractStatus(root, "CON-2026-001")
	if status != "active" {
		t.Errorf("CON-2026-001 status = %q after dry-run, want active", status)
	}
}
