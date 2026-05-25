package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Panel chrome
	panelBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("238"))

	activePanelBorder = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("99"))

	// Header bar
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("63")).
			Padding(0, 1)

	headerDimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("189")).
			Background(lipgloss.Color("63"))

	// Selected row — full-width background highlight
	selectedRowStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("255")).
				Background(lipgloss.Color("62")).
				Bold(true)

	// Normal row
	normalRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	// Dim / secondary text
	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))

	// Workspace items
	workspaceStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("220")).
			Bold(true)

	workspaceIconStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("226"))

	// Running indicator
	runningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Bold(true)

	// Source badges with background
	claudeBadge = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("62")).
			Padding(0, 1).
			Bold(true)

	codexBadge = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("172")).
			Padding(0, 1).
			Bold(true)

	// Source labels (no background, for list rows)
	sourceClaudeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("111"))
	sourceCodexStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("215"))

	// Preview pane
	previewTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("99"))

	previewLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("243"))

	previewValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	// Status bar
	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Background(lipgloss.Color("236")).
			Padding(0, 1)

	statusKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("117")).
			Background(lipgloss.Color("236")).
			Bold(true)

	// Date column
	dateStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))

	// Star style
	starStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))

	// Title style (used in header)
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
)
