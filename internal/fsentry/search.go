package fsentry

import (
	"slices"
	"strings"
	"unicode/utf8"
)

// Matches returns the 1-based indices of entries whose name contains query.
// Matching is case- and diacritic-insensitive. An empty query (or one that
// folds to nothing, such as a lone combining mark) yields no matches.
func Matches(entries []Entry, query string) []int {
	q := fold(query)
	if q == "" {
		return nil
	}
	var out []int
	for i, e := range entries {
		if strings.Contains(fold(e.Name), q) {
			out = append(out, i+1)
		}
	}
	return out
}

// CountMatches returns the number of entries matching query.
func CountMatches(entries []Entry, query string) int {
	return len(Matches(entries, query))
}

// HighlightRange returns the byte-offset range [start, end) of the first case-
// and diacritic-insensitive occurrence of query in s. It returns (-1, -1) when
// query is empty (or folds to nothing) or does not occur, mirroring the
// matching rules used by Matches.
func HighlightRange(s, query string) (int, int) {
	needle := []rune(fold(query))
	if len(needle) == 0 {
		return -1, -1
	}

	// Fold s rune by rune, remembering the source byte span of every folded
	// rune. This maps a match in the folded text back onto the original even
	// when a composed accent makes the folded form shorter than the original.
	type span struct{ start, end int }
	var folded []rune
	var spans []span
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		for _, fr := range foldRune(r) {
			folded = append(folded, fr)
			spans = append(spans, span{i, i + size})
		}
		i += size
	}

	at := indexRunes(folded, needle)
	if at < 0 {
		return -1, -1
	}

	start := spans[at].start
	end := spans[at+len(needle)-1].end
	// Include combining marks that follow the final matched base rune, so a
	// decomposed accent (for example "e" + U+0301) is highlighted along with
	// its base letter.
	for end < len(s) {
		r, size := utf8.DecodeRuneInString(s[end:])
		if !isMark(r) {
			break
		}
		end += size
	}
	return start, end
}

// indexRunes returns the index of the first occurrence of needle in haystack,
// or -1. An empty needle matches at index 0.
func indexRunes(haystack, needle []rune) int {
	if len(needle) == 0 {
		return 0
	}
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if slices.Equal(haystack[i:i+len(needle)], needle) {
			return i
		}
	}
	return -1
}

// NthMatch returns the 1-based index of the n-th match. If n is out of range it
// is clamped to the available matches, matching the legacy behaviour. It
// returns 0 when nothing matches.
func NthMatch(entries []Entry, query string, n int) int {
	m := Matches(entries, query)
	if len(m) == 0 {
		return 0
	}
	if n < 1 {
		n = 1
	}
	if n > len(m) {
		n = len(m)
	}
	return m[n-1]
}
