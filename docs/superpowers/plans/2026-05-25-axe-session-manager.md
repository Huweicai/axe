# axe Session Manager Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Go TUI that unifies Claude Code + Codex session management — search, resume, create new sessions from one entry point.

**Architecture:** Provider pattern isolates data source parsing (Claude/Codex) from the TUI layer. State is persisted separately in `~/.axe/`. The TUI uses Bubbletea with a split-pane layout (list + preview). `syscall.Exec` replaces the process on resume/new.

**Tech Stack:** Go 1.22+, github.com/charmbracelet/bubbletea, github.com/charmbracelet/lipgloss, github.com/charmbracelet/bubbles, github.com/sahilm/fuzzy

---

## File Structure

```
axe/
├── main.go                      ← CLI entry, flag parsing, app bootstrap
├── go.mod / go.sum
├── provider/
│   ├── provider.go              ← Session/Match types + Provider interface
│   ├── claude.go                ← Claude Code: scan ~/.claude/projects/
│   ├── claude_test.go           ← Tests with fixture data
│   ├── codex.go                 ← Codex: parse session_index.jsonl + session files
│   ├── codex_test.go            ← Tests with fixture data
│   └── testdata/
│       ├── claude/
│       │   └── projects/
│       │       └── -Users-test-myproject/
│       │           └── abc123.jsonl
│       └── codex/
│           ├── session_index.jsonl
│           └── sessions/2025/11/03/
│               └── test-session.jsonl
├── state/
│   ├── state.go                 ← Config + State load/save for ~/.axe/
│   └── state_test.go
├── process/
│   ├── detect.go                ← Running process detection (pgrep)
│   └── detect_test.go
├── tui/
│   ├── app.go                   ← Root bubbletea model, Init/Update/View
│   ├── app_test.go              ← Model update logic tests
│   ├── item.go                  ← ListItem type (workspace or session)
│   ├── filter.go                ← Fuzzy filter + path shortening helpers
│   ├── preview.go               ← Right-pane preview rendering
│   ├── search.go                ← Deep search (concurrent JSONL grep)
│   ├── search_test.go
│   ├── keys.go                  ← Keybinding definitions
│   └── styles.go                ← Lipgloss style constants
└── docs/
    └── superpowers/...
```

---

### Task 1: Go Module + Dependencies

**Files:**
- Create: `go.mod`
- Create: `go.sum` (generated)

- [ ] **Step 1: Initialize Go module**

```bash
cd /Users/huweicai/Projects/GoProjects/github.com/Huweicai/axe
go mod init github.com/Huweicai/axe
```

- [ ] **Step 2: Add dependencies**

```bash
cd /Users/huweicai/Projects/GoProjects/github.com/Huweicai/axe
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/lipgloss@latest
go get github.com/charmbracelet/bubbles@latest
go get github.com/sahilm/fuzzy@latest
```

- [ ] **Step 3: Create minimal main.go to verify build**

```go
// main.go
package main

import "fmt"

func main() {
	fmt.Println("axe — AI session manager")
}
```

- [ ] **Step 4: Verify build**

Run: `cd /Users/huweicai/Projects/GoProjects/github.com/Huweicai/axe && go build -o axe . && ./axe`
Expected: prints "axe — AI session manager"

- [ ] **Step 5: Commit**

```bash
cd /Users/huweicai/Projects/GoProjects/github.com/Huweicai/axe
git add go.mod go.sum main.go
git commit -m "feat: initialize Go module with bubbletea dependencies"
```

---

### Task 2: Provider Interface + Types

**Files:**
- Create: `provider/provider.go`

- [ ] **Step 1: Write the provider package with types and interface**

```go
// provider/provider.go
package provider

import "time"

type Session struct {
	ID        string
	Source    string // "claude" | "codex"
	Title     string
	Directory string
	CreatedAt time.Time
	UpdatedAt time.Time
	FilePath  string
	FileSize  int64
}

type Match struct {
	SessionID string
	Line      string
	LineNum   int
}

type Provider interface {
	Name() string
	ListSessions() ([]Session, error)
	SearchContent(sessionID string, keyword string) ([]Match, error)
	ResumeCommand(sessionID string) []string
	NewCommand(cwd string) []string
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd /Users/huweicai/Projects/GoProjects/github.com/Huweicai/axe && go build ./provider/`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
cd /Users/huweicai/Projects/GoProjects/github.com/Huweicai/axe
git add provider/provider.go
git commit -m "feat: add provider interface and session types"
```

---

### Task 3: Claude Provider

**Files:**
- Create: `provider/claude.go`
- Create: `provider/claude_test.go`
- Create: `provider/testdata/claude/projects/-Users-test-myproject/abc123.jsonl`

- [ ] **Step 1: Create test fixture**

The fixture mimics the real Claude session format. Line 0 has `sessionId` and `cwd`. User messages have `type: "user"` with `message.role: "user"` and `message.content` as a list of text blocks.

```bash
mkdir -p /Users/huweicai/Projects/GoProjects/github.com/Huweicai/axe/provider/testdata/claude/projects/-Users-test-myproject
```

Write file `provider/testdata/claude/projects/-Users-test-myproject/abc123.jsonl`:
```
{"type":"permission-mode","permissionMode":"default","sessionId":"abc123"}
{"parentUuid":null,"isSidechain":false,"attachment":{"type":"hook_success"},"type":"attachment","uuid":"u1","timestamp":1700000000000,"cwd":"/Users/test/myproject","sessionId":"abc123"}
{"parentUuid":"u1","isSidechain":false,"type":"user","message":{"role":"user","content":[{"type":"text","text":"Help me refactor the auth module"}]},"uuid":"u2","timestamp":1700000001000,"cwd":"/Users/test/myproject","sessionId":"abc123"}
{"parentUuid":"u2","isSidechain":false,"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"I'll help with that."}]},"uuid":"u3","timestamp":1700000002000,"cwd":"/Users/test/myproject","sessionId":"abc123"}
{"parentUuid":"u3","isSidechain":false,"type":"user","message":{"role":"user","content":[{"type":"text","text":"Now add tests for it"}]},"uuid":"u4","timestamp":1700000003000,"cwd":"/Users/test/myproject","sessionId":"abc123"}
```

- [ ] **Step 2: Write the failing test**

```go
// provider/claude_test.go
package provider

import (
	"path/filepath"
	"runtime"
	"testing"
)

func testdataDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "testdata")
}

func TestClaudeProvider_ListSessions(t *testing.T) {
	baseDir := filepath.Join(testdataDir(), "claude")
	p := NewClaudeProvider(baseDir)

	sessions, err := p.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions error: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}

	s := sessions[0]
	if s.ID != "abc123" {
		t.Errorf("ID = %q, want %q", s.ID, "abc123")
	}
	if s.Source != "claude" {
		t.Errorf("Source = %q, want %q", s.Source, "claude")
	}
	if s.Directory != "/Users/test/myproject" {
		t.Errorf("Directory = %q, want %q", s.Directory, "/Users/test/myproject")
	}
	if s.Title != "Help me refactor the auth module" {
		t.Errorf("Title = %q, want %q", s.Title, "Help me refactor the auth module")
	}
}

