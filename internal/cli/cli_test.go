package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveStartDir(t *testing.T) {
	base := t.TempDir()
	sub := filepath.Join(base, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		arg  string
		want string
	}{
		{"empty uses cwd", "", base},
		{"absolute", sub, sub},
		{"relative", "sub", sub},
		{"current directory", ".", base},
		{"trailing separator", "sub" + string(filepath.Separator), sub},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveStartDir(tt.arg, base)
			if err != nil {
				t.Fatalf("ResolveStartDir(%q) error: %v", tt.arg, err)
			}
			if got != tt.want {
				t.Errorf("ResolveStartDir(%q) = %q, want %q", tt.arg, got, tt.want)
			}
		})
	}
}

func TestResolveStartDirErrors(t *testing.T) {
	base := t.TempDir()
	file := filepath.Join(base, "file.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := ResolveStartDir("missing", base); err == nil {
		t.Error("expected an error for a missing directory")
	}
	if _, err := ResolveStartDir(file, base); err == nil {
		t.Error("expected an error for a file")
	}
}

func TestResolveStartDirWindowsRootedPath(t *testing.T) {
	if filepath.Separator != '\\' {
		t.Skip("Windows-only path semantics")
	}

	base := t.TempDir()
	got, err := ResolveStartDir(`\`, base)
	if err != nil {
		t.Fatalf(`ResolveStartDir("\\") error: %v`, err)
	}
	want := filepath.VolumeName(base) + `\`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
