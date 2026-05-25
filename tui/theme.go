package tui

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Name string

	// Panel borders
	BorderFg       lipgloss.Color
	ActiveBorderFg lipgloss.Color

	// Header
	HeaderFg    lipgloss.Color
	HeaderBg    lipgloss.Color
	HeaderDimFg lipgloss.Color

	// Selected row
	SelectedFg lipgloss.Color
	SelectedBg lipgloss.Color

	// Normal text
	NormalFg lipgloss.Color
	DimFg    lipgloss.Color

	// Source colors
	ClaudeFg lipgloss.Color
	CodexFg  lipgloss.Color

	// Accents
	WorkspaceFg lipgloss.Color
	StarFg      lipgloss.Color
	RunningFg   lipgloss.Color
	DateFg      lipgloss.Color

	// Badges (preview pane)
	ClaudeBadgeFg lipgloss.Color
	ClaudeBadgeBg lipgloss.Color
	CodexBadgeFg  lipgloss.Color
	CodexBadgeBg  lipgloss.Color

	// Preview
	PreviewTitleFg lipgloss.Color
	PreviewLabelFg lipgloss.Color
	PreviewValueFg lipgloss.Color

	// Status bar
	StatusFg    lipgloss.Color
	StatusBg    lipgloss.Color
	StatusKeyFg lipgloss.Color
}