func TestClaudeProvider_SearchContent(t *testing.T) {
	baseDir := filepath.Join(testdataDir(), "claude")
	p := NewClaudeProvider(baseDir)

	matches, err := p.SearchContent("abc123", "refactor")
	if err != nil {
		t.Fatalf("SearchContent error: %v", err)
	}
	if len(matches) == 0 {
		t.Fatal("expected at least 1 match")
	}
	if matches[0].SessionID != "abc123" {
		t.Errorf("match SessionID = %q, want %q", matches[0].SessionID, "abc123")
	}
}

func TestClaudeProvider_ResumeCommand(t *testing.T) {
	p := NewClaudeProvider("")
	cmd := p.ResumeCommand("abc123")
	expected := []string{"claude", "-r", "abc123"}
	if len(cmd) != len(expected) {
		t.Fatalf("ResumeCommand = %v, want %v", cmd, expected)
	}
	for i := range cmd {
		if cmd[i] != expected[i] {
			t.Errorf("cmd[%d] = %q, want %q", i, cmd[i], expected[i])
		}
	}
}

func TestClaudeProvider_NewCommand(t *testing.T) {
	p := NewClaudeProvider("")
	cmd := p.NewCommand("/some/path")
	expected := []string{"claude", "--cwd", "/some/path"}
	if len(cmd) != len(expected) {
		t.Fatalf("NewCommand = %v, want %v", cmd, expected)
	}
	for i := range cmd {
		if cmd[i] != expected[i] {
			t.Errorf("cmd[%d] = %q, want %q", i, cmd[i], expected[i])
		}
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `cd /Users/huweicai/Projects/GoProjects/github.com/Huweicai/axe && go test ./provider/ -run TestClaude -v`
Expected: FAIL — `NewClaudeProvider` undefined

- [ ] **Step 4: Implement Claude provider**

```go
// provider/claude.go
package provider

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ClaudeProvider struct {
	baseDir string // e.g. ~/.claude
}

func NewClaudeProvider(baseDir string) *ClaudeProvider {
	return &ClaudeProvider{baseDir: baseDir}
}

func (p *ClaudeProvider) Name() string { return "claude" }

func (p *ClaudeProvider) ListSessions() ([]Session, error) {
	projectsDir := filepath.Join(p.baseDir, "projects")
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil, err
	}

	var sessions []Session
	for _, projEntry := range entries {
		if !projEntry.IsDir() {
			continue
		}
		projPath := filepath.Join(projectsDir, projEntry.Name())
		files, err := os.ReadDir(projPath)
		if err != nil {
			continue
		}
		for _, f := range files {
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".jsonl") {
				continue
			}
			sessionID := strings.TrimSuffix(f.Name(), ".jsonl")
			filePath := filepath.Join(projPath, f.Name())

			info, _ := f.Info()
			var fileSize int64
			var modTime time.Time
			if info != nil {
				fileSize = info.Size()
				modTime = info.ModTime()
			}

			cwd, title := p.parseSessionFile(filePath)
			sessions = append(sessions, Session{
				ID:        sessionID,
				Source:    "claude",
				Title:     title,
				Directory: cwd,
				UpdatedAt: modTime,
				FilePath:  filePath,
				FileSize:  fileSize,
			})
		}
	}
	return sessions, nil
}

func (p *ClaudeProvider) parseSessionFile(path string) (cwd, title string) {
	f, err := os.Open(path)
	if err != nil {
		return "", ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(scanner.Bytes(), &raw); err != nil {
			continue
		}

		// Extract cwd from any line that has it
		if cwd == "" {
			if cwdRaw, ok := raw["cwd"]; ok {
				json.Unmarshal(cwdRaw, &cwd)
			}
		}

		// Extract title from first non-meta user message
		if title == "" {
			var typ string
			json.Unmarshal(raw["type"], &typ)
			if typ != "user" {
				continue
			}

			// Skip meta messages
			var isMeta bool
			if metaRaw, ok := raw["isMeta"]; ok {
				json.Unmarshal(metaRaw, &isMeta)
			}
			if isMeta {
				continue
			}

			// Parse message content
			title = p.extractUserText(raw["message"])
		}

		if cwd != "" && title != "" {
			break
		}
	}
	return cwd, title
}

func (p *ClaudeProvider) extractUserText(msgRaw json.RawMessage) string {
	var msg struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
	}
	if err := json.Unmarshal(msgRaw, &msg); err != nil || msg.Role != "user" {
		return ""
	}

	// content can be a list of blocks or a string
	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(msg.Content, &blocks); err == nil {
		for _, b := range blocks {
			if b.Type == "text" && b.Text != "" {
				text := b.Text
				// Skip hook/skill injected content
				if strings.HasPrefix(text, "<command-") || strings.HasPrefix(text, "Base directory") {
					continue
				}
				// Truncate for title
				if len(text) > 200 {
					text = text[:200]
				}
				return strings.TrimSpace(text)
			}
		}
	}

	var plainStr string
	if err := json.Unmarshal(msg.Content, &plainStr); err == nil {
		if len(plainStr) > 200 {
			plainStr = plainStr[:200]
		}
		return strings.TrimSpace(plainStr)
	}

	return ""
}

func (p *ClaudeProvider) SearchContent(sessionID string, keyword string) ([]Match, error) {
	filePath := p.findSessionFile(sessionID)
	if filePath == "" {
		return nil, nil
	}

	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	keyword = strings.ToLower(keyword)
	var matches []Match
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if strings.Contains(strings.ToLower(line), keyword) {
			// Extract readable text from the line
			text := p.extractTextFromLine(line)
			if text != "" && strings.Contains(strings.ToLower(text), keyword) {
				matches = append(matches, Match{
					SessionID: sessionID,
					Line:      text,
					LineNum:   lineNum,
				})
			}
		}
	}
	return matches, nil
}

func (p *ClaudeProvider) extractTextFromLine(line string) string {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return ""
	}
	if msgRaw, ok := raw["message"]; ok {
		return p.extractUserText(msgRaw)
	}
	return ""
}

func (p *ClaudeProvider) findSessionFile(sessionID string) string {
	projectsDir := filepath.Join(p.baseDir, "projects")
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		path := filepath.Join(projectsDir, e.Name(), sessionID+".jsonl")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

func (p *ClaudeProvider) ResumeCommand(sessionID string) []string {
	return []string{"claude", "-r", sessionID}
}

func (p *ClaudeProvider) NewCommand(cwd string) []string {
	return []string{"claude", "--cwd", cwd}
}
```

- [ ] **Step 5: Run tests**

Run: `cd /Users/huweicai/Projects/GoProjects/github.com/Huweicai/axe && go test ./provider/ -run TestClaude -v`
Expected: all PASS

- [ ] **Step 6: Commit**

```bash
cd /Users/huweicai/Projects/GoProjects/github.com/Huweicai/axe
git add provider/claude.go provider/claude_test.go provider/testdata/
git commit -m "feat: implement Claude Code session provider"
```

---

### Task 4: Codex Provider

**Files:**
- Create: `provider/codex.go`
- Create: `provider/codex_test.go`
- Create: `provider/testdata/codex/session_index.jsonl`
- Create: `provider/testdata/codex/sessions/2025/11/03/test-session-id.jsonl`

- [ ] **Step 1: Create test fixtures**

`provider/testdata/codex/session_index.jsonl`:
```
{"id":"test-session-id","thread_name":"优化下单链路 cache","updated_at":"2026-03-06T05:46:31.888912Z"}
{"id":"another-session","thread_name":"Add dashboard aggregation","updated_at":"2026-03-05T07:44:55.966656Z"}
```

```bash
mkdir -p /Users/huweicai/Projects/GoProjects/github.com/Huweicai/axe/provider/testdata/codex/sessions/2025/11/03
```

`provider/testdata/codex/sessions/2025/11/03/test-session-id.jsonl`:
```
{"timestamp":"2025-11-03T11:47:06.400Z","type":"session_meta","payload":{"id":"test-session-id","timestamp":"2025-11-03T11:47:06.359Z","cwd":"/Users/test/spectra","originator":"codex_cli","cli_version":"1.0.0"}}
{"timestamp":"2025-11-03T11:47:10.000Z","type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":"帮我优化下单链路的 cache miss"}]}}
{"timestamp":"2025-11-03T11:47:15.000Z","type":"response_item","payload":{"type":"message","role":"assistant","content":[{"type":"output_text","text":"I'll analyze the cache usage."}]}}
```

- [ ] **Step 2: Write the failing test**

```go
// provider/codex_test.go
package provider

import (
	"path/filepath"
	"testing"
)

func TestCodexProvider_ListSessions(t *testing.T) {
	baseDir := filepath.Join(testdataDir(), "codex")
	p := NewCodexProvider(baseDir)

	sessions, err := p.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions error: %v", err)
	}
	if len(sessions) < 1 {
		t.Fatalf("expected at least 1 session, got %d", len(sessions))
	}

	// Find our test session
	var found *Session
	for i := range sessions {
		if sessions[i].ID == "test-session-id" {
			found = &sessions[i]
			break
		}
	}
	if found == nil {
		t.Fatal("test-session-id not found in sessions")
	}

	if found.Source != "codex" {
		t.Errorf("Source = %q, want %q", found.Source, "codex")
	}
	if found.Title != "优化下单链路 cache" {
		t.Errorf("Title = %q, want %q", found.Title, "优化下单链路 cache")
	}
	if found.Directory != "/Users/test/spectra" {
		t.Errorf("Directory = %q, want %q", found.Directory, "/Users/test/spectra")
	}
}

func TestCodexProvider_SearchContent(t *testing.T) {
	baseDir := filepath.Join(testdataDir(), "codex")
	p := NewCodexProvider(baseDir)

	matches, err := p.SearchContent("test-session-id", "cache")
	if err != nil {
		t.Fatalf("SearchContent error: %v", err)
	}
	if len(matches) == 0 {
		t.Fatal("expected at least 1 match")
	}
}

func TestCodexProvider_ResumeCommand(t *testing.T) {
	p := NewCodexProvider("")
	cmd := p.ResumeCommand("test-session-id")
	expected := []string{"codex", "resume", "test-session-id"}
	if len(cmd) != len(expected) {
		t.Fatalf("ResumeCommand = %v, want %v", cmd, expected)
	}
	for i := range cmd {
		if cmd[i] != expected[i] {
			t.Errorf("cmd[%d] = %q, want %q", i, cmd[i], expected[i])
		}
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `cd /Users/huweicai/Projects/GoProjects/github.com/Huweicai/axe && go test ./provider/ -run TestCodex -v`
Expected: FAIL — `NewCodexProvider` undefined

- [ ] **Step 4: Implement Codex provider**

```go
// provider/codex.go
package provider

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type CodexProvider struct {
	baseDir string // e.g. ~/.codex
}

func NewCodexProvider(baseDir string) *CodexProvider {
	return &CodexProvider{baseDir: baseDir}
}

func (p *CodexProvider) Name() string { return "codex" }

func (p *CodexProvider) ListSessions() ([]Session, error) {
	indexPath := filepath.Join(p.baseDir, "session_index.jsonl")
	f, err := os.Open(indexPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	type indexEntry struct {
		ID         string `json:"id"`
		ThreadName string `json:"thread_name"`
		UpdatedAt  string `json:"updated_at"`
	}

	var sessions []Session
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var entry indexEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}

		updatedAt, _ := time.Parse(time.RFC3339Nano, entry.UpdatedAt)

		// Find session file to get cwd
		filePath := p.findSessionFile(entry.ID)
		var cwd string
		var fileSize int64
		if filePath != "" {
			cwd = p.extractCwd(filePath)
			if info, err := os.Stat(filePath); err == nil {
				fileSize = info.Size()
			}
		}

		sessions = append(sessions, Session{
			ID:        entry.ID,
			Source:    "codex",
			Title:     entry.ThreadName,
			Directory: cwd,
			UpdatedAt: updatedAt,
			FilePath:  filePath,
			FileSize:  fileSize,
		})
	}
	return sessions, nil
}

func (p *CodexProvider) findSessionFile(id string) string {
	sessionsDir := filepath.Join(p.baseDir, "sessions")
	var found string
	filepath.Walk(sessionsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.Contains(filepath.Base(path), id) && strings.HasSuffix(path, ".jsonl") {
			found = path
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

func (p *CodexProvider) extractCwd(filePath string) string {
	f, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	if scanner.Scan() {
		var meta struct {
			Type    string `json:"type"`
			Payload struct {
				Cwd string `json:"cwd"`
			} `json:"payload"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &meta); err == nil {
			return meta.Payload.Cwd
		}
	}
	return ""
}

func (p *CodexProvider) SearchContent(sessionID string, keyword string) ([]Match, error) {
	filePath := p.findSessionFile(sessionID)
	if filePath == "" {
		return nil, nil
	}

	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	keyword = strings.ToLower(keyword)
	var matches []Match
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		lower := strings.ToLower(line)
		if strings.Contains(lower, keyword) {
			text := p.extractTextFromLine(line)
			if text != "" && strings.Contains(strings.ToLower(text), keyword) {
				matches = append(matches, Match{
					SessionID: sessionID,
					Line:      text,
					LineNum:   lineNum,
				})
			}
		}
	}
	return matches, nil
}

func (p *CodexProvider) extractTextFromLine(line string) string {
	var raw struct {
		Payload struct {
			Type    string `json:"type"`
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"payload"`
	}
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return ""
	}
	for _, c := range raw.Payload.Content {
		if (c.Type == "input_text" || c.Type == "output_text") && c.Text != "" {
			text := c.Text
			if len(text) > 200 {
				text = text[:200]
			}
			return text
		}
	}
	return ""
}

func (p *CodexProvider) ResumeCommand(sessionID string) []string {
	return []string{"codex", "resume", sessionID}
}

func (p *CodexProvider) NewCommand(cwd string) []string {
	return []string{"codex", "--cwd", cwd}
}
```

- [ ] **Step 5: Run tests**

Run: `cd /Users/huweicai/Projects/GoProjects/github.com/Huweicai/axe && go test ./provider/ -run TestCodex -v`
Expected: all PASS

- [ ] **Step 6: Commit**

```bash
cd /Users/huweicai/Projects/GoProjects/github.com/Huweicai/axe
git add provider/codex.go provider/codex_test.go provider/testdata/codex/
git commit -m "feat: implement Codex session provider"
```

---

### Task 5: State Management

**Files:**
- Create: `state/state.go`
- Create: `state/state_test.go`

- [ ] **Step 1: Write the failing test**

```go
// state/state_test.go
package state

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_Default(t *testing.T) {
	dir := t.TempDir()
	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("LoadConfig error: %v", err)
	}
	if cfg.DefaultTool != "claude" {
		t.Errorf("DefaultTool = %q, want %q", cfg.DefaultTool, "claude")
	}
}

func TestLoadConfig_FromFile(t *testing.T) {
	dir := t.TempDir()
	content := `{"default_tool":"codex","pinned_workspaces":[{"alias":"test","path":"~/test"}]}`
	os.WriteFile(filepath.Join(dir, "config.json"), []byte(content), 0644)

	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("LoadConfig error: %v", err)
	}
	if cfg.DefaultTool != "codex" {
		t.Errorf("DefaultTool = %q, want %q", cfg.DefaultTool, "codex")
	}
	if len(cfg.PinnedWorkspaces) != 1 {
		t.Fatalf("PinnedWorkspaces len = %d, want 1", len(cfg.PinnedWorkspaces))
	}
	if cfg.PinnedWorkspaces[0].Alias != "test" {
		t.Errorf("Alias = %q, want %q", cfg.PinnedWorkspaces[0].Alias, "test")
	}
}

func TestState_MarkDone(t *testing.T) {
	dir := t.TempDir()
	s, _ := LoadState(dir)

	s.MarkDone("claude:abc123")
	if !s.IsDone("claude:abc123") {
		t.Error("session should be marked done")
	}

	s.UndoDone("claude:abc123")
	if s.IsDone("claude:abc123") {
		t.Error("session should not be done after undo")
	}
}

func TestState_Notes(t *testing.T) {
	dir := t.TempDir()
	s, _ := LoadState(dir)

	s.SetNote("codex:xyz", "still iterating")
	if got := s.GetNote("codex:xyz"); got != "still iterating" {
		t.Errorf("note = %q, want %q", got, "still iterating")
	}
}

func TestState_Persistence(t *testing.T) {
	dir := t.TempDir()
	s, _ := LoadState(dir)
	s.MarkDone("claude:abc")
	s.SetNote("claude:abc", "done with this")

	if err := s.Save(); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	s2, err := LoadState(dir)
	if err != nil {
		t.Fatalf("LoadState error: %v", err)
	}
	if !s2.IsDone("claude:abc") {
		t.Error("done state not persisted")
	}
	if s2.GetNote("claude:abc") != "done with this" {
		t.Error("note not persisted")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/huweicai/Projects/GoProjects/github.com/Huweicai/axe && go test ./state/ -v`
Expected: FAIL — package not found

- [ ] **Step 3: Implement state package**

```go
// state/state.go
package state

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Workspace struct {
	Alias string `json:"alias"`
	Path  string `json:"path"`
}

type Config struct {
	DefaultTool      string      `json:"default_tool"`
	PinnedWorkspaces []Workspace `json:"pinned_workspaces"`
}

type SessionState struct {
	Done bool   `json:"done,omitempty"`
	Note string `json:"note,omitempty"`
}

type State struct {
	Sessions map[string]*SessionState `json:"sessions"`
	dir      string
}

func LoadConfig(dir string) (*Config, error) {
	cfg := &Config{DefaultTool: "claude"}
	path := filepath.Join(dir, "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func LoadState(dir string) (*State, error) {
	s := &State{
		Sessions: make(map[string]*SessionState),
		dir:      dir,
	}
	path := filepath.Join(dir, "state.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, s); err != nil {
		return nil, err
	}
	s.dir = dir
	return s, nil
}

func (s *State) Save() error {
	if err := os.MkdirAll(s.dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.dir, "state.json"), data, 0644)
}

func (s *State) getOrCreate(key string) *SessionState {
	if ss, ok := s.Sessions[key]; ok {
		return ss
	}
	ss := &SessionState{}
	s.Sessions[key] = ss
	return ss
}

func (s *State) MarkDone(key string) {
	s.getOrCreate(key).Done = true
}

func (s *State) UndoDone(key string) {
	if ss, ok := s.Sessions[key]; ok {
		ss.Done = false
	}
}

func (s *State) IsDone(key string) bool {
	if ss, ok := s.Sessions[key]; ok {
		return ss.Done
	}
	return false
}

func (s *State) SetNote(key, note string) {
	s.getOrCreate(key).Note = note
}

func (s *State) GetNote(key string) string {
	if ss, ok := s.Sessions[key]; ok {
		return ss.Note
	}
	return ""
}
```

- [ ] **Step 4: Run tests**

Run: `cd /Users/huweicai/Projects/GoProjects/github.com/Huweicai/axe && go test ./state/ -v`
Expected: all PASS

- [ ] **Step 5: Commit**

```bash
cd /Users/huweicai/Projects/GoProjects/github.com/Huweicai/axe
git add state/
git commit -m "feat: implement state persistence (config, done, notes)"
```

---

### Task 6: Running Process Detection

**Files:**
- Create: `process/detect.go`
- Create: `process/detect_test.go`

- [ ] **Step 1: Write the failing test**

```go
// process/detect_test.go
package process

import (
	"testing"
)

func TestParseProcessList(t *testing.T) {
	// Simulated output from pgrep
	output := `12345 claude -r abc123 --cwd /Users/test/project1
67890 codex resume xyz --cwd /Users/test/project2
11111 codex --cwd /Users/test/project3`

	running := ParseProcessList(output)

	if !running["/Users/test/project1"] {
		t.Error("expected /Users/test/project1 to be running")
	}
	if !running["/Users/test/project2"] {
		t.Error("expected /Users/test/project2 to be running")
	}
	if !running["/Users/test/project3"] {
		t.Error("expected /Users/test/project3 to be running")
	}
	if running["/nonexistent"] {
		t.Error("expected /nonexistent to not be running")
	}
}

func TestParseProcessList_Empty(t *testing.T) {
	running := ParseProcessList("")
	if len(running) != 0 {
		t.Errorf("expected empty map, got %d entries", len(running))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/huweicai/Projects/GoProjects/github.com/Huweicai/axe && go test ./process/ -v`
Expected: FAIL — package not found

- [ ] **Step 3: Implement process detection**

```go
// process/detect.go
package process

import (
	"os/exec"
	"strings"
)

// DetectRunning runs pgrep and returns a set of cwds with active claude/codex processes.
func DetectRunning() map[string]bool {
	out, err := exec.Command("pgrep", "-af", "claude|codex").Output()
	if err != nil {
		return make(map[string]bool)
	}
	return ParseProcessList(string(out))
}

// ParseProcessList parses pgrep output and extracts --cwd values.
func ParseProcessList(output string) map[string]bool {
	running := make(map[string]bool)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		for i, p := range parts {
			if p == "--cwd" && i+1 < len(parts) {
				running[parts[i+1]] = true
				break
			}
		}
	}
	return running
}
```

- [ ] **Step 4: Run tests**

Run: `cd /Users/huweicai/Projects/GoProjects/github.com/Huweicai/axe && go test ./process/ -v`
Expected: all PASS

- [ ] **Step 5: Commit**

```bash
cd /Users/huweicai/Projects/GoProjects/github.com/Huweicai/axe
git add process/
git commit -m "feat: implement running process detection via pgrep"
```

---

### Task 7: TUI — Core Model + Styles + Keybindings

**Files:**
- Create: `tui/styles.go`
- Create: `tui/keys.go`
- Create: `tui/item.go`
- Create: `tui/app.go`
- Create: `tui/app_test.go`

- [ ] **Step 1: Create styles**

```go
// tui/styles.go
package tui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")).
			Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	workspaceStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("220"))

	runningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82"))

	sourceClaudeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("117"))

	sourceCodexStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("215"))

	previewBorder = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(lipgloss.Color("237")).
			PaddingLeft(1)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Background(lipgloss.Color("236")).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
)
```

- [ ] **Step 2: Create keybinding definitions**

```go
// tui/keys.go
package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Enter     key.Binding
	AltEnter  key.Binding
	Done      key.Binding
	Undo      key.Binding
	Note      key.Binding
	Search    key.Binding
	Tab       key.Binding
	Group     key.Binding
	ShowDone  key.Binding
	Quit      key.Binding
	Esc       key.Binding
}

