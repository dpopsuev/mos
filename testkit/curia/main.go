package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dpopsuev/mos/moslib/survey"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-h", "--help":
			printHelp()
			os.Exit(0)
		}
	}

	rc := ResolveConfig(os.Args[1:])
	if rc.Path == "" {
		printHelp()
		os.Exit(0)
	}

	sc := &survey.AutoScanner{Override: rc.ScannerBackend, LSPCmd: rc.LSPCmd}
	mod, err := sc.Scan(rc.Path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "curia: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(newAppModel(mod, rc.Keymap, rc.Styles), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "curia: %v\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("Usage: curia [flags] [path]")
	fmt.Println()
	fmt.Println("Interactive module tree and dependency graph explorer.")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --keymap  <preset>   vim, emacs, vscode, or auto (default: auto)")
	fmt.Println("  --theme   <preset>   dark, light, or auto (default: auto)")
	fmt.Println("  --scanner <backend>  auto, go, packages, or lsp (default: auto)")
	fmt.Println("  --lsp-cmd <command>  override LSP server command (e.g. \"rust-analyzer\")")
	fmt.Println("  -h, --help           show this help")
	fmt.Println()
	fmt.Println("Config: ~/.config/curia/config.toml")
	fmt.Println()
	fmt.Println("Keys (default vscode preset):")
	fmt.Println("  ↑/↓, j/k    navigate")
	fmt.Println("  enter        expand/collapse package")
	fmt.Println("  tab          switch panel")
	fmt.Println("  pgup/pgdn   page scroll")
	fmt.Println("  home/end     jump to top/bottom")
	fmt.Println("  q, ctrl+c   quit")
}
