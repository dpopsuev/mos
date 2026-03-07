package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dpopsuev/mos/moslib/artifact"
	"github.com/dpopsuev/mos/moslib/harness"
)

func setupScaffold(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	content := "module github.com/test/scaffold\n\ngo 1.25.7\n"
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte(content), 0644); err != nil {
		t.Fatalf("writing go.mod: %v", err)
	}
	if err := artifact.Init(root, artifact.InitOpts{Name: "scaffold", Model: "bdfl", Scope: "cabinet"}); err != nil {
		t.Fatalf("scaffold setup failed: %v", err)
	}
	return root
}

func TestCON125_SprintTransitiveMembership(t *testing.T) {
	root := setupScaffold(t)
	configPath := filepath.Join(root, ".mos", "config.mos")
	data, _ := os.ReadFile(configPath)
	extra := `
  artifact_type "sprint" {
    directory = "sprints"
    fields {
      title { required = true }
      goal { required = true }
      contracts {}
      status {
        required = true
        enum = ["planned", "active", "complete", "cancelled"]
      }
    }
    lifecycle {
      active_states = ["planned", "active"]
      archive_states = ["complete", "cancelled"]
    }
  }

  artifact_type "batch" {
    directory = "batches"
    fields {
      title { required = true }
      goal { required = true }
      sprint {
        link = true
        ref_kind = "sprint"
      }
      status {
        required = true
        enum = ["planned", "ready", "promoted", "cancelled"]
      }
    }
    lifecycle {
      active_states = ["planned", "ready"]
      archive_states = ["promoted", "cancelled"]
    }
  }
`
	patched := strings.Replace(string(data), "\n}\n", extra+"\n}\n", 1)
	os.WriteFile(configPath, []byte(patched), 0644)

	for _, d := range []string{"sprints", "batches"} {
		os.MkdirAll(filepath.Join(root, ".mos", d, "active"), 0755)
		os.MkdirAll(filepath.Join(root, ".mos", d, "archive"), 0755)
	}

	sprintDir := filepath.Join(root, ".mos", "sprints", "active", "SPR-001")
	os.MkdirAll(sprintDir, 0755)
	os.WriteFile(filepath.Join(sprintDir, "sprint.mos"),
		[]byte(`sprint "SPR-001" { title = "Sprint 1" goal = "test" status = "active" contracts = "CON-LISTED" }`), 0644)

	batchDir := filepath.Join(root, ".mos", "batches", "active", "BAT-001")
	os.MkdirAll(batchDir, 0755)
	os.WriteFile(filepath.Join(batchDir, "batch.mos"),
		[]byte(`batch "BAT-001" { title = "Batch" goal = "test" status = "planned" sprint = "SPR-001" }`), 0644)

	for _, c := range []struct{ id, extra string }{
		{"CON-DIRECT", `sprint = "SPR-001"`},
		{"CON-LISTED", ""},
		{"CON-BATCHED", `batch = "BAT-001"`},
		{"CON-ORPHAN", ""},
	} {
		cdir := filepath.Join(root, ".mos", "contracts", "active", c.id)
		os.MkdirAll(cdir, 0755)
		os.WriteFile(filepath.Join(cdir, "contract.mos"),
			[]byte(fmt.Sprintf(`contract %q { title = "C" status = "draft" %s }`, c.id, c.extra)), 0644)
	}

	report, err := RunAudit(root, AuditOpts{})
	if err != nil {
		t.Fatalf("RunAudit: %v", err)
	}
	if len(report.SprintStatus) != 1 {
		t.Fatalf("expected 1 sprint, got %d", len(report.SprintStatus))
	}
	s := report.SprintStatus[0]
	if s.Total < 3 {
		t.Errorf("expected at least 3 sprint members (direct + listed + batched), got %d", s.Total)
	}
}

// --- CON-2026-236: Fast Governance Mode ---

func TestRunAuditNoHarnessWithCachedSnapshot(t *testing.T) {
	root := setupScaffold(t)
	mosDir := filepath.Join(root, ".mos")

	snapshotDir := filepath.Join(mosDir, "snapshots")
	os.MkdirAll(snapshotDir, 0755)
	snap := harness.StateSnapshot{
		Timestamp:      time.Now(),
		IntegrityScore: 0.85,
		Vectors: []harness.VectorResult{
			{Kind: "functional", Score: 0.9, Pass: true},
		},
	}
	data, _ := json.Marshal(snap)
	os.WriteFile(filepath.Join(snapshotDir, "snap-001.json"), data, 0644)

	report, err := RunAudit(root, AuditOpts{NoHarness: true})
	if err != nil {
		t.Fatalf("RunAudit: %v", err)
	}
	if report.IntegrityScore != 0.85 {
		t.Errorf("IntegrityScore = %f, want 0.85", report.IntegrityScore)
	}
	if len(report.VectorScores) != 1 {
		t.Errorf("VectorScores count = %d, want 1", len(report.VectorScores))
	}
}

func TestRunAuditNoHarnessNoSnapshots(t *testing.T) {
	root := setupScaffold(t)

	report, err := RunAudit(root, AuditOpts{NoHarness: true})
	if err != nil {
		t.Fatalf("RunAudit: %v", err)
	}
	if report.IntegrityScore != -1 {
		t.Errorf("IntegrityScore = %f, want -1 (N/A)", report.IntegrityScore)
	}
}

func TestFormatReportNoHarnessNA(t *testing.T) {
	report := &Report{IntegrityScore: -1}
	text := FormatReport(report, false)
	if !strings.Contains(text, "N/A") {
		t.Errorf("expected N/A in output, got: %s", text)
	}
}

func TestAuditReportJSONFields(t *testing.T) {
	report := &Report{
		LintErrors:      2,
		LintWarnings:    5,
		LintInfos:       1,
		SprintStatus:    []SprintSummary{{ID: "SPR-001", Total: 3, Complete: 1}},
		OrphanContracts: []string{"CON-ORPHAN"},
		ArchViolations:  []string{"forbidden edge"},
		DriftViolations: []string{"drift issue"},
		IntegrityScore:  0.75,
	}
	data, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	for _, field := range []string{
		"lint_errors", "lint_warnings", "sprint_status",
		"orphan_contracts", "arch_violations", "drift_violations",
		"integrity_score",
	} {
		if _, ok := m[field]; !ok {
			t.Errorf("JSON missing field %q", field)
		}
	}
}
