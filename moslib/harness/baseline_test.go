package harness

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBaselineRoundTrip(t *testing.T) {
	dir := t.TempDir()
	mosDir := filepath.Join(dir, ".mos")
	if err := os.MkdirAll(mosDir, 0o755); err != nil {
		t.Fatal(err)
	}

	metrics := []MetricResult{
		{Name: "ns_op", Value: 1000, Unit: "ns"},
		{Name: "allocs_op", Value: 5, Unit: "allocs"},
	}

	if err := StoreBaseline(mosDir, "bench-rule", metrics); err != nil {
		t.Fatalf("StoreBaseline: %v", err)
	}

	bl, err := LoadBaseline(mosDir, "bench-rule")
	if err != nil {
		t.Fatalf("LoadBaseline: %v", err)
	}
	if bl == nil {
		t.Fatal("expected baseline, got nil")
	}
	if bl.RuleID != "bench-rule" {
		t.Errorf("expected rule_id bench-rule, got %s", bl.RuleID)
	}
	if bl.Metrics["ns_op"] != 1000 {
		t.Errorf("expected ns_op=1000, got %.2f", bl.Metrics["ns_op"])
	}
	if bl.Metrics["allocs_op"] != 5 {
		t.Errorf("expected allocs_op=5, got %.2f", bl.Metrics["allocs_op"])
	}
}

func TestLoadBaselineNotFound(t *testing.T) {
	dir := t.TempDir()
	bl, err := LoadBaseline(dir, "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bl != nil {
		t.Error("expected nil baseline for nonexistent rule")
	}
}

func TestCompareBaselineNoRegression(t *testing.T) {
	bl := &Baseline{
		Metrics: map[string]float64{"ns_op": 1000, "allocs_op": 5},
	}
	current := []MetricResult{
		{Name: "ns_op", Value: 1050},
		{Name: "allocs_op", Value: 5},
	}

	results := CompareBaseline(bl, current, 10)
	for _, r := range results {
		if r.Regressed {
			t.Errorf("metric %s should not regress (delta=%.1f%%)", r.MetricName, r.DeltaPct)
		}
	}
}

func TestCompareBaselineRegression(t *testing.T) {
	bl := &Baseline{
		Metrics: map[string]float64{"ns_op": 1000},
	}
	current := []MetricResult{
		{Name: "ns_op", Value: 1200},
	}

	results := CompareBaseline(bl, current, 10)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Regressed {
		t.Error("expected regression (20% > 10% threshold)")
	}
	if results[0].DeltaPct < 19 || results[0].DeltaPct > 21 {
		t.Errorf("expected ~20%% delta, got %.1f%%", results[0].DeltaPct)
	}
}
