package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"jcd-go/internal/history"
)

// quickRow is one line in the Quick Access list. Section rows (the combined
// title-and-rule headers) are never selectable.
type quickRow struct {
	section  bool
	title    string
	path     string
	favorite bool
}

const (
	quickPopupMaxWidth = 72
	quickPopupMinWidth = 40
)

// openQuickAccess builds the list from the tracked directories and shows it.
func (m *Model) openQuickAccess() {
	if m.history == nil {
		m.history = history.New("")
	}
	m.quickReturn = m.state
	m.buildQuickRows()
	m.quickCursor = -1
	m.quickOffset = 0
	m.setQuickCursor(0)
	m.state = StateQuickAccess
	m.status = ""
}

// closeQuickAccess returns to the state that opened the popup.
func (m *Model) closeQuickAccess() {
	m.state = m.quickReturn
	m.status = ""
}

// buildQuickRows flattens the favorites, recent, and frequent sections into a
// single navigable list, skipping directories that no longer exist.
func (m *Model) buildQuickRows() {
	m.quickRows = m.quickRows[:0]
	favorites, recent, frequent := m.history.Sections()

	add := func(title string, dirs []history.TrackedDirectory) {
		visible := make([]history.TrackedDirectory, 0, len(dirs))
		for _, d := range dirs {
			if m.dirExists != nil && !m.dirExists(d.Path) {
				continue
			}
			visible = append(visible, d)
		}
		if len(visible) == 0 {
			return
		}
		m.quickRows = append(m.quickRows, quickRow{section: true, title: title})
		for _, d := range visible {
			m.quickRows = append(m.quickRows, quickRow{path: d.Path, favorite: d.Favorite})
		}
	}

	add("★ Favorites", favorites)
	add("🕒 Recent", recent)
	add("🔥 Frequent", frequent)
}

// nextEntry returns the first non-section row at or beyond i in direction dir,
// or -1 when none exists.
func (m *Model) nextEntry(i, dir int) int {
	for i >= 0 && i < len(m.quickRows) {
		if !m.quickRows[i].section {
			return i
		}
		i += dir
	}
	return -1
}

// setQuickCursor selects the entry nearest to i, skipping section rows.
func (m *Model) setQuickCursor(i int) {
	if len(m.quickRows) == 0 {
		m.quickCursor = -1
		m.quickOffset = 0
		return
	}
	if i < 0 {
		i = 0
	}
	if i >= len(m.quickRows) {
		i = len(m.quickRows) - 1
	}
	if m.quickRows[i].section {
		if f := m.nextEntry(i, 1); f >= 0 {
			i = f
		} else if b := m.nextEntry(i, -1); b >= 0 {
			i = b
		} else {
			m.quickCursor = -1
			m.quickOffset = 0
			return
		}
	}
	m.quickCursor = i
	m.ensureQuickVisible()
}

// moveQuick moves the selection delta entries, skipping section headers.
func (m *Model) moveQuick(delta int) {
	if len(m.quickRows) == 0 {
		return
	}
	if m.quickCursor < 0 {
		m.setQuickCursor(0)
		return
	}
	step := 1
	if delta < 0 {
		step = -1
		delta = -delta
	}
	i := m.quickCursor
	for n := 0; n < delta; n++ {
		j := m.nextEntry(i+step, step)
		if j < 0 {
			break
		}
		i = j
	}
	m.quickCursor = i
	m.ensureQuickVisible()
}

// ensureQuickVisible scrolls the popup so the selected row is on screen.
func (m *Model) ensureQuickVisible() {
	if m.quickCursor < 0 {
		m.quickOffset = 0
		return
	}
	rows := m.quickVisibleRows()
	if m.quickCursor < m.quickOffset {
		m.quickOffset = m.quickCursor
	}
	if m.quickCursor >= m.quickOffset+rows {
		m.quickOffset = m.quickCursor - rows + 1
	}
	maxOffset := len(m.quickRows) - rows
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.quickOffset > maxOffset {
		m.quickOffset = maxOffset
	}
	if m.quickOffset < 0 {
		m.quickOffset = 0
	}
}

// quickVisibleRows is the number of list rows that fit in the popup.
func (m Model) quickVisibleRows() int {
	rows := m.height - 6 // popup border, title, and footer
	if rows < 1 {
		rows = 1
	}
	if rows > len(m.quickRows) {
		rows = len(m.quickRows)
	}
	if rows < 1 {
		rows = 1
	}
	return rows
}

// toggleQuickFavorite flips the selected entry and refreshes the list, keeping
// the cursor on the same path as it moves between sections.
func (m *Model) toggleQuickFavorite() {
	if m.quickCursor < 0 || m.quickCursor >= len(m.quickRows) {
		return
	}
	row := m.quickRows[m.quickCursor]
	if row.section {
		return
	}

	m.history.ToggleFavorite(row.path)
	_ = m.history.Save()
	m.buildQuickRows()
	m.focusQuickPath(row.path)
}

// focusQuickPath selects the first row showing path, or keeps the current
// selection when the path is gone.
func (m *Model) focusQuickPath(path string) {
	for i, row := range m.quickRows {
		if !row.section && row.path == path {
			m.quickCursor = i
			m.ensureQuickVisible()
			return
		}
	}
	m.setQuickCursor(m.quickCursor)
}

func (m Model) updateQuickAccess(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m.quit()
	case "esc", "f4":
		m.closeQuickAccess()
		return m, nil
	case "up":
		m.moveQuick(-1)
	case "down":
		m.moveQuick(1)
	case "pgup":
		m.moveQuick(-m.quickVisibleRows())
	case "pgdown":
		m.moveQuick(m.quickVisibleRows())
	case "home":
		m.setQuickCursor(0)
	case "end":
		m.setQuickCursor(len(m.quickRows) - 1)
	case "enter":
		if m.quickCursor >= 0 && m.quickCursor < len(m.quickRows) {
			if row := m.quickRows[m.quickCursor]; !row.section && row.path != "" {
				m.chosen = row.path
				m.quitting = true
				return m, tea.Quit
			}
		}
	case "f", "F":
		m.toggleQuickFavorite()
	}
	return m, nil
}

