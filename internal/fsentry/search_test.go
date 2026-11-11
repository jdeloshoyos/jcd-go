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

func TestMatchesIgnoresAccents(t *testing.T) {
	entries := []Entry{
		{Name: "imágenes"},   // 1
		{Name: "café"},       // 2
		{Name: "cafe\u0301"}, // 3: decomposed café (e + combining acute)
		{Name: "naïve"},      // 4
		{Name: "über"},       // 5
		{Name: "plain"},      // 6
	}

	tests := []struct {
		name  string
		query string
		want  []int
	}{
		{"unaccented query matches accented name", "imagenes", []int{1}},
		{"accented query matches same name", "IMÁGENES", []int{1}},
		{"partial unaccented", "ima", []int{1}},
		{"unaccented query matches precomposed and decomposed", "cafe", []int{2, 3}},
		{"accented query matches precomposed and decomposed", "café", []int{2, 3}},
		{"diaeresis", "naive", []int{4}},
		{"accented diaeresis query", "naïve", []int{4}},
		{"umlaut", "uber", []int{5}},
		{"uppercase accented query", "ÜBER", []int{5}},
		{"combining mark alone matches nothing", "\u0301", nil},
		{"no match", "zzz", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Matches(entries, tt.query); !slices.Equal(got, tt.want) {
				t.Errorf("Matches(%q) = %v, want %v", tt.query, got, tt.want)
			}
		})
	}
}

func TestHighlightRangeIgnoresAccents(t *testing.T) {
	tests := []struct {
		name  string
		s     string
		query string
		start int
		end   int
	}{
		{"unaccented query over accent", "imágenes", "ima", 0, 4},
		{"after the accent", "imágenes", "genes", 4, 9},
		{"whole name", "imágenes", "imagenes", 0, 9},
		{"single accented rune", "imágenes", "á", 2, 4},
		{"uppercase single accented rune", "imágenes", "Á", 2, 4},
		{"precomposed accent", "café", "cafe", 0, 5},
		{"decomposed accent includes the mark", "cafe\u0301", "café", 0, 6},
		{"diaeresis in the middle", "naïve", "naive", 0, 6},
		{"leading umlaut", "über", "uber", 0, 5},
		{"no match", "plain", "zzz", -1, -1},
		{"empty query", "plain", "", -1, -1},
		{"combining mark alone", "plain", "\u0301", -1, -1},
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
