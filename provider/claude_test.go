package provider

import (
	"path/filepath"
	"runtime"
	"testing"
)

func testdataDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "testdata")
}

func TestClaudeProvider_ListSessions(t *testing.T) {
	p := NewClaudeProvider(filepath.Join(testdataDir(), "claude"))
	sessions, err := p.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions error: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	s := sessions[0]
	if s.ID != "abc123" {
		t.Errorf("ID: got %q, want %q", s.ID, "abc123")
	}
	if s.Source != "claude" {
		t.Errorf("Source: got %q, want %q", s.Source, "claude")
	}
	if s.Directory != "/Users/test/myproject" {
		t.Errorf("Directory: got %q, want %q", s.Directory, "/Users/test/myproject")
	}
	if s.Title != "Help me refactor the auth module" {
		t.Errorf("Title: got %q, want %q", s.Title, "Help me refactor the auth module")
	}
}

func TestClaudeProvider_SearchContent(t *testing.T) {
	p := NewClaudeProvider(filepath.Join(testdataDir(), "claude"))
	matches, err := p.SearchContent("abc123", "refactor")
	if err != nil {
		t.Fatalf("SearchContent error: %v", err)
	}
	if len(matches) == 0 {
		t.Fatal("expected at least one match, got none")
	}
	if matches[0].SessionID != "abc123" {
		t.Errorf("SessionID: got %q, want %q", matches[0].SessionID, "abc123")
	}
}

func TestClaudeProvider_ResumeCommand(t *testing.T) {
	p := NewClaudeProvider("")
	cmd := p.ResumeCommand("abc123")
	want := []string{"claude", "-r", "abc123"}
	if len(cmd) != len(want) {
		t.Fatalf("ResumeCommand: got %v, want %v", cmd, want)
	}
	for i := range want {
		if cmd[i] != want[i] {
			t.Errorf("ResumeCommand[%d]: got %q, want %q", i, cmd[i], want[i])
		}
	}
}

func TestClaudeProvider_NewCommand(t *testing.T) {
	p := NewClaudeProvider("")
	cmd := p.NewCommand("/some/path")
	want := []string{"claude", "--cwd", "/some/path"}
	if len(cmd) != len(want) {
		t.Fatalf("NewCommand: got %v, want %v", cmd, want)
	}
	for i := range want {
		if cmd[i] != want[i] {
			t.Errorf("NewCommand[%d]: got %q, want %q", i, cmd[i], want[i])
		}
	}
}
