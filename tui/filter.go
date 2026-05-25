package tui

import (
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// truncate trims s to maxLen runes, appending ".." if truncated.
func truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	count := utf8.RuneCountInString(s)
	if count <= maxLen {
		return s
	}
	if maxLen <= 2 {
		return ".."
	}
	runes := []rune(s)
	return string(runes[:maxLen-2]) + ".."
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
