package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
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

func TestState_MarkDone(t *testing.T) {
	dir := t.TempDir()
	s, err := LoadState(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	key := "/some/project"
	if s.IsDone(key) {
		t.Error("expected not done initially")
	}
	s.MarkDone(key)
	if !s.IsDone(key) {
		t.Error("expected done after MarkDone")
	}
	s.UndoDone(key)
	if s.IsDone(key) {
		t.Error("expected not done after UndoDone")
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
	s.MarkDone(key)
	s.SetNote(key, "persisted note")
	if err := s.Save(); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	s2, err := LoadState(dir)
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	if !s2.IsDone(key) {
		t.Error("expected done after reload")
	}
	if s2.GetNote(key) != "persisted note" {
		t.Errorf("expected 'persisted note', got %q", s2.GetNote(key))
	}
}
