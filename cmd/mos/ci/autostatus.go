package ci

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dpopsuev/mos/moslib/mesh"
)

var AutoStatusCmd = &cobra.Command{
	Use:   "auto-status",
	Short: "Infer contract status transitions from git commits",
	Long: `Scans git commit messages for contract ID references and proposes
status transitions based on heuristics:
  - A draft contract mentioned in any commit → active
  - An active contract mentioned on the main branch → complete

By default runs in dry-run mode. Use --apply to commit changes.`,
	RunE: runAutoStatus,
}

var (
	autoStatusApply  bool
	autoStatusFormat string
)

func init() {
	AutoStatusCmd.Flags().BoolVar(&autoStatusApply, "apply", false, "Apply inferred status changes")
	AutoStatusCmd.Flags().StringVar(&autoStatusFormat, "format", "text", "Output format: text or json")
	Cmd.AddCommand(AutoStatusCmd)
}

func runAutoStatus(cmd *cobra.Command, args []string) error {
	changes, err := mesh.InferStatusChanges(".")
	if err != nil {
		return err
	}

	if len(changes) == 0 {
		if autoStatusFormat == "json" {
			fmt.Println("[]")
		} else {
			fmt.Println("No status changes inferred from git history.")
		}
		return nil
	}

	if autoStatusFormat == "json" {
		data, _ := json.MarshalIndent(changes, "", "  ")
		fmt.Println(string(data))
	} else {
		action := "Proposed"
		if autoStatusApply {
			action = "Applied"
		}
		fmt.Printf("%s status changes:\n", action)
		for _, c := range changes {
			fmt.Printf("  %s: %s → %s  (%s, %s)\n", c.ContractID, c.From, c.To, c.Reason, c.CommitHash[:8])
		}
	}

	if autoStatusApply {
		return mesh.ApplyStatusChanges(".", changes)
	}
	return nil
}