var keys = keyMap{
	Enter:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open")),
	AltEnter: key.NewBinding(key.WithKeys("x"), key.WithHelp("x+enter", "alt tool")),
	Done:     key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "done")),
	Undo:     key.NewBinding(key.WithKeys("u"), key.WithHelp("u", "undo")),
	Note:     key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "note")),
	Search:   key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "deep-search")),
	Tab:      key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "filter")),
	Group:    key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "group")),
	ShowDone: key.NewBinding(key.WithKeys("ctrl+a"), key.WithHelp("ctrl+a", "show done")),
	Quit:     key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
	Esc:      key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back/quit")),
}
```

- [ ] **Step 3: Create list item type**

```go
// tui/item.go
package tui

import (
	"time"

	"github.com/Huweicai/axe/provider"
)

type ItemKind int

const (
	KindWorkspace ItemKind = iota
	KindSession
)

type ListItem struct {
	Kind      ItemKind
	Alias     string // workspace alias
	Path      string // workspace path or session directory
	Session   *provider.Session
	Done      bool
	Note      string
	Running   bool
	MatchText string // for fuzzy match: title + dir + note combined
}

func (i ListItem) Title() string {
	if i.Kind == KindWorkspace {
		return i.Alias
	}
	if i.Session != nil {
		return i.Session.Title
	}
	return ""
}

