package tui

import (
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
)

// truncate trims s to maxLen visual columns, appending ".." if truncated.
func truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	w := lipgloss.Width(s)
	if w <= maxLen {
		return s
	}
	if maxLen <= 2 {
		return ".."
	}
	runes := []rune(s)
	result := ""
	for _, r := range runes {
		test := result + string(r)
		if lipgloss.Width(test) > maxLen-2 {
			return result + ".."
		}
		result = test
	}
	return result + ".."
}

// shortenPath replaces $HOME with ~ and collapses middle segments if too long.
func shortenPath(path string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	home, _ := os.UserHomeDir()
	if home != "" && strings.HasPrefix(path, home) {
		path = "~" + path[len(home):]
	}
	if utf8.RuneCountInString(path) <= maxLen {
		return path
	}
	// Collapse middle segments: ~/a/b/c/d → ~/a/../d
	parts := strings.Split(path, string(filepath.Separator))
	if len(parts) <= 3 {
		return truncate(path, maxLen)
	}
	short := parts[0] + "/" + parts[1] + "/../" + parts[len(parts)-1]
	if utf8.RuneCountInString(short) > maxLen {
		return truncate(path, maxLen)
	}
	return short
}
