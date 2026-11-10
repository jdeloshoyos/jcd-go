package fsentry

import (
	"slices"
	"testing"
	"time"
)

func date(day int) time.Time {
	return time.Date(2021, time.January, day, 0, 0, 0, 0, time.UTC)
}

// sampleEntries returns a deliberately unsorted mix of files and directories.
func sampleEntries() []Entry {
	return []Entry{
		{Name: "beta.txt", Kind: KindFile, Size: 300, ModTime: date(2)},
		{Name: "Alpha.txt", Kind: KindFile, Size: 100, ModTime: date(3)},
		{Name: "gamma", Kind: KindDir, Size: 0, ModTime: date(1)},
		{Name: "delta", Kind: KindDir, Size: 0, ModTime: date(4)},
		{Name: "alpha2.txt", Kind: KindFile, Size: 200, ModTime: date(1)},
	}
}

func names(entries []Entry) []string {
	out := make([]string, len(entries))
	for i, e := range entries {
		out[i] = e.Name
	}
	return out
}

func TestSortModes(t *testing.T) {
	tests := []struct {
		mode SortMode
		want []string
	}{
		{
			SortNameAsc,
			[]string{"delta", "gamma", "Alpha.txt", "alpha2.txt", "beta.txt"},
		},
		{
			SortDateAsc,
			[]string{"gamma", "delta", "alpha2.txt", "beta.txt", "Alpha.txt"},
		},
		{
			SortSizeAsc,
			[]string{"delta", "gamma", "Alpha.txt", "alpha2.txt", "beta.txt"},
		},
		{
			SortNameDesc,
			[]string{"gamma", "delta", "beta.txt", "alpha2.txt", "Alpha.txt"},
		},
		{
			SortDateDesc,
			[]string{"delta", "gamma", "Alpha.txt", "beta.txt", "alpha2.txt"},
		},
		{
			SortSizeDesc,
			[]string{"delta", "gamma", "beta.txt", "alpha2.txt", "Alpha.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.mode.String(), func(t *testing.T) {
			entries := sampleEntries()
			Sort(entries, tt.mode)
			if got := names(entries); !slices.Equal(got, tt.want) {
				t.Errorf("Sort(%v) = %v, want %v", tt.mode, got, tt.want)
			}
		})
	}
}

func TestSortDirectoriesAlwaysFirst(t *testing.T) {
	entries := []Entry{
		{Name: "zzz-file", Kind: KindFile, Size: 1},
		{Name: "aaa-dir", Kind: KindDir},
		{Name: "mmm-file", Kind: KindFile, Size: 2},
		{Name: "bbb-dir", Kind: KindDir},
	}
	for _, mode := range []SortMode{
		SortNameAsc, SortDateAsc, SortSizeAsc,
		SortNameDesc, SortDateDesc, SortSizeDesc,
	} {
		mode := mode
		t.Run(mode.String(), func(t *testing.T) {
			got := slices.Clone(entries)
			Sort(got, mode)
			if !got[0].IsDirLike() || !got[1].IsDirLike() {
				t.Fatalf("%v: directories not first: %v", mode, names(got))
			}
			if got[2].IsDirLike() || got[3].IsDirLike() {
				t.Fatalf("%v: files not last: %v", mode, names(got))
			}
		})
	}
}

func TestSortDrivesCountAsDirectories(t *testing.T) {
	entries := []Entry{
		{Name: "a-file", Kind: KindFile},
		{Name: "C:", Kind: KindDrive},
	}
	Sort(entries, SortNameAsc)
	if entries[0].Kind != KindDrive {
		t.Fatalf("drive should sort before file, got %v", names(entries))
	}
}

func TestSortNameCaseInsensitive(t *testing.T) {
	entries := []Entry{
		{Name: "banana", Kind: KindFile},
		{Name: "Apple", Kind: KindFile},
		{Name: "cherry", Kind: KindFile},
	}
	Sort(entries, SortNameAsc)
	want := []string{"Apple", "banana", "cherry"}
	if got := names(entries); !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestSortModeNext(t *testing.T) {
	sequence := []SortMode{
		SortNameAsc, SortDateAsc, SortSizeAsc,
		SortNameDesc, SortDateDesc, SortSizeDesc, SortNameAsc,
	}
	for i := 0; i < len(sequence)-1; i++ {
		if got := sequence[i].Next(); got != sequence[i+1] {
			t.Errorf("%v.Next() = %v, want %v", sequence[i], got, sequence[i+1])
		}
	}
	if got := SortMode(0).Next(); got != SortNameAsc {
		t.Errorf("invalid Next() = %v, want %v", got, SortNameAsc)
	}
}

func TestSortModeNormalize(t *testing.T) {
	tests := map[SortMode]SortMode{
		0:            SortNameAsc,
		SortNameAsc:  SortNameAsc,
		SortSizeAsc:  SortSizeAsc,
		SortSizeDesc: SortSizeDesc,
		7:            SortNameAsc,
	}
	for in, want := range tests {
		if got := in.Normalize(); got != want {
			t.Errorf("%v.Normalize() = %v, want %v", in, got, want)
		}
	}
}
