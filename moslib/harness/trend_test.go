package harness

import (
	"math"
	"strings"
	"testing"
	"time"
)

func TestComputeTrend_Deltas(t *testing.T) {
	snapshots := []StateSnapshot{
		{Timestamp: time.Unix(100, 0), IntegrityScore: 50},
		{Timestamp: time.Unix(200, 0), IntegrityScore: 60},
		{Timestamp: time.Unix(300, 0), IntegrityScore: 55},
	}

	points := ComputeTrend(snapshots)
	if len(points) != 3 {
		t.Fatalf("len = %d, want 3", len(points))
	}
	if points[0].Delta != 0 {
		t.Errorf("points[0].Delta = %.1f, want 0", points[0].Delta)
	}
	if points[1].Delta != 10 {
		t.Errorf("points[1].Delta = %.1f, want 10", points[1].Delta)
	}
	if points[2].Delta != -5 {
		t.Errorf("points[2].Delta = %.1f, want -5", points[2].Delta)
	}
}

func TestComputeTrend_Empty(t *testing.T) {
	points := ComputeTrend(nil)
	if points != nil {
		t.Errorf("expected nil, got %v", points)
	}
}

func TestComputeVelocity_Basic(t *testing.T) {
	snapshots := []StateSnapshot{
		{
			Timestamp:      time.Unix(100, 0),
			IntegrityScore: 40,
			Vectors: []VectorResult{
				{Kind: VectorFunctional, Score: 60},
				{Kind: VectorStructural, Score: 30},
				{Kind: VectorPerformance, Score: 30},
			},
		},
		{
			Timestamp:      time.Unix(200, 0),
			IntegrityScore: 60,
			Vectors: []VectorResult{
				{Kind: VectorFunctional, Score: 80},
				{Kind: VectorStructural, Score: 50},
				{Kind: VectorPerformance, Score: 50},
			},
		},
	}

	vel := ComputeVelocity(snapshots)
	if vel.Snapshots != 2 {
		t.Errorf("Snapshots = %d, want 2", vel.Snapshots)
	}
	if math.Abs(vel.Overall-20.0) > 0.01 {
		t.Errorf("Overall = %.2f, want 20.0", vel.Overall)
	}
	if math.Abs(vel.PerVector[VectorFunctional]-20.0) > 0.01 {
		t.Errorf("functional velocity = %.2f, want 20.0", vel.PerVector[VectorFunctional])
	}
	if vel.ETA == "" {
		t.Error("ETA should not be empty")
	}
}

func TestComputeVelocity_InsufficientData(t *testing.T) {
	vel := ComputeVelocity([]StateSnapshot{{Timestamp: time.Unix(100, 0)}})
	if vel.Snapshots != 1 {
		t.Errorf("Snapshots = %d, want 1", vel.Snapshots)
	}
	if vel.Overall != 0 {
		t.Errorf("Overall = %.2f, want 0", vel.Overall)
	}
}

func TestComputeVelocity_Diverging(t *testing.T) {
	snapshots := []StateSnapshot{
		{Timestamp: time.Unix(100, 0), IntegrityScore: 80},
		{Timestamp: time.Unix(200, 0), IntegrityScore: 60},
	}
	vel := ComputeVelocity(snapshots)
	if vel.ETA != "diverging" {
		t.Errorf("ETA = %q, want %q", vel.ETA, "diverging")
	}
}

func TestFormatTrendText_Sparkline(t *testing.T) {
	points := []TrendPoint{
		{IntegrityScore: 10},
		{IntegrityScore: 50},
		{IntegrityScore: 90},
	}
	text := FormatTrendText(points)
	if !strings.Contains(text, "Integrity Trend:") {
		t.Error("missing header")
	}
	if !strings.Contains(text, "↑") {
		t.Error("missing upward arrow for increasing trend")
	}
}

func TestFormatTrendText_Empty(t *testing.T) {
	text := FormatTrendText(nil)
	if !strings.Contains(text, "No trend data") {
		t.Errorf("got %q, expected no-data message", text)
	}
}

func TestFormatVelocityText_InsufficientData(t *testing.T) {
	report := &VelocityReport{Snapshots: 1, PerVector: map[string]float64{}}
	text := FormatVelocityText(report)
	if !strings.Contains(text, "insufficient data") {
		t.Errorf("got %q, expected insufficient data message", text)
	}
}

func TestFormatVelocityText_Normal(t *testing.T) {
	report := &VelocityReport{
		Overall:   5.0,
		PerVector: map[string]float64{VectorFunctional: 3.0, VectorStructural: 5.0, VectorPerformance: 7.0},
		Snapshots: 5,
		TimeSpan:  2 * time.Hour,
		ETA:       "~4 snapshots to 100",
	}
	text := FormatVelocityText(report)
	if !strings.Contains(text, "Velocity Report") {
		t.Error("missing header")
	}
	if !strings.Contains(text, "+5.00") {
		t.Errorf("missing overall velocity in %q", text)
	}
	if !strings.Contains(text, "ETA") {
		t.Error("missing ETA")
	}
}
