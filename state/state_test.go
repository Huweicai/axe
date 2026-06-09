package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadConfig_Default(t *testing.T) {
	dir := t.TempDir()
	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DefaultTool != "claude" {
		t.Errorf("expected DefaultTool=claude, got %q", cfg.DefaultTool)
	}
}

func TestLoadConfig_FromFile(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		DefaultTool: "codex",
		PinnedWorkspaces: []Workspace{
			{Alias: "myproj", Path: "/home/user/proj"},
		},
	}
	data, _ := json.Marshal(cfg)
	os.WriteFile(filepath.Join(dir, "config.json"), data, 0644)

	loaded, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loaded.DefaultTool != "codex" {
		t.Errorf("expected codex, got %q", loaded.DefaultTool)
	}
	if len(loaded.PinnedWorkspaces) != 1 || loaded.PinnedWorkspaces[0].Alias != "myproj" {
		t.Errorf("pinned workspaces mismatch: %+v", loaded.PinnedWorkspaces)
	}
}

func TestState_Archive(t *testing.T) {
	dir := t.TempDir()
	s, err := LoadState(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	key := "/some/project"
	if s.IsArchived(key) {
		t.Error("expected not archived initially")
	}
	s.MarkArchived(key)
	if !s.IsArchived(key) {
		t.Error("expected archived after MarkArchived")
	}
	s.UnmarkArchived(key)
	if s.IsArchived(key) {
		t.Error("expected not archived after UnmarkArchived")
	}
}

func TestState_DeleteAndRestore(t *testing.T) {
	s, _ := LoadState(t.TempDir())
	key := "claude:abc"

	if s.IsDeleted(key) {
		t.Error("expected not deleted initially")
	}
	now := time.Now()
	s.MarkDeleted(key, "/tmp/abc.jsonl", now)
	if !s.IsDeleted(key) {
		t.Fatal("expected deleted after MarkDeleted")
	}
	if got := s.DeletedAt(key); got != now.Unix() {
		t.Errorf("DeletedAt: got %d, want %d", got, now.Unix())
	}
	s.Restore(key)
	if s.IsDeleted(key) {
		t.Error("expected not deleted after Restore")
	}
}

func TestState_TakeExpiredDeletions(t *testing.T) {
	s, _ := LoadState(t.TempDir())
	now := time.Now()

	s.MarkDeleted("claude:old", "/tmp/old.jsonl", now.Add(-8*24*time.Hour))
	s.MarkDeleted("claude:fresh", "/tmp/fresh.jsonl", now.Add(-1*time.Hour))

	files := s.TakeExpiredDeletions(now, TrashTTL)
	if len(files) != 1 || files[0] != "/tmp/old.jsonl" {
		t.Fatalf("expected [/tmp/old.jsonl], got %v", files)
	}
	if _, ok := s.Sessions["claude:old"]; ok {
		t.Error("expired entry should have been removed from state")
	}
	if !s.IsDeleted("claude:fresh") {
		t.Error("fresh entry should still be in the recycle bin")
	}
}

func TestState_MigrateDoneToArchived(t *testing.T) {
	dir := t.TempDir()
	// Write legacy state with the old "done" field.
	legacy := `{"sessions":{"claude:x":{"done":true,"note":"keep"}}}`
	os.WriteFile(filepath.Join(dir, "state.json"), []byte(legacy), 0644)

	s, err := LoadState(dir)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if !s.IsArchived("claude:x") {
		t.Error("legacy done=true should migrate to archived")
	}
	if s.Sessions["claude:x"].Done {
		t.Error("legacy Done field should be cleared after migration")
	}
	if s.GetNote("claude:x") != "keep" {
		t.Error("note should survive migration")
	}
}

func TestState_Notes(t *testing.T) {
	dir := t.TempDir()
	s, _ := LoadState(dir)

	key := "/some/project"
	if s.GetNote(key) != "" {
		t.Error("expected empty note initially")
	}
	s.SetNote(key, "hello world")
	if s.GetNote(key) != "hello world" {
		t.Errorf("expected 'hello world', got %q", s.GetNote(key))
	}
}

func TestState_Persistence(t *testing.T) {
	dir := t.TempDir()
	s, _ := LoadState(dir)

	key := "/persist/project"
	s.MarkArchived(key)
	s.SetNote(key, "persisted note")
	if err := s.Save(); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	s2, err := LoadState(dir)
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	if !s2.IsArchived(key) {
		t.Error("expected archived after reload")
	}
	if s2.GetNote(key) != "persisted note" {
		t.Errorf("expected 'persisted note', got %q", s2.GetNote(key))
	}
}
