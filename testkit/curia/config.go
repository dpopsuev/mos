package main

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Keymap  keymapConfig  `toml:"keymap"`
	Theme   themeConfig   `toml:"theme"`
	Scanner scannerConfig `toml:"scanner"`
}

type scannerConfig struct {
	Backend string `toml:"backend"`
	LSPCmd  string `toml:"lsp_cmd"`
}

type keymapConfig struct {
	Preset   string              `toml:"preset"`
	Bindings map[string][]string `toml:"bindings"`
}

type themeConfig struct {
	Preset  string           `toml:"preset"`
	Palette PaletteOverrides `toml:"palette"`
}

func DefaultConfig() Config {
	return Config{
		Keymap: keymapConfig{Preset: "auto"},
		Theme:  themeConfig{Preset: "auto"},
	}
}

func configPath() string {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "curia", "config.toml")
}

func LoadConfig() Config {
	cfg := DefaultConfig()
	path := configPath()
	if path == "" {
		return cfg
	}
	if _, err := os.Stat(path); err != nil {
		return cfg
	}
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return DefaultConfig()
	}
	if cfg.Keymap.Preset == "" {
		cfg.Keymap.Preset = "auto"
	}
	if cfg.Theme.Preset == "" {
		cfg.Theme.Preset = "auto"
	}
	return cfg
}

type cliFlags struct {
	keymap  string
	theme   string
	scanner string
	lspCmd  string
	path    string
}

func parseFlags(args []string) cliFlags {
	f := cliFlags{path: "."}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--keymap":
			if i+1 < len(args) {
				i++
				f.keymap = args[i]
			}
		case "--theme":
			if i+1 < len(args) {
				i++
				f.theme = args[i]
			}
		case "--scanner":
			if i+1 < len(args) {
				i++
				f.scanner = args[i]
			}
		case "--lsp-cmd":
			if i+1 < len(args) {
				i++
				f.lspCmd = args[i]
			}
		case "-h", "--help":
			f.path = ""
		default:
			if args[i] != "" && args[i][0] != '-' {
				f.path = args[i]
			}
		}
	}
	return f
}

// ResolvedConfig holds the resolved configuration for the TUI.
type ResolvedConfig struct {
	Keymap         *Keymap
	Styles         *Styles
	Path           string
	ScannerBackend string
	LSPCmd         string
}

// ResolveConfig loads the TOML config, applies CLI flag overrides, and returns
// the final resolved configuration.
func ResolveConfig(args []string) ResolvedConfig {
	flags := parseFlags(args)
	if flags.path == "" {
		return ResolvedConfig{}
	}

	cfg := LoadConfig()

	keymapPreset := cfg.Keymap.Preset
	if flags.keymap != "" {
		keymapPreset = flags.keymap
	}
	km := ResolveKeymap(keymapPreset)
	if len(cfg.Keymap.Bindings) > 0 && flags.keymap == "" {
		km.ApplyOverrides(cfg.Keymap.Bindings)
	}

	themePreset := cfg.Theme.Preset
	if flags.theme != "" {
		themePreset = flags.theme
	}
	theme := ResolveTheme(themePreset)
	theme.ApplyPaletteOverrides(cfg.Theme.Palette)
	styles := BuildStyles(theme)

	scannerBackend := cfg.Scanner.Backend
	if flags.scanner != "" {
		scannerBackend = flags.scanner
	}
	if scannerBackend == "" {
		scannerBackend = "auto"
	}

	lspCmd := cfg.Scanner.LSPCmd
	if flags.lspCmd != "" {
		lspCmd = flags.lspCmd
	}

	return ResolvedConfig{
		Keymap:         km,
		Styles:         styles,
		Path:           flags.path,
		ScannerBackend: scannerBackend,
		LSPCmd:         lspCmd,
	}
}
