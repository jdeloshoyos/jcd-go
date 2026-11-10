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

// Sort orders entries in place according to mode. Directories and drives are
// always placed before files; remaining ties are broken by name (ascending).
func Sort(entries []Entry, mode SortMode) {
	mode = mode.Normalize()
	slices.SortStableFunc(entries, func(a, b Entry) int {
		return compare(a, b, mode)
	})
}

func compare(a, b Entry, mode SortMode) int {
	// Directories (and drives) always come first.
	if a.IsDirLike() != b.IsDirLike() {
		if a.IsDirLike() {
			return -1
		}
		return 1
	}

	switch mode {
	case SortNameAsc, SortNameDesc:
		c := strings.Compare(foldName(a.Name), foldName(b.Name))
		if mode == SortNameDesc {
			c = -c
		}
		if c != 0 {
			return c
		}
	case SortDateAsc, SortDateDesc:
		c := a.ModTime.Compare(b.ModTime)
		if mode == SortDateDesc {
			c = -c
		}
		if c != 0 {
			return c
		}
	case SortSizeAsc, SortSizeDesc:
		c := cmp.Compare(a.Size, b.Size)
		if mode == SortSizeDesc {
			c = -c
		}
		if c != 0 {
			return c
		}
	}

	// Name is the ascending tiebreaker for every mode.
	if c := strings.Compare(foldName(a.Name), foldName(b.Name)); c != 0 {
		return c
	}
	return strings.Compare(a.Name, b.Name)
}

func foldName(s string) string {
	return strings.ToLower(s)
}
