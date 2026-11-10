//go:build !windows

package platform

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAttrsUnix(t *testing.T) {
	dir := t.TempDir()

	t.Run("directory", func(t *testing.T) {
		sub := filepath.Join(dir, "sub")
		if err := os.Mkdir(sub, 0o755); err != nil {
			t.Fatal(err)
		}
		info, err := os.Stat(sub)
		if err != nil {
			t.Fatal(err)
		}
		if a := Attrs(sub, info); !a.Directory {
			t.Error("Directory = false, want true")
		}
	})

	t.Run("hidden dotfile", func(t *testing.T) {
		path := filepath.Join(dir, ".hidden")
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if a := Attrs(path, info); !a.Hidden {
			t.Error("Hidden = false, want true")
		}
	})

	t.Run("visible file", func(t *testing.T) {
		path := filepath.Join(dir, "visible.txt")
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if a := Attrs(path, info); a.Hidden {
			t.Error("Hidden = true, want false")
		}
	})

	t.Run("read-only", func(t *testing.T) {
		path := filepath.Join(dir, "readonly.txt")
		if err := os.WriteFile(path, []byte("x"), 0o444); err != nil {
			t.Fatal(err)
		}
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if a := Attrs(path, info); !a.ReadOnly {
			t.Error("ReadOnly = false, want true")
		}
	})

	t.Run("symlink", func(t *testing.T) {
		target := filepath.Join(dir, "target.txt")
		if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		link := filepath.Join(dir, "link.txt")
		if err := os.Symlink(target, link); err != nil {
			t.Skipf("symlink not permitted: %v", err)
		}
		info, err := os.Lstat(link)
		if err != nil {
			t.Fatal(err)
		}
		if a := Attrs(link, info); !a.Symlink {
			t.Error("Symlink = false, want true")
		}
	})
}
