package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Tokens da nova Identidade Visual do p2t (Matematica + Dinheiro + Telemetria)
var (
	CyanPrimary    = lipgloss.Color("#00F2FE") // Ciano Eletrico (Matematica / Telemetria)
	GreenSecondary = lipgloss.Color("#00E676") // Verde Esmeralda (Dinheiro / Liquidez Real)
	CyberPurple    = lipgloss.Color("#7D56F4") // Roxo Cyber (Titulos & Destaque)
	MutedGray      = lipgloss.Color("#6272A4") // Cinza Slate (Textos Secundarios)
	DarkBg         = lipgloss.Color("#12131C") // Fundo Dark Contrastado

	GreenZoneColor  = lipgloss.Color("#50FA7B") // Neon Green
	YellowZoneColor = lipgloss.Color("#FFB800") // Ambar Ouro
	RedZoneColor    = lipgloss.Color("#FF3366") // Crimson Neon

	AppTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(CyberPurple).
			Padding(0, 2).
			MarginBottom(1)

	LogoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(CyanPrimary).
			MarginBottom(1)

	TabStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(MutedGray).
			Padding(0, 2).
			MarginRight(1)

	ActiveTabStyle = TabStyle.Copy().
			BorderForeground(CyanPrimary).
			Foreground(CyanPrimary).
			Bold(true)

	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(CyberPurple).
			Padding(1, 2).
			MarginBottom(1)

	MetricLabelStyle = lipgloss.NewStyle().
				Foreground(MutedGray).
				Bold(true)

	MetricMoneyStyle = lipgloss.NewStyle().
				Foreground(GreenSecondary).
				Bold(true)

	MetricMathStyle = lipgloss.NewStyle().
				Foreground(CyanPrimary).
				Bold(true)

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
			Foreground(MutedGray).
			MarginTop(1)

	FormBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(CyanPrimary).
			Padding(1, 2).
			MarginRight(2)

	ResultBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(CyberPurple).
			Padding(1, 2)

	FieldFocusedStyle = lipgloss.NewStyle().
				Foreground(CyanPrimary).
				Bold(true)

	FieldBlurStyle = lipgloss.NewStyle().
			Foreground(MutedGray)
)

// RenderProgressBar gera uma barra de progresso em texto/lipgloss dada uma porcentagem (0-100%).
func RenderProgressBar(percent float64, width int, zoneColor lipgloss.Color) string {
	if width < 10 {
		width = 20
	}
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	filledLength := int((percent / 100.0) * float64(width))
	if filledLength > width {
		filledLength = width
	}
	emptyLength := width - filledLength

	filledStyle := lipgloss.NewStyle().Foreground(zoneColor).Bold(true)
	emptyStyle := lipgloss.NewStyle().Foreground(MutedGray)

	bar := filledStyle.Render(strings.Repeat("█", filledLength)) + emptyStyle.Render(strings.Repeat("░", emptyLength))
	return bar
}

