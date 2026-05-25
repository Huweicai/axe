package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Huweicai/axe/process"
	"github.com/Huweicai/axe/provider"
	"github.com/Huweicai/axe/state"
	"github.com/Huweicai/axe/tui"
)

func main() {
	showAll := flag.Bool("all", false, "show done sessions")
	flag.Parse()

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	axeDir := filepath.Join(home, ".axe")
	os.MkdirAll(axeDir, 0755)

	cfg, err := state.LoadConfig(axeDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		os.Exit(1)
	}

	st, err := state.LoadState(axeDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading state: %v\n", err)
		os.Exit(1)
	}

	claudeDir := filepath.Join(home, ".claude")
	codexDir := filepath.Join(home, ".codex")

	var providers []provider.Provider
	if _, err := os.Stat(filepath.Join(claudeDir, "projects")); err == nil {
		providers = append(providers, provider.NewClaudeProvider(claudeDir))
	}
	if _, err := os.Stat(filepath.Join(codexDir, "session_index.jsonl")); err == nil {
		providers = append(providers, provider.NewCodexProvider(codexDir))
	}

	if len(providers) == 0 {
		fmt.Fprintln(os.Stderr, "No Claude or Codex sessions found")
		os.Exit(1)
	}

	// Task 11: Bootstrap config on first run
	configPath := filepath.Join(axeDir, "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		cfg.PinnedWorkspaces = inferWorkspaces(providers)
		data, _ := json.MarshalIndent(cfg, "", "  ")
		os.WriteFile(configPath, data, 0644)
	}

	running := process.DetectRunning()

	model := tui.New(providers, cfg, st, running)
	if *showAll {
		model.SetShowDone(true)
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if m, ok := finalModel.(tui.Model); ok {
		if args := m.ExecArgs(); len(args) > 0 {
			if err := tui.ExecReplace(args); err != nil {
				fmt.Fprintf(os.Stderr, "exec error: %v\n", err)
				os.Exit(1)
			}
		}
	}
}

func inferWorkspaces(providers []provider.Provider) []state.Workspace {
	dirCount := make(map[string]int)
	for _, p := range providers {
		sessions, err := p.ListSessions()
		if err != nil {
			continue
		}
		for _, s := range sessions {
			if s.Directory != "" {
				dirCount[s.Directory]++
			}
		}
	}

	type kv struct {
		dir   string
		count int
	}
	var sorted []kv
	for d, c := range dirCount {
		sorted = append(sorted, kv{d, c})
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].count > sorted[j].count })

	var workspaces []state.Workspace
	for i, kv := range sorted {
		if i >= 5 {
			break
		}
		alias := filepath.Base(kv.dir)
		workspaces = append(workspaces, state.Workspace{Alias: alias, Path: kv.dir})
	}
	return workspaces
}