func (i ListItem) UpdatedAt() time.Time {
	if i.Session != nil {
		return i.Session.UpdatedAt
	}
	return time.Time{}
}

func (i ListItem) Source() string {
	if i.Session != nil {
		return i.Session.Source
	}
	return ""
}

func (i ListItem) StateKey() string {
	if i.Session != nil {
		return i.Session.Source + ":" + i.Session.ID
	}
	return ""
}
```

- [ ] **Step 4: Create core app model**

```go
// tui/app.go
package tui

import (
	"os"
	"os/exec"
	"strings"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Huweicai/axe/provider"
	"github.com/Huweicai/axe/state"
)

type sourceFilter int

const (
	filterAll sourceFilter = iota
	filterClaude
	filterCodex
)

type Model struct {
	items     []ListItem
	filtered  []int
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

	execArgs []string // set before quitting to exec after tea exits
}

func New(providers []provider.Provider, cfg *state.Config, st *state.State, running map[string]bool) Model {
	m := Model{
		providers: providers,
		config:    cfg,
		state:     st,
		running:   running,
		grouped:   true,
		filter:    filterAll,
	}
	m.loadItems()
	m.applyFilter()
	return m
}

func (m *Model) loadItems() {
	m.items = nil

	for _, ws := range m.config.PinnedWorkspaces {
		path := expandHome(ws.Path)
		m.items = append(m.items, ListItem{
			Kind:      KindWorkspace,
			Alias:     ws.Alias,
			Path:      path,
			MatchText: strings.ToLower(ws.Alias + " " + path),
		})
	}

	for _, p := range m.providers {
		sessions, err := p.ListSessions()
		if err != nil {
			continue
		}
		for i := range sessions {
			s := &sessions[i]
			key := s.Source + ":" + s.ID
			item := ListItem{
				Kind:      KindSession,
				Path:      s.Directory,
				Session:   s,
				Done:      m.state.IsDone(key),
				Note:      m.state.GetNote(key),
				Running:   m.running[s.Directory],
				MatchText: strings.ToLower(s.Title + " " + s.Directory + " " + m.state.GetNote(key)),
			}
			m.items = append(m.items, item)
		}
	}
}

