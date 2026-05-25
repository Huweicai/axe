package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Huweicai/axe/process"
	"github.com/Huweicai/axe/provider"
	"github.com/Huweicai/axe/state"
	"github.com/Huweicai/axe/tui"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "ls":
			cmdLs(os.Args[2:])
			return
		case "search":
			cmdSearch(os.Args[2:])
			return
		case "done":
			cmdDone(os.Args[2:])
			return
		case "undone":
			cmdUndone(os.Args[2:])
			return
		case "note":
			cmdNote(os.Args[2:])
			return
		case "config":
			cmdConfig()
			return
		case "help", "--help", "-h":
			cmdHelp()
			return
		}
	}
	cmdTUI()
}

func cmdHelp() {
	fmt.Println(`axe — AI session manager

Usage:
  axe                        Launch TUI (default)
  axe ls [-a|--all]          List sessions
  axe search <keyword>       Deep search session content
  axe done <id-prefix>       Mark session as done
  axe undone <id-prefix>     Undo done mark
  axe note <id-prefix> <txt> Add/replace note on session
  axe config                 Show current config
  axe help                   Show this help

In TUI:
  enter  open/resume    d  mark done     /  deep search
  x      alt tool       u  undo done     tab  filter source
  n      add note       q  quit          ^a   show done`)
}

type context struct {
	providers []provider.Provider
	cfg       *state.Config
	st        *state.State
	axeDir    string
}

func loadContext() *context {
	home, _ := os.UserHomeDir()
	axeDir := filepath.Join(home, ".axe")
	os.MkdirAll(axeDir, 0755)

	cfg, _ := state.LoadConfig(axeDir)
	st, _ := state.LoadState(axeDir)

	claudeDir := filepath.Join(home, ".claude")
	codexDir := filepath.Join(home, ".codex")

	var providers []provider.Provider
	if _, err := os.Stat(filepath.Join(claudeDir, "projects")); err == nil {
		providers = append(providers, provider.NewClaudeProvider(claudeDir))
	}
	if _, err := os.Stat(filepath.Join(codexDir, "session_index.jsonl")); err == nil {
		providers = append(providers, provider.NewCodexProvider(codexDir))
	}

	configPath := filepath.Join(axeDir, "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		cfg.PinnedWorkspaces = inferWorkspaces(providers)
		data, _ := json.MarshalIndent(cfg, "", "  ")
		os.WriteFile(configPath, data, 0644)
	}

	return &context{providers: providers, cfg: cfg, st: st, axeDir: axeDir}
}

func allSessions(ctx *context) []provider.Session {
	var sessions []provider.Session
	for _, p := range ctx.providers {
		ss, err := p.ListSessions()
		if err != nil {
			continue
		}
		sessions = append(sessions, ss...)
	}
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].UpdatedAt.After(sessions[j].UpdatedAt)
	})
	return sessions
}

func findSession(ctx *context, prefix string) (*provider.Session, error) {
	prefix = strings.ToLower(prefix)
	var matches []provider.Session
	for _, s := range allSessions(ctx) {
		key := s.Source + ":" + s.ID
		if strings.HasPrefix(strings.ToLower(key), prefix) || strings.HasPrefix(strings.ToLower(s.ID), prefix) {
			matches = append(matches, s)
		}
	}
	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("no session matching %q", prefix)
	case 1:
		return &matches[0], nil
	default:
		var ids []string
		for _, m := range matches {
			ids = append(ids, m.Source+":"+m.ID[:8])
		}
		return nil, fmt.Errorf("ambiguous prefix %q, matches: %s", prefix, strings.Join(ids, ", "))
	}
}

func shortenDir(dir string) string {
	home, _ := os.UserHomeDir()
	if home != "" && strings.HasPrefix(dir, home) {
		return "~" + dir[len(home):]
	}
	return dir
}

// --- Commands ---

func cmdLs(args []string) {
	showAll := false
	for _, a := range args {
		if a == "--all" || a == "-a" {
			showAll = true
		}
	}

	ctx := loadContext()
	sessions := allSessions(ctx)

	for _, s := range sessions {
		key := s.Source + ":" + s.ID
		done := ctx.st.IsDone(key)
		if done && !showAll {
			continue
		}
		note := ctx.st.GetNote(key)

		title := s.Title
		if title == "" {
			title = "(no title)"
		}
		runes := []rune(title)
		if len(runes) > 50 {
			title = string(runes[:50]) + ".."
		}

		mark := " "
		if done {
			mark = "✓"
		}

		date := s.UpdatedAt.Format("01/02 15:04")
		dir := shortenDir(s.Directory)

		line := fmt.Sprintf("%s %-6s %s  %-52s  %s", mark, s.Source, date, title, dir)
		if note != "" {
			line += fmt.Sprintf("  [%s]", note)
		}
		fmt.Println(line)
	}
}

