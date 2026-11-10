package fsentry

import (
	"os"
	"path/filepath"
)

// AttrFunc derives the portable attribute model for a path and its FileInfo.
// Platform packages provide the real implementation; tests can inject a fake.
type AttrFunc func(path string, info os.FileInfo) Attributes

// Listing is the result of scanning a directory.
type Listing struct {
	Dir        string
	Entries    []Entry
	SortMode   SortMode
	NumDirs    int
	TotalBytes int64
}

// NumFiles returns the number of non-directory entries.
func (l *Listing) NumFiles() int {
	n := len(l.Entries) - l.NumDirs
	if n < 0 {
		return 0
	}
	return n
}

// ScanDir reads dir, builds entries for every child (including hidden items),
// computes summary counts, and sorts the result using mode.
//
// attr may be nil, in which case entries carry zero-value attributes. The
// Directory and Symlink flags are always kept consistent with the entry kind.
func ScanDir(dir string, attr AttrFunc, mode SortMode) (*Listing, error) {
	dirents, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	listing := &Listing{
		Dir:      dir,
		Entries:  make([]Entry, 0, len(dirents)),
		SortMode: mode.Normalize(),
	}

	for _, de := range dirents {
		info, err := de.Info()
		if err != nil {
			// The entry disappeared or is unreadable. Skipping it is preferable
			// to failing the entire listing.
			continue
		}

		entry := Entry{
			Name:    de.Name(),
			Path:    filepath.Join(dir, de.Name()),
			Kind:    kindOf(info),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		}
		if attr != nil {
			entry.Attrs = attr(entry.Path, info)
		}
		// Keep the kind and mask consistent even with a zero or partial
		// attribute provider.
		if entry.Kind == KindDir {
			entry.Attrs.Directory = true
		}
		if entry.Kind == KindSymlink {
			entry.Attrs.Symlink = true
		}

		listing.Entries = append(listing.Entries, entry)
		if entry.IsDirLike() {
			listing.NumDirs++
		}
		// Only regular files contribute to "used" size, keeping the value
		// stable across platforms (directory sizes differ by filesystem).
		if entry.Kind == KindFile {
			listing.TotalBytes += entry.Size
		}
	}

	Sort(listing.Entries, listing.SortMode)
	return listing, nil
}

func kindOf(info os.FileInfo) Kind {
	switch {
	case info.IsDir():
		return KindDir
	case info.Mode()&os.ModeSymlink != 0:
		return KindSymlink
	case info.Mode().IsRegular():
		return KindFile
	default:
		return KindOther
	}
}
