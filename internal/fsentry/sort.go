package fsentry

import (
	"cmp"
	"slices"
	"strings"
)

// SortMode selects how a listing is ordered. The values match the legacy F2
// cycle: 1..6.
type SortMode int

const (
	SortNameAsc SortMode = iota + 1
	SortDateAsc
	SortSizeAsc
	SortNameDesc
	SortDateDesc
	SortSizeDesc
)

// String returns the label used in the status bar.
func (m SortMode) String() string {
	switch m {
	case SortDateAsc:
		return "Date Asc"
	case SortSizeAsc:
		return "Size Asc"
	case SortNameDesc:
		return "Name Desc"
	case SortDateDesc:
		return "Date Desc"
	case SortSizeDesc:
		return "Size Desc"
	default:
		return "Name Asc"
	}
}

// Next returns the next sort mode in the F2 cycle, wrapping after the last.
func (m SortMode) Next() SortMode {
	if m < SortNameAsc || m >= SortSizeDesc {
		return SortNameAsc
	}
	return m + 1
}

// Normalize returns a valid sort mode, defaulting to SortNameAsc.
func (m SortMode) Normalize() SortMode {
	if m < SortNameAsc || m > SortSizeDesc {
		return SortNameAsc
	}
	return m
}

// sortableEntry carries the case-folded name so each entry is folded once per
// sort rather than once per comparison.
type sortableEntry struct {
	entry  Entry
	folded string
}

// Sort orders entries in place according to mode. Directories and drives are
// always placed before files; remaining ties are broken by name (ascending).
func Sort(entries []Entry, mode SortMode) {
	mode = mode.Normalize()
	if len(entries) < 2 {
		return
	}

	decorated := make([]sortableEntry, len(entries))
	for i, e := range entries {
		decorated[i] = sortableEntry{entry: e, folded: strings.ToLower(e.Name)}
	}

	slices.SortStableFunc(decorated, func(a, b sortableEntry) int {
		return compareSorted(a, b, mode)
	})

	for i := range decorated {
		entries[i] = decorated[i].entry
	}
}

func compareSorted(a, b sortableEntry, mode SortMode) int {
	// Directories (and drives) always come first.
	if a.entry.IsDirLike() != b.entry.IsDirLike() {
		if a.entry.IsDirLike() {
			return -1
		}
		return 1
	}

	switch mode {
	case SortNameAsc, SortNameDesc:
		c := strings.Compare(a.folded, b.folded)
		if mode == SortNameDesc {
			c = -c
		}
		if c != 0 {
			return c
		}
	case SortDateAsc, SortDateDesc:
		c := a.entry.ModTime.Compare(b.entry.ModTime)
		if mode == SortDateDesc {
			c = -c
		}
		if c != 0 {
			return c
		}
	case SortSizeAsc, SortSizeDesc:
		c := cmp.Compare(a.entry.Size, b.entry.Size)
		if mode == SortSizeDesc {
			c = -c
		}
		if c != 0 {
			return c
		}
	}

	// Name is the ascending tiebreaker for every mode.
	if c := strings.Compare(a.folded, b.folded); c != 0 {
		return c
	}
	return strings.Compare(a.entry.Name, b.entry.Name)
}
