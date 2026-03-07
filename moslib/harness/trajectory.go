package harness

import (
	"fmt"
	"math"
	"strings"
)

// TrajectoryStatus classifies the movement pattern on an axis.
type TrajectoryStatus string

const (
	StatusConverging   TrajectoryStatus = "converging"
	StatusStalled      TrajectoryStatus = "stalled"
	StatusOscillating  TrajectoryStatus = "oscillating"
	StatusRegressing   TrajectoryStatus = "regressing"
	StatusInsufficient TrajectoryStatus = "insufficient_data"
)

// AxisTrajectory describes the trajectory classification for a single axis.
type AxisTrajectory struct {
	Axis      string           `json:"axis"`
	Status    TrajectoryStatus `json:"status"`
	Rate      float64          `json:"rate"`
	Window    int              `json:"window"`
	Sparkline string           `json:"sparkline"`
}

// TrajectoryReport is the full trajectory analysis across all axes.
type TrajectoryReport struct {
	Overall   TrajectoryStatus `json:"overall"`
	Axes      []AxisTrajectory `json:"axes"`
	Snapshots int              `json:"snapshots_analyzed"`
}

const defaultTrajectoryWindow = 10

// AnalyzeTrajectory classifies the trajectory for each quality axis and
// the aggregate integrity score over the given snapshot series.
func AnalyzeTrajectory(snapshots []StateSnapshot, window int) *TrajectoryReport {
	if window <= 0 {
		window = defaultTrajectoryWindow
	}
	report := &TrajectoryReport{Snapshots: len(snapshots)}

	if len(snapshots) < 3 {
		report.Overall = StatusInsufficient
		return report
	}

	tail := snapshots
	if len(tail) > window {
		tail = tail[len(tail)-window:]
	}

	axisNames := []string{VectorFunctional, VectorStructural, VectorPerformance, "lint_errors", "drift"}
	for _, axis := range axisNames {
		scores := extractAxisSeries(tail, axis)
		at := classifyAxis(axis, scores, len(tail))
		report.Axes = append(report.Axes, at)
	}

	report.Overall = deriveOverall(report.Axes)
	return report
}

func extractAxisSeries(snapshots []StateSnapshot, axis string) []float64 {
	scores := make([]float64, len(snapshots))
	for i, s := range snapshots {
		switch axis {
		case VectorFunctional, VectorStructural, VectorPerformance:
			scores[i] = vectorScore(s.Vectors, axis)
		case "lint_errors":
			scores[i] = float64(s.LintErrors)
		case "drift":
			scores[i] = float64(s.DriftViolations)
		default:
			scores[i] = s.IntegrityScore
		}
	}
	return scores
}

func classifyAxis(axis string, scores []float64, window int) AxisTrajectory {
	at := AxisTrajectory{
		Axis:      axis,
		Window:    window,
		Sparkline: buildSparkline(scores),
	}

	if len(scores) < 3 {
		at.Status = StatusInsufficient
		return at
	}

	deltas := make([]float64, len(scores)-1)
	for i := 1; i < len(scores); i++ {
		deltas[i-1] = scores[i] - scores[i-1]
	}

	at.Rate = linearSlope(scores)

	inverted := axis == "lint_errors" || axis == "drift"

	if DetectStall(scores, 1.0) {
		at.Status = StatusStalled
	} else if DetectOscillation(deltas) {
		at.Status = StatusOscillating
	} else if (inverted && at.Rate > 0.5) || (!inverted && at.Rate < -0.5) {
		at.Status = StatusRegressing
	} else if (inverted && at.Rate < -0.5) || (!inverted && at.Rate > 0.5) {
		at.Status = StatusConverging
	} else {
		at.Status = StatusStalled
	}

	return at
}

