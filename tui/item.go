package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/Huweicai/axe/provider"
)

// ItemKind distinguishes workspace entries from session entries.
type ItemKind int

const (
	KindWorkspace ItemKind = iota
	KindSession
)

// ListItem is a unified row in the TUI list — either a pinned workspace or a session.
type ListItem struct {
	Kind      ItemKind
	Alias     string // workspace alias (KindWorkspace only)
	Path      string // workspace path or session directory
	Session   *provider.Session
	Done      bool
	Note      string
	Running   bool
	MatchText string // lowered title + dir + note for search
}

// Title returns the display title for this item.
func (li ListItem) Title() string {
	if li.Kind == KindWorkspace {
		if li.Alias != "" {
			return li.Alias
		}
		return li.Path
	}
	if li.Session != nil {
		return li.Session.Title
	}
	return ""
}

// UpdatedAt returns the last-modified time (zero for workspaces).
func (li ListItem) UpdatedAt() time.Time {
	if li.Session != nil {
		return li.Session.UpdatedAt
	}
	return time.Time{}
}

// Source returns the provider name ("claude"/"codex") or empty for workspaces.
func (li ListItem) Source() string {
	if li.Session != nil {
		return li.Session.Source
	}
	return ""
}

// StateKey returns "source:id" used as the key in state.State.Sessions.
func (li ListItem) StateKey() string {
	if li.Session != nil {
		return fmt.Sprintf("%s:%s", li.Session.Source, li.Session.ID)
	}
	return ""
}

// buildMatchText computes the lowered search text for an item.
func buildMatchText(li ListItem) string {
	parts := []string{strings.ToLower(li.Title()), strings.ToLower(li.Path)}
	if li.Note != "" {
		parts = append(parts, strings.ToLower(li.Note))
	}
	if li.Alias != "" {
		parts = append(parts, strings.ToLower(li.Alias))
	}
	return strings.Join(parts, " ")
}
