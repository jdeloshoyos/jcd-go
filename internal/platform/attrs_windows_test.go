//go:build windows

package platform

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"unsafe"

	"jcd-go/internal/fsentry"
)

func TestAttrsFromWin32(t *testing.T) {
	all := uint32(
		fileAttributeReadOnly |
			fileAttributeHidden |
			fileAttributeSystem |
			fileAttributeDirectory |
			fileAttributeArchive |
			fileAttributeReparse |
			fileAttributeCompressed |
			fileAttributeNotIndexed |
			fileAttributeEncrypted,
	)

	tests := []struct {
		name string
		attr uint32
		want fsentry.Attributes
	}{
		{"none", 0, fsentry.Attributes{}},
		{"directory", fileAttributeDirectory, fsentry.Attributes{Directory: true}},
		{"reparse", fileAttributeReparse, fsentry.Attributes{Symlink: true}},
		{
			"read-only and hidden",
			fileAttributeReadOnly | fileAttributeHidden,
			fsentry.Attributes{ReadOnly: true, Hidden: true},
		},
		{
			"all",
			all,
			fsentry.Attributes{
				Directory:  true,
				Symlink:    true,
				ReadOnly:   true,
				Hidden:     true,
				System:     true,
				Archive:    true,
				Compressed: true,
				NotIndexed: true,
				Encrypted:  true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := attrsFromWin32(tt.attr); got != tt.want {
				t.Errorf("attrsFromWin32(%#x) = %+v, want %+v", tt.attr, got, tt.want)
			}
		})
	}
}

func TestAttrsRealDirectory(t *testing.T) {
	dir := t.TempDir()
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if a := Attrs(dir, info); !a.Directory {
		t.Errorf("Attrs(dir).Directory = false, want true")
	}
}

func TestAttrsRealReadOnly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "readonly.txt")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o444); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o644) })

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if a := Attrs(path, info); !a.ReadOnly {
		t.Errorf("Attrs(read-only file).ReadOnly = false, want true")
	}
}

func TestAttrsRealHidden(t *testing.T) {
	path := filepath.Join(t.TempDir(), "hidden.txt")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := setWin32Attributes(path, fileAttributeHidden); err != nil {
		t.Fatalf("SetFileAttributes: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if a := Attrs(path, info); !a.Hidden {
		t.Errorf("Attrs(hidden file).Hidden = false, want true")
	}
}

func TestAttrsRealSymlink(t *testing.T) {
	dir := t.TempDir()
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
		t.Errorf("Attrs(symlink).Symlink = false, want true")
	}
}

var procSetFileAttributesW = syscall.NewLazyDLL("kernel32.dll").NewProc("SetFileAttributesW")

func setWin32Attributes(path string, attrs uint32) error {
	ptr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	r, _, callErr := procSetFileAttributesW.Call(uintptr(unsafe.Pointer(ptr)), uintptr(attrs))
	if r == 0 {
		return callErr
	}
	return nil
}
