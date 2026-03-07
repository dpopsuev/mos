package subsystem

import (
	"github.com/spf13/cobra"

	"github.com/dpopsuev/mos/cmd/mos/vcscmd"
)

// VCSCmd returns the "vcs" subsystem command with all version control
// subcommands registered.
func VCSCmd() *cobra.Command {
	return vcscmd.Cmd
}
