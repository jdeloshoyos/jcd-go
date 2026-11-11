package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"jcd-go/internal/fsentry"
)

const dateLayout = "2006/01/02 15:04:05"

const footerKeys = "[↑↓]=Move [Enter]=Open [Bksp]=Up [F3]=Search [F4]=Quick [F1]=Help [ESC]=Quit"

const footerDriveKeys = "[↑↓]=Select [Enter]=Open [F1]=Help [ESC]=Quit"

// View implements tea.Model.
func (m Model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	v.WindowTitle = "jcd"
	if m.mouse {
		v.MouseMode = tea.MouseModeCellMotion
	}
	return v
}

func (m Model) render() string {
	if m.width < MinWidth || m.height < MinHeight {
		return fmt.Sprintf(
			"Terminal too small (%dx%d). Minimum %dx%d. Press Esc to quit.",
			m.width, m.height, MinWidth, MinHeight,
		)
	}
	if m.state == StateHelp {
		return m.renderHelp()
	}
	if m.state == StateQuickAccess {
		return m.renderQuickAccess()
	}
	return m.renderBrowser()
}

func (m Model) renderBrowser() string {
	lines := make([]string, 0, m.height)
	lines = append(lines, m.renderPath())
	lines = append(lines, m.renderInfo())

	rows := m.visibleRows()
	for i := 0; i < rows; i++ {
		idx := m.offset + i
		if m.listing != nil && idx < len(m.listing.Entries) {
			lines = append(lines, m.renderEntry(m.listing.Entries[idx], idx == m.cursor))
		} else {
			lines = append(lines, strings.Repeat(" ", m.width))
		}
	}

	lines = append(lines, m.renderFooter())
	return strings.Join(lines, "\n")
}

func (m Model) renderPath() string {
	label := m.dir
	if m.state == StateDriveSelect {
		label = "System drives"
	}
	return m.theme.HeaderPath.Width(m.width).Render(truncatePath(label, m.width))
}

func (m Model) renderInfo() string {
	if m.state == StateDriveSelect {
		count := 0
		if m.listing != nil {
			count = len(m.listing.Entries)
		}
		pos := 0
		if count > 0 {
			pos = m.cursor + 1
		}
		return m.theme.HeaderInfo.Render(twoColumn(
			fmt.Sprintf("%d drives", count),
			fmt.Sprintf("%d/%d", pos, count),
			m.width,
		))
	}

	dirs, files, total, count := 0, 0, int64(0), 0
	if m.listing != nil {
		dirs = m.listing.NumDirs
		files = m.listing.NumFiles()
		total = m.listing.TotalBytes
		count = len(m.listing.Entries)
	}

	pos := 0
	if count > 0 {
		pos = m.cursor + 1
	}

	left := fmt.Sprintf("%d directories, %d files; %s used.", dirs, files, fsentry.HumanFileSize(total))
	right := fmt.Sprintf("[F2]=Toggle sort (%s)  %d/%d", m.sortMode, pos, count)
	return m.theme.HeaderInfo.Render(twoColumn(left, right, m.width))
}

func (m Model) renderEntry(e fsentry.Entry, selected bool) string {
	if e.Kind == fsentry.KindDrive {
		text := e.Name + e.DriveType.Label()
		if selected {
			return m.theme.Selected.Width(m.width).Render(padRight(truncate(text, m.width, "…"), m.width))
		}
		return padRight(m.theme.Drive.Render(truncate(text, m.width, "…")), m.width)
	}

	mask := e.Attrs.Mask()
	date := e.ModTime.Format(dateLayout)
	suffix := ""
	if e.Kind == fsentry.KindFile {
		suffix = " (" + fsentry.HumanFileSize(e.Size) + ")"
	}
	prefix := mask + " " + date + " "

	name := truncate(e.Name, m.width-len(prefix)-len(suffix), "…")
	start, end := m.matchRange(name)

	if selected {
		line := m.theme.Selected.Render(prefix) +
			renderHighlighted(name, m.theme.Selected, m.theme.Match, start, end) +
			m.theme.Selected.Render(suffix)
		return padRightStyled(line, m.width, m.theme.Selected)
	}

	nameStyle := m.nameStyle(e)
	line := m.theme.Meta.Render(prefix) +
		renderHighlighted(name, nameStyle, m.theme.Match, start, end) +
		m.theme.Meta.Render(suffix)
	return padRight(line, m.width)
}

// nameStyle returns the style used for an entry's name when it is not selected.
func (m Model) nameStyle(e fsentry.Entry) lipgloss.Style {
	switch e.Kind {
	case fsentry.KindDir:
		return m.theme.Dir
	case fsentry.KindSymlink:
		return m.theme.Link
	default:
		return m.theme.File
	}
}

// matchRange returns the byte range of the active search term within name, or
// (-1, -1) when no search is active or the name does not match.
func (m Model) matchRange(name string) (int, int) {
	if m.state != StateSearching {
		return -1, -1
	}
	return fsentry.HighlightRange(name, m.search.Value())
}

// renderHighlighted renders name with base, emphasizing the [start, end)
// substring with the match style. The match style is layered on top of base so
// the background (e.g. the selected row) and other base attributes are kept.
func renderHighlighted(name string, base, match lipgloss.Style, start, end int) string {
	if start < 0 || start >= end || end > len(name) {
		return base.Render(name)
	}
	return base.Render(name[:start]) +
		match.Inherit(base).Render(name[start:end]) +
		base.Render(name[end:])
}

func (m Model) renderFooter() string {
	if m.state == StateSearching {
		return m.renderSearchFooter()
	}
	if m.status != "" {
		return m.theme.Status.Width(m.width).Render(truncate(m.status, m.width, "…"))
	}

	keys := footerKeys
	if m.state == StateDriveSelect {
		keys = footerDriveKeys
	}
	return m.theme.Footer.Width(m.width).Render(truncate(keys, m.width, "…"))
}

func (m Model) renderSearchFooter() string {
	count := 0
	if m.listing != nil && m.search.Value() != "" {
		count = len(fsentry.Matches(m.listing.Entries, m.search.Value()))
	}

	left := "Search: " + m.search.View()
	right := fmt.Sprintf("%s  [Enter]=Open [Tab]=Next [Esc]=Cancel", matchCountLabel(count))
	return m.theme.Footer.Width(m.width).Render(twoColumn(left, right, m.width))
}

func matchCountLabel(n int) string {
	if n == 1 {
		return "1 match"
	}
	return fmt.Sprintf("%d matches", n)
}