func (m *Model) applyFilter() {
	m.filtered = nil
	query := strings.ToLower(m.query)

	for i, item := range m.items {
		if item.Done && !m.showDone {
			continue
		}
		if m.filter == filterClaude && item.Source() == "codex" {
			continue
		}
		if m.filter == filterCodex && item.Source() == "claude" {
			continue
		}
		if query != "" && !strings.Contains(item.MatchText, query) {
			continue
		}
		m.filtered = append(m.filtered, i)
	}

	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		if m.noteInput {
			return m.handleNoteInput(msg)
		}
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		if m.query == "" {
			return m, tea.Quit
		}
		m.query += "q"
		m.applyFilter()
	case "esc":
		if m.query != "" {
			m.query = ""
			m.applyFilter()
		} else {
			return m, tea.Quit
		}
	case "enter":
		return m.doExec(false)
	case "x":
		if len(m.filtered) > 0 && m.items[m.filtered[m.cursor]].Kind == KindWorkspace {
			return m.doExec(true)
		}
	case "d":
		m.toggleDone()
	case "u":
		m.undoDone()
	case "n":
		if len(m.filtered) > 0 && m.items[m.filtered[m.cursor]].Kind == KindSession {
			m.noteInput = true
			m.noteText = m.items[m.filtered[m.cursor]].Note
		}
	case "/":
		m.deepMode = !m.deepMode
	case "tab":
		m.filter = (m.filter + 1) % 3
		m.applyFilter()
	case "g":
		if m.query == "" {
			m.grouped = !m.grouped
		} else {
			m.query += "g"
			m.applyFilter()
		}
	case "ctrl+a":
		m.showDone = !m.showDone
		m.applyFilter()
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.filtered)-1 {
			m.cursor++
		}
	case "backspace":
		if len(m.query) > 0 {
			m.query = m.query[:len(m.query)-1]
			m.applyFilter()
		}
	default:
		r := msg.String()
		if len(r) == 1 && r[0] >= 32 {
			m.query += r
			m.applyFilter()
		}
	}
	return m, nil
}

func (m Model) handleNoteInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if len(m.filtered) > 0 {
			idx := m.filtered[m.cursor]
			item := &m.items[idx]
			key := item.StateKey()
			if key != "" {
				m.state.SetNote(key, m.noteText)
				m.state.Save()
				item.Note = m.noteText
				item.MatchText = strings.ToLower(item.Title() + " " + item.Path + " " + m.noteText)
			}
		}
		m.noteInput = false
		m.noteText = ""
	case "esc":
		m.noteInput = false
		m.noteText = ""
	case "backspace":
		if len(m.noteText) > 0 {
			m.noteText = m.noteText[:len(m.noteText)-1]
		}
	default:
		r := msg.String()
		if len(r) == 1 && r[0] >= 32 {
			m.noteText += r
		}
	}
	return m, nil
}

