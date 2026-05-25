package provider

import (
	"path/filepath"
	"testing"
)

func TestCodexProvider_ListSessions(t *testing.T) {
	p := NewCodexProvider(filepath.Join(testdataDir(), "codex"))
	sessions, err := p.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions error: %v", err)
	}

	var found *Session
	for i := range sessions {
		if sessions[i].ID == "test-session-id" {
			found = &sessions[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("session test-session-id not found; got %v", sessions)
	}
	if found.Source != "codex" {
		t.Errorf("Source: got %q, want %q", found.Source, "codex")
	}
	if found.Title != "优化下单链路 cache" {
		t.Errorf("Title: got %q, want %q", found.Title, "优化下单链路 cache")
	}
	if found.Directory != "/Users/test/spectra" {
		t.Errorf("Directory: got %q, want %q", found.Directory, "/Users/test/spectra")
	}
}

func TestCodexProvider_SearchContent(t *testing.T) {
	p := NewCodexProvider(filepath.Join(testdataDir(), "codex"))
	matches, err := p.SearchContent("test-session-id", "cache")
	if err != nil {
		t.Fatalf("SearchContent error: %v", err)
	}
	if len(matches) == 0 {
		t.Fatal("expected at least one match, got none")
	}
	if matches[0].SessionID != "test-session-id" {
		t.Errorf("SessionID: got %q, want %q", matches[0].SessionID, "test-session-id")
	}
}

func TestCodexProvider_ResumeCommand(t *testing.T) {
	p := NewCodexProvider("")
	cmd := p.ResumeCommand("test-session-id")
	want := []string{"codex", "resume", "test-session-id"}
	if len(cmd) != len(want) {
		t.Fatalf("ResumeCommand: got %v, want %v", cmd, want)
	}
	for i := range want {
		if cmd[i] != want[i] {
			t.Errorf("ResumeCommand[%d]: got %q, want %q", i, cmd[i], want[i])
		}
	}
}
