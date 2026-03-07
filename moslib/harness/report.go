package harness

import (
	"encoding/json"
	"fmt"
	"strings"
)

// FormatText formats evidence as a human-readable table.
func FormatText(results []Evidence) string {
	if len(results) == 0 {
		return "No harness specs found.\n"
	}

	var b strings.Builder

	ruleWidth := len("RULE")
	cmdWidth := len("COMMAND")
	for _, ev := range results {
		if len(ev.RuleID) > ruleWidth {
			ruleWidth = len(ev.RuleID)
		}
		if len(ev.Command) > cmdWidth {
			cmdWidth = len(ev.Command)
		}
	}

	header := fmt.Sprintf("%-*s  %-*s  %-6s  %s\n",
		ruleWidth, "RULE", cmdWidth, "COMMAND", "STATUS", "DURATION")
	b.WriteString(header)
	b.WriteString(strings.Repeat("─", len(header)-1) + "\n")

	for _, ev := range results {
		status := "PASS"
		if !ev.Pass {
			if ev.TimedOut {
				status = "TIMEOUT"
			} else {
				status = fmt.Sprintf("FAIL(%d)", ev.ExitCode)
			}
		}

		b.WriteString(fmt.Sprintf("%-*s  %-*s  %-6s  %s\n",
			ruleWidth, ev.RuleID,
			cmdWidth, ev.Command,
			status,
			ev.Duration.Round(1_000_000).String()))

		for _, mr := range ev.Metrics {
			tag := "OK"
			if !mr.Pass {
				tag = "BREACH"
			}
			b.WriteString(fmt.Sprintf("  metric %-16s  %10.2f %-6s  [%s]", mr.Name, mr.Value, mr.Unit, tag))
			if mr.Message != "" {
				b.WriteString("  " + mr.Message)
			}
			b.WriteString("\n")
		}
	}

	passed := 0
	for _, ev := range results {
		if ev.Pass {
			passed++
		}
	}
	b.WriteString(fmt.Sprintf("\n%d/%d passed\n", passed, len(results)))

	return b.String()
}

// jsonEvidence mirrors Evidence but with duration as int64 milliseconds for JSON.
type jsonEvidence struct {
	RuleID     string         `json:"rule_id"`
	Command    string         `json:"command"`
	ExitCode   int            `json:"exit_code"`
	DurationMs int64          `json:"duration_ms"`
	Pass       bool           `json:"pass"`
	TimedOut   bool           `json:"timed_out,omitempty"`
	Stdout     string         `json:"stdout,omitempty"`
	Stderr     string         `json:"stderr,omitempty"`
	Metrics    []MetricResult `json:"metrics,omitempty"`
}

// FormatJSON formats evidence as a JSON array.
func FormatJSON(results []Evidence) ([]byte, error) {
	out := make([]jsonEvidence, len(results))
	for i, ev := range results {
		out[i] = jsonEvidence{
			RuleID:     ev.RuleID,
			Command:    ev.Command,
			ExitCode:   ev.ExitCode,
			DurationMs: ev.Duration.Milliseconds(),
			Pass:       ev.Pass,
			TimedOut:   ev.TimedOut,
			Stdout:     ev.Stdout,
			Stderr:     ev.Stderr,
			Metrics:    ev.Metrics,
		}
	}
	return json.MarshalIndent(out, "", "  ")
}
