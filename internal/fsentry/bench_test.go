package fsentry

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

func benchmarkEntries(n int) []Entry {
	entries := make([]Entry, n)
	base := time.Unix(1600000000, 0)
	for i := range entries {
		kind := KindFile
		if i%10 == 0 {
			kind = KindDir
		}
		entries[i] = Entry{
			Name:    fmt.Sprintf("entry-%05d-%c", i, 'a'+rune(i%26)),
			Kind:    kind,
			Size:    int64((i * 37) % 100000),
			ModTime: base.Add(time.Duration(i) * time.Minute),
		}
	}
	return entries
}

func BenchmarkSort(b *testing.B) {
	base := benchmarkEntries(10000)
	modes := []SortMode{
		SortNameAsc, SortDateAsc, SortSizeAsc,
		SortNameDesc, SortDateDesc, SortSizeDesc,
	}

	for _, mode := range modes {
		b.Run(mode.String(), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				entries := slices.Clone(base)
				Sort(entries, mode)
			}
		})
	}
}

func BenchmarkMatches(b *testing.B) {
	entries := benchmarkEntries(10000)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		Matches(entries, "entry-099")
	}
}

func BenchmarkScanDir(b *testing.B) {
	dir := b.TempDir()
	for i := 0; i < 1000; i++ {
		name := filepath.Join(dir, fmt.Sprintf("file-%04d.txt", i))
		if err := os.WriteFile(name, []byte("x"), 0o644); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := ScanDir(dir, nil, SortNameAsc); err != nil {
			b.Fatal(err)
		}
	}
}
