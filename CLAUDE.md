# axe — AI session manager

TUI + CLI 管理 Claude / Codex 会话。Go，charmbracelet/bubbletea。

## 工作流（每次改动都要遵守）

- 改完 → 自动 `git commit`（**绝不加 Co-Authored-By**）；message 用 `feat:` / `fix:` / `docs:` 前缀，跟随 git 历史风格。
- commit 后 → 自动 `make install`（= `go install .`）编译安装。
- 提交前确保 `go build ./... && go test ./...` 通过。
- 互不相关的改动分开 commit，不要混在一起。

## 数据模型

- 会话本体是 `~/.claude` / `~/.codex` 下的 `.jsonl`，axe 只读它们。
- 元数据存 `~/.axe/state.json`，per-session：`archived` / `deleted_at` / `deleted_file` / `note` / `starred`。
- TTL 单一来源：`state.TrashTTL`（7 天）。

## archive / delete / restore

- **archive**（TUI `A` / `axe archive`）：隐藏、可恢复（即旧的 "done"）。
- **delete**（TUI `d` / `axe delete`）：软删除进回收站；7 天后**启动时清理**——同步从 state 删条目 + 异步 `os.Remove` 真实 `.jsonl`（真删磁盘）。
- **restore**（TUI `u` / `axe restore`）：从回收站或归档恢复。
- TUI 切换：`^a` 显示已归档，`^t` 显示回收站。
- 回收站清理只在 TUI 启动（`cmdTUI` → `purgeExpiredTrash`）触发，不在一次性 CLI 子命令里跑。
- 旧 `state.json` 里的 `done` 字段在 `LoadState` 时自动迁移成 `archived`。
