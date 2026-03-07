package harness

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// TrendPoint records the integrity score and its delta from the previous snapshot.
type TrendPoint struct {
	Timestamp      time.Time `json:"timestamp"`
	IntegrityScore float64   `json:"score"`
	Delta          float64   `json:"delta"`
}

// VelocityReport summarizes the rate of integrity index change.
type VelocityReport struct {
	Overall   float64            `json:"overall_velocity"`
	PerVector map[string]float64 `json:"per_vector"`
	Snapshots int                `json:"snapshots_analyzed"`
	TimeSpan  time.Duration      `json:"time_span"`
	ETA       string             `json:"eta,omitempty"`
}

// ComputeTrend converts a time-ordered list of snapshots into trend points
// with consecutive deltas.
func ComputeTrend(snapshots []StateSnapshot) []TrendPoint {
	if len(snapshots) == 0 {
		return nil
	}
	points := make([]TrendPoint, len(snapshots))
	points[0] = TrendPoint{
		Timestamp:      snapshots[0].Timestamp,
		IntegrityScore: snapshots[0].IntegrityScore,
		Delta:          0,
	}
	for i := 1; i < len(snapshots); i++ {
		points[i] = TrendPoint{
			Timestamp:      snapshots[i].Timestamp,
			IntegrityScore: snapshots[i].IntegrityScore,
			Delta:          snapshots[i].IntegrityScore - snapshots[i-1].IntegrityScore,
		}
	}
	return points
}

// ComputeVelocity analyzes snapshots to produce an overall velocity (average
// delta per snapshot), per-vector velocities, and an ETA to score 100.
func ComputeVelocity(snapshots []StateSnapshot) *VelocityReport {
	if len(snapshots) < 2 {
		return &VelocityReport{
			Snapshots: len(snapshots),
			PerVector: make(map[string]float64),
		}
	}

	n := len(snapshots)
	first := snapshots[0]
	last := snapshots[n-1]
	span := last.Timestamp.Sub(first.Timestamp)

	overallDelta := last.IntegrityScore - first.IntegrityScore
	overallVelocity := overallDelta / float64(n-1)

	perVector := make(map[string]float64)
	for _, kind := range AllVectors {
		firstScore := vectorScore(first.Vectors, kind)
		lastScore := vectorScore(last.Vectors, kind)
		perVector[kind] = (lastScore - firstScore) / float64(n-1)
	}

	report := &VelocityReport{
		Overall:   overallVelocity,
		PerVector: perVector,
		Snapshots: n,
		TimeSpan:  span,
	}

	if overallVelocity > 0 {
		remaining := 100.0 - last.IntegrityScore
		snapshotsToGoal := remaining / overallVelocity
		report.ETA = fmt.Sprintf("~%.0f snapshots to 100", math.Ceil(snapshotsToGoal))
	} else if last.IntegrityScore >= 100 {
		report.ETA = "achieved"
	} else if overallVelocity == 0 {
		report.ETA = "stalled"
	} else {
		report.ETA = "diverging"
	}

	return report
}

func vectorScore(vectors []VectorResult, kind string) float64 {
	for _, v := range vectors {
		if v.Kind == kind {
			return v.Score
		}
	}
	return 0
}

var sparkBlocks = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// FormatTrendText renders an ASCII sparkline of integrity scores over time.
func FormatTrendText(points []TrendPoint) string {
	if len(points) == 0 {
		return "No trend data available.\n"
	}

	var b strings.Builder
	b.WriteString("Integrity Trend:\n")

	minScore := 100.0
	maxScore := 0.0
	for _, p := range points {
		if p.IntegrityScore < minScore {
			minScore = p.IntegrityScore
		}
		if p.IntegrityScore > maxScore {
			maxScore = p.IntegrityScore
		}
	}

	span := maxScore - minScore
	if span == 0 {
		span = 1
	}

	b.WriteString("  ")
	for _, p := range points {
		idx := int((p.IntegrityScore - minScore) / span * float64(len(sparkBlocks)-1))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(sparkBlocks) {
			idx = len(sparkBlocks) - 1
		}
		b.WriteRune(sparkBlocks[idx])
	}
	b.WriteString("\n")

	first := points[0]
	last := points[len(points)-1]
	direction := "→"
	if last.IntegrityScore > first.IntegrityScore {
		direction = "↑"
	} else if last.IntegrityScore < first.IntegrityScore {
		direction = "↓"
	}

	fmt.Fprintf(&b, "  %s %.1f → %.1f (%d snapshots)\n",
		direction, first.IntegrityScore, last.IntegrityScore, len(points))

	return b.String()
}

// FormatVelocityText returns a human-readable velocity summary.
func FormatVelocityText(report *VelocityReport) string {
	if report.Snapshots < 2 {
		return "Velocity: insufficient data (need at least 2 snapshots)\n"
	}

	var b strings.Builder
	b.WriteString("Velocity Report:\n")
	fmt.Fprintf(&b, "  Overall:    %+.2f per snapshot (%d snapshots over %s)\n",
		report.Overall, report.Snapshots, formatDuration(report.TimeSpan))

	for _, kind := range AllVectors {
		v := report.PerVector[kind]
		fmt.Fprintf(&b, "  %-12s %+.2f per snapshot\n", kind+":", v)
	}

	fmt.Fprintf(&b, "  ETA:        %s\n", report.ETA)
	return b.String()
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return d.Truncate(time.Second).String()
	}
	if d < time.Hour {
		return fmt.Sprintf("%.0fm", d.Minutes())
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%.1fh", d.Hours())
	}
	return fmt.Sprintf("%.1fd", d.Hours()/24)
}
