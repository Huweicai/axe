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
	DefaultTool      string            `json:"default_tool"`
	Theme            string            `json:"theme,omitempty"`
	DirAliases       map[string]string `json:"dir_aliases,omitempty"`
	PinnedWorkspaces []Workspace       `json:"pinned_workspaces"`
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
	Done    bool   `json:"done,omitempty"`
	Note    string `json:"note,omitempty"`
	Starred bool   `json:"starred,omitempty"`
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

func (s *State) MarkDone(key string) {
	s.getOrCreate(key).Done = true
}

func (s *State) UndoDone(key string) {
	s.getOrCreate(key).Done = false
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