func (m *Model) toggleDone() {
	if len(m.filtered) == 0 {
		return
	}
	item := &m.items[m.filtered[m.cursor]]
	if item.Kind != KindSession {
		return
	}
	m.state.MarkDone(item.StateKey())
	m.state.Save()
	item.Done = true
	m.applyFilter()
}

func (m *Model) undoDone() {
	if len(m.filtered) == 0 {
		return
	}
	item := &m.items[m.filtered[m.cursor]]
	if item.Kind != KindSession {
		return
	}
	m.state.UndoDone(item.StateKey())
	m.state.Save()
	item.Done = false
}

func (m Model) doExec(altTool bool) (tea.Model, tea.Cmd) {
	if len(m.filtered) == 0 {
		return m, nil
	}
	item := m.items[m.filtered[m.cursor]]

	var args []string
	switch item.Kind {
	case KindWorkspace:
		tool := m.config.DefaultTool
		if altTool {
			if tool == "claude" {
				tool = "codex"
			} else {
				tool = "claude"
			}
		}
		if tool == "claude" {
			args = []string{"claude", "--cwd", item.Path}
		} else {
			args = []string{"codex", "--cwd", item.Path}
		}
	case KindSession:
		for _, p := range m.providers {
			if p.Name() == item.Session.Source {
				args = p.ResumeCommand(item.Session.ID)
				break
			}
		}
	}

	m.execArgs = args
	return m, tea.Quit
}

// ExecArgs returns the command to exec after tea exits. Called from main.
func (m Model) ExecArgs() []string {
	return m.execArgs
}

// ExecReplace replaces the current process with the given command.
func ExecReplace(args []string) error {
	binary, err := exec.LookPath(args[0])
	if err != nil {
		return err
	}
	return syscall.Exec(binary, args, os.Environ())
}

func (m Model) View() string {
	// Placeholder — implemented in Task 8
	return ""
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return home + path[1:]
	}
	return path
}
```

- [ ] **Step 5: Write model test**

```go
// tui/app_test.go
package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Huweicai/axe/provider"
	"github.com/Huweicai/axe/state"
)

type mockProvider struct {
	sessions []provider.Session
}

func (m *mockProvider) Name() string                         { return "mock" }
func (m *mockProvider) ListSessions() ([]provider.Session, error) { return m.sessions, nil }
func (m *mockProvider) SearchContent(id, kw string) ([]provider.Match, error) { return nil, nil }
func (m *mockProvider) ResumeCommand(id string) []string     { return []string{"mock", "-r", id} }
func (m *mockProvider) NewCommand(cwd string) []string       { return []string{"mock", "--cwd", cwd} }

func newTestModel() Model {
	mp := &mockProvider{
		sessions: []provider.Session{
			{ID: "s1", Source: "mock", Title: "cache optimization", Directory: "/proj/a", FileSize: 1024},
			{ID: "s2", Source: "mock", Title: "auth refactor", Directory: "/proj/b", FileSize: 2048},
			{ID: "s3", Source: "mock", Title: "cache line alignment", Directory: "/proj/a", FileSize: 512},
		},
	}
	cfg := &state.Config{
		DefaultTool: "claude",
		PinnedWorkspaces: []state.Workspace{
			{Alias: "projA", Path: "/proj/a"},
		},
	}
	st := &state.State{Sessions: map[string]*state.SessionState{}}
	return New([]provider.Provider{mp}, cfg, st, map[string]bool{"/proj/a": true})
}

func TestModel_InitialState(t *testing.T) {
	m := newTestModel()
	// 1 workspace + 3 sessions = 4 items
	if len(m.items) != 4 {
		t.Errorf("items = %d, want 4", len(m.items))
	}
	// All visible (no done)
	if len(m.filtered) != 4 {
		t.Errorf("filtered = %d, want 4", len(m.filtered))
	}
}

func TestModel_QueryFilter(t *testing.T) {
	m := newTestModel()
	// Type "cache"
	for _, ch := range "cache" {
		result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		m = result.(Model)
	}
	// Should match "cache optimization" and "cache line alignment"
	if len(m.filtered) != 2 {
		t.Errorf("filtered after 'cache' = %d, want 2", len(m.filtered))
	}
}

func TestModel_ToggleDone(t *testing.T) {
	m := newTestModel()
	// Move to first session (index 1, after workspace)
	m.cursor = 1
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = result.(Model)
	// Session should be removed from filtered (done hidden by default)
	if len(m.filtered) != 3 {
		t.Errorf("filtered after done = %d, want 3", len(m.filtered))
	}
}

func TestModel_SourceFilter(t *testing.T) {
	m := newTestModel()
	// Tab cycles: all → claude → codex → all
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = result.(Model)
	if m.filter != filterClaude {
		t.Errorf("filter = %d, want filterClaude", m.filter)
	}
}

func TestModel_ExecArgs(t *testing.T) {
	m := newTestModel()
	// Select first workspace, press enter
	m.cursor = 0
	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = result.(Model)
	_ = cmd
	if len(m.execArgs) == 0 {
		t.Fatal("execArgs should be set after enter on workspace")
	}
	if m.execArgs[0] != "claude" {
		t.Errorf("execArgs[0] = %q, want 'claude'", m.execArgs[0])
	}
}
```

- [ ] **Step 6: Run tests**

Run: `cd /Users/huweicai/Projects/GoProjects/github.com/Huweicai/axe && go test ./tui/ -v`
Expected: all PASS

- [ ] **Step 7: Commit**

```bash
cd /Users/huweicai/Projects/GoProjects/github.com/Huweicai/axe
git add tui/
git commit -m "feat: implement TUI core model with keybindings and filtering"
```

---

### Task 8: TUI — View Rendering (List + Preview)

**Files:**
- Create: `tui/preview.go`
- Modify: `tui/app.go` (the `View()` method)
- Create: `tui/filter.go`

- [ ] **Step 1: Implement filter.go for path shortening and truncation helpers**

```go
// tui/filter.go
package tui

import (
	"os"
	"strings"
)

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-2]) + ".."
}

