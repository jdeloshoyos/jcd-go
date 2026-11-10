package cli

import (
	"os"
	"path/filepath"
	"strings"
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

func TestResolveStartDirSpecialPaths(t *testing.T) {
	base := t.TempDir()
	names := []string{
		"target with spaces",
		"café-Δ-日本",
		"amp & (paren) 'quote';,=",
	}

	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			dir := filepath.Join(base, name)
			if err := os.Mkdir(dir, 0o755); err != nil {
				t.Fatalf("mkdir: %v", err)
			}

			if got, err := ResolveStartDir(name, base); err != nil {
				t.Errorf("relative resolve: %v", err)
			} else if got != dir {
				t.Errorf("relative resolve = %q, want %q", got, dir)
			}

			if got, err := ResolveStartDir(dir, base); err != nil {
				t.Errorf("absolute resolve: %v", err)
			} else if got != dir {
				t.Errorf("absolute resolve = %q, want %q", got, dir)
			}
		})
	}
}

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    Options
		wantErr bool
	}{
		{"no args", nil, Options{Action: ActionRun}, false},
		{"directory", []string{"some/dir"}, Options{Action: ActionRun, Dir: "some/dir"}, false},
		{"version long", []string{"--version"}, Options{Action: ActionVersion}, false},
		{"version short", []string{"-v"}, Options{Action: ActionVersion}, false},
		{"help long", []string{"--help"}, Options{Action: ActionHelp}, false},
		{"help short", []string{"-h"}, Options{Action: ActionHelp}, false},
		{"end of options", []string{"--", "-weird"}, Options{Action: ActionRun, Dir: "-weird"}, false},
		{"end of options only", []string{"--"}, Options{Action: ActionRun}, false},
		{"single dash path", []string{"-"}, Options{Action: ActionRun, Dir: "-"}, false},
		{"too many args", []string{"a", "b"}, Options{}, true},
		{"unknown option", []string{"-x"}, Options{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseArgs(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseArgs(%v) = %+v, want error", tt.args, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseArgs(%v): %v", tt.args, err)
			}
			if got != tt.want {
				t.Errorf("ParseArgs(%v) = %+v, want %+v", tt.args, got, tt.want)
			}
		})
	}
}

func TestUsage(t *testing.T) {
	var b strings.Builder
	Usage(&b, "9.9.9")
	out := b.String()

	for _, want := range []string{"jcd 9.9.9", "Usage:", "[directory]", "--version"} {
		if !strings.Contains(out, want) {
			t.Errorf("usage missing %q:\n%s", want, out)
		}
	}
}
