package harness

import (
	"path/filepath"
	"time"

	"github.com/dpopsuev/mos/moslib/names"
)

// LintResult captures the minimal lint output needed by LightweightScan,
// avoiding a direct dependency on the linter package.
type LintResult struct {
	Errors            int
	Warnings          int
	DriftViolations   int
	StructuralMetrics map[string]float64
}

// LintFunc is the signature for a function that runs lint and returns
// a lightweight result. Injected by the caller to break the import cycle.
type LintFunc func(root string) (*LintResult, error)

// LightweightScan runs a fast governance-only scan without executing harness
// commands. It produces a StateSnapshot capturing lint counts, drift violations,
// and the latest cached integrity score. Target latency: under 3 seconds.
//
// The lintFn parameter is injected by the caller (typically from the CLI layer)
// to avoid a circular dependency between harness and linter.
func LightweightScan(root string, lintFn LintFunc) (*StateSnapshot, error) {
	mosDir := filepath.Join(root, names.MosDir)
	snap := &StateSnapshot{
		Timestamp: time.Now().UTC(),
	}

	if lintFn != nil {
		lr, err := lintFn(root)
		if err == nil && lr != nil {
			snap.LintErrors = lr.Errors
			snap.LintWarnings = lr.Warnings
			snap.DriftViolations = lr.DriftViolations
			snap.StructuralMetrics = lr.StructuralMetrics
		}
	}

	latest, _ := LoadSnapshots(mosDir)
	if len(latest) > 0 {
		last := latest[len(latest)-1]
		snap.IntegrityScore = last.IntegrityScore
		snap.Vectors = last.Vectors
		snap.ScenariosTotal = last.ScenariosTotal
		snap.ScenariosPassing = last.ScenariosPassing
		snap.HarnessPassCount = last.HarnessPassCount
		snap.HarnessFailCount = last.HarnessFailCount
	}

	return snap, nil
}
