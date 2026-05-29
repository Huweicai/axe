package tui

import (
	"sync"
	"time"

	"github.com/Huweicai/axe/provider"
	tea "github.com/charmbracelet/bubbletea"
)

// deepSearchDays bounds deep search to sessions updated within this window.
const deepSearchDays = 7

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
	cutoff := time.Now().AddDate(0, 0, -deepSearchDays)
	return func() tea.Msg {
		results := DeepSearch(providers, q, cutoff)
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
// It spawns goroutines (bounded by a semaphore of 8) to call SearchContent
// on each session and collects matching results.
func DeepSearch(providers []provider.Provider, keyword string, cutoff time.Time) []SearchResult {
	// Collect sessions updated within the cutoff window
	type sessionRef struct {
		provider provider.Provider
		session  provider.Session
	}
	var refs []sessionRef
	for _, p := range providers {
		sessions, err := p.ListSessions()
		if err != nil {
			continue
		}
		for _, s := range sessions {
			if !cutoff.IsZero() && s.UpdatedAt.Before(cutoff) {
				continue
			}
			refs = append(refs, sessionRef{provider: p, session: s})
		}
	}

	var (
		mu      sync.Mutex
		results []SearchResult
		wg      sync.WaitGroup
		sem     = make(chan struct{}, 8) // concurrency limiter
	)

	for _, ref := range refs {
		wg.Add(1)
		go func(r sessionRef) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			matches, err := r.provider.SearchContent(r.session.ID, keyword)
			if err != nil || len(matches) == 0 {
				return
			}
			mu.Lock()
			results = append(results, SearchResult{
				SessionID: r.session.ID,
				Source:    r.session.Source,
				Matches:   matches,
			})
			mu.Unlock()
		}(ref)
	}

	wg.Wait()
	return results
}
