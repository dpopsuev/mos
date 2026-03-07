package harness

import (
	"math"
	"testing"
)

func TestComputeIntegrityIndexFromVectors_EqualWeighting(t *testing.T) {
	vectors := []VectorResult{
		{Kind: VectorFunctional, Score: 90},
		{Kind: VectorStructural, Score: 60},
		{Kind: VectorPerformance, Score: 30},
	}
	idx := ComputeIntegrityIndexFromVectors(vectors)

	want := 60.0
	if math.Abs(idx.Score-want) > 0.01 {
		t.Errorf("Score = %.2f, want %.2f", idx.Score, want)
	}
	if len(idx.Vectors) != 3 {
		t.Fatalf("Vectors len = %d, want 3", len(idx.Vectors))
	}
}

func TestComputeIntegrityIndexFromVectors_AllPerfect(t *testing.T) {
	vectors := []VectorResult{
		{Kind: VectorFunctional, Score: 100},
		{Kind: VectorStructural, Score: 100},
		{Kind: VectorPerformance, Score: 100},
	}
	idx := ComputeIntegrityIndexFromVectors(vectors)

	if idx.Score != 100.0 {
		t.Errorf("Score = %.2f, want 100.0", idx.Score)
	}
}

func TestComputeIntegrityIndexFromVectors_MissingVector(t *testing.T) {
	vectors := []VectorResult{
		{Kind: VectorFunctional, Score: 90},
	}
	idx := ComputeIntegrityIndexFromVectors(vectors)

	want := 30.0
	if math.Abs(idx.Score-want) > 0.01 {
		t.Errorf("Score = %.2f, want %.2f (missing vectors should score 0)", idx.Score, want)
	}
}

func TestComputeIntegrityIndexFromVectors_NoVectors(t *testing.T) {
	idx := ComputeIntegrityIndexFromVectors(nil)
	if idx.Score != 0.0 {
		t.Errorf("Score = %.2f, want 0.0", idx.Score)
	}
}

func TestFormatIntegrityText(t *testing.T) {
	idx := &IntegrityIndex{
		Score: 73.3,
		Vectors: []VectorResult{
			{Kind: VectorFunctional, Score: 80},
			{Kind: VectorStructural, Score: 60},
			{Kind: VectorPerformance, Score: 80},
		},
	}
	got := FormatIntegrityText(idx)
	want := "Integrity Index: 73.3/100 (functional: 80, structural: 60, performance: 80)"
	if got != want {
		t.Errorf("FormatIntegrityText =\n  %q\nwant\n  %q", got, want)
	}
}
