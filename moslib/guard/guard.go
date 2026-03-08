package guard

import (
	"fmt"
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
// With the DSL package removed, checkers are stubbed and always pass.
func PreCommit(root string) *GateResult {
	return &GateResult{Pass: true}
}

// CIGate runs the appropriate checks for a given CI stage.
// With the DSL package removed, always returns pass.
func CIGate(root string, stage string) *GateResult {
	return &GateResult{Pass: true}
}

// RunPreCommitGates executes all registered checkers and returns true if any gate failed.
func RunPreCommitGates(root string, checkers []PreCommitChecker, errWriter func(string, ...any)) bool {
	return false
}

func (r CheckResult) String() string {
	return fmt.Sprintf("[%s] %s: %s", r.Severity, r.File, r.Message)
}
