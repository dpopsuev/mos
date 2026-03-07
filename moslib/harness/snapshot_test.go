package harness

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStoreAndLoadSnapshot(t *testing.T) {
	dir := t.TempDir()
	snap := &StateSnapshot{
		Timestamp:      time.Unix(1700000000, 0).UTC(),
		IntegrityScore: 75.5,
		Vectors: []VectorResult{
			{Kind: VectorFunctional, Score: 80, Pass: true},
			{Kind: VectorStructural, Score: 70, Pass: true},
			{Kind: VectorPerformance, Score: 76.5, Pass: true},
		},
		LintErrors:   2,
		LintWarnings: 5,
	}

	if err := StoreSnapshot(dir, snap); err != nil {
		t.Fatalf("StoreSnapshot: %v", err)
	}

	path := filepath.Join(dir, snapshotDir, "1700000000.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("snapshot file not created: %v", err)
	}

	loaded, err := LoadSnapshots(dir)
	if err != nil {
		t.Fatalf("LoadSnapshots: %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("loaded %d snapshots, want 1", len(loaded))
	}
	if loaded[0].IntegrityScore != 75.5 {
		t.Errorf("IntegrityScore = %.1f, want 75.5", loaded[0].IntegrityScore)
	}
	if loaded[0].LintErrors != 2 {
		t.Errorf("LintErrors = %d, want 2", loaded[0].LintErrors)
	}
}

func TestLoadSnapshots_Ordering(t *testing.T) {
	dir := t.TempDir()

	timestamps := []time.Time{
		time.Unix(1700000300, 0).UTC(),
		time.Unix(1700000100, 0).UTC(),
		time.Unix(1700000200, 0).UTC(),
	}

	for _, ts := range timestamps {
		snap := &StateSnapshot{
			Timestamp:      ts,
			IntegrityScore: float64(ts.Unix() % 100),
		}
		if err := StoreSnapshot(dir, snap); err != nil {
			t.Fatalf("StoreSnapshot: %v", err)
		}
	}

	loaded, err := LoadSnapshots(dir)
	if err != nil {
		t.Fatalf("LoadSnapshots: %v", err)
	}
	if len(loaded) != 3 {
		t.Fatalf("loaded %d, want 3", len(loaded))
	}
	for i := 1; i < len(loaded); i++ {
		if loaded[i].Timestamp.Before(loaded[i-1].Timestamp) {
			t.Errorf("snapshot %d (%v) before %d (%v)", i, loaded[i].Timestamp, i-1, loaded[i-1].Timestamp)
		}
	}
}

func TestLoadSnapshotsInRange(t *testing.T) {
	dir := t.TempDir()

	for _, unix := range []int64{100, 200, 300, 400, 500} {
		snap := &StateSnapshot{
			Timestamp:      time.Unix(unix, 0).UTC(),
			IntegrityScore: float64(unix / 10),
		}
		if err := StoreSnapshot(dir, snap); err != nil {
			t.Fatalf("StoreSnapshot: %v", err)
		}
	}

	from := time.Unix(200, 0).UTC()
	to := time.Unix(400, 0).UTC()
	filtered, err := LoadSnapshotsInRange(dir, from, to)
	if err != nil {
		t.Fatalf("LoadSnapshotsInRange: %v", err)
	}
	if len(filtered) != 3 {
		t.Errorf("filtered %d, want 3 (200, 300, 400)", len(filtered))
	}
}

func TestLoadSnapshots_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	loaded, err := LoadSnapshots(dir)
	if err != nil {
		t.Fatalf("LoadSnapshots: %v", err)
	}
	if len(loaded) != 0 {
		t.Errorf("loaded %d, want 0", len(loaded))
	}
}

