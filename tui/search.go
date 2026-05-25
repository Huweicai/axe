package tui

import (
	"github.com/Huweicai/axe/provider"
	tea "github.com/charmbracelet/bubbletea"
)

// SearchResult holds matches found in a single session during deep search.
type SearchResult struct {
	SessionID string
	Source    string
	Matches   []provider.Match
}

// deepSearchDone is sent when a deep search completes.
type deepSearchDone struct {
	results []SearchResult
}

// startDeepSearch returns a tea.Cmd that performs deep search in the background.
func (m *Model) startDeepSearch() tea.Cmd {
	q := m.query
	providers := m.providers
	return func() tea.Msg {
		results := DeepSearch(providers, q)
		return deepSearchDone{results: results}
	}
}

// handleDeepSearchDone updates the model with deep search results.
func (m *Model) handleDeepSearchDone(msg deepSearchDone) {
	// Build a set of matched session keys
	matched := make(map[string]bool)
	for _, r := range msg.results {
		matched[r.Source+":"+r.SessionID] = true
	}

	// Filter items to only matched sessions
	m.filtered = nil
	for i, it := range m.items {
		if it.Kind == KindSession && matched[it.StateKey()] {
			m.filtered = append(m.filtered, i)
		}
	}
	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
	m.deepMode = false
}

// DeepSearch concurrently searches all sessions across all providers.
// Stub — full implementation in commit 3.
func DeepSearch(providers []provider.Provider, keyword string) []SearchResult {
	return nil
}
