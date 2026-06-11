# axe — AI session manager

TUI + CLI for managing Claude and Codex sessions. Built with Go and charmbracelet/bubbletea.

## Workflow

- After changes are complete, create a `git commit` with no `Co-Authored-By` trailer. Use the repository's existing `feat:`, `fix:`, or `docs:` message style.
- After committing, run `make install`, which builds and installs via `go install .`.
- Before committing, make sure `go build ./...` and `go test ./...` pass.
- Keep unrelated changes in separate commits.

## Data Model

- Session data lives in `.jsonl` files under `~/.claude` and `~/.codex`; axe only reads them.
- Metadata lives in `~/.axe/state.json` with per-session fields: `archived`, `deleted_at`, `deleted_file`, `note`, and `starred`.
- The single source of truth for TTL is `state.TrashTTL`, currently 7 days.

## archive / delete / restore

- **archive** (TUI `A` / `axe archive`): hide a session while keeping it restorable. This replaces the old "done" state.
- **delete** (TUI `d` / `axe delete`): soft-delete a session into trash. After 7 days, startup cleanup removes the state entry synchronously and removes the real `.jsonl` file asynchronously with `os.Remove`.
- **restore** (TUI `u` / `axe restore`): restore a session from trash or archive.
- TUI toggles: `^a` shows archived sessions, and `^t` shows trash.
- Trash cleanup only runs at TUI startup through `cmdTUI` and `purgeExpiredTrash`; one-shot CLI subcommands do not run it.
- The old `done` field in `state.json` is migrated to `archived` by `LoadState`.