func TestSnapshotFromIndex(t *testing.T) {
	idx := &IntegrityIndex{
		Score: 80,
		Vectors: []VectorResult{
			{Kind: VectorFunctional, Score: 90},
		},
		Timestamp: time.Unix(1700000000, 0).UTC(),
	}
	snap := SnapshotFromIndex(idx, 3, 7)
	if snap.IntegrityScore != 80 {
		t.Errorf("IntegrityScore = %.1f, want 80", snap.IntegrityScore)
	}
	if snap.LintErrors != 3 {
		t.Errorf("LintErrors = %d, want 3", snap.LintErrors)
	}
	if snap.LintWarnings != 7 {
		t.Errorf("LintWarnings = %d, want 7", snap.LintWarnings)
	}
}

func TestStoreAndLoadCommitSnapshot(t *testing.T) {
	dir := t.TempDir()
	snap := &StateSnapshot{
		Timestamp:       time.Unix(1700000000, 0).UTC(),
		IntegrityScore:  82.5,
		LintErrors:      1,
		DriftViolations: 3,
	}

	if err := StoreCommitSnapshot(dir, "abc123def", snap); err != nil {
		t.Fatalf("StoreCommitSnapshot: %v", err)
	}

	path := filepath.Join(dir, vcsSnapshotDir, "abc123def.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("vcs snapshot file not created: %v", err)
	}

	loaded, err := LoadCommitSnapshot(dir, "abc123def")
	if err != nil {
		t.Fatalf("LoadCommitSnapshot: %v", err)
	}
	if loaded.IntegrityScore != 82.5 {
		t.Errorf("IntegrityScore = %.1f, want 82.5", loaded.IntegrityScore)
	}
	if loaded.CommitHash != "abc123def" {
		t.Errorf("CommitHash = %s, want abc123def", loaded.CommitHash)
	}
	if loaded.DriftViolations != 3 {
		t.Errorf("DriftViolations = %d, want 3", loaded.DriftViolations)
	}
}

func TestLoadCommitSnapshots_Ordering(t *testing.T) {
	dir := t.TempDir()

	for i, ts := range []int64{300, 100, 200} {
		snap := &StateSnapshot{
			Timestamp:      time.Unix(ts, 0).UTC(),
			IntegrityScore: float64(i * 10),
		}
		hash := []string{"ccc", "aaa", "bbb"}[i]
		if err := StoreCommitSnapshot(dir, hash, snap); err != nil {
			t.Fatalf("StoreCommitSnapshot: %v", err)
		}
	}

	loaded, err := LoadCommitSnapshots(dir)
	if err != nil {
		t.Fatalf("LoadCommitSnapshots: %v", err)
	}
	if len(loaded) != 3 {
		t.Fatalf("loaded %d, want 3", len(loaded))
	}
	for i := 1; i < len(loaded); i++ {
		if loaded[i].Timestamp.Before(loaded[i-1].Timestamp) {
			t.Errorf("snapshot %d before %d", i, i-1)
		}
	}
}

func TestLoadAllSnapshots_MergesVCSAndAudit(t *testing.T) {
	dir := t.TempDir()

	auditSnap := &StateSnapshot{
		Timestamp:      time.Unix(100, 0).UTC(),
		IntegrityScore: 50,
	}
	if err := StoreSnapshot(dir, auditSnap); err != nil {
		t.Fatalf("StoreSnapshot: %v", err)
	}

	vcsSnap := &StateSnapshot{
		Timestamp:      time.Unix(200, 0).UTC(),
		IntegrityScore: 60,
	}
	if err := StoreCommitSnapshot(dir, "abc", vcsSnap); err != nil {
		t.Fatalf("StoreCommitSnapshot: %v", err)
	}

	all, err := LoadAllSnapshots(dir)
	if err != nil {
		t.Fatalf("LoadAllSnapshots: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("loaded %d, want 2", len(all))
	}
	if all[0].IntegrityScore != 50 {
		t.Errorf("first snapshot score = %.1f, want 50", all[0].IntegrityScore)
	}
	if all[1].IntegrityScore != 60 {
		t.Errorf("second snapshot score = %.1f, want 60", all[1].IntegrityScore)
	}
}
