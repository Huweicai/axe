# axe — AI Coding Agent Session Manager

## Overview

`axe` (Agent eXtension) is a TUI tool that unifies Claude Code and Codex session management into a single entry point. It replaces the workflow of manually grepping through `~/.claude/` and `~/.codex/` to find past conversations.

**Primary user flow:** type `axe` → search/browse → resume a session or start a new one.

## Tech Stack

- **Language:** Go
- **TUI:** Bubbletea + Lipgloss + Bubbles
- **Build:** Standard `go build`, single binary

## Data Sources

### Claude Code

| Item | Location |
|---|---|
| Session index | `~/.claude/history.jsonl` |
| Session data | `~/.claude/projects/<encoded-path>/<uuid>.jsonl` |
| Resume command | `claude -r <session-id>` |

Index entry format:
```json
{"display": "first user message...", "timestamp": 1761210520611, "project": "/abs/path/to/cwd"}
```

Session ID is the filename (minus `.jsonl`) under the project directory. The `history.jsonl` entries do NOT contain session IDs — they only have `display`, `timestamp`, and `project` (cwd). To list sessions:

1. Scan `~/.claude/projects/` directories
2. For each `<encoded-path>/<uuid>.jsonl`, decode the path to get `cwd`
3. Read the first few lines of the JSONL to extract the first user message as title
4. Use file mtime as `updated_at`

The `history.jsonl` is useful as a fallback for title extraction but is not the primary session index. The authoritative source is the filesystem scan of `~/.claude/projects/`.

### Codex

| Item | Location |
|---|---|
| Session index | `~/.codex/session_index.jsonl` |
| Session data | `~/.codex/sessions/YYYY/MM/DD/<id>.jsonl` |
| Resume command | `codex resume` (with session picker, or by ID) |

Index entry format:
```json
{"id": "uuid", "thread_name": "title", "updated_at": "2026-03-06T05:46:31Z"}
```

Session metadata (first JSONL line) contains `cwd`:
```json
{"type": "session_meta", "payload": {"id": "uuid", "cwd": "/abs/path", ...}}
```

## Architecture

```
axe/
├── main.go                  ← CLI entry point
├── provider/
│   ├── provider.go          ← Session/Workspace interfaces
│   ├── claude.go            ← Claude Code provider
│   └── codex.go             ← Codex provider
├── state/
│   └── state.go             ← ~/.axe/ config + state persistence
├── tui/
│   ├── app.go               ← Bubbletea root model
│   ├── list.go              ← Mixed list component (workspaces + sessions)
│   └── preview.go           ← Right-side detail panel
└── web/                     ← (future) HTTP server + frontend
```

### Provider Interface

```go
type Session struct {
    ID        string
    Source    string    // "claude" | "codex"
    Title     string    // first user msg (claude) or thread_name (codex)
    Directory string    // cwd
    CreatedAt time.Time
    UpdatedAt time.Time
    FilePath  string    // path to JSONL file
    FileSize  int64     // for stats display
}

type Match struct {
    SessionID string
    Line      string // matched line content
    LineNum   int
}

type Provider interface {
    Name() string
    ListSessions() ([]Session, error)
    SearchContent(sessionID string, keyword string) ([]Match, error)
    ResumeCommand(sessionID string) []string // e.g. ["claude", "-r", "uuid"]
    NewCommand(cwd string) []string          // e.g. ["claude", "--cwd", "/path"]
}
```

### State (`~/.axe/`)

```
~/.axe/
├── config.json   ← user preferences
└── state.json    ← session-level state (done, notes)
```

**config.json:**
```json
{
  "default_tool": "claude",
  "pinned_workspaces": [
    {"alias": "optical", "path": "~/Projects/GoProjects/github.com/spectra-fund/optical"},
    {"alias": "spectra", "path": "~/Projects/CProjects/spectra-huweicai"},
    {"alias": "temp",    "path": "~/Temp"}
  ]
}
```

**state.json:**
```json
{
  "sessions": {
    "claude:uuid1": {"done": true, "note": "T2O baseline measurement, concluded"},
    "codex:uuid2": {"done": false, "note": "还在迭代中"}
  }
}
```

Session keys use `source:id` format to avoid collisions between Claude and Codex UUIDs.

## TUI Design

### Default View — Mixed List + Preview

