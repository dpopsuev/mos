package subsystem

import (
	"github.com/spf13/cobra"

	"github.com/dpopsuev/mos/cmd/mos/vcscmd"
)

// GateCmd returns the "gate" subsystem command.
// Quality gate commands (lint, audit, harness, ci) were removed with the DSL package.
func GateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gate",
		Short: "Verification & enforcement — quality gates",
		Long: `Quality gates: hook.

Run pre-commit gate hooks. Lint, audit, harness, and ci commands
were removed with the DSL package.`,
	}

	cmd.AddCommand(vcscmd.HookCmd)

	return cmd
}
