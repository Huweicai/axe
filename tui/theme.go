package tui

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Name string

	BorderFg       lipgloss.TerminalColor
	ActiveBorderFg lipgloss.TerminalColor

	HeaderFg    lipgloss.TerminalColor
	HeaderBg    lipgloss.TerminalColor
	HeaderDimFg lipgloss.TerminalColor

	SelectedFg lipgloss.TerminalColor
	SelectedBg lipgloss.TerminalColor

	NormalFg lipgloss.TerminalColor
	DimFg    lipgloss.TerminalColor

	ClaudeFg lipgloss.TerminalColor
	CodexFg  lipgloss.TerminalColor

	WorkspaceFg lipgloss.TerminalColor
	StarFg      lipgloss.TerminalColor
	RunningFg   lipgloss.TerminalColor
	DateFg      lipgloss.TerminalColor

	ClaudeBadgeFg lipgloss.TerminalColor
	ClaudeBadgeBg lipgloss.TerminalColor
	CodexBadgeFg  lipgloss.TerminalColor
	CodexBadgeBg  lipgloss.TerminalColor

	PreviewTitleFg lipgloss.TerminalColor
	PreviewLabelFg lipgloss.TerminalColor
	PreviewValueFg lipgloss.TerminalColor

	StatusFg    lipgloss.TerminalColor
	StatusBg    lipgloss.TerminalColor
	StatusKeyFg lipgloss.TerminalColor
}

func ac(light, dark string) lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: light, Dark: dark}
}

var themes = map[string]Theme{
	"default": {
		Name:           "default",
		BorderFg:       ac("245", "238"),
		ActiveBorderFg: ac("63", "99"),
		HeaderFg:       ac("255", "255"),
		HeaderBg:       ac("63", "63"),
		HeaderDimFg:    ac("232", "189"),
		SelectedFg:     ac("255", "255"),
		SelectedBg:     ac("62", "62"),
		NormalFg:       ac("234", "252"),
		DimFg:          ac("242", "243"),
		ClaudeFg:       ac("25", "111"),
		CodexFg:        ac("130", "215"),
		WorkspaceFg:    ac("136", "220"),
		StarFg:         ac("136", "220"),
		RunningFg:      ac("28", "82"),
		DateFg:         ac("242", "243"),
		ClaudeBadgeFg:  ac("255", "255"),
		ClaudeBadgeBg:  ac("62", "62"),
		CodexBadgeFg:   ac("255", "255"),
		CodexBadgeBg:   ac("172", "172"),
		PreviewTitleFg: ac("63", "99"),
		PreviewLabelFg: ac("240", "250"),
		PreviewValueFg: ac("234", "255"),
		StatusFg:       ac("240", "245"),
		StatusBg:       ac("253", "236"),
		StatusKeyFg:    ac("25", "117"),
	},
	"monokai": {
		Name:           "monokai",
		BorderFg:       ac("245", "239"),
		ActiveBorderFg: ac("64", "148"),
		HeaderFg:       ac("232", "232"),
		HeaderBg:       ac("148", "148"),
		HeaderDimFg:    ac("236", "236"),
		SelectedFg:     ac("232", "232"),
		SelectedBg:     ac("148", "148"),
		NormalFg:       ac("234", "253"),
		DimFg:          ac("242", "242"),
		ClaudeFg:       ac("25", "81"),
		CodexFg:        ac("130", "208"),
		WorkspaceFg:    ac("100", "186"),
		StarFg:         ac("100", "186"),
		RunningFg:      ac("64", "148"),
		DateFg:         ac("242", "242"),
		ClaudeBadgeFg:  ac("232", "232"),
		ClaudeBadgeBg:  ac("81", "81"),
		CodexBadgeFg:   ac("232", "232"),
		CodexBadgeBg:   ac("208", "208"),
		PreviewTitleFg: ac("64", "148"),
		PreviewLabelFg: ac("240", "250"),
		PreviewValueFg: ac("234", "255"),
		StatusFg:       ac("240", "242"),
		StatusBg:       ac("253", "235"),
		StatusKeyFg:    ac("25", "81"),
	},
	"dracula": {
		Name:           "dracula",
		BorderFg:       ac("245", "61"),
		ActiveBorderFg: ac("98", "141"),
		HeaderFg:       ac("255", "255"),
		HeaderBg:       ac("141", "141"),
		HeaderDimFg:    ac("232", "189"),
		SelectedFg:     ac("255", "255"),
		SelectedBg:     ac("141", "141"),
		NormalFg:       ac("234", "253"),
		DimFg:          ac("242", "103"),
		ClaudeFg:       ac("25", "117"),
		CodexFg:        ac("130", "215"),
		WorkspaceFg:    ac("28", "84"),
		StarFg:         ac("136", "228"),
		RunningFg:      ac("28", "84"),
		DateFg:         ac("242", "103"),
		ClaudeBadgeFg:  ac("255", "255"),
		ClaudeBadgeBg:  ac("61", "61"),
		CodexBadgeFg:   ac("255", "255"),
		CodexBadgeBg:   ac("166", "166"),
		PreviewTitleFg: ac("98", "141"),
		PreviewLabelFg: ac("240", "250"),
		PreviewValueFg: ac("234", "255"),
		StatusFg:       ac("240", "103"),
		StatusBg:       ac("253", "236"),
		StatusKeyFg:    ac("98", "141"),
	},
	"catppuccin": {
		Name:           "catppuccin",
		BorderFg:       ac("245", "240"),
		ActiveBorderFg: ac("133", "183"),
		HeaderFg:       ac("232", "232"),
		HeaderBg:       ac("183", "183"),
		HeaderDimFg:    ac("236", "236"),
		SelectedFg:     ac("232", "232"),
		SelectedBg:     ac("183", "183"),
		NormalFg:       ac("234", "254"),
		DimFg:          ac("242", "245"),
		ClaudeFg:       ac("25", "110"),
		CodexFg:        ac("130", "216"),
		WorkspaceFg:    ac("28", "114"),
		StarFg:         ac("136", "222"),
		RunningFg:      ac("28", "114"),
		DateFg:         ac("242", "245"),
		ClaudeBadgeFg:  ac("232", "232"),
		ClaudeBadgeBg:  ac("110", "110"),
		CodexBadgeFg:   ac("232", "232"),
		CodexBadgeBg:   ac("216", "216"),
		PreviewTitleFg: ac("133", "183"),
		PreviewLabelFg: ac("240", "250"),
		PreviewValueFg: ac("234", "255"),
		StatusFg:       ac("240", "245"),
		StatusBg:       ac("253", "236"),
		StatusKeyFg:    ac("133", "183"),
	},
	"nord": {
		Name:           "nord",
		BorderFg:       ac("245", "60"),
		ActiveBorderFg: ac("25", "110"),
		HeaderFg:       ac("255", "255"),
		HeaderBg:       ac("60", "60"),
		HeaderDimFg:    ac("232", "146"),
		SelectedFg:     ac("255", "255"),
		SelectedBg:     ac("60", "60"),
		NormalFg:       ac("234", "254"),
		DimFg:          ac("242", "246"),
		ClaudeFg:       ac("25", "110"),
		CodexFg:        ac("130", "173"),
		WorkspaceFg:    ac("28", "150"),
		StarFg:         ac("136", "222"),
		RunningFg:      ac("28", "150"),
		DateFg:         ac("242", "246"),
		ClaudeBadgeFg:  ac("255", "255"),
		ClaudeBadgeBg:  ac("67", "67"),
		CodexBadgeFg:   ac("255", "255"),
		CodexBadgeBg:   ac("173", "173"),
		PreviewTitleFg: ac("25", "110"),
		PreviewLabelFg: ac("240", "250"),
		PreviewValueFg: ac("234", "255"),
		StatusFg:       ac("240", "246"),
		StatusBg:       ac("253", "236"),
		StatusKeyFg:    ac("25", "110"),
	},
}

