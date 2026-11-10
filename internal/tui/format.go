package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// truncate shortens s to at most width display cells, appending tail when it
// does not fit. ANSI sequences and wide characters are handled correctly.
func truncate(s string, width int, tail string) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	return ansi.Truncate(s, width, tail)
}

// truncatePath shortens a path from the left, keeping the (more useful) tail.
func truncatePath(s string, width int) string {
	if width <= 0 {
		return ""
	}
	w := lipgloss.Width(s)
	if w <= width {
		return s
	}
	if width <= 3 {
		return ansi.Truncate(s, width, "")
	}
	return ansi.TruncateLeft(s, w-(width-3), "...")
}

// padRight pads s with spaces to width display cells.
func padRight(s string, width int) string {
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

// padRightStyled pads s with spaces rendered in style to width display cells.
// It keeps a full-width background (for example, on the selected row).
func padRightStyled(s string, width int, style lipgloss.Style) string {
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}
	return s + style.Render(strings.Repeat(" ", width-w))
}

// twoColumn lays out left and right on a single line of the given width,
// truncating the left side if there is not enough room.
func twoColumn(left, right string, width int) string {
	rw := lipgloss.Width(right)
	gap := width - lipgloss.Width(left) - rw
	if gap < 1 {
		left = truncate(left, max(0, width-rw-1), "…")
		gap = width - lipgloss.Width(left) - rw
		if gap < 0 {
			gap = 0
		}
	}
	return left + strings.Repeat(" ", gap) + right
}
