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
	if len(sessions) != 3 {
		t.Fatalf("expected 3 sessions, got %d", len(sessions))
	}

	byID := make(map[string]Session)
	for _, s := range sessions {
		byID[s.ID] = s
	}

	s, ok := byID["abc123"]
	if !ok {
		t.Fatalf("session abc123 not found; got %v", sessions)
	}
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

	if got := byID["custom-title"].Title; got != "东证白云机器对时比较性能" {
		t.Errorf("custom-title Title: got %q, want %q", got, "东证白云机器对时比较性能")
	}
	if got := byID["auto-title"].Title; got != "MySQL 迁移到 CK" {
		t.Errorf("auto-title Title: got %q, want %q", got, "MySQL 迁移到 CK")
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
