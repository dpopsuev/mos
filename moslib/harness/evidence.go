package harness

import "time"

// HarnessSpec is a discovered harness block from a rule artifact.
type HarnessSpec struct {
	RuleID      string
	Command     string
	Timeout     time.Duration
	Enforcement string // "error" or "warning"
	Trigger     string // "manual" (default), "pre-commit", "post-commit"
	Vector      string // "functional", "structural", "performance", or "" (untagged)
	Thresholds  []MetricThreshold
}

// FilterByVector returns specs matching the given vector value.
func FilterByVector(specs []HarnessSpec, vector string) []HarnessSpec {
	var out []HarnessSpec
	for _, s := range specs {
		if s.Vector == vector {
			out = append(out, s)
		}
	}
	return out
}

// Filter returns a subset of specs matching the given criteria.
// Empty ruleIDs means "all rules". Empty minEnforcement means "all levels".
func Filter(specs []HarnessSpec, ruleIDs []string, minEnforcement string) []HarnessSpec {
	if len(ruleIDs) == 0 && minEnforcement == "" {
		return specs
	}

	idSet := make(map[string]bool, len(ruleIDs))
	for _, id := range ruleIDs {
		idSet[id] = true
	}

	var out []HarnessSpec
	for _, s := range specs {
		if len(idSet) > 0 && !idSet[s.RuleID] {
			continue
		}
		if minEnforcement == "error" && s.Enforcement != "error" {
			continue
		}
		out = append(out, s)
	}
	return out
}

// FilterByTrigger returns specs matching the given trigger value (e.g. "pre-commit").
func FilterByTrigger(specs []HarnessSpec, trigger string) []HarnessSpec {
	var out []HarnessSpec
	for _, s := range specs {
		if s.Trigger == trigger {
			out = append(out, s)
		}
	}
	return out
}

// Evidence is the result of executing a single harness command.
type Evidence struct {
	RuleID   string         `json:"rule_id"`
	Command  string         `json:"command"`
	ExitCode int            `json:"exit_code"`
	Stdout   string         `json:"stdout,omitempty"`
	Stderr   string         `json:"stderr,omitempty"`
	Duration time.Duration  `json:"duration_ms"`
	Pass     bool           `json:"pass"`
	TimedOut bool           `json:"timed_out,omitempty"`
	Metrics  []MetricResult `json:"metrics,omitempty"`
}
