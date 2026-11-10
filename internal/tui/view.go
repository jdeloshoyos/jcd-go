package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"jcd-go/internal/fsentry"
)

const dateLayout = "2006/01/02 15:04:05"

const footerKeys = "[↑↓]=Move [Enter]=Open [Bksp]=Up [/]=Search [F2]=Sort [ESC]=Quit"

// View implements tea.Model.
func (m Model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	v.WindowTitle = "jcd"
	return v
}

func (m Model) render() string {
	if m.width < MinWidth || m.height < MinHeight {
		return fmt.Sprintf(
			"Terminal too small (%dx%d). Minimum %dx%d. Press Esc to quit.",
			m.width, m.height, MinWidth, MinHeight,
		)
	}

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
	return styleHeaderPath.Width(m.width).Render(truncatePath(m.dir, m.width))
}

func (m Model) renderInfo() string {
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
	return styleHeaderInfo.Render(twoColumn(left, right, m.width))
}

func (m Model) renderEntry(e fsentry.Entry, selected bool) string {
	if selected {
		return styleSelected.Width(m.width).Render(padRight(truncate(m.entryText(e), m.width, "…"), m.width))
	}
	return m.entryStyled(e)
}

func (m Model) entryText(e fsentry.Entry) string {
	if e.Kind == fsentry.KindDrive {
		return e.Name + e.DriveType.Label()
	}

	mask := e.Attrs.Mask()
	date := e.ModTime.Format(dateLayout)
	suffix := ""
	if e.Kind == fsentry.KindFile {
		suffix = " (" + fsentry.HumanFileSize(e.Size) + ")"
	}

	prefix := mask + " " + date + " "
	name := truncate(e.Name, m.width-len(prefix)-len(suffix), "…")
	return prefix + name + suffix
}

func (m Model) entryStyled(e fsentry.Entry) string {
	if e.Kind == fsentry.KindDrive {
		return padRight(styleDrive.Render(truncate(m.entryText(e), m.width, "…")), m.width)
	}

	mask := e.Attrs.Mask()
	date := e.ModTime.Format(dateLayout)
	suffix := ""
	if e.Kind == fsentry.KindFile {
		suffix = " (" + fsentry.HumanFileSize(e.Size) + ")"
	}

	prefix := mask + " " + date + " "
	name := truncate(e.Name, m.width-len(prefix)-len(suffix), "…")

	nameStyle := styleFile
	switch e.Kind {
	case fsentry.KindDir:
		nameStyle = styleDir
	case fsentry.KindSymlink:
		nameStyle = styleLink
	}

	rendered := styleMeta.Render(prefix) + nameStyle.Render(name) + styleMeta.Render(suffix)
	return padRight(rendered, m.width)
}

func (m Model) renderFooter() string {
	if m.state == StateSearching {
		return m.renderSearchFooter()
	}
	if m.status != "" {
		return styleStatus.Width(m.width).Render(truncate(m.status, m.width, "…"))
	}
	return styleFooter.Width(m.width).Render(truncate(footerKeys, m.width, "…"))
}

func (m Model) renderSearchFooter() string {
	count := 0
	if m.listing != nil && m.search.Value() != "" {
		count = len(fsentry.Matches(m.listing.Entries, m.search.Value()))
	}

	left := "Search: " + m.search.View()
	right := fmt.Sprintf("%d matches  [Enter]=Jump [Tab]=Next [Esc]=Cancel", count)
	return styleFooter.Width(m.width).Render(twoColumn(left, right, m.width))
}
