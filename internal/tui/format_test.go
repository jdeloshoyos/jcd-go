package tui

import (
	"testing"

	"charm.land/lipgloss/v2"
)

func TestTruncateMiddlePath(t *testing.T) {
	tests := []struct {
		name  string
		path  string
		width int
		want  string
	}{
		{
			name:  "fits unchanged",
			path:  `C:\users\jaime`,
			width: 40,
			want:  `C:\users\jaime`,
		},
		{
			name:  "keeps beginning and end",
			path:  `C:\users\jaime\documents\projects\doccompose\backend`,
			width: 37,
			want:  `C:\users\jaime\...\doccompose\backend`,
		},
		{
			name:  "removes one leading segment",
			path:  `C:\users\jaime\documents\projects\doccompose\backend`,
			width: 31,
			want:  `C:\users\...\doccompose\backend`,
		},
		{
			name:  "keeps only the root and tail",
			path:  `C:\users\jaime\documents\projects\doccompose\backend`,
			width: 25,
			want:  `C:\...\doccompose\backend`,
		},
		{
			name:  "drops the parent when necessary",
			path:  `C:\users\jaime\documents\projects\doccompose\backend`,
			width: 14,
			want:  `C:\...\backend`,
		},
		{
			name:  "last resort truncates the leaf",
			path:  `C:\users\jaime\documents\projects\doccompose\backend`,
			width: 10,
			want:  `...backend`,
		},
		{
			name:  "unix path",
			path:  `/home/jaime/projects/doccompose/backend`,
			width: 30,
			want:  `/home/.../doccompose/backend`,
		},
		{
			name:  "zero width",
			path:  `C:\projects`,
			width: 0,
			want:  ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := truncateMiddlePath(tt.path, tt.width); got != tt.want {
				t.Errorf("truncateMiddlePath(%q, %d) = %q, want %q", tt.path, tt.width, got, tt.want)
			}
		})
	}
}

func TestTruncateMiddlePathNeverExceedsWidth(t *testing.T) {
	paths := []string{
		`C:\users\jaime\documents\projects\doccompose\backend`,
		`\\server\share\team\projects\doccompose\backend`,
		`/home/jaime/projects/doccompose/backend`,
		`/a`,
		`C:\`,
	}
	for _, p := range paths {
		for width := 1; width <= 60; width++ {
			got := truncateMiddlePath(p, width)
			if w := lipgloss.Width(got); w > width {
				t.Errorf("truncateMiddlePath(%q, %d) = %q (%d cells)", p, width, got, w)
			}
		}
	}
}

func TestSplitPathParts(t *testing.T) {
	tests := []struct {
		path     string
		wantRoot string
		wantSegs []string
	}{
		{`C:\a\b`, `C:\`, []string{"a", "b"}},
		{`C:/a/b`, `C:/`, []string{"a", "b"}},
		{`/a/b`, `/`, []string{"a", "b"}},
		{`\\server\share\a\b`, `\\server\share\`, []string{"a", "b"}},
		{`relative\a\b`, ``, []string{"relative", "a", "b"}},
	}
	for _, tt := range tests {
		root, segs := splitPathParts(tt.path)
		if root != tt.wantRoot {
			t.Errorf("splitPathParts(%q) root = %q, want %q", tt.path, root, tt.wantRoot)
		}
		if len(segs) != len(tt.wantSegs) {
			t.Errorf("splitPathParts(%q) segs = %v, want %v", tt.path, segs, tt.wantSegs)
			continue
		}
		for i := range segs {
			if segs[i] != tt.wantSegs[i] {
				t.Errorf("splitPathParts(%q) segs = %v, want %v", tt.path, segs, tt.wantSegs)
				break
			}
		}
	}
}