func cmdSearch(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: axe search <keyword>")
		os.Exit(1)
	}
	keyword := strings.Join(args, " ")
	ctx := loadContext()

	results := tui.DeepSearch(ctx.providers, keyword)
	if len(results) == 0 {
		fmt.Println("No matches found.")
		return
	}

	// Enrich with session metadata
	sessionMap := make(map[string]*provider.Session)
	for _, s := range allSessions(ctx) {
		s := s
		sessionMap[s.Source+":"+s.ID] = &s
	}

	for _, r := range results {
		key := r.Source + ":" + r.SessionID
		s := sessionMap[key]
		header := key[:14]
		if s != nil {
			title := s.Title
			if len([]rune(title)) > 40 {
				title = string([]rune(title)[:40]) + ".."
			}
			header = fmt.Sprintf("%s  %s  %s", r.Source, s.UpdatedAt.Format("01/02"), title)
		}
		fmt.Printf("── %s ──\n", header)
		shown := 0
		for _, m := range r.Matches {
			if shown >= 3 {
				fmt.Printf("  ... and %d more matches\n", len(r.Matches)-3)
				break
			}
			line := strings.TrimSpace(m.Line)
			line = strings.ReplaceAll(line, "\n", " ")
			line = strings.Join(strings.Fields(line), " ")
			if len([]rune(line)) > 100 {
				line = string([]rune(line)[:100]) + ".."
			}
			fmt.Printf("  %s\n", line)
			shown++
		}
	}
}

func cmdDone(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: axe done <session-id-prefix>")
		os.Exit(1)
	}
	ctx := loadContext()
	s, err := findSession(ctx, args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	key := s.Source + ":" + s.ID
	ctx.st.MarkDone(key)
	if err := ctx.st.Save(); err != nil {
		fmt.Fprintln(os.Stderr, "save error:", err)
		os.Exit(1)
	}
	title := s.Title
	if len([]rune(title)) > 50 {
		title = string([]rune(title)[:50]) + ".."
	}
	fmt.Printf("Marked done: %s:%s  %s\n", s.Source, s.ID[:8], title)
}

func cmdUndone(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: axe undone <session-id-prefix>")
		os.Exit(1)
	}
	ctx := loadContext()
	s, err := findSession(ctx, args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	key := s.Source + ":" + s.ID
	ctx.st.UndoDone(key)
	if err := ctx.st.Save(); err != nil {
		fmt.Fprintln(os.Stderr, "save error:", err)
		os.Exit(1)
	}
	title := s.Title
	if len([]rune(title)) > 50 {
		title = string([]rune(title)[:50]) + ".."
	}
	fmt.Printf("Unmarked done: %s:%s  %s\n", s.Source, s.ID[:8], title)
}

func cmdNote(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: axe note <session-id-prefix> <text>")
		os.Exit(1)
	}
	ctx := loadContext()
	s, err := findSession(ctx, args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	note := strings.Join(args[1:], " ")
	key := s.Source + ":" + s.ID
	ctx.st.SetNote(key, note)
	if err := ctx.st.Save(); err != nil {
		fmt.Fprintln(os.Stderr, "save error:", err)
		os.Exit(1)
	}
	fmt.Printf("Note set on %s:%s → %s\n", s.Source, s.ID[:8], note)
}

func cmdConfig() {
	ctx := loadContext()
	data, _ := json.MarshalIndent(ctx.cfg, "", "  ")
	fmt.Println(string(data))
}

// --- TUI ---

func cmdTUI() {
	ctx := loadContext()

	if len(ctx.providers) == 0 {
		fmt.Fprintln(os.Stderr, "No Claude or Codex sessions found")
		os.Exit(1)
	}

	running := process.DetectRunning()
	model := tui.New(ctx.providers, ctx.cfg, ctx.st, running)

	p := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if m, ok := finalModel.(tui.Model); ok {
		if args := m.ExecArgs(); len(args) > 0 {
			if err := tui.ExecReplace(m.ExecDir(), args); err != nil {
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