func shortenPath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	home, _ := os.UserHomeDir()
	if strings.HasPrefix(path, home) {
		path = "~" + path[len(home):]
	}
	if len(path) <= maxLen {
		return path
	}
	parts := strings.Split(path, "/")
	if len(parts) > 4 {
		path = parts[0] + "/" + parts[1] + "/.../" + parts[len(parts)-2] + "/" + parts[len(parts)-1]
	}
	if len(path) > maxLen {
		return path[:maxLen-2] + ".."
	}
	return path
}
```

- [ ] **Step 2: Implement preview.go**

```go
// tui/preview.go
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderPreview(width int) string {
	if len(m.filtered) == 0 {
		return dimStyle.Render("No sessions")
	}

	item := m.items[m.filtered[m.cursor]]
	var b strings.Builder

	switch item.Kind {
	case KindWorkspace:
		b.WriteString(titleStyle.Render("Workspace"))
		b.WriteString("\n\n")
		b.WriteString(fmt.Sprintf("  Alias: %s\n", item.Alias))
		b.WriteString(fmt.Sprintf("  Path:  %s\n", item.Path))
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("  enter = new session ("+m.config.DefaultTool+")"))
		b.WriteString("\n")
		altTool := "codex"
		if m.config.DefaultTool == "codex" {
			altTool = "claude"
		}
		b.WriteString(dimStyle.Render("  x = new session ("+altTool+")"))

	case KindSession:
		s := item.Session
		sourceLabel := sourceClaudeStyle.Render("Claude Code")
		if s.Source == "codex" {
			sourceLabel = sourceCodexStyle.Render("Codex")
		}
		b.WriteString(fmt.Sprintf("  Source:  %s\n", sourceLabel))
		b.WriteString(fmt.Sprintf("  Dir:     %s\n", shortenPath(s.Directory, width-10)))
		b.WriteString(fmt.Sprintf("  Updated: %s\n", s.UpdatedAt.Format("2006-01-02 15:04")))
		b.WriteString(fmt.Sprintf("  Size:    %s\n", formatSize(s.FileSize)))

		if item.Running {
			b.WriteString(fmt.Sprintf("  Status:  %s\n", runningStyle.Render("running")))
		}
		if item.Note != "" {
			b.WriteString(fmt.Sprintf("\n  Note: %s\n", item.Note))
		}
		// Show truncated title as first message preview
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("  > "+truncate(s.Title, width-6)))
	}

	return lipgloss.NewStyle().Width(width).Render(b.String())
}

func formatSize(bytes int64) string {
	switch {
	case bytes >= 1024*1024:
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
	case bytes >= 1024:
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
```

- [ ] **Step 3: Implement View() in app.go**

Replace the placeholder `View()` method in `tui/app.go`:

```go
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	listWidth := m.width * 55 / 100
	previewWidth := m.width - listWidth - 3 // border

	// Header
	filterLabel := "all"
	if m.filter == filterClaude {
		filterLabel = "claude"
	} else if m.filter == filterCodex {
		filterLabel = "codex"
	}
	prompt := "> " + m.query
	if m.noteInput {
		prompt = "note> " + m.noteText
	}
	if m.deepMode {
		prompt = "/search> " + m.query
	}
	header := fmt.Sprintf(" %s%s [%d] %s",
		prompt,
		strings.Repeat(" ", max(0, listWidth-len(prompt)-20)),
		len(m.filtered),
		filterLabel)
	header = titleStyle.Render(truncate(header, listWidth))

	// List
	listHeight := m.height - 4 // header + footer
	var listLines []string

	// Determine visible range
	start := 0
	if m.cursor >= listHeight {
		start = m.cursor - listHeight + 1
	}
	end := min(start+listHeight, len(m.filtered))

	prevDir := ""
	for vi := start; vi < end; vi++ {
		item := m.items[m.filtered[vi]]
		line := m.renderListLine(item, vi == m.cursor, listWidth, &prevDir)
		listLines = append(listLines, line)
	}

	// Pad remaining height
	for len(listLines) < listHeight {
		listLines = append(listLines, "")
	}

	listView := header + "\n" + strings.Join(listLines, "\n")

	// Preview
	preview := previewBorder.Width(previewWidth).Height(m.height - 2).Render(
		m.renderPreview(previewWidth))

	// Combine
	body := lipgloss.JoinHorizontal(lipgloss.Top, listView, preview)

	// Footer
	footer := statusBarStyle.Width(m.width).Render(
		"enter open │ d done │ / search │ tab filter │ g group │ ? help")

	return body + "\n" + footer
}

func (m Model) renderListLine(item ListItem, selected bool, width int, prevDir *string) string {
	var line string

	switch item.Kind {
	case KindWorkspace:
		alias := workspaceStyle.Render("☆ " + item.Alias)
		path := dimStyle.Render(shortenPath(item.Path, width-len(item.Alias)-6))
		line = fmt.Sprintf(" %s  %s", alias, path)

	case KindSession:
		s := item.Session
		prefix := "  "
		if item.Running {
			prefix = runningStyle.Render("● ")
		}

		source := sourceClaudeStyle.Render("claude")
		if s.Source == "codex" {
			source = sourceCodexStyle.Render("codex ")
		}

		date := s.UpdatedAt.Format("1/02")
		title := truncate(s.Title, width-20)

		if m.grouped && s.Directory != *prevDir && s.Directory != "" {
			*prevDir = s.Directory
			dirLine := dimStyle.Render("  " + shortenPath(s.Directory, width-4))
			line = dirLine + "\n"
		}

		line += fmt.Sprintf(" %s%s %s  %s", prefix, source, title, dimStyle.Render(date))
	}

	if selected {
		line = selectedStyle.Render(line)
	}

	return line
}
```

- [ ] **Step 4: Add missing `fmt` and `os` imports to app.go if not already present**

Ensure `tui/app.go` imports include:
```go
import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Huweicai/axe/provider"
	"github.com/Huweicai/axe/state"
)
```

- [ ] **Step 5: Verify build**

Run: `cd /Users/huweicai/Projects/GoProjects/github.com/Huweicai/axe && go build ./tui/`
Expected: compiles without errors

- [ ] **Step 6: Commit**

```bash
cd /Users/huweicai/Projects/GoProjects/github.com/Huweicai/axe
git add tui/preview.go tui/filter.go tui/app.go
git commit -m "feat: implement TUI view rendering with list and preview panes"
```

---

### Task 9: TUI — Deep Search (Concurrent JSONL Grep)

**Files:**
- Create: `tui/search.go`
- Create: `tui/search_test.go`
- Modify: `tui/app.go` (integrate deep search messages)

- [ ] **Step 1: Write the failing test**

```go
// tui/search_test.go
package tui

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Huweicai/axe/provider"
)

func testdataDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "provider", "testdata")
}

