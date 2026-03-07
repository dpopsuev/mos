package artifact

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupWatchTest(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	mos := filepath.Join(root, ".mos")

	cfg := `config {
  mos { version = 1 }
  backend { type = "git" }

  artifact_type "watch" {
    directory = "watches"
    fields {
      title   { required = true }
      status  { required = true
        enum = ["active" "triggered" "expired" "dismissed"]
      }
      text    {}
      target  {}
      expires {}
      trigger {}
    }
    lifecycle {
      active_states  = ["active"]
      archive_states = ["triggered" "expired" "dismissed"]
    }
  }
}
`
	os.MkdirAll(filepath.Join(mos, "watches", "active"), 0o755)
	os.MkdirAll(filepath.Join(mos, "watches", "archive"), 0o755)
	os.WriteFile(filepath.Join(mos, "config.mos"), []byte(cfg), 0o644)
	return root
}

func TestEvaluateWatchTriggers_Expired(t *testing.T) {
	root := setupWatchTest(t)
	mos := filepath.Join(root, ".mos")

	watchDir := filepath.Join(mos, "watches", "active", "WATCH-2026-001")
	os.MkdirAll(watchDir, 0o755)
	os.WriteFile(filepath.Join(watchDir, "watch.mos"), []byte(`watch "WATCH-2026-001" {
  title   = "Expiry test"
  status  = "active"
  expires = "2026-01-01T00:00:00Z"
}
`), 0o644)

	now := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	results, err := EvaluateWatchTriggers(root, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Action != "expired" {
		t.Errorf("expected action=expired, got %s", results[0].Action)
	}
	if results[0].WatchID != "WATCH-2026-001" {
		t.Errorf("expected WatchID=WATCH-2026-001, got %s", results[0].WatchID)
	}
}

func TestEvaluateWatchTriggers_NotExpiredYet(t *testing.T) {
	root := setupWatchTest(t)
	mos := filepath.Join(root, ".mos")

	watchDir := filepath.Join(mos, "watches", "active", "WATCH-2026-002")
	os.MkdirAll(watchDir, 0o755)
	os.WriteFile(filepath.Join(watchDir, "watch.mos"), []byte(`watch "WATCH-2026-002" {
  title   = "Future watch"
  status  = "active"
  expires = "2027-01-01T00:00:00Z"
}
`), 0o644)

	now := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	results, err := EvaluateWatchTriggers(root, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestEvaluateWatchTriggers_WhenTrigger(t *testing.T) {
	root := setupWatchTest(t)
	mos := filepath.Join(root, ".mos")

	conDir := filepath.Join(mos, "contracts", "active", "CON-2026-100")
	os.MkdirAll(conDir, 0o755)
	os.WriteFile(filepath.Join(conDir, "contract.mos"), []byte(`contract "CON-2026-100" {
  title  = "Target contract"
  status = "complete"
}
`), 0o644)

	watchDir := filepath.Join(mos, "watches", "active", "WATCH-2026-003")
	os.MkdirAll(watchDir, 0o755)
	os.WriteFile(filepath.Join(watchDir, "watch.mos"), []byte(`watch "WATCH-2026-003" {
  title   = "Trigger test"
  status  = "active"
  trigger = "when CON-2026-100 complete"
}
`), 0o644)

	now := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	results, err := EvaluateWatchTriggers(root, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Action != "triggered" {
		t.Errorf("expected action=triggered, got %s", results[0].Action)
	}
}

func TestEvaluateWatchTriggers_WhenTriggerNotMet(t *testing.T) {
	root := setupWatchTest(t)
	mos := filepath.Join(root, ".mos")

	conDir := filepath.Join(mos, "contracts", "active", "CON-2026-101")
	os.MkdirAll(conDir, 0o755)
	os.WriteFile(filepath.Join(conDir, "contract.mos"), []byte(`contract "CON-2026-101" {
  title  = "Still active"
  status = "active"
}
`), 0o644)

	watchDir := filepath.Join(mos, "watches", "active", "WATCH-2026-004")
	os.MkdirAll(watchDir, 0o755)
	os.WriteFile(filepath.Join(watchDir, "watch.mos"), []byte(`watch "WATCH-2026-004" {
  title   = "Waiting"
  status  = "active"
  trigger = "when CON-2026-101 complete"
}
`), 0o644)

	now := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	results, err := EvaluateWatchTriggers(root, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}
