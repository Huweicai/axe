package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) renderPreviewContent(width int) string {
	sel := m.selected()
	if sel == nil {
		return dimStyle.Render("\n  No selection")
	}

	var lines []string
	kv := func(label, value string) string {
		return previewLabelStyle.Render(label+": ") + previewValueStyle.Render(value)
	}

	switch sel.Kind {
	case KindWorkspace:
		lines = append(lines,
			workspaceIconStyle.Render("★")+" "+workspaceStyle.Render(sel.Alias),
			"",
			kv("Path", shortenPath(sel.Path, width-8)),
		)
		if sel.Running {
			lines = append(lines, runningStyle.Render("● running"))
		}
		lines = append(lines, "",
			dimStyle.Render("  enter → "+m.defaultTool()),
			dimStyle.Render("  x     → "+m.altTool()),
		)

	case KindSession:
		s := sel.Session
		if s == nil {
			break
		}

		badge := claudeBadge.Render(s.Source)
		if s.Source == "codex" {
			badge = codexBadge.Render(s.Source)
		}
		lines = append(lines,
			badge,
			"",
			kv("Dir    ", shortenPath(s.Directory, width-12)),
			kv("Updated", s.UpdatedAt.Format("2006-01-02 15:04")),
			kv("Size   ", formatSize(s.FileSize)),
		)
		if sel.Running {
			lines = append(lines, "", runningStyle.Render("● running"))
		}
		if sel.Note != "" {
			lines = append(lines, "", previewLabelStyle.Render("Note: ")+previewValueStyle.Render(sel.Note))
		}
		if sel.Done {
			lines = append(lines, dimStyle.Render("  ✓ done"))
		}
		if s.Title != "" {
			lines = append(lines, "",
				dimStyle.Render("─── first message ───"),
				previewValueStyle.Render(truncate(s.Title, width-2)),
			)
		}
	}

	return strings.Join(lines, "\n")
}

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

func (m *Model) renderHeader(width int) string {
	prompt := "❯ "
	if m.deepMode {
		prompt = "/ "
	}
	query := m.query + "▏"
	if m.noteInput {
		query = m.noteText + "▏" + dimStyle.Render(" (note)")
	}

	left := prompt + query
	right := fmt.Sprintf("[%d]", len(m.filtered))
	if m.filter != filterAll {
		right += " " + m.filter.String()
	}

	gap := width - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if gap < 1 {
		gap = 1
	}

	bar := left + strings.Repeat(" ", gap) + headerDimStyle.Render(right)
	return headerStyle.Width(width).Render(bar)
}

func (m *Model) renderStatusBar(width int) string {
	items := []struct{ key, desc string }{
		{"enter", "open"},
		{"d", "done"},
		{"/", "search"},
		{"tab", "filter"},
		{"^a", "show done"},
		{"q", "quit"},
	}
	var parts []string
	for _, it := range items {
		parts = append(parts, statusKeyStyle.Render(it.key)+" "+statusBarStyle.Render(it.desc))
	}
	bar := strings.Join(parts, statusBarStyle.Render("  "))
	return statusBarStyle.Width(width).Render(bar)
}

func (m *Model) renderListRow(idx int, listWidth int, isCurrent bool) string {
	it := m.items[idx]
	contentWidth := listWidth - 4

	var row string
	switch it.Kind {
	case KindWorkspace:
		icon := workspaceIconStyle.Render("★")
		alias := workspaceStyle.Render(truncate(it.Alias, 14))
		path := dimStyle.Render(shortenPath(it.Path, contentWidth-18))
		row = fmt.Sprintf(" %s %s  %s", icon, alias, path)

	case KindSession:
		src := sourceClaudeStyle.Render("claude")
		if it.Source() == "codex" {
			src = sourceCodexStyle.Render("codex ")
		}

		titleMax := contentWidth - 22
		if titleMax < 8 {
			titleMax = 8
		}
		title := truncate(it.Title(), titleMax)
		date := dateStyle.Render(it.UpdatedAt().Format("1/02"))

		indicator := " "
		if it.Running {
			indicator = runningStyle.Render("●")
		}

		row = fmt.Sprintf(" %s %s  %s  %s", indicator, src, title, date)
	}

	if isCurrent {
		return selectedRowStyle.Width(listWidth).Render(row)
	}
	return normalRowStyle.Width(listWidth).Render(row)
}
