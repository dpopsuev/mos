package guard

import (
	"testing"
)

type passChecker struct{}

func (c *passChecker) Check(root string) ([]CheckResult, error) {
	return nil, nil
}

type failChecker struct{}

func (c *failChecker) Check(root string) ([]CheckResult, error) {
	return []CheckResult{
		{File: "a.mos", Severity: "error", Message: "test error", Rule: "test-rule"},
		{File: "b.mos", Severity: "warning", Message: "test warning", Rule: "test-rule2"},
	}, nil
}

func TestRunPreCommitGates_AllPass(t *testing.T) {
	failed := RunPreCommitGates(".", []PreCommitChecker{&passChecker{}}, func(string, ...any) {})
	if failed {
		t.Error("expected pass, got fail")
	}
}

func TestRunPreCommitGates_HasError(t *testing.T) {
	failed := RunPreCommitGates(".", []PreCommitChecker{&failChecker{}}, func(string, ...any) {})
	if !failed {
		t.Error("expected fail, got pass")
	}
}

func TestGateResult_FromCheckers(t *testing.T) {
	result := runCheckers([]PreCommitChecker{&failChecker{}})
	if result.Pass {
		t.Error("expected fail")
	}
	if result.ErrorCount != 1 {
		t.Errorf("ErrorCount = %d, want 1", result.ErrorCount)
	}
	if result.WarningCount != 1 {
		t.Errorf("WarningCount = %d, want 1", result.WarningCount)
	}
	if len(result.Diagnostics) != 2 {
		t.Errorf("Diagnostics count = %d, want 2", len(result.Diagnostics))
	}
}

func TestGateResult_AllPass(t *testing.T) {
	result := runCheckers([]PreCommitChecker{&passChecker{}})
	if !result.Pass {
		t.Error("expected pass")
	}
	if result.ErrorCount != 0 {
		t.Errorf("ErrorCount = %d, want 0", result.ErrorCount)
	}
}

func TestCIGate_UnknownStage(t *testing.T) {
	result := CIGate(".", "nonexistent")
	if result.Pass {
		t.Error("expected fail for unknown stage")
	}
}

func TestCheckResult_String(t *testing.T) {
	r := CheckResult{File: "x.mos", Severity: "error", Message: "bad"}
	s := r.String()
	if s != "[error] x.mos: bad" {
		t.Errorf("String() = %q", s)
	}
}

func runCheckers(checkers []PreCommitChecker) *GateResult {
	result := &GateResult{Pass: true}
	for _, checker := range checkers {
		checks, err := checker.Check(".")
		if err != nil {
			result.Pass = false
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
