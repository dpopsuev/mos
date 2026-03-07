package linter

import (
	"fmt"
	"path/filepath"

	"github.com/dpopsuev/mos/moslib/harness"
)

const (
	minSnapshotsForTrajectory = 5
	stallWindow               = 8
	churnWindow               = 6
	regressionWindow          = 5
)

// validateTrajectory loads commit snapshots and flags trajectory anomalies
// as lint warnings. Gracefully returns nil if insufficient history exists.
func validateTrajectory(root, mosDir string) []Diagnostic {
	snapshots, err := harness.LoadAllSnapshots(mosDir)
	if err != nil || len(snapshots) < minSnapshotsForTrajectory {
		return nil
	}

	report := harness.AnalyzeTrajectory(snapshots, 0)
	if report == nil {
		return nil
	}

	var diags []Diagnostic
	for _, ax := range report.Axes {
		switch ax.Status {
		case harness.StatusStalled:
			if ax.Window >= stallWindow {
				diags = append(diags, Diagnostic{
					File:     filepath.Join(mosDir, "vcs", "snapshots"),
					Severity: SeverityWarning,
					Rule:     "trajectory-stall",
					Message:  fmt.Sprintf("axis %q stalled for %d snapshots (rate: %+.2f)", ax.Axis, ax.Window, ax.Rate),
				})
			}
		case harness.StatusOscillating:
			if ax.Window >= churnWindow {
				diags = append(diags, Diagnostic{
					File:     filepath.Join(mosDir, "vcs", "snapshots"),
					Severity: SeverityWarning,
					Rule:     "trajectory-churn",
					Message:  fmt.Sprintf("axis %q oscillating over %d snapshots (rate: %+.2f)", ax.Axis, ax.Window, ax.Rate),
				})
			}
		case harness.StatusRegressing:
			if ax.Window >= regressionWindow {
				diags = append(diags, Diagnostic{
					File:     filepath.Join(mosDir, "vcs", "snapshots"),
					Severity: SeverityWarning,
					Rule:     "trajectory-regression",
					Message:  fmt.Sprintf("axis %q regressing over %d snapshots (rate: %+.2f)", ax.Axis, ax.Window, ax.Rate),
				})
			}
		}
	}
	return diags
}
