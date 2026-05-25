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
