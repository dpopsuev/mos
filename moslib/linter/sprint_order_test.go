package linter

import (
	"os"
	"path/filepath"
	"testing"
)

func writeSprint(t *testing.T, mosDir, id, status string) {
	t.Helper()
	dir := filepath.Join(mosDir, "sprints", "active", id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `sprint "` + id + `" {
  title = "Test Sprint"
  status = "` + status + `"
}
`
	if err := os.WriteFile(filepath.Join(dir, "sprint.mos"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestValidateSprintOrdering_LinearOrder(t *testing.T) {
	mosDir := t.TempDir()
	writeSprint(t, mosDir, "SPR-2026-001", "closed")
	writeSprint(t, mosDir, "SPR-2026-002", "closed")
	writeSprint(t, mosDir, "SPR-2026-003", "planned")

	ctx := &ProjectContext{Root: mosDir}
	diags := validateSprintOrdering(ctx)
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics for linear order, got %d: %v", len(diags), diags)
	}
}

func TestValidateSprintOrdering_SkippedSprint(t *testing.T) {
	mosDir := t.TempDir()
	writeSprint(t, mosDir, "SPR-2026-001", "closed")
	writeSprint(t, mosDir, "SPR-2026-002", "planned")
	writeSprint(t, mosDir, "SPR-2026-003", "closed")

	ctx := &ProjectContext{Root: mosDir}
	diags := validateSprintOrdering(ctx)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic for skipped sprint, got %d: %v", len(diags), diags)
	}
	if diags[0].Rule != "sprint-order" {
		t.Errorf("expected rule=sprint-order, got %s", diags[0].Rule)
	}
	if diags[0].Severity != SeverityError {
		t.Errorf("expected severity=error, got %s", diags[0].Severity)
	}
}

func TestValidateSprintOrdering_CancelledSkipped(t *testing.T) {
	mosDir := t.TempDir()
	writeSprint(t, mosDir, "SPR-2026-001", "closed")
	writeSprint(t, mosDir, "SPR-2026-002", "cancelled")
	writeSprint(t, mosDir, "SPR-2026-003", "closed")

	ctx := &ProjectContext{Root: mosDir}
	diags := validateSprintOrdering(ctx)
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics when skipped sprint is cancelled, got %d: %v", len(diags), diags)
	}
}

func TestParseSprintSeq(t *testing.T) {
	tests := []struct {
		id       string
		expected int
	}{
		{"SPR-2026-033", 2026033},
		{"SPR-2026-001", 2026001},
		{"SPR-2025-999", 2025999},
		{"invalid", -1},
		{"SPR-abc-001", -1},
	}
	for _, tc := range tests {
		got := parseSprintSeq(tc.id)
		if got != tc.expected {
			t.Errorf("parseSprintSeq(%q) = %d, want %d", tc.id, got, tc.expected)
		}
	}
}
