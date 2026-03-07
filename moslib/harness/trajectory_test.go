package harness

import (
	"strings"
	"testing"
	"time"
)

func makeSnapshots(scores ...float64) []StateSnapshot {
	snaps := make([]StateSnapshot, len(scores))
	for i, s := range scores {
		snaps[i] = StateSnapshot{
			Timestamp:      time.Unix(int64(100+i*100), 0).UTC(),
			IntegrityScore: s,
			Vectors: []VectorResult{
				{Kind: VectorFunctional, Score: s},
				{Kind: VectorStructural, Score: s},
				{Kind: VectorPerformance, Score: s},
			},
		}
	}
	return snaps
}

func TestAnalyzeTrajectory_Converging(t *testing.T) {
	snaps := makeSnapshots(90, 85, 80, 75, 70, 65, 60, 55, 50, 45)
	report := AnalyzeTrajectory(snaps, 0)

	if report.Snapshots != 10 {
		t.Errorf("Snapshots = %d, want 10", report.Snapshots)
	}
	found := false
	for _, ax := range report.Axes {
		if ax.Axis == "lint_errors" {
			continue
		}
		if ax.Axis == "drift" {
			continue
		}
		if ax.Status != StatusConverging && ax.Status != StatusRegressing {
			continue
		}
		found = true
	}
	if !found {
		t.Errorf("expected at least one non-stalled axis, got all stalled or insufficient")
	}
}

func TestAnalyzeTrajectory_Stalled(t *testing.T) {
	snaps := makeSnapshots(50, 50, 50, 50, 50, 50, 50, 50, 50, 50)
	report := AnalyzeTrajectory(snaps, 0)

	for _, ax := range report.Axes {
		if ax.Axis == VectorFunctional && ax.Status != StatusStalled {
			t.Errorf("functional status = %s, want stalled", ax.Status)
		}
	}
}

func TestAnalyzeTrajectory_Oscillating(t *testing.T) {
	snaps := makeSnapshots(50, 60, 50, 60, 50, 60, 50, 60, 50, 60)
	report := AnalyzeTrajectory(snaps, 0)

	found := false
	for _, ax := range report.Axes {
		if ax.Status == StatusOscillating {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected at least one oscillating axis")
	}
}

func TestAnalyzeTrajectory_InsufficientData(t *testing.T) {
	snaps := makeSnapshots(50, 60)
	report := AnalyzeTrajectory(snaps, 0)

	if report.Overall != StatusInsufficient {
		t.Errorf("Overall = %s, want insufficient_data", report.Overall)
	}
}

func TestDetectStall(t *testing.T) {
	if !DetectStall([]float64{5, 5, 5, 5}, 1.0) {
		t.Error("expected stall for flat series")
	}
	if DetectStall([]float64{5, 10, 5, 10}, 1.0) {
		t.Error("expected non-stall for varying series")
	}
	if DetectStall([]float64{5}, 1.0) {
		t.Error("expected non-stall for single element")
	}
}

func TestDetectOscillation(t *testing.T) {
	if !DetectOscillation([]float64{5, -5, 5, -5, 5}) {
		t.Error("expected oscillation for alternating deltas")
	}
	if DetectOscillation([]float64{5, 5, 5, 5, 5}) {
		t.Error("expected no oscillation for monotone deltas")
	}
	if DetectOscillation([]float64{5, -5}) {
		t.Error("expected no oscillation for too few data points")
	}
}

func TestBuildSparkline(t *testing.T) {
	s := buildSparkline([]float64{0, 50, 100})
	if len(s) == 0 {
		t.Error("sparkline should not be empty")
	}
	if s[0] == s[len(s)-1] {
		t.Error("sparkline endpoints should differ for 0->100")
	}
}

func TestFormatTrajectoryText_InsufficientData(t *testing.T) {
	report := &TrajectoryReport{Snapshots: 2}
	text := FormatTrajectoryText(report)
	if !strings.Contains(text, "insufficient data") {
		t.Errorf("expected insufficient data message, got %q", text)
	}
}

func TestFormatTrajectoryText_Normal(t *testing.T) {
	report := &TrajectoryReport{
		Overall:   StatusConverging,
		Snapshots: 10,
		Axes: []AxisTrajectory{
			{Axis: VectorFunctional, Status: StatusConverging, Rate: -2.1, Sparkline: "▁▂▃▄▅▆▇█"},
			{Axis: VectorStructural, Status: StatusStalled, Rate: 0.0, Sparkline: "▄▄▄▄▄▄▄▄"},
		},
	}
	text := FormatTrajectoryText(report)
	if !strings.Contains(text, "Trajectory (10 snapshots)") {
		t.Error("missing header")
	}
	if !strings.Contains(text, "converging") {
		t.Error("missing converging status")
	}
	if !strings.Contains(text, "stalled") {
		t.Error("missing stalled status")
	}
}

func TestLinearSlope(t *testing.T) {
	slope := linearSlope([]float64{0, 1, 2, 3, 4})
	if slope < 0.99 || slope > 1.01 {
		t.Errorf("slope = %f, want ~1.0", slope)
	}

	slope = linearSlope([]float64{4, 3, 2, 1, 0})
	if slope > -0.99 || slope < -1.01 {
		t.Errorf("slope = %f, want ~-1.0", slope)
	}

	slope = linearSlope([]float64{5, 5, 5})
	if slope != 0 {
		t.Errorf("slope = %f, want 0", slope)
	}
}
