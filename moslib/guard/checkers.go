package guard

import (
	"fmt"
	"path/filepath"

	"github.com/dpopsuev/mos/moslib/harness"
	"github.com/dpopsuev/mos/moslib/linter"
	"github.com/dpopsuev/mos/moslib/names"
)

// LintChecker implements PreCommitChecker using the linter.
type LintChecker struct{}

func (c *LintChecker) Check(root string) ([]CheckResult, error) {
	l := &linter.Linter{}
	diags, err := l.Lint(root)
	if err != nil {
		return nil, err
	}
	var results []CheckResult
	for _, d := range diags {
		results = append(results, CheckResult{
			File:     d.File,
			Severity: d.Severity.String(),
			Message:  d.Message,
			Rule:     d.Rule,
		})
	}
	return results, nil
}

// HarnessChecker implements PreCommitChecker using harness specs.
type HarnessChecker struct{}

func (c *HarnessChecker) Check(root string) ([]CheckResult, error) {
	mosDir := filepath.Join(root, names.MosDir)
	specs, err := harness.Discover(mosDir)
	if err != nil {
		return nil, nil
	}
	preCommit := harness.FilterByTrigger(specs, "pre-commit")
	if len(preCommit) == 0 {
		return nil, nil
	}
	var results []CheckResult
	evResults := harness.Run(root, preCommit)
	for _, ev := range evResults {
		if !ev.Pass {
			msg := fmt.Sprintf("harness %q: FAIL (exit %d)", ev.RuleID, ev.ExitCode)
			if ev.Stderr != "" {
				msg += ": " + ev.Stderr
			}
			results = append(results, CheckResult{
				Severity: "error",
				Message:  msg,
				Rule:     ev.RuleID,
			})
		}
	}
	return results, nil
}

// LintResultFromDiags runs a lint pass and returns a LintResult for snapshot creation.
func LintResultFromDiags(root string) (*harness.LintResult, error) {
	l := &linter.Linter{}
	diags, err := l.Lint(root)
	if err != nil {
		return nil, err
	}
	result := &harness.LintResult{}
	structMetrics := map[string]float64{}
	for _, d := range diags {
		switch d.Severity {
		case linter.SeverityError:
			result.Errors++
		case linter.SeverityWarning:
			result.Warnings++
		}
		if d.Rule == "structural-drift" || d.Rule == "arch-forbidden" {
			result.DriftViolations++
		}
		switch d.Rule {
		case "nesting-depth", "function-length", "params-per-function",
			"fan-out", "loc-per-file":
			structMetrics[d.Rule]++
		}
	}
	if len(structMetrics) > 0 {
		result.StructuralMetrics = structMetrics
	}
	return result, nil
}
