package harness

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLightweightScan_WithLintFn(t *testing.T) {
	dir := t.TempDir()
	mosDir := filepath.Join(dir, ".mos")
	os.MkdirAll(mosDir, 0o755)

	lintFn := func(root string) (*LintResult, error) {
		return &LintResult{
			Errors:          3,
			Warnings:        7,
			DriftViolations: 2,
			StructuralMetrics: map[string]float64{
				"nesting-depth": 1,
			},
		}, nil
	}

	snap, err := LightweightScan(dir, lintFn)
	if err != nil {
		t.Fatalf("LightweightScan: %v", err)
	}
	if snap.LintErrors != 3 {
		t.Errorf("LintErrors = %d, want 3", snap.LintErrors)
	}
	if snap.LintWarnings != 7 {
		t.Errorf("LintWarnings = %d, want 7", snap.LintWarnings)
	}
	if snap.DriftViolations != 2 {
		t.Errorf("DriftViolations = %d, want 2", snap.DriftViolations)
	}
	if snap.StructuralMetrics["nesting-depth"] != 1 {
		t.Errorf("StructuralMetrics[nesting-depth] = %v", snap.StructuralMetrics["nesting-depth"])
	}
}

func TestLightweightScan_NilLintFn(t *testing.T) {
	dir := t.TempDir()
	mosDir := filepath.Join(dir, ".mos")
	os.MkdirAll(mosDir, 0o755)

	snap, err := LightweightScan(dir, nil)
	if err != nil {
		t.Fatalf("LightweightScan: %v", err)
	}
	if snap.LintErrors != 0 {
		t.Errorf("LintErrors = %d, want 0", snap.LintErrors)
	}
}

func TestLightweightScan_InheritsCachedIntegrity(t *testing.T) {
	dir := t.TempDir()
	mosDir := filepath.Join(dir, ".mos")
	os.MkdirAll(mosDir, 0o755)

	cached := &StateSnapshot{
		IntegrityScore: 75.5,
		Vectors: []VectorResult{
			{Kind: VectorFunctional, Score: 80},
		},
		ScenariosTotal:   10,
		ScenariosPassing: 8,
	}
	cached.Timestamp = cached.Timestamp.UTC()
	if err := StoreSnapshot(mosDir, cached); err != nil {
		t.Fatalf("StoreSnapshot: %v", err)
	}

	snap, err := LightweightScan(dir, nil)
	if err != nil {
		t.Fatalf("LightweightScan: %v", err)
	}
	if snap.IntegrityScore != 75.5 {
		t.Errorf("IntegrityScore = %.1f, want 75.5", snap.IntegrityScore)
	}
	if snap.ScenariosTotal != 10 {
		t.Errorf("ScenariosTotal = %d, want 10", snap.ScenariosTotal)
	}
}
