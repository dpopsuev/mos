package subsystem

import "github.com/spf13/cobra"

// GovCmd returns the deprecated "gov" subsystem command.
// Artifact management has moved to the scribe CLI.
func GovCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "gov",
		Short: "[DEPRECATED] Use scribe for artifact management",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("mgov is deprecated. Use 'scribe' for artifact management.")
			cmd.Println()
			cmd.Println("  scribe create --kind contract --title \"...\"")
			cmd.Println("  scribe show <ID>")
			cmd.Println("  scribe list [--kind ...] [--scope ...]")
			cmd.Println("  scribe status <ID> <status>")
			cmd.Println("  scribe set <ID> <field> <value>")
			return nil
		},
	}
}