// quickBoxWidths returns the outer popup width and the content width inside
// its border and horizontal padding.
func (m Model) quickBoxWidths() (boxWidth, innerWidth int) {
	boxWidth = m.width - 8
	if boxWidth > quickPopupMaxWidth {
		boxWidth = quickPopupMaxWidth
	}
	if boxWidth < quickPopupMinWidth {
		boxWidth = quickPopupMinWidth
	}
	if boxWidth > m.width {
		boxWidth = m.width
	}
	innerWidth = boxWidth - 4 // border (2) + horizontal padding (2)
	if innerWidth < 1 {
		innerWidth = 1
	}
	return boxWidth, innerWidth
}

func (m Model) renderQuickAccess() string {
	boxWidth, innerWidth := m.quickBoxWidths()
	rows := m.quickVisibleRows()

	lines := make([]string, 0, rows+2)
	lines = append(lines, m.theme.HeaderPath.Render(padRight(truncate("Quick Access", innerWidth, "…"), innerWidth)))

	if len(m.quickRows) == 0 {
		msg := "No tracked directories yet."
		lines = append(lines, m.theme.Meta.Render(padRight(truncate(msg, innerWidth, "…"), innerWidth)))
	} else {
		end := m.quickOffset + rows
		if end > len(m.quickRows) {
			end = len(m.quickRows)
		}
		for i := m.quickOffset; i < end; i++ {
			lines = append(lines, m.renderQuickRow(m.quickRows[i], i == m.quickCursor, innerWidth))
		}
	}

	footer := "[↑↓]=Move [Enter]=Go [F]=Favorite [Esc]=Close"
	lines = append(lines, m.theme.Footer.Render(padRight(truncate(footer, innerWidth, "…"), innerWidth)))

	content := strings.Join(lines, "\n")
	box := m.theme.HelpBox.Render(content)

	// Composite the popup over the normal browser so the listing stays visible
	// around it. The popup itself is opaque: its rendered characters replace the
	// listing in the covered region.
	base := strings.Split(m.renderBrowser(), "\n")
	boxLines := strings.Split(box, "\n")

	left := (m.width - boxWidth) / 2
	if left < 0 {
		left = 0
	}
	top := (m.height - len(boxLines)) / 2
	if top < 0 {
		top = 0
	}

	return strings.Join(overlayBlock(base, boxLines, left, top), "\n")
}

func (m Model) renderQuickRow(row quickRow, selected bool, width int) string {
	if row.section {
		return m.renderQuickSection(row.title, width)
	}

	prefix := "  "
	if row.favorite {
		prefix = "★ "
	}
	avail := width - lipgloss.Width(prefix)
	if avail < 1 {
		avail = 1
	}
	path := truncateMiddlePath(row.path, avail)
	text := prefix + path

	if selected {
		return m.theme.Selected.Render(padRight(truncate(text, width, "…"), width))
	}
	line := m.theme.Favorite.Render(prefix) + m.theme.Dir.Render(path)
	return padRight(line, width)
}

// renderQuickSection draws a section header as a single line: a dim "- " prefix,
// a bright title, and a dim dashed rule that fills the remaining width.
func (m Model) renderQuickSection(title string, width int) string {
	const prefix = "- "
	prefixWidth := lipgloss.Width(prefix)
	labelWidth := prefixWidth + lipgloss.Width(title)

	// Leave room for a separating space and at least one rule character; if the
	// label is too wide for that, fall back to a plain padded label.
	if labelWidth+2 > width {
		avail := width - prefixWidth
		if avail < 1 {
			avail = 1
		}
		return m.theme.Meta.Render(prefix) +
			m.theme.Section.Render(padRight(truncate(title, avail, "…"), avail))
	}

	fill := width - labelWidth - 1
	return m.theme.Meta.Render(prefix) +
		m.theme.Section.Render(title) +
		m.theme.Meta.Render(" "+strings.Repeat("-", fill))
}

// overlayBlock returns a copy of base with block drawn over it, the block's
// top-left corner at (left, top). Rows of block that fall outside base are
// ignored.
func overlayBlock(base, block []string, left, top int) []string {
	out := make([]string, len(base))
	copy(out, base)
	for i, line := range block {
		y := top + i
		if y < 0 || y >= len(out) {
			continue
		}
		out[y] = spliceLine(out[y], line, left)
	}
	return out
}

// spliceLine replaces the display cells of base starting at column left with
// overlay. It works in terminal cells rather than bytes, so ANSI styling in
// both strings is preserved. An explicit reset separates the three parts to
// stop styling from bleeding across the seams.
func spliceLine(base, overlay string, left int) string {
	if left < 0 {
		left = 0
	}
	baseWidth := lipgloss.Width(base)
	if left >= baseWidth {
		return base
	}

	overlayWidth := lipgloss.Width(overlay)
	if left+overlayWidth > baseWidth {
		overlay = truncate(overlay, baseWidth-left, "")
		overlayWidth = baseWidth - left
	}

	var b strings.Builder
	if left > 0 {
		b.WriteString(ansi.Truncate(base, left, ""))
		b.WriteString("\x1b[0m")
	}
	b.WriteString(overlay)
	b.WriteString("\x1b[0m")
	if left+overlayWidth < baseWidth {
		b.WriteString(ansi.TruncateLeft(base, left+overlayWidth, ""))
	}
	return b.String()
}
