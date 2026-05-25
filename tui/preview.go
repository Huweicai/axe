package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderPreviewContent renders the content for the right-side detail pane (no border).
func (m *Model) renderPreviewContent(width int) string {
	sel := m.selected()
	if sel == nil {
		return dimStyle.Render("no selection")
	}

	var lines []string

	switch sel.Kind {
	case KindWorkspace:
		lines = append(lines,
			workspaceStyle.Render("☆ Workspace"),
			"",
			fmt.Sprintf("Alias: %s", sel.Alias),
			fmt.Sprintf("Path:  %s", shortenPath(sel.Path, width-7)),
		)
		if sel.Running {
			lines = append(lines, "", runningStyle.Render("● running"))
		}
		lines = append(lines, "",
			dimStyle.Render("enter → open with "+m.defaultTool()),
			dimStyle.Render("x     → open with "+m.altTool()),
		)

	case KindSession:
		s := sel.Session
		if s == nil {
			break
		}

		srcStyle := sourceClaudeStyle
		if s.Source == "codex" {
			srcStyle = sourceCodexStyle
		}
		lines = append(lines,
			srcStyle.Render("Source: "+s.Source),
			"",
			fmt.Sprintf("Dir:     %s", shortenPath(s.Directory, width-10)),
			fmt.Sprintf("Updated: %s", s.UpdatedAt.Format("2006-01-02 15:04")),
			fmt.Sprintf("Size:    %s", formatSize(s.FileSize)),
		)
		if sel.Running {
			lines = append(lines, runningStyle.Render("● running"))
		}
		if sel.Note != "" {
			lines = append(lines, "", "Note: "+sel.Note)
		}
		if sel.Done {
			lines = append(lines, "", dimStyle.Render("✓ done"))
		}
		if s.Title != "" {
			lines = append(lines, "", dimStyle.Render("> "+truncate(s.Title, width-3)))
		}
	}

	return strings.Join(lines, "\n")
}

// formatSize formats a byte count into a human-readable string.
func formatSize(bytes int64) string {
	switch {
	case bytes >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(1<<30))
	case bytes >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(1<<20))
	case bytes >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(1<<10))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// renderHeader builds the top header line with query, count, and filter.
func (m *Model) renderHeader(width int) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("> "))
	if m.deepMode {
		b.WriteString(titleStyle.Render("/"))
	}
	b.WriteString(m.query)
	if m.noteInput {
		b.WriteString(titleStyle.Render(" [note: " + m.noteText + "_]"))
	} else {
		b.WriteString(titleStyle.Render("_"))
	}

	count := fmt.Sprintf(" [%d]", len(m.filtered))
	b.WriteString(dimStyle.Render(count))

	if m.filter != filterAll {
		b.WriteString(" ")
		if m.filter == filterClaude {
			b.WriteString(sourceClaudeStyle.Render("claude"))
		} else {
			b.WriteString(sourceCodexStyle.Render("codex"))
		}
	}

	line := b.String()
	// Pad to full width
	return lipgloss.NewStyle().Width(width).Render(line)
}

// renderStatusBar builds the bottom status bar.
func (m *Model) renderStatusBar(width int) string {
	parts := []string{
		"enter open",
		"d done",
		"/ search",
		"tab filter",
		"g group",
		"^a show done",
		"q quit",
	}
	bar := strings.Join(parts, " │ ")
	return statusBarStyle.Width(width).Render(bar)
}

// renderListRow renders a single row for the list pane.
func (m *Model) renderListRow(idx int, listWidth int, isCurrent bool) string {
	it := m.items[idx]
	var row string

	switch it.Kind {
	case KindWorkspace:
		prefix := "  "
		if isCurrent {
			prefix = selectedStyle.Render("> ")
		}
		alias := workspaceStyle.Render("☆ " + truncate(it.Alias, 15))
		path := dimStyle.Render(shortenPath(it.Path, listWidth-22))
		row = prefix + alias + "  " + path

	case KindSession:
		prefix := "  "
		if isCurrent {
			prefix = selectedStyle.Render("> ")
		}
		// Source label
		srcLabel := it.Source()
		srcStyle := sourceClaudeStyle
		if srcLabel == "codex" {
			srcStyle = sourceCodexStyle
		}
		src := srcStyle.Render(fmt.Sprintf("%-6s", srcLabel))

		// Title
		titleMaxLen := listWidth - 24
		if titleMaxLen < 10 {
			titleMaxLen = 10
		}
		title := truncate(it.Title(), titleMaxLen)
		if it.Done {
			title = dimStyle.Render(title)
		}

		// Date
		date := it.UpdatedAt().Format("1/02")

		// Running indicator
		indicator := " "
		if it.Running {
			indicator = runningStyle.Render("●")
		}

		row = fmt.Sprintf("%s%s %s  %s  %s", prefix, indicator, src, title, dimStyle.Render(date))
	}

	return row
}
