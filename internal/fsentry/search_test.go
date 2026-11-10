package fsentry

import (
	"slices"
	"testing"
)

func searchEntries() []Entry {
	return []Entry{
		{Name: "README.md"},
		{Name: "readme.txt"},
		{Name: "cat.png"},
		{Name: "dog.png"},
		{Name: "Catalog"},
	}
}

func TestMatches(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  []int
	}{
		{"case-insensitive substring", "read", []int{1, 2}},
		{"uppercase query", "CAT", []int{3, 5}},
		{"anywhere in name", "alog", []int{5}},
		{"no match", "zzz", nil},
		{"empty query", "", nil},
		{"full name", "dog.png", []int{4}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Matches(searchEntries(), tt.query); !slices.Equal(got, tt.want) {
				t.Errorf("Matches(%q) = %v, want %v", tt.query, got, tt.want)
			}
		})
	}
}

func TestCountMatches(t *testing.T) {
	tests := []struct {
		query string
		want  int
	}{
		{"read", 2},
		{"cat", 2},
		{"zzz", 0},
		{"", 0},
	}
	for _, tt := range tests {
		if got := CountMatches(searchEntries(), tt.query); got != tt.want {
			t.Errorf("CountMatches(%q) = %d, want %d", tt.query, got, tt.want)
		}
	}
}

func TestHighlightRange(t *testing.T) {
	tests := []struct {
		name  string
		s     string
		query string
		start int
		end   int
	}{
		{"match in middle", "main.go", "in", 2, 4},
		{"case-insensitive", "Catalog", "CAT", 0, 3},
		{"match at end", "README.md", "md", 7, 9},
		{"first occurrence only", "banana", "an", 1, 3},
		{"no match", "dog.png", "zzz", -1, -1},
		{"empty query", "README.md", "", -1, -1},
		{"whole name", "docs", "docs", 0, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := HighlightRange(tt.s, tt.query)
			if start != tt.start || end != tt.end {
				t.Errorf("HighlightRange(%q, %q) = (%d, %d), want (%d, %d)",
					tt.s, tt.query, start, end, tt.start, tt.end)
			}
		})
	}
}

func TestNthMatch(t *testing.T) {
	entries := searchEntries()
	tests := []struct {
		query string
		n     int
		want  int
	}{
		{"cat", 1, 3},
		{"cat", 2, 5},
		{"cat", 3, 5},  // clamped to last match (legacy behaviour)
		{"cat", 0, 3},  // clamped to first match
		{"cat", -1, 3}, // clamped to first match
		{"zzz", 1, 0},
		{"", 1, 0},
	}
	for _, tt := range tests {
		if got := NthMatch(entries, tt.query, tt.n); got != tt.want {
			t.Errorf("NthMatch(%q, %d) = %d, want %d", tt.query, tt.n, got, tt.want)
		}
	}
}