var themes = map[string]Theme{
	"default": {
		Name:           "default",
		BorderFg:       lipgloss.Color("238"),
		ActiveBorderFg: lipgloss.Color("99"),
		HeaderFg:       lipgloss.Color("255"),
		HeaderBg:       lipgloss.Color("63"),
		HeaderDimFg:    lipgloss.Color("189"),
		SelectedFg:     lipgloss.Color("255"),
		SelectedBg:     lipgloss.Color("62"),
		NormalFg:       lipgloss.Color("252"),
		DimFg:          lipgloss.Color("243"),
		ClaudeFg:       lipgloss.Color("111"),
		CodexFg:        lipgloss.Color("215"),
		WorkspaceFg:    lipgloss.Color("220"),
		StarFg:         lipgloss.Color("220"),
		RunningFg:      lipgloss.Color("82"),
		DateFg:         lipgloss.Color("243"),
		ClaudeBadgeFg:  lipgloss.Color("255"),
		ClaudeBadgeBg:  lipgloss.Color("62"),
		CodexBadgeFg:   lipgloss.Color("255"),
		CodexBadgeBg:   lipgloss.Color("172"),
		PreviewTitleFg: lipgloss.Color("99"),
		PreviewLabelFg: lipgloss.Color("250"),
		PreviewValueFg: lipgloss.Color("255"),
		StatusFg:       lipgloss.Color("245"),
		StatusBg:       lipgloss.Color("236"),
		StatusKeyFg:    lipgloss.Color("117"),
	},
	"monokai": {
		Name:           "monokai",
		BorderFg:       lipgloss.Color("239"),
		ActiveBorderFg: lipgloss.Color("148"),
		HeaderFg:       lipgloss.Color("232"),
		HeaderBg:       lipgloss.Color("148"),
		HeaderDimFg:    lipgloss.Color("236"),
		SelectedFg:     lipgloss.Color("232"),
		SelectedBg:     lipgloss.Color("148"),
		NormalFg:       lipgloss.Color("253"),
		DimFg:          lipgloss.Color("242"),
		ClaudeFg:       lipgloss.Color("81"),
		CodexFg:        lipgloss.Color("208"),
		WorkspaceFg:    lipgloss.Color("186"),
		StarFg:         lipgloss.Color("186"),
		RunningFg:      lipgloss.Color("148"),
		DateFg:         lipgloss.Color("242"),
		ClaudeBadgeFg:  lipgloss.Color("232"),
		ClaudeBadgeBg:  lipgloss.Color("81"),
		CodexBadgeFg:   lipgloss.Color("232"),
		CodexBadgeBg:   lipgloss.Color("208"),
		PreviewTitleFg: lipgloss.Color("148"),
		PreviewLabelFg: lipgloss.Color("250"),
		PreviewValueFg: lipgloss.Color("255"),
		StatusFg:       lipgloss.Color("242"),
		StatusBg:       lipgloss.Color("235"),
		StatusKeyFg:    lipgloss.Color("81"),
	},
	"dracula": {
		Name:           "dracula",
		BorderFg:       lipgloss.Color("61"),
		ActiveBorderFg: lipgloss.Color("141"),
		HeaderFg:       lipgloss.Color("255"),
		HeaderBg:       lipgloss.Color("141"),
		HeaderDimFg:    lipgloss.Color("189"),
		SelectedFg:     lipgloss.Color("255"),
		SelectedBg:     lipgloss.Color("141"),
		NormalFg:       lipgloss.Color("253"),
		DimFg:          lipgloss.Color("103"),
		ClaudeFg:       lipgloss.Color("117"),
		CodexFg:        lipgloss.Color("215"),
		WorkspaceFg:    lipgloss.Color("84"),
		StarFg:         lipgloss.Color("228"),
		RunningFg:      lipgloss.Color("84"),
		DateFg:         lipgloss.Color("103"),
		ClaudeBadgeFg:  lipgloss.Color("255"),
		ClaudeBadgeBg:  lipgloss.Color("61"),
		CodexBadgeFg:   lipgloss.Color("255"),
		CodexBadgeBg:   lipgloss.Color("166"),
		PreviewTitleFg: lipgloss.Color("141"),
		PreviewLabelFg: lipgloss.Color("250"),
		PreviewValueFg: lipgloss.Color("255"),
		StatusFg:       lipgloss.Color("103"),
		StatusBg:       lipgloss.Color("236"),
		StatusKeyFg:    lipgloss.Color("141"),
	},
	"catppuccin": {
		Name:           "catppuccin",
		BorderFg:       lipgloss.Color("240"),
		ActiveBorderFg: lipgloss.Color("183"),
		HeaderFg:       lipgloss.Color("232"),
		HeaderBg:       lipgloss.Color("183"),
		HeaderDimFg:    lipgloss.Color("236"),
		SelectedFg:     lipgloss.Color("232"),
		SelectedBg:     lipgloss.Color("183"),
		NormalFg:       lipgloss.Color("254"),
		DimFg:          lipgloss.Color("245"),
		ClaudeFg:       lipgloss.Color("110"),
		CodexFg:        lipgloss.Color("216"),
		WorkspaceFg:    lipgloss.Color("114"),
		StarFg:         lipgloss.Color("222"),
		RunningFg:      lipgloss.Color("114"),
		DateFg:         lipgloss.Color("245"),
		ClaudeBadgeFg:  lipgloss.Color("232"),
		ClaudeBadgeBg:  lipgloss.Color("110"),
		CodexBadgeFg:   lipgloss.Color("232"),
		CodexBadgeBg:   lipgloss.Color("216"),
		PreviewTitleFg: lipgloss.Color("183"),
		PreviewLabelFg: lipgloss.Color("250"),
		PreviewValueFg: lipgloss.Color("255"),
		StatusFg:       lipgloss.Color("245"),
		StatusBg:       lipgloss.Color("236"),
		StatusKeyFg:    lipgloss.Color("183"),
	},
	"nord": {
		Name:           "nord",
		BorderFg:       lipgloss.Color("60"),
		ActiveBorderFg: lipgloss.Color("110"),
		HeaderFg:       lipgloss.Color("255"),
		HeaderBg:       lipgloss.Color("60"),
		HeaderDimFg:    lipgloss.Color("146"),
		SelectedFg:     lipgloss.Color("255"),
		SelectedBg:     lipgloss.Color("60"),
		NormalFg:       lipgloss.Color("254"),
		DimFg:          lipgloss.Color("246"),
		ClaudeFg:       lipgloss.Color("110"),
		CodexFg:        lipgloss.Color("173"),
		WorkspaceFg:    lipgloss.Color("150"),
		StarFg:         lipgloss.Color("222"),
		RunningFg:      lipgloss.Color("150"),
		DateFg:         lipgloss.Color("246"),
		ClaudeBadgeFg:  lipgloss.Color("255"),
		ClaudeBadgeBg:  lipgloss.Color("67"),
		CodexBadgeFg:   lipgloss.Color("255"),
		CodexBadgeBg:   lipgloss.Color("173"),
		PreviewTitleFg: lipgloss.Color("110"),
		PreviewLabelFg: lipgloss.Color("246"),
		PreviewValueFg: lipgloss.Color("254"),
		StatusFg:       lipgloss.Color("246"),
		StatusBg:       lipgloss.Color("236"),
		StatusKeyFg:    lipgloss.Color("110"),
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
