package harness

import (
	"fmt"
	"time"
)

// IntegrityIndex is the scalar projection of the three quality vector scores
// into a single 0-100 number summarizing project health.
type IntegrityIndex struct {
	Score      float64        `json:"score"`
	Vectors    []VectorResult `json:"vectors"`
	Timestamp  time.Time      `json:"timestamp"`
	CommitHash string         `json:"commit_hash,omitempty"`
}

// ComputeIntegrityIndex evaluates all three quality vectors and aggregates
// them into a single score. Each vector contributes equally (1/3). Vectors
// with no rules score 0, lowering the index.
func ComputeIntegrityIndex(root, mosDir string) (*IntegrityIndex, error) {
	vectors, err := EvaluateVectors(root, mosDir)
	if err != nil {
		return nil, fmt.Errorf("evaluating vectors: %w", err)
	}
	return ComputeIntegrityIndexFromVectors(vectors), nil
}

// ComputeIntegrityIndexFromVectors computes the index from pre-evaluated vectors.
func ComputeIntegrityIndexFromVectors(vectors []VectorResult) *IntegrityIndex {
	total := 0.0
	count := len(AllVectors)
	vectorMap := make(map[string]float64, len(vectors))
	for _, v := range vectors {
		vectorMap[v.Kind] = v.Score
	}
	for _, kind := range AllVectors {
		total += vectorMap[kind]
	}

	score := 0.0
	if count > 0 {
		score = total / float64(count)
	}

	return &IntegrityIndex{
		Score:     score,
		Vectors:   vectors,
		Timestamp: time.Now().UTC(),
	}
}

// FormatIntegrityText returns a one-line summary of the integrity index.
func FormatIntegrityText(idx *IntegrityIndex) string {
	details := ""
	for i, v := range idx.Vectors {
		if i > 0 {
			details += ", "
		}
		details += fmt.Sprintf("%s: %.0f", v.Kind, v.Score)
	}
	return fmt.Sprintf("Integrity Index: %.1f/100 (%s)", idx.Score, details)
}
