package subsystem

import (
	"github.com/spf13/cobra"

	"github.com/dpopsuev/mos/cmd/mos/tracecmd"
)

// TraceCmd returns the "trace" subsystem command with all traceability
// subcommands registered.
func TraceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trace",
		Short: "Traceability across the full SE stack",
		Long: `Cross-cutting traceability connecting needs, specs, architecture,
implementation, and documentation.

Commands: mesh, survey, clone, architecture sync`,
	}

	cmd.AddCommand(
		tracecmd.MeshCmd,
		tracecmd.CloneCmd,
		tracecmd.SurveyCmd,
		tracecmd.ArchSyncCmd,
	)

	return cmd
}
