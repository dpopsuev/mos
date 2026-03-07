package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

type Palette struct {
	Subtle    string
	Highlight string
	Green     string
	Orange    string
	Bright    string
	Text      string
}

type Theme struct {
	Name   string
	Dark   Palette
	Light  Palette
	IsDark bool
}

func (t Theme) Active() Palette {
	if t.IsDark {
		return t.Dark
	}
	return t.Light
}

var DarkPalette = Palette{
	Subtle:    "243",
	Highlight: "39",
	Green:     "76",
	Orange:    "208",
	Bright:    "255",
	Text:      "252",
}

var LightPalette = Palette{
	Subtle:    "242",
	Highlight: "25",
	Green:     "22",
	Orange:    "166",
	Bright:    "0",
	Text:      "235",
}

func DefaultTheme() Theme {
	return Theme{
		Name:   "default",
		Dark:   DarkPalette,
		Light:  LightPalette,
		IsDark: termenv.HasDarkBackground(),
	}
}

func DarkTheme() Theme {
	return Theme{Name: "dark", Dark: DarkPalette, Light: DarkPalette, IsDark: true}
}

func LightTheme() Theme {
	return Theme{Name: "light", Dark: LightPalette, Light: LightPalette, IsDark: false}
}

func ResolveTheme(preset string) Theme {
	switch strings.ToLower(preset) {
	case "dark":
		return DarkTheme()
	case "light":
		return LightTheme()
	default:
		return DefaultTheme()
	}
}

// Styles holds all lipgloss styles derived from a Theme. Replaces the
// package-level style vars that were previously hardcoded.
type Styles struct {
	ActiveBorder   lipgloss.Style
	InactiveBorder lipgloss.Style

	Title      lipgloss.Style
	Package    lipgloss.Style
	File       lipgloss.Style
	Exported   lipgloss.Style
	Unexported lipgloss.Style
	Cursor     lipgloss.Style

	SectionHeader lipgloss.Style
	InternalEdge  lipgloss.Style
	ExternalEdge  lipgloss.Style
	EdgeTag       lipgloss.Style
	NoData        lipgloss.Style

	HelpKey  lipgloss.Style
	HelpDesc lipgloss.Style
}

func BuildStyles(t Theme) *Styles {
	p := t.Active()

	subtle := lipgloss.Color(p.Subtle)
	highlight := lipgloss.Color(p.Highlight)
	green := lipgloss.Color(p.Green)
	orange := lipgloss.Color(p.Orange)
	bright := lipgloss.Color(p.Bright)
	text := lipgloss.Color(p.Text)

	return &Styles{
		ActiveBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(highlight),
		InactiveBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(subtle),

		Title:      lipgloss.NewStyle().Bold(true).Foreground(bright),
		Package:    lipgloss.NewStyle().Bold(true).Foreground(highlight),
		File:       lipgloss.NewStyle().Foreground(text),
		Exported:   lipgloss.NewStyle().Foreground(green),
		Unexported: lipgloss.NewStyle().Foreground(subtle),
		Cursor:     lipgloss.NewStyle().Bold(true).Foreground(highlight),

		SectionHeader: lipgloss.NewStyle().Bold(true).Underline(true).Foreground(bright),
		InternalEdge:  lipgloss.NewStyle().Foreground(highlight),
		ExternalEdge:  lipgloss.NewStyle().Foreground(orange),
		EdgeTag:       lipgloss.NewStyle().Foreground(subtle),
		NoData:        lipgloss.NewStyle().Italic(true).Foreground(subtle),

		HelpKey:  lipgloss.NewStyle().Bold(true).Foreground(highlight),
		HelpDesc: lipgloss.NewStyle().Foreground(subtle),
	}
}

// ApplyPaletteOverrides patches a theme's active palette with any non-empty
// overrides from the config file.
func (t *Theme) ApplyPaletteOverrides(overrides PaletteOverrides) {
	p := &t.Dark
	if !t.IsDark {
		p = &t.Light
	}
	if overrides.Subtle != "" {
		p.Subtle = overrides.Subtle
	}
	if overrides.Highlight != "" {
		p.Highlight = overrides.Highlight
	}
	if overrides.Green != "" {
		p.Green = overrides.Green
	}
	if overrides.Orange != "" {
		p.Orange = overrides.Orange
	}
	if overrides.Bright != "" {
		p.Bright = overrides.Bright
	}
	if overrides.Text != "" {
		p.Text = overrides.Text
	}
}

type PaletteOverrides struct {
	Subtle    string `toml:"subtle"`
	Highlight string `toml:"highlight"`
	Green     string `toml:"green"`
	Orange    string `toml:"orange"`
	Bright    string `toml:"bright"`
	Text      string `toml:"text"`
}
