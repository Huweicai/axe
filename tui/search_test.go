package tui

import (
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/Huweicai/axe/provider"
)

func testdataDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "provider", "testdata")
}

func TestDeepSearch_Claude(t *testing.T) {
	baseDir := filepath.Join(testdataDir(), "claude")
	p := provider.NewClaudeProvider(baseDir)

	results := DeepSearch([]provider.Provider{p}, "refactor", time.Time{})

	if len(results) == 0 {
		t.Fatal("expected at least 1 result, got 0")
	}

	found := false
	for _, r := range results {
		if r.SessionID == "abc123" && r.Source == "claude" {
			found = true
			if len(r.Matches) == 0 {
				t.Error("expected matches for session abc123")
			}
		}
	}
	if !found {
		t.Error("expected result with SessionID=abc123 and Source=claude")
	}
}

func TestDeepSearch_NoMatch(t *testing.T) {
	baseDir := filepath.Join(testdataDir(), "claude")
	p := provider.NewClaudeProvider(baseDir)

	results := DeepSearch([]provider.Provider{p}, "zzz_nonexistent_keyword_zzz", time.Time{})

	if len(results) != 0 {
		t.Fatalf("expected 0 results for nonsense keyword, got %d", len(results))
	}
}

func TestDeepSearch_MultipleProviders(t *testing.T) {
	claudeDir := filepath.Join(testdataDir(), "claude")
	codexDir := filepath.Join(testdataDir(), "codex")
	cp := provider.NewClaudeProvider(claudeDir)
	xp := provider.NewCodexProvider(codexDir)

	// "cache" appears in codex testdata.
	results := DeepSearch([]provider.Provider{cp, xp}, "cache", time.Time{})

	if len(results) == 0 {
		t.Fatal("expected at least 1 result for 'cache'")
	}

	// Should have a codex match
	foundCodex := false
	for _, r := range results {
		if r.Source == "codex" {
			foundCodex = true
		}
	}
	if !foundCodex {
		t.Error("expected a codex result for 'cache'")
	}
}
