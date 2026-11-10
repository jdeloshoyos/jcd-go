package tui

import "charm.land/lipgloss/v2"

// Styles used by the interface. Colors are ANSI indices so they adapt to the
// terminal palette.
var (
	styleHeaderPath = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true)
	styleHeaderInfo = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	styleFooter     = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	styleStatus     = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))

	styleDir      = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	styleFile     = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	styleLink     = lipgloss.NewStyle().Foreground(lipgloss.Color("13"))
	styleDrive    = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	styleMeta     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	styleSelected = lipgloss.NewStyle().Background(lipgloss.Color("7")).Foreground(lipgloss.Color("0"))
)
