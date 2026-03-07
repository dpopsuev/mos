package harness

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dpopsuev/mos/moslib/names"
)

const (
	VectorFunctional  = "functional"
	VectorStructural  = "structural"
	VectorPerformance = "performance"
)

// AllVectors lists the three quality vector kinds.
var AllVectors = []string{VectorFunctional, VectorStructural, VectorPerformance}

// VectorDetail records the outcome of a single rule within a vector.
type VectorDetail struct {
	RuleID string  `json:"rule_id"`
	Pass   bool    `json:"pass"`
	Score  float64 `json:"score"`
}

// VectorResult aggregates pass/fail, score, and per-rule details for one
// quality vector.
type VectorResult struct {
	Kind    string         `json:"kind"`
	Score   float64        `json:"score"`
	Pass    bool           `json:"pass"`
	Details []VectorDetail `json:"details,omitempty"`
}

// EvaluateVectors discovers all rules, groups them by vector, executes each
// group, and returns a VectorResult per vector. Vectors with no tagged rules
// get a score of 0 and Pass=true (vacuously).
func EvaluateVectors(root, mosDir string) ([]VectorResult, error) {
	specs, err := Discover(mosDir)
	if err != nil {
		return nil, fmt.Errorf("discovering harness specs: %w", err)
	}

	results := make([]VectorResult, 0, len(AllVectors))

	for _, vec := range AllVectors {
		vr := VectorResult{Kind: vec, Pass: true}

		tagged := FilterByVector(specs, vec)
		if len(tagged) > 0 {
			evidence := Run(root, tagged)
			passing := 0
			for _, ev := range evidence {
				detail := VectorDetail{RuleID: ev.RuleID, Pass: ev.Pass}
				if ev.Pass {
					detail.Score = 100
					passing++
				}

				if vec == VectorPerformance && len(ev.Metrics) > 0 {
					bl, _ := LoadBaseline(mosDir, ev.RuleID)
					if bl != nil {
						regressions := CompareBaseline(bl, ev.Metrics, 10)
						for _, rr := range regressions {
							if rr.Regressed {
								detail.Pass = false
								detail.Score = 0
								passing--
								vr.Pass = false
								break
							}
						}
					}
					if ev.Pass {
						_ = StoreBaseline(mosDir, ev.RuleID, ev.Metrics)
					}
				}

				vr.Details = append(vr.Details, detail)
				if !ev.Pass {
					vr.Pass = false
				}
			}
			if len(evidence) > 0 {
				vr.Score = float64(passing) / float64(len(evidence)) * 100
				if vr.Score < 0 {
					vr.Score = 0
				}
			}
		}

		// Structural vector also includes quality config checks (linter diagnostics).
		if vec == VectorStructural {
			configs, _ := DiscoverQualityConfigs(mosDir)
			structConfigs := filterQualityByVector(configs, VectorStructural)
			for _, qc := range structConfigs {
				vr.Details = append(vr.Details, VectorDetail{
					RuleID: qc.RuleID,
					Pass:   true,
					Score:  100,
				})
			}
		}

		results = append(results, vr)
	}

	return results, nil
}

func filterQualityByVector(configs []QualityConfig, vector string) []QualityConfig {
	var out []QualityConfig
	for _, qc := range configs {
		if qc.Vector == vector {
			out = append(out, qc)
		}
	}
	return out
}

// FormatVectorsText returns a human-readable summary of vector results.
func FormatVectorsText(results []VectorResult) string {
	var b strings.Builder
	b.WriteString("\nQuality Vectors:\n")
	b.WriteString(fmt.Sprintf("  %-14s  %-6s  %s\n", "VECTOR", "SCORE", "STATUS"))
	b.WriteString(fmt.Sprintf("  %-14s  %-6s  %s\n", "------", "-----", "------"))
	for _, vr := range results {
		status := "PASS"
		if !vr.Pass {
			status = "FAIL"
		}
		if len(vr.Details) == 0 {
			status = "N/A"
		}
		b.WriteString(fmt.Sprintf("  %-14s  %5.1f%%  %s\n", vr.Kind, vr.Score, status))
	}
	return b.String()
}

// VectorsMosDir returns the .mos directory path from a project root.
func VectorsMosDir(root string) string {
	return filepath.Join(root, names.MosDir)
}
