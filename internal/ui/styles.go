package ui

import "github.com/charmbracelet/lipgloss"

const (
	colorAccent = lipgloss.Color("#BD93F9")
	colorMuted  = lipgloss.Color("#666666")
	colorText   = lipgloss.Color("#DDDDDD")
	colorDim    = lipgloss.Color("#444444")
	colorBg     = lipgloss.Color("#141414")
)

var (
	navBarStyle = lipgloss.NewStyle().Background(colorBg)

	navItemStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Background(colorBg).
			Padding(0, 2)

	navItemActiveStyle = lipgloss.NewStyle().
				Foreground(colorAccent).
				Background(colorBg).
				Bold(true).
				Padding(0, 2)

	footerStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			Background(colorBg).
			Padding(0, 1)

	accentStyle = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	mutedStyle  = lipgloss.NewStyle().Foreground(colorMuted)
	dimStyle    = lipgloss.NewStyle().Foreground(colorDim)

	selectedStyle = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	normalStyle   = lipgloss.NewStyle().Foreground(colorText)

	inputLabelStyle = lipgloss.NewStyle().Foreground(colorMuted)
	errorStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555"))
	successStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B"))
)
