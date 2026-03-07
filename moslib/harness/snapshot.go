package harness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const snapshotDir = "snapshots"

// StateSnapshot captures the governance health state at a point in time.
type StateSnapshot struct {
	Timestamp      time.Time      `json:"timestamp"`
	CommitHash     string         `json:"commit_hash,omitempty"`
	IntegrityScore float64        `json:"integrity_score"`
	Vectors        []VectorResult `json:"vectors"`
	LintErrors     int            `json:"lint_errors"`
	LintWarnings   int            `json:"lint_warnings"`

	ScenariosTotal    int                `json:"scenarios_total,omitempty"`
	ScenariosPassing  int                `json:"scenarios_passing,omitempty"`
	DriftViolations   int                `json:"drift_violations,omitempty"`
	HarnessPassCount  int                `json:"harness_pass_count,omitempty"`
	HarnessFailCount  int                `json:"harness_fail_count,omitempty"`
	StructuralMetrics map[string]float64 `json:"structural_metrics,omitempty"`
}

func snapshotPath(mosDir string, ts time.Time) string {
	return filepath.Join(mosDir, snapshotDir, fmt.Sprintf("%d.json", ts.Unix()))
}

// StoreSnapshot writes a state snapshot to .mos/snapshots/<unix-timestamp>.json.
func StoreSnapshot(mosDir string, snap *StateSnapshot) error {
	dir := filepath.Join(mosDir, snapshotDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating snapshots dir: %w", err)
	}
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(snapshotPath(mosDir, snap.Timestamp), data, 0o644)
}

// LoadSnapshots reads all snapshot files from .mos/snapshots/ and returns
// them sorted by timestamp ascending.
func LoadSnapshots(mosDir string) ([]StateSnapshot, error) {
	dir := filepath.Join(mosDir, snapshotDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var snapshots []StateSnapshot
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var snap StateSnapshot
		if err := json.Unmarshal(data, &snap); err != nil {
			continue
		}
		snapshots = append(snapshots, snap)
	}

	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].Timestamp.Before(snapshots[j].Timestamp)
	})
	return snapshots, nil
}

// LoadSnapshotsInRange returns snapshots within the given time window (inclusive).
func LoadSnapshotsInRange(mosDir string, from, to time.Time) ([]StateSnapshot, error) {
	all, err := LoadSnapshots(mosDir)
	if err != nil {
		return nil, err
	}
	var filtered []StateSnapshot
	for _, s := range all {
		if (s.Timestamp.Equal(from) || s.Timestamp.After(from)) &&
			(s.Timestamp.Equal(to) || s.Timestamp.Before(to)) {
			filtered = append(filtered, s)
		}
	}
	return filtered, nil
}

// SnapshotFromIndex creates a StateSnapshot from an IntegrityIndex and
// optional lint counts.
func SnapshotFromIndex(idx *IntegrityIndex, lintErrors, lintWarnings int) *StateSnapshot {
	return &StateSnapshot{
		Timestamp:      idx.Timestamp,
		CommitHash:     idx.CommitHash,
		IntegrityScore: idx.Score,
		Vectors:        idx.Vectors,
		LintErrors:     lintErrors,
		LintWarnings:   lintWarnings,
	}
}

// ParseSnapshotTimestamp extracts the unix timestamp from a snapshot filename.
func ParseSnapshotTimestamp(name string) (time.Time, bool) {
	base := strings.TrimSuffix(name, ".json")
	unix, err := strconv.ParseInt(base, 10, 64)
	if err != nil {
		return time.Time{}, false
	}
	return time.Unix(unix, 0).UTC(), true
}

const vcsSnapshotDir = "vcs/snapshots"

// StoreCommitSnapshot writes a state snapshot keyed by commit hash
// to .mos/vcs/snapshots/<hash>.json.
func StoreCommitSnapshot(mosDir, commitHash string, snap *StateSnapshot) error {
	dir := filepath.Join(mosDir, vcsSnapshotDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating vcs snapshots dir: %w", err)
	}
	snap.CommitHash = commitHash
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, commitHash+".json"), data, 0o644)
}

// LoadCommitSnapshot reads a single snapshot by commit hash.
func LoadCommitSnapshot(mosDir, hash string) (*StateSnapshot, error) {
	path := filepath.Join(mosDir, vcsSnapshotDir, hash+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var snap StateSnapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, err
	}
	return &snap, nil
}

// LoadCommitSnapshots reads all VCS snapshots and returns them sorted by
// timestamp ascending.
func LoadCommitSnapshots(mosDir string) ([]StateSnapshot, error) {
	dir := filepath.Join(mosDir, vcsSnapshotDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var snapshots []StateSnapshot
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var snap StateSnapshot
		if err := json.Unmarshal(data, &snap); err != nil {
			continue
		}
		snapshots = append(snapshots, snap)
	}
	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].Timestamp.Before(snapshots[j].Timestamp)
	})
	return snapshots, nil
}

// LoadAllSnapshots loads both VCS commit snapshots and timestamp-based audit
// snapshots, merged and sorted by timestamp. VCS snapshots are preferred.
func LoadAllSnapshots(mosDir string) ([]StateSnapshot, error) {
	vcsSnaps, _ := LoadCommitSnapshots(mosDir)
	auditSnaps, _ := LoadSnapshots(mosDir)

	seen := make(map[string]bool, len(vcsSnaps))
	for _, s := range vcsSnaps {
		if s.CommitHash != "" {
			seen[s.CommitHash] = true
		}
	}

	all := append([]StateSnapshot{}, vcsSnaps...)
	for _, s := range auditSnaps {
		if s.CommitHash != "" && seen[s.CommitHash] {
			continue
		}
		all = append(all, s)
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].Timestamp.Before(all[j].Timestamp)
	})
	return all, nil
}
