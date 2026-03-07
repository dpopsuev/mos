package vcscmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var HookCmd = &cobra.Command{
	Use:   "hook",
	Short: "Manage git hooks",
	Long:  "Install or uninstall git pre-commit hook that runs mos ci --fast",
}

func init() {
	HookCmd.AddCommand(hookInstallCmd, hookUninstallCmd)
}

var hookInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install git pre-commit hook that runs mos ci --fast",
	RunE:  runHookInstall,
}

var hookUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove the git pre-commit hook",
	RunE:  runHookUninstall,
}

const preCommitScript = `#!/bin/sh
# Installed by: mos hook install
# Runs the fast CI pipeline before each commit.
if [ -x ./mos ]; then
  ./mos ci --fast
else
  go run ./cmd/mos ci --fast
fi
`

func runHookInstall(cmd *cobra.Command, args []string) error {
	hookDir := filepath.Join(".git", "hooks")
	if _, err := os.Stat(hookDir); os.IsNotExist(err) {
		return fmt.Errorf("mos hook install: .git/hooks directory not found (is this a git repo?)")
	}
	hookPath := filepath.Join(hookDir, "pre-commit")
	if _, err := os.Stat(hookPath); err == nil {
		return fmt.Errorf("mos hook install: pre-commit hook already exists; remove it first with 'mos hook uninstall'")
	}
	if err := os.WriteFile(hookPath, []byte(preCommitScript), 0755); err != nil {
		return fmt.Errorf("mos hook install: %w", err)
	}
	fmt.Println("Installed pre-commit hook: .git/hooks/pre-commit")
	return nil
}

func runHookUninstall(cmd *cobra.Command, args []string) error {
	hookPath := filepath.Join(".git", "hooks", "pre-commit")
	data, err := os.ReadFile(hookPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("mos hook uninstall: no pre-commit hook found")
		}
		return fmt.Errorf("mos hook uninstall: %w", err)
	}
	if !strings.Contains(string(data), "mos ci") {
		return fmt.Errorf("mos hook uninstall: pre-commit hook was not installed by mos; refusing to remove")
	}
	if err := os.Remove(hookPath); err != nil {
		return fmt.Errorf("mos hook uninstall: %w", err)
	}
	fmt.Println("Removed pre-commit hook")
	return nil
}
