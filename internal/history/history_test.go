package history

import (
	"path/filepath"
	"testing"
	"time"
)

func TestRecordNewAndExisting(t *testing.T) {
	s := New("")
	s.Record(`C:\one`)
	s.Record(`C:\two`)
	s.Record(`C:\one`)

	if s.Len() != 2 {
		t.Fatalf("Len() = %d, want 2", s.Len())
	}

	var one TrackedDirectory
	for _, d := range s.dirs {
		if d.Path == `C:\one` {
			one = d
		}
	}
	if one.VisitCount != 2 {
		t.Errorf("visit count = %d, want 2", one.VisitCount)
	}
	if one.LastVisitedUtc.IsZero() {
		t.Error("last visited time is zero")
	}
}

func TestRecordDedupsCaseInsensitiveOnWindows(t *testing.T) {
	if filepath.Separator != '\\' {
		t.Skip("case-insensitive paths are a Windows behaviour")
	}
	s := New("")
	s.Record(`C:\Projects\JCD`)
	s.Record(`c:\projects\jcd`)

	if s.Len() != 1 {
		t.Fatalf("Len() = %d, want 1 (case-insensitive dedup)", s.Len())
	}
	if s.dirs[0].VisitCount != 2 {
		t.Errorf("visit count = %d, want 2", s.dirs[0].VisitCount)
	}
}

func TestToggleFavorite(t *testing.T) {
	s := New("")
	s.Record("/work")

	if !s.ToggleFavorite("/work") {
		t.Error("toggle did not set favorite")
	}
	if s.dirs[0].Favorite != true {
		t.Error("favorite = false after toggle, want true")
	}
	if s.ToggleFavorite("/work") {
		t.Error("toggle did not clear favorite")
	}
	if s.dirs[0].Favorite != false {
		t.Error("favorite = true after second toggle, want false")
	}
}

func TestToggleFavoriteUnknownPath(t *testing.T) {
	s := New("")
	if s.ToggleFavorite("/nope") {
		t.Error("toggle of unknown path returned true")
	}
}

func TestSections(t *testing.T) {
	s := New("")
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	s.dirs = []TrackedDirectory{
		{Path: "/old-fav", VisitCount: 1, LastVisitedUtc: base.Add(1 * time.Hour), Favorite: true},
		{Path: "/new", VisitCount: 1, LastVisitedUtc: base.Add(5 * time.Hour)},
		{Path: "/frequent", VisitCount: 9, LastVisitedUtc: base.Add(2 * time.Hour)},
		{Path: "/middle", VisitCount: 3, LastVisitedUtc: base.Add(4 * time.Hour)},
		{Path: "/new-fav", VisitCount: 2, LastVisitedUtc: base.Add(3 * time.Hour), Favorite: true},
	}

	favorites, recent, frequent := s.Sections()

	if len(favorites) != 2 {
		t.Fatalf("favorites = %d, want 2", len(favorites))
	}
	if favorites[0].Path != "/new-fav" || favorites[1].Path != "/old-fav" {
		t.Errorf("favorites order = %v, want most recent first", paths(favorites))
	}

	if len(recent) != 5 {
		t.Fatalf("recent = %d, want 5", len(recent))
	}
	if recent[0].Path != "/new" {
		t.Errorf("recent[0] = %q, want /new", recent[0].Path)
	}

	if len(frequent) != 5 {
		t.Fatalf("frequent = %d, want 5", len(frequent))
	}
	if frequent[0].Path != "/frequent" {
		t.Errorf("frequent[0] = %q, want /frequent", frequent[0].Path)
	}
}

func TestSectionsLimits(t *testing.T) {
	s := New("")
	for i := 0; i < 30; i++ {
		s.Record(filepath.Join("/dir", string(rune('a'+i))))
	}

	_, recent, frequent := s.Sections()
	if len(recent) > RecentLimit {
		t.Errorf("recent = %d, want <= %d", len(recent), RecentLimit)
	}
	if len(frequent) > FrequentLimit {
		t.Errorf("frequent = %d, want <= %d", len(frequent), FrequentLimit)
	}
}

func TestPruneKeepsFavoritesAndFrequent(t *testing.T) {
	s := New("")
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	// Fill past the limit with ordinary, older entries.
	for i := 0; i < MaxTracked; i++ {
		s.dirs = append(s.dirs, TrackedDirectory{
			Path:           filepath.Join("/filler", string(rune('a'+i%26))+string(rune('a'+i/26))),
			VisitCount:     1,
			LastVisitedUtc: base.Add(time.Duration(i) * time.Minute),
		})
	}
	// A favorite is always kept, even with a single visit.
	s.dirs = append(s.dirs, TrackedDirectory{Path: "/favorite", VisitCount: 1, LastVisitedUtc: base, Favorite: true})
	// Trigger a prune by recording a new path.
	s.Record("/newest")

	if s.Len() > MaxTracked {
		t.Fatalf("Len() = %d, want <= %d", s.Len(), MaxTracked)
	}

	found := false
	for _, d := range s.dirs {
		if d.Path == "/favorite" {
			found = true
		}
	}
	if !found {
		t.Error("favorite was evicted by pruning")
	}
}

func TestPruneKeepsMostVisited(t *testing.T) {
	s := New("")
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	s.dirs = append(s.dirs, TrackedDirectory{Path: "/popular", VisitCount: 50, LastVisitedUtc: base})

	for i := 0; i < MaxTracked; i++ {
		s.dirs = append(s.dirs, TrackedDirectory{
			Path:           filepath.Join("/filler", string(rune('a'+i%26))+string(rune('a'+i/26))),
			VisitCount:     1,
			LastVisitedUtc: base.Add(time.Duration(i) * time.Minute),
		})
	}
	s.Record("/new")

	found := false
	for _, d := range s.dirs {
		if d.Path == "/popular" {
			found = true
		}
	}
	if !found {
		t.Error("frequently visited directory was evicted")
	}
}

func TestPruneMissing(t *testing.T) {
	s := New("")
	s.Record("/keep")
	s.Record("/drop")

	s.PruneMissing(func(p string) bool { return p == "/keep" })

	if s.Len() != 1 || s.dirs[0].Path != "/keep" {
		t.Errorf("after prune: %v, want [/keep]", paths(s.dirs))
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tracked.json")
	s := New(path)
	s.Record(`C:\projects\jcd`)
	s.ToggleFavorite(`C:\projects\jcd`)
	if err := s.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded := New(path)
	if err := loaded.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Len() != 1 {
		t.Fatalf("Len() = %d, want 1", loaded.Len())
	}
	d := loaded.dirs[0]
	if d.Path != `C:\projects\jcd` || d.VisitCount != 1 || !d.Favorite {
		t.Errorf("round trip = %+v", d)
	}
	if d.LastVisitedUtc.IsZero() {
		t.Error("last visited time is zero after round trip")
	}
}

func TestLoadMissingAndEmptyPath(t *testing.T) {
	if err := New(filepath.Join(t.TempDir(), "missing.json")).Load(); err != nil {
		t.Errorf("Load(missing): %v", err)
	}
	if err := New("").Load(); err != nil {
		t.Errorf("Load(empty): %v", err)
	}
	if err := New("").Save(); err != nil {
		t.Errorf("Save(empty): %v", err)
	}
}

func paths(dirs []TrackedDirectory) []string {
	out := make([]string, len(dirs))
	for i, d := range dirs {
		out[i] = d.Path
	}
	return out
}
