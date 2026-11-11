package fsentry

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// fold normalizes s for case- and diacritic-insensitive comparison. It
// lowercases the text, decomposes composed characters into a base plus
// combining marks (NFD), and drops the combining marks. As a result, "imá",
// "imà", and "ïma" all fold so that a search for "ima" matches them.
func fold(s string) string {
	if s == "" {
		return ""
	}
	decomposed := norm.NFD.String(strings.ToLower(s))
	var b strings.Builder
	b.Grow(len(decomposed))
	for _, r := range decomposed {
		if isMark(r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// foldRune folds a single rune to its lowercase base form, returning zero or
// more runes. A combining mark on its own folds to nothing.
func foldRune(r rune) []rune {
	decomposed := norm.NFD.String(strings.ToLower(string(r)))
	var out []rune
	for _, d := range decomposed {
		if isMark(d) {
			continue
		}
		out = append(out, d)
	}
	return out
}

// isMark reports whether r is a combining (nonspacing) mark, which NFD splits
// off from a composed character.
func isMark(r rune) bool {
	return unicode.Is(unicode.Mn, r)
}
