package guard

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dpopsuev/mos/moslib/harness"
	"github.com/dpopsuev/mos/moslib/linter"
	"github.com/dpopsuev/mos/moslib/names"
)

// CheckResult represents a single diagnostic from a pre-commit check.
type CheckResult struct {
	File     string
	Severity string // "error", "warning", "info"
	Message  string
	Rule     string
}

// PreCommitChecker validates project state before a commit.
type PreCommitChecker interface {
	Check(root string) ([]CheckResult, error)
}

// GateResult aggregates the outcome of an enforcement gate.
type GateResult struct {
	Pass         bool          `json:"pass"`
	Diagnostics  []CheckResult `json:"diagnostics,omitempty"`
	ErrorCount   int           `json:"error_count"`
	WarningCount int           `json:"warning_count"`
	Output       string        `json:"output,omitempty"`
}

// PreCommit runs lint + pre-commit harness checks and returns a unified result.
func PreCommit(root string) *GateResult {
	checkers := []PreCommitChecker{
		&LintChecker{},
		&HarnessChecker{},
	}
	result := &GateResult{Pass: true}
	for _, checker := range checkers {
		checks, err := checker.Check(root)
		if err != nil {
			result.Pass = false
			result.Diagnostics = append(result.Diagnostics, CheckResult{
				Severity: "error",
				Message:  fmt.Sprintf("pre-commit check: %v", err),
			})
			result.ErrorCount++
			continue
		}
		for _, r := range checks {
			result.Diagnostics = append(result.Diagnostics, r)
			switch r.Severity {
			case "error":
				result.ErrorCount++
				result.Pass = false
			case "warning":
				result.WarningCount++
			}
		}
	}
	return result
}

// CIGate runs the appropriate checks for a given CI stage.
// Supported stages: "lint", "audit", "harness", "vectors".
func CIGate(root string, stage string) *GateResult {
	switch stage {
	case "lint":
		return ciLint(root)
	case "harness":
		return ciHarness(root)
	case "vectors":
		return ciVectors(root)
	default:
		return &GateResult{
			Pass:   false,
			Output: fmt.Sprintf("unknown CI stage: %s", stage),
		}
	}
}

func ciLint(root string) *GateResult {
	l := &linter.Linter{}
	diags, err := l.Lint(root)
	if err != nil {
		return &GateResult{Pass: false, Output: err.Error(), ErrorCount: 1}
	}
	result := &GateResult{Pass: true}
	var buf strings.Builder
	for _, d := range diags {
		cr := CheckResult{
			File:     d.File,
			Severity: d.Severity.String(),
			Message:  d.Message,
			Rule:     d.Rule,
		}
		result.Diagnostics = append(result.Diagnostics, cr)
		if d.Severity == linter.SeverityError {
			result.ErrorCount++
			result.Pass = false
			fmt.Fprintf(&buf, "%s: %s [%s] %s\n", d.File, d.Severity, d.Rule, d.Message)
		} else if d.Severity == linter.SeverityWarning {
			result.WarningCount++
		}
	}
	result.Output = buf.String()
	return result
}

func ciHarness(root string) *GateResult {
	mosDir := filepath.Join(root, names.MosDir)
	specs, err := harness.Discover(mosDir)
	if err != nil {
		return &GateResult{Pass: false, Output: err.Error(), ErrorCount: 1}
	}
	results := harness.Run(root, specs)
	gate := &GateResult{Pass: true}
	for _, ev := range results {
		if !ev.Pass {
			gate.Pass = false
			gate.ErrorCount++
			gate.Diagnostics = append(gate.Diagnostics, CheckResult{
				Severity: "error",
				Message:  fmt.Sprintf("harness %q: FAIL (exit %d)", ev.RuleID, ev.ExitCode),
				Rule:     ev.RuleID,
			})
		}
	}
	gate.Output = harness.FormatText(results)
	return gate
}

func ciVectors(root string) *GateResult {
	mosDir := filepath.Join(root, names.MosDir)
	vectors, err := harness.EvaluateVectors(root, mosDir)
	if err != nil {
		return &GateResult{Pass: false, Output: err.Error(), ErrorCount: 1}
	}
	gate := &GateResult{Pass: true}
	for _, v := range vectors {
		if !v.Pass {
			gate.Pass = false
			gate.ErrorCount++
			gate.Diagnostics = append(gate.Diagnostics, CheckResult{
				Severity: "error",
				Message:  fmt.Sprintf("vector %q: FAIL (%.1f)", v.Kind, v.Score),
				Rule:     "vector-" + v.Kind,
			})
		}
	}
	gate.Output = harness.FormatVectorsText(vectors)
	return gate
}

// RunPreCommitGates executes all registered checkers and returns true if any gate failed.
func RunPreCommitGates(root string, checkers []PreCommitChecker, errWriter func(string, ...any)) bool {
	failed := false
	for _, checker := range checkers {
		results, err := checker.Check(root)
		if err != nil {
			errWriter("pre-commit check: %v\n", err)
			failed = true
			continue
		}
		for _, r := range results {
			if r.Severity == "error" {
				errWriter("  [%s] %s: %s\n", r.Severity, r.File, r.Message)
				failed = true
			}
		}
	}
	return failed
}

func (r CheckResult) String() string {
	return fmt.Sprintf("[%s] %s: %s", r.Severity, r.File, r.Message)
}
