package fsentry

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func mustWrite(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.Mkdir(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func findEntry(entries []Entry, name string) *Entry {
	for i := range entries {
		if entries[i].Name == name {
			return &entries[i]
		}
	}
	return nil
}

func TestScanDirEmpty(t *testing.T) {
	dir := t.TempDir()

	listing, err := ScanDir(dir, nil, SortNameAsc)
	if err != nil {
		t.Fatalf("ScanDir: %v", err)
	}
	if len(listing.Entries) != 0 {
		t.Errorf("len(Entries) = %d, want 0", len(listing.Entries))
	}
	if listing.NumDirs != 0 {
		t.Errorf("NumDirs = %d, want 0", listing.NumDirs)
	}
	if listing.NumFiles() != 0 {
		t.Errorf("NumFiles() = %d, want 0", listing.NumFiles())
	}
	if listing.TotalBytes != 0 {
		t.Errorf("TotalBytes = %d, want 0", listing.TotalBytes)
	}
	if listing.Dir != dir {
		t.Errorf("Dir = %q, want %q", listing.Dir, dir)
	}
	if listing.SortMode != SortNameAsc {
		t.Errorf("SortMode = %v, want %v", listing.SortMode, SortNameAsc)
	}
}

func TestScanDirCountsAndSort(t *testing.T) {
	dir := t.TempDir()
	mustMkdir(t, filepath.Join(dir, "sub"))
	mustWrite(t, filepath.Join(dir, "a.txt"), []byte("hello"))  // 5 bytes
	mustWrite(t, filepath.Join(dir, "B.txt"), []byte("world!")) // 6 bytes
	mustWrite(t, filepath.Join(dir, ".hidden"), []byte("x"))    // 1 byte

	listing, err := ScanDir(dir, nil, SortNameAsc)
	if err != nil {
		t.Fatalf("ScanDir: %v", err)
	}

	want := []string{"sub", ".hidden", "a.txt", "B.txt"}
	if got := names(listing.Entries); !slices.Equal(got, want) {
		t.Errorf("entry order = %v, want %v", got, want)
	}
	if listing.NumDirs != 1 {
		t.Errorf("NumDirs = %d, want 1", listing.NumDirs)
	}
	if listing.NumFiles() != 3 {
		t.Errorf("NumFiles() = %d, want 3", listing.NumFiles())
	}
	if listing.TotalBytes != 12 {
		t.Errorf("TotalBytes = %d, want 12 (regular files only)", listing.TotalBytes)
	}
}

func TestScanDirKinds(t *testing.T) {
	dir := t.TempDir()
	mustMkdir(t, filepath.Join(dir, "sub"))
	mustWrite(t, filepath.Join(dir, "file.txt"), []byte("data"))

	listing, err := ScanDir(dir, nil, SortNameAsc)
	if err != nil {
		t.Fatalf("ScanDir: %v", err)
	}

	sub := findEntry(listing.Entries, "sub")
	if sub == nil {
		t.Fatal("missing sub directory")
	}
	if sub.Kind != KindDir || !sub.Attrs.Directory {
		t.Errorf("sub kind = %v, Directory = %v; want dir/true", sub.Kind, sub.Attrs.Directory)
	}

	file := findEntry(listing.Entries, "file.txt")
	if file == nil {
		t.Fatal("missing file.txt")
	}
	if file.Kind != KindFile {
		t.Errorf("file kind = %v, want file", file.Kind)
	}
	if file.Size != 4 {
		t.Errorf("file size = %d, want 4", file.Size)
	}
}

func TestScanDirInjectedAttributes(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, ".secret"), []byte("x"))
	mustWrite(t, filepath.Join(dir, "public.txt"), []byte("y"))

	fake := func(_ string, info os.FileInfo) Attributes {
		return Attributes{Hidden: strings.HasPrefix(info.Name(), ".")}
	}

	listing, err := ScanDir(dir, fake, SortNameAsc)
	if err != nil {
		t.Fatalf("ScanDir: %v", err)
	}

	secret := findEntry(listing.Entries, ".secret")
	if secret == nil {
		t.Fatal("missing .secret")
	}
	if !secret.Attrs.Hidden {
		t.Errorf(".secret Hidden = false, want true")
	}
	if got, want := secret.Attrs.Mask(), "---h-----"; got != want {
		t.Errorf(".secret mask = %q, want %q", got, want)
	}

	public := findEntry(listing.Entries, "public.txt")
	if public == nil {
		t.Fatal("missing public.txt")
	}
	if public.Attrs.Hidden {
		t.Errorf("public.txt Hidden = true, want false")
	}
}

func TestScanDirSortModeApplied(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "b.txt"), []byte("22")) // size 2
	mustWrite(t, filepath.Join(dir, "a.txt"), []byte("1"))  // size 1

	listing, err := ScanDir(dir, nil, SortSizeDesc)
	if err != nil {
		t.Fatalf("ScanDir: %v", err)
	}
	want := []string{"b.txt", "a.txt"}
	if got := names(listing.Entries); !slices.Equal(got, want) {
		t.Errorf("entry order = %v, want %v", got, want)
	}
	if listing.SortMode != SortSizeDesc {
		t.Errorf("SortMode = %v, want %v", listing.SortMode, SortSizeDesc)
	}
}

func TestScanDirSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.txt")
	mustWrite(t, target, []byte("x"))

	link := filepath.Join(dir, "link.txt")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks not supported on this platform: %v", err)
	}

	listing, err := ScanDir(dir, nil, SortNameAsc)
	if err != nil {
		t.Fatalf("ScanDir: %v", err)
	}

	entry := findEntry(listing.Entries, "link.txt")
	if entry == nil {
		t.Fatal("missing link.txt")
	}
	if entry.Kind != KindSymlink {
		t.Errorf("link kind = %v, want symlink", entry.Kind)
	}
	if !entry.Attrs.Symlink {
		t.Errorf("link Symlink = false, want true")
	}
}

func TestScanDirMissing(t *testing.T) {
	dir := t.TempDir()
	_, err := ScanDir(filepath.Join(dir, "does-not-exist"), nil, SortNameAsc)
	if err == nil {
		t.Fatal("ScanDir on a missing directory returned nil error")
	}
}

func TestScanDirHiddenAndSystemEntriesListed(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, ".gitignore"), []byte("x"))
	mustWrite(t, filepath.Join(dir, "visible.txt"), []byte("y"))

	listing, err := ScanDir(dir, nil, SortNameAsc)
	if err != nil {
		t.Fatalf("ScanDir: %v", err)
	}
	if findEntry(listing.Entries, ".gitignore") == nil {
		t.Error("hidden dotfile was not listed")
	}
	if findEntry(listing.Entries, "visible.txt") == nil {
		t.Error("visible file was not listed")
	}
}
