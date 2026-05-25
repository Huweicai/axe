package tui

import (
	"os/exec"
	"sort"
	"strings"
	"syscall"

	"github.com/Huweicai/axe/provider"
	"github.com/Huweicai/axe/state"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type sourceFilter int

const (
	filterAll sourceFilter = iota
	filterClaude
	filterCodex
)

func (f sourceFilter) String() string {
	switch f {
	case filterClaude:
		return "claude"
	case filterCodex:
		return "codex"
	default:
		return "all"
	}
}

// Model is the top-level Bubbletea model for axe.
type Model struct {
	items     []ListItem
	filtered  []int // indices into items
	cursor    int
	query     string
	width     int
	height    int
	filter    sourceFilter
	showDone  bool
	grouped   bool
	deepMode  bool
	noteInput bool
	noteText  string
	providers []provider.Provider
	state     *state.State
	config    *state.Config
	running   map[string]bool
	execArgs  []string // set before quitting to exec after tea exits
}

// New creates a Model populated with workspace and session items.
func New(providers []provider.Provider, cfg *state.Config, st *state.State, running map[string]bool) Model {
	m := Model{
		providers: providers,
		config:    cfg,
		state:     st,
		running:   running,
		width:     80,
		height:    24,
	}
	m.loadItems()
	m.applyFilter()
	return m
}

// SetShowDone toggles the --all flag for showing done sessions.
func (m *Model) SetShowDone(v bool) {
	m.showDone = v
	m.applyFilter()
}

// ExecArgs returns the command + args set by the last quit action.
func (m Model) ExecArgs() []string {
	return m.execArgs
}

// ExecReplace replaces the current process with the given command.
func ExecReplace(args []string) error {
	if len(args) == 0 {
		return nil
	}
	bin, err := exec.LookPath(args[0])
	if err != nil {
		return err
	}
	return syscall.Exec(bin, args, syscall.Environ())
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case deepSearchDone:
		m.handleDeepSearchDone(msg)
		return m, nil

	case tea.KeyMsg:
		if m.noteInput {
			return m.updateNoteInput(msg)
		}
		return m.updateNormal(msg)
	}
	return m, nil
}

// updateNoteInput handles keys while editing a note.
func (m Model) updateNoteInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Enter):
		if sel := m.selected(); sel != nil && sel.Kind == KindSession {
			m.state.SetNote(sel.StateKey(), m.noteText)
			sel.Note = m.noteText
			sel.MatchText = buildMatchText(*sel)
			_ = m.state.Save()
		}
		m.noteInput = false
		m.noteText = ""
		return m, nil

	case key.Matches(msg, keys.Escape):
		m.noteInput = false
		m.noteText = ""
		return m, nil

	case key.Matches(msg, keys.Backspace):
		if len(m.noteText) > 0 {
			m.noteText = m.noteText[:len(m.noteText)-1]
		}
		return m, nil

	default:
		if msg.Type == tea.KeyRunes {
			m.noteText += string(msg.Runes)
		}
		return m, nil
	}
}

// updateNormal handles keys in the main browsing mode.
func (m Model) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Special keys first
	switch {
	case key.Matches(msg, keys.Escape):
		if m.query != "" {
			m.query = ""
			m.deepMode = false
			m.applyFilter()
			return m, nil
		}
		return m, tea.Quit

	case key.Matches(msg, keys.Enter):
		if m.deepMode && m.query != "" {
			return m, m.startDeepSearch()
		}
		return m.execSelected(false)

	case key.Matches(msg, keys.Backspace):
		if len(m.query) > 0 {
			m.query = m.query[:len(m.query)-1]
			m.applyFilter()
		}
		return m, nil

	case key.Matches(msg, keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil

	case key.Matches(msg, keys.Down):
		if m.cursor < len(m.filtered)-1 {
			m.cursor++
		}
		return m, nil

	case key.Matches(msg, keys.TabFilter):
		m.filter = (m.filter + 1) % 3
		m.applyFilter()
		return m, nil

	case key.Matches(msg, keys.ToggleDone):
		m.showDone = !m.showDone
		m.applyFilter()
		return m, nil
	}

	// Rune-based keys — only when query is empty (otherwise they type into search)
	if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 {
		ch := msg.Runes[0]

		if m.query == "" {
			switch ch {
			case 'q':
				return m, tea.Quit
			case 'x':
				return m.execSelected(true)
			case 'd':
				m.markDone()
				return m, nil
			case 'u':
				m.undoDone()
				return m, nil
			case 'n':
				if sel := m.selected(); sel != nil && sel.Kind == KindSession {
					m.noteInput = true
					m.noteText = sel.Note
				}
				return m, nil
			case '/':
				m.deepMode = !m.deepMode
				return m, nil
			case 'g':
				m.grouped = !m.grouped
				return m, nil
			case 'j':
				if m.cursor < len(m.filtered)-1 {
					m.cursor++
				}
				return m, nil
			case 'k':
				if m.cursor > 0 {
					m.cursor--
				}
				return m, nil
			}
		}

		// Not a command key — treat as search input
		m.query += string(msg.Runes)
		m.applyFilter()
		return m, nil
	}

	return m, nil
}

