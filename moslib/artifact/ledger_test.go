package artifact

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCON031_LedgerTraitParsed(t *testing.T) {
	root := setupScaffold(t)
	configPath := filepath.Join(root, ".mos", "config.mos")
	data, _ := os.ReadFile(configPath)
	extra := `
  artifact_type "incident" {
    directory = "incidents"
    ledger = true
    fields {}
    lifecycle {
      active_states = ["open"]
      archive_states = ["resolved"]
    }
  }
`
	patched := strings.Replace(string(data), "\n}\n", extra+"\n}\n", 1)
	os.WriteFile(configPath, []byte(patched), 0644)

	reg, err := LoadRegistry(root)
	if err != nil {
		t.Fatalf("LoadRegistry: %v", err)
	}
	td := reg.Types["incident"]
	if !td.Ledger {
		t.Error("expected incident type to have Ledger=true")
	}
}

func TestCON031_LedgerCreatedOnFirstMutation(t *testing.T) {
	root := setupScaffold(t)
	createHookTestContract(t, root, "CON-LED-001", "draft", map[string]string{
		"Step A": "pending",
	})

	ledgerPath, err := LedgerPathForContract(root, "CON-LED-001")
	if err != nil {
		t.Fatalf("LedgerPathForContract: %v", err)
	}
	if _, err := os.Stat(ledgerPath); !os.IsNotExist(err) {
		t.Fatal("expected ledger to not exist before mutation")
	}

	SetScenarioStatus(root, "CON-LED-001", "Step A", "implemented")

	if _, err := os.Stat(ledgerPath); err != nil {
		t.Fatalf("expected ledger to exist after mutation: %v", err)
	}

	entries, err := ReadLedger(ledgerPath)
	if err != nil {
		t.Fatalf("ReadLedger: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least one ledger entry")
	}
}

func TestCON031_SubsequentMutationsAppend(t *testing.T) {
	root := setupScaffold(t)
	createHookTestContract(t, root, "CON-LED-002", "active", map[string]string{
		"Step A": "pending",
		"Step B": "pending",
	})

	SetScenarioStatus(root, "CON-LED-002", "Step A", "implemented")
	SetScenarioStatus(root, "CON-LED-002", "Step B", "implemented")

	ledgerPath, _ := LedgerPathForContract(root, "CON-LED-002")
	entries, err := ReadLedger(ledgerPath)
	if err != nil {
		t.Fatalf("ReadLedger: %v", err)
	}
	if len(entries) < 2 {
		t.Fatalf("expected at least 2 ledger entries, got %d", len(entries))
	}
}

func TestCON031_LedgerRecordsOldNewValues(t *testing.T) {
	root := setupScaffold(t)
	createHookTestContract(t, root, "CON-LED-003", "active", map[string]string{
		"Step A": "pending",
	})

	SetScenarioStatus(root, "CON-LED-003", "Step A", "implemented")

	ledgerPath, _ := LedgerPathForContract(root, "CON-LED-003")
	entries, _ := ReadLedger(ledgerPath)

	found := false
	for _, e := range entries {
		if e.Event == "scenario_status_changed" && e.OldValue == "pending" && e.NewValue == "implemented" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected ledger entry with old=pending new=implemented, got entries: %+v", entries)
	}
}

func TestCON031_HistoryReconstructsTimeline(t *testing.T) {
	root := setupScaffold(t)
	createHookTestContract(t, root, "CON-LED-004", "active", map[string]string{
		"Step A": "pending",
	})

	SetScenarioStatus(root, "CON-LED-004", "Step A", "implemented")
	UpdateContractStatus(root, "CON-LED-004", "complete")

	ledgerPath, _ := LedgerPathForContract(root, "CON-LED-004")
	entries, _ := ReadLedger(ledgerPath)

	history := FormatHistory(entries)
	if !strings.Contains(history, "scenario_status_changed") {
		t.Errorf("expected timeline to contain scenario_status_changed, got:\n%s", history)
	}
	if !strings.Contains(history, "status_changed") {
		t.Errorf("expected timeline to contain status_changed, got:\n%s", history)
	}
}

func TestCON031_DefaultContractHasLedgerEnabled(t *testing.T) {
	root := setupScaffold(t)
	reg := loadTestRegistry(t, root)
	td := reg.Types["contract"]
	if !td.Ledger {
		t.Error("expected contract type to have Ledger=true")
	}
}

func TestCON031_ConcurrentAppendsPreserveAll(t *testing.T) {
	root := setupScaffold(t)
	createHookTestContract(t, root, "CON-LED-005", "active", map[string]string{
		"Step A": "pending",
		"Step B": "pending",
		"Step C": "pending",
	})

	ledgerPath, _ := LedgerPathForContract(root, "CON-LED-005")

	done := make(chan struct{})
	for _, name := range []string{"Step A", "Step B", "Step C"} {
		go func(n string) {
			defer func() { done <- struct{}{} }()
			AppendLedger(ledgerPath, LedgerEntry{
				Event:        "scenario_status_changed",
				Field:        "status",
				OldValue:     "pending",
				NewValue:     "implemented",
				ScenarioName: n,
			})
		}(name)
	}
	for i := 0; i < 3; i++ {
		<-done
	}

	entries, err := ReadLedger(ledgerPath)
	if err != nil {
		t.Fatalf("ReadLedger: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("expected 3 ledger entries after concurrent appends, got %d", len(entries))
	}
}

// --- CON-2026-033: Need -- The Sensation Primitive ---
