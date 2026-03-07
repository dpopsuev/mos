package harness

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"
)

// Baseline stores the last-known-good metric values for a rule.
type Baseline struct {
	RuleID    string             `json:"rule_id"`
	Metrics   map[string]float64 `json:"metrics"`
	Timestamp time.Time          `json:"timestamp"`
}

// RegressionResult records whether a metric regressed beyond the threshold.
type RegressionResult struct {
	MetricName    string  `json:"metric_name"`
	BaselineValue float64 `json:"baseline_value"`
	CurrentValue  float64 `json:"current_value"`
	DeltaPct      float64 `json:"delta_pct"`
	Regressed     bool    `json:"regressed"`
}

const baselineDir = "baselines"

func baselinePath(mosDir, ruleID string) string {
	return filepath.Join(mosDir, baselineDir, ruleID+".json")
}

// StoreBaseline writes metric results as the new baseline for a rule.
func StoreBaseline(mosDir, ruleID string, metrics []MetricResult) error {
	dir := filepath.Join(mosDir, baselineDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating baselines dir: %w", err)
	}

	bl := Baseline{
		RuleID:    ruleID,
		Metrics:   make(map[string]float64, len(metrics)),
		Timestamp: time.Now().UTC(),
	}
	for _, m := range metrics {
		bl.Metrics[m.Name] = m.Value
	}

	data, err := json.MarshalIndent(bl, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(baselinePath(mosDir, ruleID), data, 0o644)
}

// LoadBaseline reads the stored baseline for a rule. Returns nil if none exists.
func LoadBaseline(mosDir, ruleID string) (*Baseline, error) {
	data, err := os.ReadFile(baselinePath(mosDir, ruleID))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var bl Baseline
	if err := json.Unmarshal(data, &bl); err != nil {
		return nil, err
	}
	return &bl, nil
}

// CompareBaseline checks each current metric against the baseline.
// A metric is "regressed" if (current - baseline) / baseline > thresholdPct/100.
// For metrics where lower is better (ns_op, allocs_op), positive delta = regression.
func CompareBaseline(bl *Baseline, current []MetricResult, thresholdPct float64) []RegressionResult {
	var results []RegressionResult
	for _, m := range current {
		base, ok := bl.Metrics[m.Name]
		if !ok {
			continue
		}
		rr := RegressionResult{
			MetricName:    m.Name,
			BaselineValue: base,
			CurrentValue:  m.Value,
		}
		if base != 0 {
			rr.DeltaPct = (m.Value - base) / math.Abs(base) * 100
		}
		if rr.DeltaPct > thresholdPct {
			rr.Regressed = true
		}
		results = append(results, rr)
	}
	return results
}
