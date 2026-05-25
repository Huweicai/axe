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

type Snippet struct {
	Role string // "user" | "assistant"
	Text string
}

type Provider interface {
	Name() string
	ListSessions() ([]Session, error)
	SearchContent(sessionID string, keyword string) ([]Match, error)
	GetSnippets(sessionID string, headN, tailN int) ([]Snippet, int, error) // returns snippets, totalMsgCount, error
	ResumeCommand(sessionID string) []string
	NewCommand(cwd string) []string
}

// mergeHeadTail returns the first headN and last tailN snippets from all.
// If total <= headN+tailN, returns all without dedup.
func mergeHeadTail(all []Snippet, headN, tailN int) []Snippet {
	total := len(all)
	if total == 0 {
		return nil
	}
	if total <= headN+tailN {
		return all
	}
	var result []Snippet
	if headN > 0 {
		end := headN
		if end > total {
			end = total
		}
		result = append(result, all[:end]...)
	}
	if tailN > 0 {
		start := total - tailN
		if start < 0 {
			start = 0
		}
		result = append(result, all[start:]...)
	}
	return result
}
