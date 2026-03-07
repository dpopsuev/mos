package subsystem

import (
	"github.com/spf13/cobra"

	"github.com/dpopsuev/mos/cmd/mos/storecmd"
)

// StoreCmd returns the "store" subsystem command.
func StoreCmd() *cobra.Command {
	return storecmd.StoreCmd
}
