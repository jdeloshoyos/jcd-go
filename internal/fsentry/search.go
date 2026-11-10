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
