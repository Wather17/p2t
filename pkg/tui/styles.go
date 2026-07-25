package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Estilos de UI do p2t usando Lip Gloss com visual moderno e sofisticado
var (
	PrimaryColor   = lipgloss.Color("#7D56F4") // Violeta elegante
	SecondaryColor = lipgloss.Color("#04B575") // Verde esmeralda
	AccentColor    = lipgloss.Color("#FF7675") // Coral suave
	MutedColor     = lipgloss.Color("#6272A4") // Azul acinzentado

	GreenZoneColor  = lipgloss.Color("#50FA7B")
	YellowZoneColor = lipgloss.Color("#F1FA8C")
	RedZoneColor    = lipgloss.Color("#FF5555")

	AppTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(PrimaryColor).
			Padding(0, 2).
			MarginBottom(1)

	TabStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(MutedColor).
			Padding(0, 2).
			MarginRight(1)

	ActiveTabStyle = TabStyle.Copy().
			BorderForeground(PrimaryColor).
			Foreground(PrimaryColor).
			Bold(true)

	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(MutedColor).
			Padding(1, 2).
			MarginBottom(1)

	BadgeGreenStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#000000")).
			Background(GreenZoneColor).
			Padding(0, 1)

	BadgeYellowStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#000000")).
			Background(YellowZoneColor).
			Padding(0, 1)

	BadgeRedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(RedZoneColor).
			Padding(0, 1)

	HelpStyle = lipgloss.NewStyle().
			Foreground(MutedColor).
			MarginTop(1)
)