```
┌─ axe ───────────────────────────────┬──────────────────────────┐
│ > _                          [628]  │                          │
│─────────────────────────────────────│                          │
│ ☆ optical    ~/Pro../optical        │  Source: Claude Code     │
│ ☆ spectra    ~/Pro../spectra-h..    │  Dir: ~/Projects/C..     │
│ ☆ temp       ~/Temp                 │  Updated: 2026-03-06     │
│─────────────────────────────────────│  Msgs: 47  Size: 2.1MB  │
│ ● claude 收集 FuturesBase.. 3/06   │  Status: running 🟢     │
│   codex  优化下单链路..     2/28   │                          │
│   claude T2O延迟分析        2/15   │  Note: baseline done     │
│   codex  SharedMemory..     1/20   │                          │
│                                     │  > 帮我分析一下          │
│                                     │  > FuturesBaseStrategy   │
│                                     │  > 里 cache miss 高的..  │
│─────────────────────────────────────┴──────────────────────────│
│ enter open │ d done │ / deep-search │ tab filter │ ? help      │
└────────────────────────────────────────────────────────────────┘
```

List ordering:
1. Pinned workspaces (from config, with ☆ prefix)
2. Sessions sorted by `updated_at` descending
3. Done sessions hidden by default

Running indicator: `●` prefix for sessions whose source tool currently has a process running in the corresponding cwd. Detection: run `pgrep -af "claude|codex"` at startup, parse each process's `--cwd` or working directory, match against session cwds. Cached at startup, not polled.

Directory grouping (default on): sessions from the same `cwd` are visually grouped with the directory shown once, subsequent sessions indented. Press `g` to toggle flat/grouped view:

```
  ~/Projects/CProjects/spectra-huweicai
    claude  收集 FuturesBase..     3/06
    codex   优化下单链路 cache     2/28
    claude  T2O cache miss 分析    2/15
  ~/Projects/GoProjects/.../optical
    codex   Add dashboard..        3/06
    claude  合约交易参数抽提..     3/05
```

### Keybindings

| Key | On Session | On Workspace |
|---|---|---|
| `enter` | `exec` resume session | `exec` new session (default tool) |
| `x`+`enter` | — | `exec` new session (alternate tool) |
| `d` | mark done | — |
| `u` | undo done | — |
| `n` | add/edit note | — |
| `/` | deep search (grep full JSONL) | — |
| `tab` | cycle source filter: all → claude → codex | — |
| `g` | toggle flat/grouped view | — |
| `ctrl+a` | toggle show/hide done sessions | — |
| `q` / `esc` | quit | — |

### Deep Search (`/`)

Normal input filters by title + directory + note (in-memory, instant). Pressing `/` enters deep-search mode:

1. User types keyword
2. Goroutines scan JSONL files in parallel, extracting user message text
3. Matching sessions stream into the list in real-time
4. Match context shown in preview pane

### Resume / New Session

Resume:
```
claude session  → syscall.Exec("claude", ["-r", sessionID], env)
codex session   → syscall.Exec("codex", ["resume", "--id", sessionID], env)
```

New session:
```
default=claude  → syscall.Exec("claude", ["--cwd", path], env)
alternate=codex → syscall.Exec("codex", ["--cwd", path], env)
```

Uses `syscall.Exec` (not `os/exec`) to replace the process, so the user's terminal seamlessly transitions into the AI tool.

## v1 Features

### Core
- [x] Unified session listing from Claude Code + Codex
- [x] Real-time fuzzy search across title, directory, alias, note
- [x] Resume session via `exec`
- [x] New session from workspace via `exec`
- [x] Pinned workspaces with aliases
- [x] Done marking with toggle

### Extras (v1)
- [x] **Session notes** — press `n` to add a one-line note, persisted in state.json, searchable
- [x] **Running indicator** — detect active claude/codex processes per cwd, show `●` prefix
- [x] **Stats** — message count (line count heuristic), file size, displayed in preview
- [x] **Directory grouping** — sessions from same cwd grouped visually

### Deferred
- [ ] Batch done (multi-select)
- [ ] Web UI
- [ ] Session export (markdown summary)
- [ ] Cost/token tracking (if data available)
- [ ] Turn-level annotations (inspired by aimux)

## Web (Future)

Provider layer is decoupled from TUI. Future `web/` package will:
- Serve a single-page app on `localhost:<port>`
- Reuse the same `provider.Provider` interface for data
- Add HTTP handlers for search, state mutations
- `axe web` subcommand to start the server

## CLI Interface

```
axe              → launch TUI (default)
axe --all        → TUI with done sessions visible
axe search <kw>  → non-interactive search, print results (future)
```

## Non-Goals (v1)

- Modifying Claude/Codex session files (read-only)
- Real-time monitoring of active sessions
- Multi-user support
- Remote/cloud session management
