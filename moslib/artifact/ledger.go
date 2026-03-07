package artifact

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dpopsuev/mos/moslib/dsl"
)

// LedgerEntry represents a single mutation event in the append-only ledger.
type LedgerEntry struct {
	Event        string
	Field        string
	OldValue     string
	NewValue     string
	ScenarioName string
	Timestamp    string
}

// AppendLedger atomically appends an entry to the ledger file.
func AppendLedger(ledgerPath string, entry LedgerEntry) error {
	mu := fileMutex(ledgerPath)
	mu.Lock()
	defer mu.Unlock()

	if entry.Timestamp == "" {
		entry.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}

	entryItems := []dsl.Node{
		&dsl.Field{Key: "event", Value: &dsl.StringVal{Text: entry.Event}},
		&dsl.Field{Key: "timestamp", Value: &dsl.StringVal{Text: entry.Timestamp}},
	}
	if entry.Field != "" {
		entryItems = append(entryItems, &dsl.Field{Key: "field", Value: &dsl.StringVal{Text: entry.Field}})
	}
	if entry.OldValue != "" {
		entryItems = append(entryItems, &dsl.Field{Key: "old_value", Value: &dsl.StringVal{Text: entry.OldValue}})
	}
	if entry.NewValue != "" {
		entryItems = append(entryItems, &dsl.Field{Key: "new_value", Value: &dsl.StringVal{Text: entry.NewValue}})
	}
	if entry.ScenarioName != "" {
		entryItems = append(entryItems, &dsl.Field{Key: "scenario", Value: &dsl.StringVal{Text: entry.ScenarioName}})
	}

	entryBlock := &dsl.Block{
		Name:  "entry",
		Items: entryItems,
	}

	if _, err := os.Stat(ledgerPath); err == nil {
		return dsl.WithArtifact(ledgerPath, func(ab *dsl.ArtifactBlock) error {
			ab.Items = append(ab.Items, entryBlock)
			return nil
		})
	}

	f := &dsl.File{
		Artifact: &dsl.ArtifactBlock{
			Kind:  "ledger",
			Name:  "log",
			Items: []dsl.Node{entryBlock},
		},
	}
	dir := filepath.Dir(ledgerPath)
	if err := os.MkdirAll(dir, DirPerm); err != nil {
		return fmt.Errorf("creating ledger directory: %w", err)
	}
	if err := writeArtifact(ledgerPath, f); err != nil {
		return fmt.Errorf("AppendLedger: %w", err)
	}
	return nil
}

// ReadLedger parses the ledger file and returns all entries.
func ReadLedger(ledgerPath string) ([]LedgerEntry, error) {
	ab, err := dsl.ReadArtifact(ledgerPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading ledger: %w", err)
	}

	var entries []LedgerEntry
	for _, item := range ab.Items {
		blk, ok := item.(*dsl.Block)
		if !ok || blk.Name != "entry" {
			continue
		}
		le := LedgerEntry{}
		le.Event, _ = dsl.FieldString(blk.Items, "event")
		le.Timestamp, _ = dsl.FieldString(blk.Items, "timestamp")
		le.Field, _ = dsl.FieldString(blk.Items, "field")
		le.OldValue, _ = dsl.FieldString(blk.Items, "old_value")
		le.NewValue, _ = dsl.FieldString(blk.Items, "new_value")
		le.ScenarioName, _ = dsl.FieldString(blk.Items, "scenario")
		entries = append(entries, le)
	}
	return entries, nil
}

// FormatHistory renders ledger entries as a human-readable timeline.
func FormatHistory(entries []LedgerEntry) string {
	if len(entries) == 0 {
		return "(no history)\n"
	}
	var sb strings.Builder
	for _, e := range entries {
		sb.WriteString(fmt.Sprintf("[%s] %s", e.Timestamp, e.Event))
		if e.Field != "" {
			sb.WriteString(fmt.Sprintf(" field=%s", e.Field))
		}
		if e.OldValue != "" || e.NewValue != "" {
			sb.WriteString(fmt.Sprintf(" %s -> %s", e.OldValue, e.NewValue))
		}
		if e.ScenarioName != "" {
			sb.WriteString(fmt.Sprintf(" scenario=%q", e.ScenarioName))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// LedgerPathForContract returns the ledger file path for a contract.
func LedgerPathForContract(root, id string) (string, error) {
	contractPath, err := FindContractPath(root, id)
	if err != nil {
		return "", fmt.Errorf("LedgerPathForContract: %w", err)
	}
	return filepath.Join(filepath.Dir(contractPath), "ledger.mos"), nil
}

// AppendContractLedger appends a ledger entry for a contract if ledger is enabled.
func AppendContractLedger(root, id string, entry LedgerEntry) {
	reg, err := LoadRegistry(root)
	if err != nil {
		return
	}
	td, ok := reg.Types[KindContract]
	if !ok || !td.Ledger {
		return
	}
	ledgerPath, err := LedgerPathForContract(root, id)
	if err != nil {
		return
	}
	AppendLedger(ledgerPath, entry)
}
