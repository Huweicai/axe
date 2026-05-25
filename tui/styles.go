package tui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	selectedStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Bold(true)
	dimStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	workspaceStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	runningStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	sourceClaudeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("117"))
	sourceCodexStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("215"))
	previewBorder     = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, false, true).BorderForeground(lipgloss.Color("237")).PaddingLeft(1)
	statusBarStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Background(lipgloss.Color("236")).Padding(0, 1)
)
