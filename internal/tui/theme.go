package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// Theme holds the styles used to render the interface.
type Theme struct {
	HeaderPath lipgloss.Style
	HeaderInfo lipgloss.Style
	Footer     lipgloss.Style
	Status     lipgloss.Style
	Dir        lipgloss.Style
	File       lipgloss.Style
	Link       lipgloss.Style
	Drive      lipgloss.Style
	Meta       lipgloss.Style
	Selected   lipgloss.Style
	Match      lipgloss.Style
	HelpBox    lipgloss.Style
}

// DefaultTheme is the color theme.
func DefaultTheme() Theme {
	return Theme{
		HeaderPath: lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true),
		HeaderInfo: lipgloss.NewStyle().Foreground(lipgloss.Color("15")),
		Footer:     lipgloss.NewStyle().Foreground(lipgloss.Color("14")),
		Status:     lipgloss.NewStyle().Foreground(lipgloss.Color("11")),
		Dir:        lipgloss.NewStyle().Foreground(lipgloss.Color("12")),
		File:       lipgloss.NewStyle().Foreground(lipgloss.Color("15")),
		Link:       lipgloss.NewStyle().Foreground(lipgloss.Color("13")),
		Drive:      lipgloss.NewStyle().Foreground(lipgloss.Color("10")),
		Meta:       lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
		Selected:   lipgloss.NewStyle().Background(lipgloss.Color("7")).Foreground(lipgloss.Color("0")),
		Match:      lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Bold(true),
		HelpBox: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("14")).
			Padding(0, 1),
	}
}

// MonoTheme uses no colors; it relies on bold, underline, and reverse video.
func MonoTheme() Theme {
	return Theme{
		HeaderPath: lipgloss.NewStyle().Bold(true),
		HeaderInfo: lipgloss.NewStyle(),
		Footer:     lipgloss.NewStyle(),
		Status:     lipgloss.NewStyle().Bold(true),
		Dir:        lipgloss.NewStyle().Bold(true),
		File:       lipgloss.NewStyle(),
		Link:       lipgloss.NewStyle().Underline(true),
		Drive:      lipgloss.NewStyle().Bold(true).Underline(true),
		Meta:       lipgloss.NewStyle(),
		Selected:   lipgloss.NewStyle().Reverse(true),
		Match:      lipgloss.NewStyle().Underline(true),
		HelpBox:    lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1),
	}
}

// ThemeByName returns a theme by name. The empty string selects the default.
func ThemeByName(name string) (Theme, bool) {
	switch strings.ToLower(name) {
	case "", "default":
		return DefaultTheme(), true
	case "mono", "monochrome":
		return MonoTheme(), true
	default:
		return Theme{}, false
	}
}
