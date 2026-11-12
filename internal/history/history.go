// Package history persists the directories the user has selected through JCD
// so the Quick Access popup can offer favorites, recent, and frequent paths.
//
// The package keeps a single collection and derives every view from it. It has
// no terminal dependencies and stores a small JSON file next to the config.
package history

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// MaxTracked is the maximum number of directories kept in the collection.
const MaxTracked = 100

// Limits for the derived Recent and Frequent sections.
const (
	RecentLimit   = 10
	FrequentLimit = 10
)

// TrackedDirectory is one persisted directory entry.
type TrackedDirectory struct {
	Path           string    `json:"path"`
	VisitCount     int       `json:"visit_count"`
	LastVisitedUtc time.Time `json:"last_visited_utc"`
	Favorite       bool      `json:"favorite"`
}

// fileData is the on-disk representation. The wrapper object leaves room for
// future fields without breaking existing files.
type fileData struct {
	Directories []TrackedDirectory `json:"directories"`
}

// Store owns the in-memory collection and its backing file.
type Store struct {
	path string
	dirs []TrackedDirectory
}

// New returns an empty store backed by path. An empty path makes Save a no-op.
func New(path string) *Store {
	return &Store{path: path}
}

// Path returns the backing file path.
func (s *Store) Path() string { return s.path }

// Len returns the number of tracked directories.
func (s *Store) Len() int { return len(s.dirs) }

// Load reads the collection from disk. A missing file yields an empty store.
func (s *Store) Load() error {
	s.dirs = nil
	if s.path == "" {
		return nil
	}

	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}
	if len(data) == 0 {
		return nil
	}

	var fd fileData
	if err := json.Unmarshal(data, &fd); err != nil {
		return err
	}
	s.dirs = fd.Directories
	return nil
}

// Save writes the collection to disk, creating its parent directory.
func (s *Store) Save() error {
	if s.path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}

	dirs := s.dirs
	if dirs == nil {
		dirs = []TrackedDirectory{}
	}
	data, err := json.MarshalIndent(fileData{Directories: dirs}, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(s.path, data, 0o644)
}

// Record registers a completed directory selection: it increments the visit
// count and refreshes the timestamp, or inserts a new entry. It then trims the
// collection to MaxTracked entries.
func (s *Store) Record(path string) {
	now := time.Now().UTC()
	key := pathKey(path)
	for i := range s.dirs {
		if pathKey(s.dirs[i].Path) == key {
			s.dirs[i].VisitCount++
			s.dirs[i].LastVisitedUtc = now
			s.prune()
			return
		}
	}

	s.dirs = append(s.dirs, TrackedDirectory{
		Path:           path,
		VisitCount:     1,
		LastVisitedUtc: now,
	})
	s.prune()
}

// ToggleFavorite flips the favorite flag for path and returns the new value.
// It returns false when the path is not tracked.
func (s *Store) ToggleFavorite(path string) bool {
	key := pathKey(path)
	for i := range s.dirs {
		if pathKey(s.dirs[i].Path) == key {
			s.dirs[i].Favorite = !s.dirs[i].Favorite
			return s.dirs[i].Favorite
		}
	}
	return false
}

// Replace sets the whole collection, copying dirs. It is useful for importing
// data and for deterministic tests.
func (s *Store) Replace(dirs []TrackedDirectory) {
	s.dirs = append([]TrackedDirectory(nil), dirs...)
}

// PruneMissing removes entries for which exists reports false.
func (s *Store) PruneMissing(exists func(string) bool) {
	if exists == nil {
		return
	}
	kept := s.dirs[:0]
	for _, d := range s.dirs {
		if exists(d.Path) {
			kept = append(kept, d)
		}
	}
	s.dirs = kept
}

// Sections derives the three popup views from the single collection. Favorites
// are ordered by recency. Recent and Frequent exclude favorites, so a directory
// appears in Favorites or in those sections, but never in both: Recent holds the
// most recently visited non-favorites and Frequent holds the most visited
// non-favorites.
func (s *Store) Sections() (favorites, recent, frequent []TrackedDirectory) {
	nonFavorites := make([]TrackedDirectory, 0, len(s.dirs))
	for _, d := range s.dirs {
		if d.Favorite {
			favorites = append(favorites, d)
		} else {
			nonFavorites = append(nonFavorites, d)
		}
	}
	sortNewestFirst(favorites)

	recent = append([]TrackedDirectory(nil), nonFavorites...)
	sortNewestFirst(recent)
	recent = limit(recent, RecentLimit)

	frequent = append([]TrackedDirectory(nil), nonFavorites...)
	sort.SliceStable(frequent, func(i, j int) bool {
		if frequent[i].VisitCount != frequent[j].VisitCount {
			return frequent[i].VisitCount > frequent[j].VisitCount
		}
		return frequent[i].LastVisitedUtc.After(frequent[j].LastVisitedUtc)
	})
	frequent = limit(frequent, FrequentLimit)

	return favorites, recent, frequent
}

func (s *Store) prune() {
	if len(s.dirs) <= MaxTracked {
		return
	}
	sort.SliceStable(s.dirs, func(i, j int) bool {
		return rankBefore(s.dirs[i], s.dirs[j])
	})
	s.dirs = s.dirs[:MaxTracked]
}

// rankBefore reports whether a should be kept before b when trimming the
// collection: favorites first, then more visits, then more recent.
func rankBefore(a, b TrackedDirectory) bool {
	if a.Favorite != b.Favorite {
		return a.Favorite
	}
	if a.VisitCount != b.VisitCount {
		return a.VisitCount > b.VisitCount
	}
	return a.LastVisitedUtc.After(b.LastVisitedUtc)
}

func sortNewestFirst(dirs []TrackedDirectory) {
	sort.SliceStable(dirs, func(i, j int) bool {
		return dirs[i].LastVisitedUtc.After(dirs[j].LastVisitedUtc)
	})
}

func limit(dirs []TrackedDirectory, n int) []TrackedDirectory {
	if len(dirs) <= n {
		return dirs
	}
	return dirs[:n]
}

// pathKey normalizes a path for comparison. Windows paths are compared
// case-insensitively because the filesystem is.
func pathKey(p string) string {
	clean := filepath.Clean(p)
	if filepath.Separator == '\\' {
		return strings.ToLower(clean)
	}
	return clean
}