var currentTheme = themes["default"]

func SetTheme(name string) {
	if t, ok := themes[name]; ok {
		currentTheme = t
		applyTheme()
	}
}

func ThemeNames() []string {
	return []string{"default", "monokai", "dracula", "catppuccin", "nord"}
}

func applyTheme() {
	t := currentTheme

	panelBorder = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.BorderFg)

	activePanelBorder = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.ActiveBorderFg)

	headerStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.HeaderFg).
		Background(t.HeaderBg).
		Padding(0, 1)

	headerDimStyle = lipgloss.NewStyle().
		Foreground(t.HeaderDimFg).
		Background(t.HeaderBg)

	selectedRowStyle = lipgloss.NewStyle().
		Foreground(t.SelectedFg).
		Background(t.SelectedBg).
		Bold(true)

	normalRowStyle = lipgloss.NewStyle().
		Foreground(t.NormalFg)

	dimStyle = lipgloss.NewStyle().
		Foreground(t.DimFg)

	workspaceStyle = lipgloss.NewStyle().
		Foreground(t.WorkspaceFg).
		Bold(true)

	workspaceIconStyle = lipgloss.NewStyle().
		Foreground(t.WorkspaceFg)

	runningStyle = lipgloss.NewStyle().
		Foreground(t.RunningFg).
		Bold(true)

	claudeBadge = lipgloss.NewStyle().
		Foreground(t.ClaudeBadgeFg).
		Background(t.ClaudeBadgeBg).
		Padding(0, 1).
		Bold(true)

	codexBadge = lipgloss.NewStyle().
		Foreground(t.CodexBadgeFg).
		Background(t.CodexBadgeBg).
		Padding(0, 1).
		Bold(true)

	sourceClaudeStyle = lipgloss.NewStyle().Foreground(t.ClaudeFg)
	sourceCodexStyle = lipgloss.NewStyle().Foreground(t.CodexFg)

	previewTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.PreviewTitleFg)

	previewLabelStyle = lipgloss.NewStyle().
		Foreground(t.PreviewLabelFg)

	previewValueStyle = lipgloss.NewStyle().
		Foreground(t.PreviewValueFg)

	statusBarStyle = lipgloss.NewStyle().
		Foreground(t.StatusFg).
		Background(t.StatusBg).
		Padding(0, 1)

	statusKeyStyle = lipgloss.NewStyle().
		Foreground(t.StatusKeyFg).
		Background(t.StatusBg).
		Bold(true)

	dateStyle = lipgloss.NewStyle().Foreground(t.DateFg)
	starStyle = lipgloss.NewStyle().Foreground(t.StarFg)
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(t.ActiveBorderFg)
}