// loadItems builds the full item list from providers and config.
func (m *Model) loadItems() {
	m.items = nil

	// Pinned workspaces
	if m.config != nil {
		for _, ws := range m.config.PinnedWorkspaces {
			li := ListItem{
				Kind:    KindWorkspace,
				Alias:   ws.Alias,
				Path:    ws.Path,
				Running: m.running[ws.Path],
			}
			li.MatchText = buildMatchText(li)
			m.items = append(m.items, li)
		}
	}

	// Sessions from all providers
	for _, p := range m.providers {
		sessions, err := p.ListSessions()
		if err != nil {
			continue
		}
		for i := range sessions {
			s := &sessions[i]
			sk := s.Source + ":" + s.ID
			li := ListItem{
				Kind:    KindSession,
				Path:    s.Directory,
				Session: s,
				Done:    m.state.IsDone(sk),
				Note:    m.state.GetNote(sk),
				Running: m.running[s.Directory],
			}
			li.MatchText = buildMatchText(li)
			m.items = append(m.items, li)
		}
	}

	// Sort sessions by UpdatedAt descending (workspaces stay at top)
	sort.SliceStable(m.items, func(i, j int) bool {
		a, b := m.items[i], m.items[j]
		if a.Kind != b.Kind {
			return a.Kind < b.Kind // workspaces first
		}
		if a.Kind == KindSession {
			return a.UpdatedAt().After(b.UpdatedAt())
		}
		return false
	})
}

// applyFilter rebuilds m.filtered from query, source filter, and done status.
func (m *Model) applyFilter() {
	q := strings.ToLower(m.query)
	m.filtered = nil

	for i, it := range m.items {
		// source filter
		if m.filter == filterClaude && it.Source() != "" && it.Source() != "claude" {
			continue
		}
		if m.filter == filterCodex && it.Source() != "" && it.Source() != "codex" {
			continue
		}
		// done filter
		if !m.showDone && it.Done {
			continue
		}
		// query filter
		if q != "" && !strings.Contains(it.MatchText, q) {
			continue
		}
		m.filtered = append(m.filtered, i)
	}

	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
}

// selected returns a pointer to the currently selected item, or nil.
func (m *Model) selected() *ListItem {
	if m.cursor < 0 || m.cursor >= len(m.filtered) {
		return nil
	}
	return &m.items[m.filtered[m.cursor]]
}

// markDone marks the selected session as done.
func (m *Model) markDone() {
	sel := m.selected()
	if sel == nil || sel.Kind != KindSession {
		return
	}
	m.state.MarkDone(sel.StateKey())
	sel.Done = true
	_ = m.state.Save()
	m.applyFilter()
}

// undoDone clears done status on the selected session.
func (m *Model) undoDone() {
	sel := m.selected()
	if sel == nil || sel.Kind != KindSession {
		return
	}
	m.state.UndoDone(sel.StateKey())
	sel.Done = false
	_ = m.state.Save()
	m.applyFilter()
}

// execSelected sets execArgs for the selected item and quits.
func (m Model) execSelected(altTool bool) (tea.Model, tea.Cmd) {
	sel := m.selected()
	if sel == nil {
		return m, nil
	}

	switch sel.Kind {
	case KindWorkspace:
		tool := m.defaultTool()
		if altTool {
			tool = m.altTool()
		}
		m.execArgs = providerNewCommand(tool, sel.Path)
	case KindSession:
		if altTool {
			// no alt for sessions — ignore
			return m, nil
		}
		m.execArgs = m.providerResumeCommand(sel.Session)
	}

	return m, tea.Quit
}

func (m Model) defaultTool() string {
	if m.config != nil && m.config.DefaultTool != "" {
		return m.config.DefaultTool
	}
	return "claude"
}

func (m Model) altTool() string {
	if m.defaultTool() == "claude" {
		return "codex"
	}
	return "claude"
}

func providerNewCommand(tool, cwd string) []string {
	if tool == "codex" {
		return []string{"codex", "--cwd", cwd}
	}
	return []string{"claude", "--cwd", cwd}
}

func (m Model) providerResumeCommand(s *provider.Session) []string {
	for _, p := range m.providers {
		if p.Name() == s.Source {
			return p.ResumeCommand(s.ID)
		}
	}
	return nil
}

// View implements tea.Model.
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	header := m.renderHeader(m.width)
	statusBar := m.renderStatusBar(m.width)

	// Split: 55% list, 45% preview
	listWidth := m.width * 55 / 100
	previewWidth := m.width - listWidth
	bodyHeight := m.height - 3 // header + status bar + separator

	// Build list rows
	var listLines []string
	if m.grouped {
		listLines = m.renderGroupedList(listWidth, bodyHeight)
	} else {
		for i, fi := range m.filtered {
			if i >= bodyHeight {
				break
			}
			listLines = append(listLines, m.renderListRow(fi, listWidth, i == m.cursor))
		}
	}
	// Pad list to bodyHeight
	for len(listLines) < bodyHeight {
		listLines = append(listLines, "")
	}

	listPane := strings.Join(listLines, "\n")
	previewPane := m.renderPreview(previewWidth)

	body := lipgloss.JoinHorizontal(lipgloss.Top, listPane, previewPane)

	return header + "\n" + body + "\n" + statusBar
}

// renderGroupedList renders sessions grouped by directory with headers.
func (m *Model) renderGroupedList(listWidth, maxRows int) []string {
	var lines []string
	lastDir := ""
	row := 0

	for i, fi := range m.filtered {
		if row >= maxRows {
			break
		}
		it := m.items[fi]

		if it.Kind == KindSession && it.Path != lastDir {
			// Directory header
			if row > 0 {
				lines = append(lines, "")
				row++
				if row >= maxRows {
					break
				}
			}
			dirLabel := dimStyle.Render("── " + shortenPath(it.Path, listWidth-5) + " ──")
			lines = append(lines, dirLabel)
			row++
			lastDir = it.Path
			if row >= maxRows {
				break
			}
		}

		lines = append(lines, m.renderListRow(fi, listWidth, i == m.cursor))
		row++
	}
	return lines
}
