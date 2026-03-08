package main

import (
	"context"
	"fmt"
	"os"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"

	"github.com/dpopsuev/mos/moslib/arch"
)

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

var rootCmd = &cobra.Command{
	Use:   "mcontext [path]",
	Short: "Zero-ceremony codebase context for any repository",
	Long: `mcontext scans any repository and emits structured context:
architecture, dependency graph with weights, git history, churn,
hot spots, and exported symbol signatures.

No .mos directory or governance adoption required.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := "."
		if len(args) > 0 {
			root = args[0]
		}
		return runContext(root)
	},
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the MCP server (stdio transport)",
	Long: `Start an MCP server that exposes codebase context tools via stdio.

Tools: scan_project, suggest_depth, get_hot_spots, get_dependencies.

Configure in your MCP client (e.g. Cursor, Claude Code):
  { "command": "mcontext", "args": ["serve"] }`,
	RunE: func(cmd *cobra.Command, args []string) error {
		srv := arch.NewMCPServer()
		return srv.Run(context.Background(), &sdkmcp.StdioTransport{})
	},
}

func init() {
	rootCmd.Flags().StringVar(&flags.format, "format", "json", "Output format: json, md, mermaid")
	rootCmd.Flags().StringVar(&flags.scanner, "scanner", "auto", "Scanner: auto, go, packages, rust, typescript, composite, ctags, lsp")
	rootCmd.Flags().IntVar(&flags.depth, "depth", 0, "Group namespaces by first N directory segments")
	rootCmd.Flags().IntVar(&flags.churnDays, "churn-days", 30, "Overlay file churn from last N days of git history (0 = disabled)")
	rootCmd.Flags().IntVar(&flags.gitDays, "git-days", 30, "Recent commits window in days")
	rootCmd.Flags().BoolVar(&flags.authors, "authors", false, "Include author ownership data")
	rootCmd.Flags().BoolVar(&flags.includeExternal, "include-external", false, "Include external (third-party) dependencies")
	rootCmd.Flags().BoolVar(&flags.includeTests, "include-tests", false, "Include test packages")
	rootCmd.Flags().IntVar(&flags.budget, "budget", 0, "Cap output to N tokens (rank by importance, 0 = unlimited)")
	rootCmd.AddCommand(serveCmd)
}

func runContext(root string) error {
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
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
