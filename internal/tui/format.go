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

// truncateMiddlePath shortens a filesystem path to at most width display cells
// by removing whole middle segments and replacing them with "...". It keeps the
// beginning (volume/root plus as many leading segments as fit) and the end
// (ideally the parent and leaf). Segment names are never cut while a shorter
// combination still fits; only as a last resort is the leaf itself truncated.
func truncateMiddlePath(path string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(path) <= width {
		return path
	}

	root, segments := splitPathParts(path)
	if len(segments) == 0 {
		return truncate(path, width, "…")
	}

	sep := "/"
	switch {
	case strings.HasSuffix(root, `\`):
		sep = `\`
	case strings.HasSuffix(root, "/"):
		sep = "/"
	case strings.Contains(path, `\`):
		sep = `\`
	}

	// Prefer keeping the last two segments (parent and leaf); fall back to the
	// leaf alone if even that does not fit.
	maxTail := len(segments)
	if maxTail > 2 {
		maxTail = 2
	}
	for tail := maxTail; tail >= 1; tail-- {
		for lead := len(segments) - tail; lead >= 0; lead-- {
			candidate := joinTruncated(root, segments, lead, tail, sep)
			if lipgloss.Width(candidate) <= width {
				return candidate
			}
		}
	}

	// Last resort: drop the root entirely and left-truncate the leaf.
	leaf := segments[len(segments)-1]
	if width <= 3 {
		return ansi.Truncate(leaf, width, "")
	}
	if w := lipgloss.Width(leaf); w > width-3 {
		return ansi.TruncateLeft(leaf, w-(width-3), "...")
	}
	return "..." + leaf
}

// splitPathParts separates a path into its root (volume or leading separator)
// and its remaining segments. It tolerates both Windows and Unix separators so
// that truncation can be tested on any platform.
func splitPathParts(path string) (root string, segments []string) {
	rest := path
	if len(rest) >= 2 && isPathSep(rest[0]) && isPathSep(rest[1]) {
		// UNC path: \\server\share
		i := 2
		for i < len(rest) && !isPathSep(rest[i]) {
			i++
		}
		if i < len(rest) {
			i++ // separator after the server
			for i < len(rest) && !isPathSep(rest[i]) {
				i++
			}
			if i < len(rest) {
				i++ // separator after the share
			}
		}
		return rest[:i], splitSegments(rest[i:])
	}
	if len(rest) >= 2 && rest[1] == ':' {
		// Drive path: C:\ or C:/
		i := 2
		if i < len(rest) && isPathSep(rest[i]) {
			i++
		}
		return rest[:i], splitSegments(rest[i:])
	}
	if len(rest) > 0 && isPathSep(rest[0]) {
		return string(rest[0]), splitSegments(rest[1:])
	}
	return "", splitSegments(rest)
}

func splitSegments(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		return r == '\\' || r == '/'
	})
}

func isPathSep(b byte) bool { return b == '\\' || b == '/' }

// joinTruncated assembles root + leading segments + "..." + trailing segments,
// omitting the ellipsis when nothing was dropped.
func joinTruncated(root string, segments []string, lead, tail int, sep string) string {
	parts := make([]string, 0, len(segments)+2)
	if root != "" {
		parts = append(parts, strings.TrimRight(root, `\/`))
	}
	parts = append(parts, segments[:lead]...)
	if lead < len(segments)-tail {
		parts = append(parts, "...")
	}
	parts = append(parts, segments[len(segments)-tail:]...)
	return strings.Join(parts, sep)
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