// DetectStall returns true if the series is flat (max-min < epsilon).
func DetectStall(scores []float64, epsilon float64) bool {
	if len(scores) < 2 {
		return false
	}
	mn, mx := scores[0], scores[0]
	for _, v := range scores[1:] {
		if v < mn {
			mn = v
		}
		if v > mx {
			mx = v
		}
	}
	return (mx - mn) < epsilon
}

// DetectOscillation returns true if consecutive deltas alternate sign
// for at least 4 consecutive pairs.
func DetectOscillation(deltas []float64) bool {
	if len(deltas) < 4 {
		return false
	}
	consecutive := 0
	for i := 1; i < len(deltas); i++ {
		if (deltas[i-1] > 0 && deltas[i] < 0) || (deltas[i-1] < 0 && deltas[i] > 0) {
			consecutive++
			if consecutive >= 4 {
				return true
			}
		} else {
			consecutive = 0
		}
	}
	return false
}

func linearSlope(values []float64) float64 {
	n := float64(len(values))
	if n < 2 {
		return 0
	}
	var sumX, sumY, sumXY, sumX2 float64
	for i, y := range values {
		x := float64(i)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	denom := n*sumX2 - sumX*sumX
	if math.Abs(denom) < 1e-12 {
		return 0
	}
	return (n*sumXY - sumX*sumY) / denom
}

func buildSparkline(scores []float64) string {
	if len(scores) == 0 {
		return ""
	}
	mn, mx := scores[0], scores[0]
	for _, v := range scores[1:] {
		if v < mn {
			mn = v
		}
		if v > mx {
			mx = v
		}
	}
	span := mx - mn
	if span == 0 {
		span = 1
	}
	var b strings.Builder
	for _, v := range scores {
		idx := int((v - mn) / span * float64(len(sparkBlocks)-1))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(sparkBlocks) {
			idx = len(sparkBlocks) - 1
		}
		b.WriteRune(sparkBlocks[idx])
	}
	return b.String()
}

func deriveOverall(axes []AxisTrajectory) TrajectoryStatus {
	counts := map[TrajectoryStatus]int{}
	for _, a := range axes {
		counts[a.Status]++
	}
	if counts[StatusRegressing] > 0 {
		return StatusRegressing
	}
	if counts[StatusOscillating] > len(axes)/2 {
		return StatusOscillating
	}
	if counts[StatusConverging] > len(axes)/2 {
		return StatusConverging
	}
	if counts[StatusStalled] > len(axes)/2 {
		return StatusStalled
	}
	if counts[StatusInsufficient] == len(axes) {
		return StatusInsufficient
	}
	return StatusStalled
}

// ANSI color helpers for terminal output.
const (
	ansiReset  = "\033[0m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiRed    = "\033[31m"
	ansiCyan   = "\033[36m"
)

func statusColor(s TrajectoryStatus) string {
	switch s {
	case StatusConverging:
		return ansiGreen
	case StatusStalled:
		return ansiYellow
	case StatusRegressing:
		return ansiRed
	case StatusOscillating:
		return ansiCyan
	default:
		return ""
	}
}

// FormatTrajectoryText renders the trajectory report as colored ASCII sparklines.
func FormatTrajectoryText(report *TrajectoryReport) string {
	if report.Snapshots < 3 {
		return "Trajectory: insufficient data (need at least 3 snapshots)\n"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Trajectory (%d snapshots):\n", report.Snapshots)
	fmt.Fprintf(&b, "  %-14s  %-10s  %-13s  %s\n", "AXIS", "TREND", "STATUS", "RATE")
	fmt.Fprintf(&b, "  %-14s  %-10s  %-13s  %s\n", "----", "-----", "------", "----")

	for _, ax := range report.Axes {
		color := statusColor(ax.Status)
		reset := ansiReset
		if color == "" {
			reset = ""
		}
		fmt.Fprintf(&b, "  %-14s  %s%-10s%s  %-13s  %+.1f/snap\n",
			ax.Axis, color, ax.Sparkline, reset, ax.Status, ax.Rate)
	}

	return b.String()
}
