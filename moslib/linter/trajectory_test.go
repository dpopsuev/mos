package linter

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dpopsuev/mos/moslib/harness"
)

func TestValidateTrajectory_InsufficientSnapshots(t *testing.T) {
	dir := t.TempDir()
	mosDir := filepath.Join(dir, ".mos")
	os.MkdirAll(filepath.Join(mosDir, "vcs", "snapshots"), 0o755)

	for i := 0; i < 3; i++ {
		snap := &harness.StateSnapshot{
			Timestamp:      time.Unix(int64(100+i*100), 0).UTC(),
			IntegrityScore: 50,
		}
		harness.StoreCommitSnapshot(mosDir, "hash"+string(rune('a'+i)), snap)
	}

	diags := validateTrajectory(dir, mosDir)
	if len(diags) != 0 {
		t.Errorf("expected 0 diags with insufficient snapshots, got %d", len(diags))
	}
}

func TestValidateTrajectory_StallDetected(t *testing.T) {
	dir := t.TempDir()
	mosDir := filepath.Join(dir, ".mos")
	os.MkdirAll(filepath.Join(mosDir, "vcs", "snapshots"), 0o755)

	for i := 0; i < 10; i++ {
		snap := &harness.StateSnapshot{
			Timestamp:      time.Unix(int64(100+i*100), 0).UTC(),
			IntegrityScore: 50,
			Vectors: []harness.VectorResult{
				{Kind: harness.VectorFunctional, Score: 50},
				{Kind: harness.VectorStructural, Score: 50},
				{Kind: harness.VectorPerformance, Score: 50},
			},
			LintErrors: 5,
		}
		harness.StoreCommitSnapshot(mosDir, "hash"+string(rune('a'+i)), snap)
	}

	diags := validateTrajectory(dir, mosDir)
	foundStall := false
	for _, d := range diags {
		if d.Rule == "trajectory-stall" {
			foundStall = true
			break
		}
	}
	if !foundStall {
		t.Error("expected trajectory-stall diagnostic for flat series")
	}
}

func TestValidateTrajectory_RegressionDetected(t *testing.T) {
	dir := t.TempDir()
	mosDir := filepath.Join(dir, ".mos")
	os.MkdirAll(filepath.Join(mosDir, "vcs", "snapshots"), 0o755)

	for i := 0; i < 10; i++ {
		snap := &harness.StateSnapshot{
			Timestamp:       time.Unix(int64(100+i*100), 0).UTC(),
			IntegrityScore:  50,
			DriftViolations: i * 3,
			Vectors: []harness.VectorResult{
				{Kind: harness.VectorFunctional, Score: 50},
				{Kind: harness.VectorStructural, Score: 50},
				{Kind: harness.VectorPerformance, Score: 50},
			},
		}
		harness.StoreCommitSnapshot(mosDir, "hash"+string(rune('a'+i)), snap)
	}

	diags := validateTrajectory(dir, mosDir)
	foundRegression := false
	for _, d := range diags {
		if d.Rule == "trajectory-regression" {
			foundRegression = true
			break
		}
	}
	if !foundRegression {
		t.Error("expected trajectory-regression diagnostic for growing drift violations")
	}
}
