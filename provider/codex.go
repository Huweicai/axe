package provider

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CodexProvider reads sessions from ~/.codex/
type CodexProvider struct {
	baseDir string
}

func NewCodexProvider(baseDir string) *CodexProvider {
	return &CodexProvider{baseDir: baseDir}
}

func (p *CodexProvider) Name() string { return "codex" }

type codexIndexEntry struct {
	ID         string    `json:"id"`
	ThreadName string    `json:"thread_name"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type codexSessionLine struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type codexSessionMeta struct {
	Cwd string `json:"cwd"`
}

type codexMessagePayload struct {
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

func (p *CodexProvider) ListSessions() ([]Session, error) {
	indexPath := filepath.Join(p.baseDir, "session_index.jsonl")
	f, err := os.Open(indexPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []codexIndexEntry
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	for scanner.Scan() {
		var e codexIndexEntry
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			continue
		}
		entries = append(entries, e)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	var sessions []Session
	for _, e := range entries {
		filePath, err := p.findSessionFile(e.ID)
		if err != nil {
			// session file missing — still include with empty dir
			sessions = append(sessions, Session{
				ID:        e.ID,
				Source:    "codex",
				Title:     e.ThreadName,
				UpdatedAt: e.UpdatedAt,
			})
			continue
		}

		info, err := os.Stat(filePath)
		var size int64
		if err == nil {
			size = info.Size()
		}

		cwd := codexExtractCwd(filePath)
		sessions = append(sessions, Session{
			ID:        e.ID,
			Source:    "codex",
			Title:     e.ThreadName,
			Directory: cwd,
			UpdatedAt: e.UpdatedAt,
			FilePath:  filePath,
			FileSize:  size,
		})
	}
	return sessions, nil
}

// codexExtractCwd reads the first session_meta line and returns cwd.
func codexExtractCwd(filePath string) string {
	f, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	for scanner.Scan() {
		var line codexSessionLine
		if err := json.Unmarshal(scanner.Bytes(), &line); err != nil {
			continue
		}
		if line.Type == "session_meta" {
			var meta codexSessionMeta
			if err := json.Unmarshal(line.Payload, &meta); err == nil && meta.Cwd != "" {
				return meta.Cwd
			}
		}
	}
	return ""
}

func (p *CodexProvider) findSessionFile(sessionID string) (string, error) {
	sessionsDir := filepath.Join(p.baseDir, "sessions")
	var found string
	err := filepath.WalkDir(sessionsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".jsonl") && strings.Contains(d.Name(), sessionID) {
			found = path
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if found == "" {
		return "", os.ErrNotExist
	}
	return found, nil
}

func (p *CodexProvider) SearchContent(sessionID, keyword string) ([]Match, error) {
	filePath, err := p.findSessionFile(sessionID)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	lower := strings.ToLower(keyword)
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	var matches []Match
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		var line codexSessionLine
		if err := json.Unmarshal(scanner.Bytes(), &line); err != nil {
			continue
		}
		if line.Type != "response_item" {
			continue
		}
		var payload codexMessagePayload
		if err := json.Unmarshal(line.Payload, &payload); err != nil {
			continue
		}
		for _, c := range payload.Content {
			if (c.Type == "input_text" || c.Type == "output_text") &&
				strings.Contains(strings.ToLower(c.Text), lower) {
				matches = append(matches, Match{
					SessionID: sessionID,
					Line:      c.Text,
					LineNum:   lineNum,
				})
			}
		}
	}
	return matches, scanner.Err()
}

func (p *CodexProvider) ResumeCommand(sessionID string) []string {
	return []string{"codex", "resume", sessionID}
}

func (p *CodexProvider) NewCommand(cwd string) []string {
	return []string{"codex", "--cwd", cwd}
}

// ensure interface compliance
var _ Provider = (*CodexProvider)(nil)
