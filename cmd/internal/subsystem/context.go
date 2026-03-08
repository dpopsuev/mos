package subsystem

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dpopsuev/mos/moslib/arch"
)

// ContextCmd returns the "context" subsystem command that provides
// zero-ceremony codebase context scanning (the mos-integrated form
// of the standalone mcontext binary).
func ContextCmd() *cobra.Command {
	var flags struct {
		format          string
		scanner         string
		depth           int
		churnDays       int
		gitDays         int
		authors         bool
		includeExternal bool
		includeTests    bool
		budget          int
	}

	cmd := &cobra.Command{
		Use:   "context [path]",
		Short: "Scan any repo and emit structured codebase context",
		Long: `Scan source code and emit structured context: architecture,
dependency graph with weights, git history, churn, hot spots, and
exported symbol signatures. No .mos directory required.

Equivalent to the standalone mcontext binary.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			if len(args) > 0 {
				root = args[0]
			}
			report, err := arch.ScanAndBuild(root, arch.ScanOpts{
				ScannerOverride: flags.scanner,
				ExcludeTests:    !flags.includeTests,
				IncludeExternal: flags.includeExternal,
				Depth:           flags.depth,
				ChurnDays:       flags.churnDays,
				GitDays:         flags.gitDays,
				Authors:         flags.authors,
				Budget:          flags.budget,
			})
			if err != nil {
				return err
			}

			switch flags.format {
			case "json":
				data, err := arch.RenderJSON(report)
				if err != nil {
					return fmt.Errorf("render JSON: %w", err)
				}
				fmt.Println(string(data))
			case "md":
				fmt.Print(arch.RenderArchMarkdown(report.Architecture))
			case "mermaid":
				fmt.Print(arch.RenderMermaid(report.Architecture))
			default:
				return fmt.Errorf("unknown format %q (use json, md, or mermaid)", flags.format)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&flags.format, "format", "json", "Output format: json, md, mermaid")
	cmd.Flags().StringVar(&flags.scanner, "scanner", "auto", "Scanner: auto, go, packages, rust, typescript, composite, ctags, lsp")
	cmd.Flags().IntVar(&flags.depth, "depth", 0, "Group namespaces by first N directory segments")
	cmd.Flags().IntVar(&flags.churnDays, "churn-days", 30, "Overlay file churn from last N days of git history (0 = disabled)")
	cmd.Flags().IntVar(&flags.gitDays, "git-days", 30, "Recent commits window in days")
	cmd.Flags().BoolVar(&flags.authors, "authors", false, "Include author ownership data")
	cmd.Flags().BoolVar(&flags.includeExternal, "include-external", false, "Include external (third-party) dependencies")
	cmd.Flags().BoolVar(&flags.includeTests, "include-tests", false, "Include test packages")
	cmd.Flags().IntVar(&flags.budget, "budget", 0, "Cap output to N tokens (rank by importance, 0 = unlimited)")

	return cmd
}
