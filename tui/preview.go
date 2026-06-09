package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/Huweicai/axe/state"
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
		if sel.Active {
			status := sel.ActiveStatus
			if status == "" {
				status = "idle"
			}
			lines = append(lines, "", runningStyle.Render(fmt.Sprintf("● active (%s)", status)))
		} else if sel.Running {
			lines = append(lines, "", runningStyle.Render("● running"))
		}
		if sel.Starred {
			lines = append(lines, starStyle.Render("★ starred"))
		}
		if sel.Note != "" {
			lines = append(lines, "", previewLabelStyle.Render("Note: ")+previewValueStyle.Render(sel.Note))
		}
		if sel.Deleted {
			lines = append(lines, dimStyle.Render(fmt.Sprintf("  🗑 deleted · %dd left", trashDaysLeft(sel.DeletedAt))))
		} else if sel.Archived {
			lines = append(lines, dimStyle.Render("  ✓ archived"))
		}

		// Conversation snippets
		if s.ID != m.previewSessionID {
			lines = append(lines, "", dimStyle.Render("loading..."))
		} else if len(m.previewSnippets) > 0 {
			lines = append(lines, "",
				dimStyle.Render(fmt.Sprintf("─── conversation (%d messages) ───", m.previewMsgCount)),
			)
			headN := 3
			tailN := 2
			showEllipsis := m.previewMsgCount > headN+tailN
			for i, sn := range m.previewSnippets {
				icon := "👤"
				if sn.Role == "assistant" {
					icon = "🤖"
				}
				lines = append(lines, previewValueStyle.Render(icon+" "+truncate(sn.Text, width-6)))
				if showEllipsis && i == headN-1 {
					lines = append(lines, dimStyle.Render("   ···"))
				}
			}
		} else if s.Title != "" {
			lines = append(lines, "",
				dimStyle.Render("─── first message ───"),
				previewValueStyle.Render(truncate(s.Title, width-2)),
			)
		}
	}

	return strings.Join(lines, "\n")
}

// trashDaysLeft returns whole days remaining before a deleted session is purged
// (7-day retention), rounded up; never negative.
func trashDaysLeft(deletedAt int64) int {
	if deletedAt == 0 {
		return 0
	}
	expiry := time.Unix(deletedAt, 0).Add(state.TrashTTL)
	left := time.Until(expiry)
	if left <= 0 {
		return 0
	}
	return int((left + 24*time.Hour - time.Nanosecond) / (24 * time.Hour))
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
	if m.dirFilter != "" {
		name := m.config.DirAlias(m.dirFilter)
		if name == "" {
			name = filepath.Base(m.dirFilter)
		}
		right += " 📁" + name
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
		{"s", "star"},
		{"A", "arch"},
		{"d", "del"},
		{"u", "restore"},
		{"/", "search"},
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

	if isCurrent {
		return selectedRowStyle.Width(listWidth).Render(m.renderRowPlain(it, contentWidth))
	}
	return normalRowStyle.Width(listWidth).Render(m.renderRowStyled(it, contentWidth))
}

type rowLayout struct {
	ind, src, star, dirName, title, date string
	titleWidth                           int
}

func (m *Model) sessionRowLayout(it ListItem, w int) rowLayout {
	const dirWidth = 12
	const dateCols = 6

	r := rowLayout{}
	r.ind = " "
	switch {
	case it.Running:
		r.ind = "●"
	case it.Deleted:
		r.ind = "✗"
	case it.Archived:
		r.ind = "✓"
	}
	r.src = fmt.Sprintf("%-6s", it.Source())
	if it.Starred {
		r.star = "★ "
	}
	r.date = it.UpdatedAt().Format("1/02")
	r.title = it.Title()

	dir := ""
	if m.config != nil {
		dir = m.config.DirAlias(it.Path)
	}
	if dir == "" {
		dir = filepath.Base(it.Path)
		if dir == "." || dir == "/" {
			dir = ""
		}
	}
	r.dirName = truncate(dir, dirWidth)

	starCols := 0
	if it.Starred {
		starCols = 3
	}
	r.titleWidth = w - 11 - dirWidth - 2 - dateCols - starCols
	if r.titleWidth < 8 {
		r.titleWidth = 8
	}
	return r
}

func padToWidth(s string, targetWidth int) string {
	vis := lipgloss.Width(s)
	if vis >= targetWidth {
		return s
	}
	return s + strings.Repeat(" ", targetWidth-vis)
}

func (m *Model) renderRowPlain(it ListItem, w int) string {
	switch it.Kind {
	case KindWorkspace:
		return fmt.Sprintf(" ★ %s  %s", truncate(it.Alias, 14), shortenPath(it.Path, w-18))
	case KindSession:
		r := m.sessionRowLayout(it, w)
		dir := padToWidth(r.dirName, 12)
		title := padToWidth(truncate(r.title, r.titleWidth), r.titleWidth)
		return fmt.Sprintf(" %s %s %s%s %s %s", r.ind, r.src, r.star, dir, title, r.date)
	}
	return ""
}

func (m *Model) renderRowStyled(it ListItem, w int) string {
	switch it.Kind {
	case KindWorkspace:
		icon := workspaceIconStyle.Render("★")
		alias := workspaceStyle.Render(truncate(it.Alias, 14))
		path := dimStyle.Render(shortenPath(it.Path, w-18))
		return fmt.Sprintf(" %s %s  %s", icon, alias, path)
	case KindSession:
		r := m.sessionRowLayout(it, w)

		src := sourceClaudeStyle.Render("claude")
		if it.Source() == "codex" {
			src = sourceCodexStyle.Render("codex ")
		}
		star := ""
		if it.Starred {
			star = starStyle.Render("★") + " "
		}
		ind := " "
		switch {
		case it.Running:
			ind = runningStyle.Render("●")
		case it.Deleted:
			ind = dimStyle.Render("✗")
		case it.Archived:
			ind = dimStyle.Render("✓")
		}

		dir := dimStyle.Render(padToWidth(r.dirName, 12))
		title := padToWidth(truncate(r.title, r.titleWidth), r.titleWidth)
		date := dateStyle.Render(r.date)
		return fmt.Sprintf(" %s %s %s%s %s %s", ind, src, star, dir, title, date)
	}
	return ""
}
