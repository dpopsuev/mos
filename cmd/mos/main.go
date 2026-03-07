package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/dpopsuev/mos/cmd/internal/subsystem"
	"github.com/dpopsuev/mos/cmd/internal/wire"
	"github.com/dpopsuev/mos/cmd/mos/cliutil"
	"github.com/dpopsuev/mos/moslib/lsp"
)

var rootCmd = &cobra.Command{
	Use:   "mos",
	Short: "Governance management system",
	Long: `mos — governance management system

Subsystems:
  gov     Governance authoring & lifecycle
  vcs     Version control for governance artifacts
  gate    Verification & enforcement — quality gates
  trace   Traceability across the full SE stack
  store   Object store plumbing

Cross-cutting:
  lsp         Start the LSP server
  completion  Generate shell completions

Use "mos <subsystem> --help" for more information.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

var lspCmd = &cobra.Command{
	Use:   "lsp",
	Short: "Start the LSP server",
	RunE: func(cmd *cobra.Command, args []string) error {
		srv := lsp.NewServer(os.Stdin, os.Stdout)
		return srv.Run()
	},
}

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completions",
	Long:  "Generate shell completion scripts.\n\nExamples:\n  mos completion bash > /etc/bash_completion.d/mos\n  mos completion zsh  > \"${fpath[1]}/_mos\"\n  mos completion fish > ~/.config/fish/completions/mos.fish",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			return cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			return cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		default:
			return fmt.Errorf("unsupported shell %q (use bash, zsh, fish, or powershell)", args[0])
		}
	},
}

func init() {
	wire.Init()

	rootCmd.AddCommand(
		subsystem.GovCmd(),
		subsystem.VCSCmd(),
		subsystem.GateCmd(),
		subsystem.TraceCmd(),
		subsystem.StoreCmd(),
		lspCmd,
		completionCmd,
	)
}

func main() {
	if cliutil.IsAgentMode() {
		output, err := cliutil.CaptureStdout(func() error {
			return rootCmd.Execute()
		})
		cliutil.EmitAgentEnvelope(output, err)
		if err != nil {
			os.Exit(1)
		}
		return
	}

	if err := rootCmd.Execute(); err != nil {
		if errors.Is(err, cliutil.ErrInternalLint) {
			os.Exit(2)
		}
		if !errors.Is(err, cliutil.ErrNonZeroExit) {
			fmt.Fprintf(os.Stderr, "mos: %v\n", err)
		}
		os.Exit(1)
	}
}