func TestDeepSearch(t *testing.T) {
	baseDir := filepath.Join(testdataDir(), "claude")
	p := provider.NewClaudeProvider(baseDir)

	results := DeepSearch([]provider.Provider{p}, "refactor")
	if len(results) == 0 {
		t.Fatal("expected at least 1 deep search result")
	}
	if results[0].SessionID != "abc123" {
		t.Errorf("SessionID = %q, want 'abc123'", results[0].SessionID)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/huweicai/Projects/GoProjects/github.com/Huweicai/axe && go test ./tui/ -run TestDeepSearch -v`
Expected: FAIL — `DeepSearch` undefined

- [ ] **Step 3: Implement deep search**

```go
// tui/search.go
package tui

import (
	"sync"

	"github.com/Huweicai/axe/provider"
)

type SearchResult struct {
	SessionID string
	Source    string
	Matches  []provider.Match
}

// DeepSearch searches all sessions from all providers concurrently.
func DeepSearch(providers []provider.Provider, keyword string) []SearchResult {
	var allSessions []provider.Session
	providerMap := make(map[string]provider.Provider)

	for _, p := range providers {
		sessions, err := p.ListSessions()
		if err != nil {
			continue
		}
		for _, s := range sessions {
			allSessions = append(allSessions, s)
		}
		providerMap[p.Name()] = p
	}

	var mu sync.Mutex
	var results []SearchResult
	var wg sync.WaitGroup

	// Limit concurrency
	sem := make(chan struct{}, 8)

	for _, s := range allSessions {
		s := s
		p := providerMap[s.Source]
		if p == nil {
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			matches, err := p.SearchContent(s.ID, keyword)
			if err != nil || len(matches) == 0 {
				return
			}
			mu.Lock()
			results = append(results, SearchResult{
				SessionID: s.ID,
				Source:    s.Source,
				Matches:  matches,
			})
			mu.Unlock()
		}()
	}
	wg.Wait()
	return results
}
```

- [ ] **Step 4: Run test**

Run: `cd /Users/huweicai/Projects/GoProjects/github.com/Huweicai/axe && go test ./tui/ -run TestDeepSearch -v`
Expected: PASS

- [ ] **Step 5: Integrate deep search into the TUI model**

Add to `tui/app.go` — a new message type and handling in Update:

```go
// Add these types near the top of app.go:
type deepSearchDone struct {
	results []SearchResult
}

// Add this method:
func (m Model) startDeepSearch() tea.Cmd {
	query := m.query
	providers := m.providers
	return func() tea.Msg {
		results := DeepSearch(providers, query)
		return deepSearchDone{results: results}
	}
}
```

In the `Update` method, add a case for `deepSearchDone`:
```go
case deepSearchDone:
    // Highlight sessions that matched deep search
    matched := make(map[string]bool)
    for _, r := range msg.results {
        matched[r.Source+":"+r.SessionID] = true
    }
    // Re-filter to only show matched sessions
    m.filtered = nil
    for i, item := range m.items {
        if item.Kind == KindSession && matched[item.StateKey()] {
            m.filtered = append(m.filtered, i)
        }
    }
    if m.cursor >= len(m.filtered) {
        m.cursor = max(0, len(m.filtered)-1)
    }
    m.deepMode = false
```

In `handleKey`, update the `"enter"` case when in deepMode:
```go
case "enter":
    if m.deepMode && m.query != "" {
        return m, m.startDeepSearch()
    }
    return m.doExec(false)
```

- [ ] **Step 6: Verify build**

Run: `cd /Users/huweicai/Projects/GoProjects/github.com/Huweicai/axe && go build ./tui/`
Expected: compiles

- [ ] **Step 7: Commit**

```bash
cd /Users/huweicai/Projects/GoProjects/github.com/Huweicai/axe
git add tui/search.go tui/search_test.go tui/app.go
git commit -m "feat: implement concurrent deep search across session files"
```

---

### Task 10: Main Entry Point

**Files:**
- Modify: `main.go`

- [ ] **Step 1: Implement main.go**

```go
// main.go
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

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

	running := process.DetectRunning()

	model := tui.New(providers, cfg, st, running)
	if *showAll {
		// Send a ctrl+a to show done sessions at startup
		// Simpler: expose a method
		model.SetShowDone(true)
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// After tea exits, exec into the selected tool if requested
	if m, ok := finalModel.(tui.Model); ok {
		if args := m.ExecArgs(); len(args) > 0 {
			if err := tui.ExecReplace(args); err != nil {
				fmt.Fprintf(os.Stderr, "exec error: %v\n", err)
				os.Exit(1)
			}
		}
	}
}
```

- [ ] **Step 2: Add SetShowDone method to tui/app.go**

```go
func (m *Model) SetShowDone(v bool) {
	m.showDone = v
	m.applyFilter()
}
```

- [ ] **Step 3: Verify full build**

Run: `cd /Users/huweicai/Projects/GoProjects/github.com/Huweicai/axe && go build -o axe .`
Expected: produces `axe` binary

- [ ] **Step 4: Smoke test against real data**

Run: `cd /Users/huweicai/Projects/GoProjects/github.com/Huweicai/axe && ./axe`
Expected: TUI launches showing workspaces and sessions. Press `q` to quit.

- [ ] **Step 5: Commit**

```bash
cd /Users/huweicai/Projects/GoProjects/github.com/Huweicai/axe
git add main.go tui/app.go
git commit -m "feat: wire up main entry point with providers and TUI"
```

---

### Task 11: Default Config Bootstrap

**Files:**
- Modify: `main.go` (create default config if missing)

- [ ] **Step 1: Add config bootstrap logic**

In `main.go`, after loading config, if `config.json` doesn't exist and we have session history, auto-generate pinned workspaces from the most frequent directories:

```go
// Add after cfg is loaded, before creating model:
configPath := filepath.Join(axeDir, "config.json")
if _, err := os.Stat(configPath); os.IsNotExist(err) {
	// Auto-generate pinned workspaces from history
	cfg.PinnedWorkspaces = inferWorkspaces(providers)
	// Save it
	data, _ := json.MarshalIndent(cfg, "", "  ")
	os.WriteFile(configPath, data, 0644)
}
```

Add the `inferWorkspaces` function:

```go
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

	// Sort by frequency, take top 5
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
```

- [ ] **Step 2: Add imports to main.go**

Ensure imports include `encoding/json`, `sort`, `path/filepath`.

- [ ] **Step 3: Verify build**

Run: `cd /Users/huweicai/Projects/GoProjects/github.com/Huweicai/axe && go build -o axe .`
Expected: compiles

- [ ] **Step 4: Test bootstrap**

Run: `rm -f ~/.axe/config.json && cd /Users/huweicai/Projects/GoProjects/github.com/Huweicai/axe && ./axe`
Expected: TUI launches, `~/.axe/config.json` is created with inferred workspaces. Press `q` to quit, then `cat ~/.axe/config.json` to verify.

- [ ] **Step 5: Commit**

```bash
cd /Users/huweicai/Projects/GoProjects/github.com/Huweicai/axe
git add main.go
git commit -m "feat: auto-generate config with inferred workspaces on first run"
```

---

### Task 12: Polish + Install

**Files:**
- Create: `Makefile`
- Create: `README.md` (minimal)

- [ ] **Step 1: Create Makefile**

```makefile
# Makefile
.PHONY: build install test clean

build:
	go build -o axe .

install:
	go install .

test:
	go test ./... -v

clean:
	rm -f axe
```

- [ ] **Step 2: Run all tests**

Run: `cd /Users/huweicai/Projects/GoProjects/github.com/Huweicai/axe && go test ./... -v`
Expected: all tests pass

- [ ] **Step 3: Install and test from PATH**

Run: `cd /Users/huweicai/Projects/GoProjects/github.com/Huweicai/axe && go install . && axe`
Expected: `axe` command works from any directory

- [ ] **Step 4: Commit**

```bash
cd /Users/huweicai/Projects/GoProjects/github.com/Huweicai/axe
git add Makefile
git commit -m "feat: add Makefile for build/install/test"
```

---

## Summary

| Task | What | Tests |
|---|---|---|
| 1 | Go module + deps | build check |
| 2 | Provider interface | compile check |
| 3 | Claude provider | 4 tests |
| 4 | Codex provider | 3 tests |
| 5 | State management | 5 tests |
| 6 | Process detection | 2 tests |
| 7 | TUI core model | 4 tests |
| 8 | TUI view rendering | build check |
| 9 | Deep search | 1 test + integration |
| 10 | Main entry point | smoke test |
| 11 | Config bootstrap | manual verify |
| 12 | Polish + install | all tests pass |
