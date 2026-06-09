package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// TrashTTL is how long a soft-deleted session stays recoverable in the recycle
// bin before its underlying file is purged from disk.
const TrashTTL = 7 * 24 * time.Hour

type Workspace struct {
	Alias string `json:"alias"`
	Path  string `json:"path"`
}

type ToolConfig struct {
	ExtraArgs []string          `json:"extra_args,omitempty"`
	Env       map[string]string `json:"env,omitempty"`
}

type Config struct {
	DefaultTool      string                 `json:"default_tool"`
	Theme            string                 `json:"theme,omitempty"`
	DirAliases       map[string]string      `json:"dir_aliases,omitempty"`
	Tools            map[string]*ToolConfig `json:"tools,omitempty"`
	PinnedWorkspaces []Workspace            `json:"pinned_workspaces"`
}

func (c *Config) GetToolConfig(tool string) *ToolConfig {
	if c.Tools != nil {
		if tc, ok := c.Tools[tool]; ok {
			return tc
		}
	}
	return nil
}

func (c *Config) DirAlias(path string) string {
	if c.DirAliases != nil {
		if alias, ok := c.DirAliases[path]; ok {
			return alias
		}
	}
	for _, ws := range c.PinnedWorkspaces {
		if ws.Path == path {
			return ws.Alias
		}
	}
	return ""
}

type SessionState struct {
	Archived    bool   `json:"archived,omitempty"`
	DeletedAt   int64  `json:"deleted_at,omitempty"`   // unix seconds; 0 = not deleted (in recycle bin)
	DeletedFile string `json:"deleted_file,omitempty"` // underlying session file to purge on expiry
	Note        string `json:"note,omitempty"`
	Starred     bool   `json:"starred,omitempty"`

	// Done is the legacy pre-split field. It is loaded only to migrate old
	// state into Archived (see LoadState) and is never written back.
	Done bool `json:"done,omitempty"`
}

type State struct {
	Sessions map[string]*SessionState `json:"sessions"`
	dir      string
}

func LoadConfig(dir string) (*Config, error) {
	path := filepath.Join(dir, "config.json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Config{DefaultTool: "claude"}, nil
	}
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func LoadState(dir string) (*State, error) {
	path := filepath.Join(dir, "state.json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &State{Sessions: make(map[string]*SessionState), dir: dir}, nil
	}
	if err != nil {
		return nil, err
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	s.dir = dir
	if s.Sessions == nil {
		s.Sessions = make(map[string]*SessionState)
	}
	// Migrate legacy "done" → "archived".
	for _, ss := range s.Sessions {
		if ss.Done {
			ss.Archived = true
			ss.Done = false
		}
	}
	return &s, nil
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

func (s *State) MarkArchived(key string) {
	s.getOrCreate(key).Archived = true
}

func (s *State) UnmarkArchived(key string) {
	s.getOrCreate(key).Archived = false
}

func (s *State) IsArchived(key string) bool {
	if ss, ok := s.Sessions[key]; ok {
		return ss.Archived
	}
	return false
}

// MarkDeleted soft-deletes a session into the recycle bin, recording when and
// which underlying file to purge once the retention window passes.
func (s *State) MarkDeleted(key, file string, now time.Time) {
	ss := s.getOrCreate(key)
	ss.DeletedAt = now.Unix()
	ss.DeletedFile = file
}

// Restore takes a session back out of the recycle bin.
func (s *State) Restore(key string) {
	if ss, ok := s.Sessions[key]; ok {
		ss.DeletedAt = 0
		ss.DeletedFile = ""
	}
}

func (s *State) IsDeleted(key string) bool {
	if ss, ok := s.Sessions[key]; ok {
		return ss.DeletedAt > 0
	}
	return false
}

func (s *State) DeletedAt(key string) int64 {
	if ss, ok := s.Sessions[key]; ok {
		return ss.DeletedAt
	}
	return 0
}

// TakeExpiredDeletions removes state entries whose deletion is older than ttl
// and returns the underlying session files that should be purged from disk.
func (s *State) TakeExpiredDeletions(now time.Time, ttl time.Duration) []string {
	cutoff := now.Add(-ttl).Unix()
	var files []string
	for k, ss := range s.Sessions {
		if ss.DeletedAt > 0 && ss.DeletedAt <= cutoff {
			if ss.DeletedFile != "" {
				files = append(files, ss.DeletedFile)
			}
			delete(s.Sessions, k)
		}
	}
	return files
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

func (s *State) ToggleStar(key string) bool {
	ss := s.getOrCreate(key)
	ss.Starred = !ss.Starred
	return ss.Starred
}

func (s *State) IsStarred(key string) bool {
	if ss, ok := s.Sessions[key]; ok {
		return ss.Starred
	}
	return false
}
