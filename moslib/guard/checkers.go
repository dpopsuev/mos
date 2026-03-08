package guard

// LintResult is the result of a lint pass for snapshot creation.
// With the linter package removed, this is a stub type.
type LintResult struct {
	Errors             int
	Warnings           int
	DriftViolations    int
	StructuralMetrics map[string]float64
}

// LintChecker implements PreCommitChecker using the linter.
// With the linter package removed, always returns no diagnostics.
type LintChecker struct{}

func (c *LintChecker) Check(root string) ([]CheckResult, error) {
	return nil, nil
}

// HarnessChecker implements PreCommitChecker using harness specs.
// With the harness package removed, always returns no diagnostics.
type HarnessChecker struct{}

func (c *HarnessChecker) Check(root string) ([]CheckResult, error) {
	return nil, nil
}

// LintResultFromDiags runs a lint pass and returns a LintResult for snapshot creation.
// With the linter package removed, returns nil (no snapshot).
func LintResultFromDiags(root string) (*LintResult, error) {
	return nil, nil
}
