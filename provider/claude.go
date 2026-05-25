package provider

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ClaudeProvider reads sessions from ~/.claude/projects/
type ClaudeProvider struct {
	baseDir string
}

func NewClaudeProvider(baseDir string) *ClaudeProvider {
	return &ClaudeProvider{baseDir: baseDir}
}

func (p *ClaudeProvider) Name() string { return "claude" }

// claudeMessage is a flexible envelope for each JSONL line in a Claude session file.
type claudeMessage struct {
	Type        string          `json:"type"`
	IsSidechain bool            `json:"isSidechain"`
	IsMeta      bool            `json:"isMeta"`
	Cwd         string          `json:"cwd"`
	Message     json.RawMessage `json:"message"`
}

type claudeInnerMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

type claudeContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (m *claudeInnerMessage) extractText() string {
	// content can be a plain string or a list of blocks
	var plainStr string
	if err := json.Unmarshal(m.Content, &plainStr); err == nil {
		return plainStr
	}
	var blocks []claudeContentBlock
	if err := json.Unmarshal(m.Content, &blocks); err == nil {
		for _, b := range blocks {
			if b.Type == "text" && b.Text != "" {
				return b.Text
			}
		}
	}
	return ""
}

func (p *ClaudeProvider) ListSessions() ([]Session, error) {
	projectsDir := filepath.Join(p.baseDir, "projects")
	projectDirs, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil, err
	}

	var sessions []Session
	for _, pd := range projectDirs {
		if !pd.IsDir() {
			continue
		}
		dir := filepath.Join(projectsDir, pd.Name())
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".jsonl") {
				continue
			}
			filePath := filepath.Join(dir, e.Name())
			sessionID := strings.TrimSuffix(e.Name(), ".jsonl")

			info, err := e.Info()
			if err != nil {
				continue
			}

			cwd, title := claudeExtractMeta(filePath)
			sessions = append(sessions, Session{
				ID:        sessionID,
				Source:    "claude",
				Title:     title,
				Directory: cwd,
				UpdatedAt: info.ModTime(),
				FilePath:  filePath,
				FileSize:  info.Size(),
			})
		}
	}
	return sessions, nil
}

// claudeExtractMeta reads a session file and returns (cwd, title).
// title = first non-meta user message text.
func claudeExtractMeta(filePath string) (cwd, title string) {
	f, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		var msg claudeMessage
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			continue
		}
		if cwd == "" && msg.Cwd != "" {
			cwd = msg.Cwd
		}
		if title == "" && msg.Type == "user" && !msg.IsMeta {
			var inner claudeInnerMessage
			if err := json.Unmarshal(msg.Message, &inner); err == nil {
				t := strings.TrimSpace(inner.extractText())
				if t != "" {
					skip := strings.HasPrefix(t, "<") ||
						strings.HasPrefix(t, "[Request interrupted") ||
						strings.HasPrefix(t, "[Image") ||
						strings.HasPrefix(t, "Base directory") ||
						strings.HasPrefix(t, "$superpowers:")
					if !skip {
						// First line only, replace remaining newlines
						t = strings.ReplaceAll(t, "\n", " ")
						t = strings.Join(strings.Fields(t), " ")
						if len([]rune(t)) > 100 {
							t = string([]rune(t)[:100]) + ".."
						}
						title = t
					}
				}
			}
		}
		if cwd != "" && title != "" {
			break
		}
	}
	return
}

func (p *ClaudeProvider) SearchContent(sessionID, keyword string) ([]Match, error) {
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
		raw := scanner.Bytes()
		var msg claudeMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			continue
		}
		if msg.Type != "user" || msg.IsMeta {
			continue
		}
		var inner claudeInnerMessage
		if err := json.Unmarshal(msg.Message, &inner); err != nil {
			continue
		}
		text := inner.extractText()
		if text != "" && strings.Contains(strings.ToLower(text), lower) {
			matches = append(matches, Match{
				SessionID: sessionID,
				Line:      text,
				LineNum:   lineNum,
			})
		}
	}
	return matches, scanner.Err()
}

func (p *ClaudeProvider) findSessionFile(sessionID string) (string, error) {
	projectsDir := filepath.Join(p.baseDir, "projects")
	var found string
	err := filepath.WalkDir(projectsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && d.Name() == sessionID+".jsonl" {
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

func (p *ClaudeProvider) ResumeCommand(sessionID string) []string {
	return []string{"claude", "-r", sessionID}
}

func (p *ClaudeProvider) NewCommand(cwd string) []string {
	return []string{"claude", "--cwd", cwd}
}

// ensure interface compliance
var _ Provider = (*ClaudeProvider)(nil)

// ensure time package used
var _ = time.Time{}
