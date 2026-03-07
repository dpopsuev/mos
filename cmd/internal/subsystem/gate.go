package subsystem

import (
	"github.com/spf13/cobra"

	"github.com/dpopsuev/mos/cmd/mos/ci"
	"github.com/dpopsuev/mos/cmd/mos/gatecmd"
	"github.com/dpopsuev/mos/cmd/mos/vcscmd"
)

// GateCmd returns the "gate" subsystem command with all verification and
// enforcement subcommands registered.
func GateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gate",
		Short: "Verification & enforcement — quality gates",
		Long: `Quality gates: lint, validate, audit, harness, doctor, ci, hook.

Run checks against governance artifacts, execute test harnesses,
and manage pre-commit gate hooks.`,
	}

	cmd.AddCommand(
		gatecmd.LintCmd,
		gatecmd.ValidateCmd,
		gatecmd.AuditCmd,
		gatecmd.HarnessCmd,
		gatecmd.DoctorCmd,
		ci.Cmd,
		vcscmd.HookCmd,
	)

	return cmd
}
