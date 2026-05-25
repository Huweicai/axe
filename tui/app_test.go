package tui

import (
	"testing"
	"time"

	"github.com/Huweicai/axe/provider"
	"github.com/Huweicai/axe/state"
	tea "github.com/charmbracelet/bubbletea"
)

// mockProvider implements provider.Provider with canned data.
type mockProvider struct {
	name     string
	sessions []provider.Session
}

func (p *mockProvider) Name() string { return p.name }
func (p *mockProvider) ListSessions() ([]provider.Session, error) {
	return p.sessions, nil
}
func (p *mockProvider) SearchContent(sessionID, keyword string) ([]provider.Match, error) {
	return nil, nil
}
func (p *mockProvider) ResumeCommand(sessionID string) []string {
	return []string{p.name, "-r", sessionID}
}
func (p *mockProvider) NewCommand(cwd string) []string {
	return []string{p.name, "--cwd", cwd}
}

func newTestModel() Model {
	now := time.Now()
	mp := &mockProvider{
		name: "claude",
		sessions: []provider.Session{
			{ID: "s1", Source: "claude", Title: "Fix cache invalidation", Directory: "/home/user/proj1", UpdatedAt: now.Add(-1 * time.Hour), FileSize: 2200000},
			{ID: "s2", Source: "claude", Title: "Add cache warmup logic", Directory: "/home/user/proj2", UpdatedAt: now.Add(-2 * time.Hour), FileSize: 1100000},
			{ID: "s3", Source: "claude", Title: "Refactor auth module", Directory: "/home/user/proj3", UpdatedAt: now.Add(-3 * time.Hour), FileSize: 500000},
		},
	}

	cfg := &state.Config{
		DefaultTool:      "claude",
		PinnedWorkspaces: []state.Workspace{{Alias: "mywork", Path: "/home/user/work"}},
	}
	st := &state.State{Sessions: make(map[string]*state.SessionState)}
	running := map[string]bool{"/home/user/proj1": true}

	return New([]provider.Provider{mp}, cfg, st, running)
}

func TestModel_InitialState(t *testing.T) {
	m := newTestModel()

	if len(m.items) != 4 {
		t.Fatalf("expected 4 items, got %d", len(m.items))
	}
	if len(m.filtered) != 4 {
		t.Fatalf("expected 4 filtered, got %d", len(m.filtered))
	}
	// First item should be workspace
	if m.items[m.filtered[0]].Kind != KindWorkspace {
		t.Error("first filtered item should be workspace")
	}
	// Sessions should follow
	for i := 1; i < len(m.filtered); i++ {
		if m.items[m.filtered[i]].Kind != KindSession {
			t.Errorf("filtered[%d] should be session", i)
		}
	}
	// First session (s1) should be running
	s1 := m.items[m.filtered[1]]
	if !s1.Running {
		t.Error("s1 should be running")
	}
}

func TestModel_QueryFilter(t *testing.T) {
	m := newTestModel()

	// Type "cache" — should match s1 ("Fix cache invalidation") and s2 ("Add cache warmup logic")
	for _, ch := range "cache" {
		var cmd tea.Cmd
		tm, c := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		m = tm.(Model)
		cmd = c
		_ = cmd
	}

	if m.query != "cache" {
		t.Fatalf("query: got %q, want %q", m.query, "cache")
	}
	if len(m.filtered) != 2 {
		t.Fatalf("expected 2 filtered after 'cache', got %d", len(m.filtered))
	}
}

func TestModel_ToggleDone(t *testing.T) {
	m := newTestModel()
	initialCount := len(m.filtered)

	// Move cursor to first session (index 1 = first session after workspace)
	m.cursor = 1

	// Press 'd' to mark done
	tm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = tm.(Model)

	if len(m.filtered) != initialCount-1 {
		t.Fatalf("expected %d filtered after done, got %d", initialCount-1, len(m.filtered))
	}

	// The done session should still exist in items
	found := false
	for _, it := range m.items {
		if it.Done {
			found = true
			break
		}
	}
	if !found {
		t.Error("no item marked as done in items")
	}
}

func TestModel_SourceFilter(t *testing.T) {
	m := newTestModel()

	// Press tab — should switch to filterClaude
	tm, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = tm.(Model)

	if m.filter != filterClaude {
		t.Fatalf("filter: got %d, want filterClaude (%d)", m.filter, filterClaude)
	}

	// Press tab again — should switch to filterCodex
	tm, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = tm.(Model)

	if m.filter != filterCodex {
		t.Fatalf("filter: got %d, want filterCodex (%d)", m.filter, filterCodex)
	}

	// With codex filter, only workspace should show (no codex sessions in test data)
	// Workspaces pass through source filter (source == "")
	for _, idx := range m.filtered {
		it := m.items[idx]
		if it.Kind == KindSession {
			t.Error("no session should pass codex filter with claude-only data")
		}
	}
}

func TestModel_ExecArgs(t *testing.T) {
	m := newTestModel()

	// Cursor is at 0 = workspace
	tm, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = tm.(Model)

	if cmd == nil {
		t.Fatal("expected tea.Quit cmd, got nil")
	}

	args := m.ExecArgs()
	if len(args) != 3 {
		t.Fatalf("execArgs length: got %d, want 3", len(args))
	}
	if args[0] != "claude" || args[1] != "--cwd" || args[2] != "/home/user/work" {
		t.Errorf("execArgs: got %v, want [claude --cwd /home/user/work]", args)
	}
}

func TestModel_ExecAltTool(t *testing.T) {
	m := newTestModel()

	// Cursor at 0 = workspace, press 'x' for alt tool
	tm, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	m = tm.(Model)

	if cmd == nil {
		t.Fatal("expected tea.Quit cmd, got nil")
	}

	args := m.ExecArgs()
	if len(args) != 3 {
		t.Fatalf("execArgs length: got %d, want 3", len(args))
	}
	if args[0] != "codex" {
		t.Errorf("alt tool: got %q, want %q", args[0], "codex")
	}
}

func TestModel_NoteInput(t *testing.T) {
	m := newTestModel()
	m.cursor = 1 // first session

	// Press 'n' to enter note mode
	tm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = tm.(Model)

	if !m.noteInput {
		t.Fatal("should be in note input mode")
	}

	// Type "hello"
	for _, ch := range "hello" {
		tm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		m = tm.(Model)
	}
	if m.noteText != "hello" {
		t.Fatalf("noteText: got %q, want %q", m.noteText, "hello")
	}

	// Press enter to save
	tm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = tm.(Model)

	if m.noteInput {
		t.Error("should have exited note input mode")
	}
	sel := m.items[m.filtered[1]]
	if sel.Note != "hello" {
		t.Errorf("note: got %q, want %q", sel.Note, "hello")
	}
}
