package fsentry

import "strings"

// Matches returns the 1-based indices of entries whose name contains query,
// case-insensitively. An empty query yields no matches.
func Matches(entries []Entry, query string) []int {
	if query == "" {
		return nil
	}
	q := strings.ToLower(query)
	var out []int
	for i, e := range entries {
		if strings.Contains(strings.ToLower(e.Name), q) {
			out = append(out, i+1)
		}
	}
	return out
}

// CountMatches returns the number of entries matching query.
func CountMatches(entries []Entry, query string) int {
	return len(Matches(entries, query))
}

// HighlightRange returns the byte-offset range [start, end) of the first
// case-insensitive occurrence of query in s. It returns (-1, -1) when query is
// empty or does not occur, mirroring the matching rules used by Matches.
func HighlightRange(s, query string) (int, int) {
	if query == "" {
		return -1, -1
	}
	start := strings.Index(strings.ToLower(s), strings.ToLower(query))
	if start < 0 {
		return -1, -1
	}
	return start, start + len(query)
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
