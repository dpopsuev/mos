package storecmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dpopsuev/mos/moslib/store"
)

// StoreCmd is the top-level `mos store` command.
var StoreCmd = &cobra.Command{
	Use:   "store",
	Short: "Manage the object store backend",
}

var storeStatusFormat string

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Print store backend type and audit log stats",
	RunE:  runStatus,
}

func init() {
	statusCmd.Flags().StringVar(&storeStatusFormat, "format", "text", "Output format: text or json")
	StoreCmd.AddCommand(statusCmd, verifyCmd, logCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	entries, _ := store.ReadAuditLog(".")

	if storeStatusFormat == "json" {
		data, _ := json.MarshalIndent(map[string]interface{}{
			"backend":    "fsstore",
			"policies":   []string{"audit-log"},
			"audit_entries": len(entries),
		}, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	fmt.Println("Backend:       fsstore")
	fmt.Println("Policies:      audit-log")
	fmt.Printf("Audit entries: %d\n", len(entries))
	return nil
}

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Run store integrity verification",
	RunE: func(cmd *cobra.Command, args []string) error {
		obj := store.DefaultObjectStore
		errs, err := obj.Verify(".")
		if err != nil {
			return err
		}
		if len(errs) == 0 {
			fmt.Println("Verification passed: no integrity errors.")
			return nil
		}
		for _, e := range errs {
			fmt.Printf("  %s: %s\n", e.Path, e.Message)
		}
		return fmt.Errorf("%d integrity error(s) found", len(errs))
	},
}

var logTail int

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Print recent audit log entries",
	RunE: func(cmd *cobra.Command, args []string) error {
		entries, err := store.ReadAuditLog(".")
		if err != nil {
			return err
		}
		if len(entries) == 0 {
			fmt.Println("No audit log entries.")
			return nil
		}
		start := 0
		if logTail > 0 && logTail < len(entries) {
			start = len(entries) - logTail
		}
		for _, e := range entries[start:] {
			fmt.Printf("%s %-12s %-8s %s %s\n", e.Timestamp, e.Actor, e.Operation, e.Path, e.SHA256)
		}
		return nil
	},
}

func init() {
	logCmd.Flags().IntVar(&logTail, "tail", 0, "Show only the last N entries")
}
