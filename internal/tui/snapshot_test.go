package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"jcd-go/internal/fsentry"
	"jcd-go/internal/history"
)

var snapshotTime = time.Date(2021, time.January, 2, 15, 4, 5, 0, time.UTC)

// snapshotModel builds a deterministic model with a fixed path and fixed
// entries, independent of the host filesystem.
func snapshotModel() Model {
	entries := []fsentry.Entry{
		{Name: "docs", Kind: fsentry.KindDir, ModTime: snapshotTime, Attrs: fsentry.Attributes{Directory: true}},
		{Name: "src", Kind: fsentry.KindDir, ModTime: snapshotTime, Attrs: fsentry.Attributes{Directory: true}},
		{Name: "latest", Kind: fsentry.KindSymlink, ModTime: snapshotTime, Attrs: fsentry.Attributes{Symlink: true}},
		{Name: "main.go", Kind: fsentry.KindFile, Size: 512, ModTime: snapshotTime, Attrs: fsentry.Attributes{Archive: true}},
		{Name: "README.md", Kind: fsentry.KindFile, Size: 2048, ModTime: snapshotTime, Attrs: fsentry.Attributes{Archive: true}},
	}
	listing := &fsentry.Listing{
		Dir:        "/demo/project",
		Entries:    entries,
		SortMode:   fsentry.SortNameAsc,
		NumDirs:    2,
		TotalBytes: 2560,
	}
	return Model{
		dir:       "/demo/project",
		listing:   listing,
		sortMode:  fsentry.SortNameAsc,
		state:     StateBrowsing,
		width:     80,
		height:    24,
		search:    newSearchInput(),
		help:      viewport.New(),
		history:   history.New(""),
		theme:     DefaultTheme(),
		dirExists: func(string) bool { return true },
		openFile:  func(string) error { return nil },
	}
}

func TestSnapshots(t *testing.T) {
	tests := []struct {
		name  string
		setup func(m *Model)
	}{
		{"browsing", func(*Model) {}},
		{"browsing_selected_file", func(m *Model) { m.cursor = 3 }},
		{"search", func(m *Model) {
			m.startSearch()
			m.search.SetValue("ma")
			m.followFirstMatch()
		}},
		{"help", func(m *Model) { m.startHelp(StateBrowsing) }},
		{"drive_select", func(m *Model) {
			m.state = StateDriveSelect
			m.listing = &fsentry.Listing{
				Entries: []fsentry.Entry{
					{Name: "C:", Path: `C:\`, Kind: fsentry.KindDrive, DriveType: fsentry.DriveFixed},
					{Name: "D:", Path: `D:\`, Kind: fsentry.KindDrive, DriveType: fsentry.DriveCDROM},
				},
				NumDirs:  2,
				SortMode: fsentry.SortNameAsc,
			}
			m.cursor = 0
		}},
		{"too_small", func(m *Model) {
			m.width = 40
			m.height = 10
		}},
		{"quick_access", func(m *Model) {
			base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
			m.history.Replace([]history.TrackedDirectory{
				{Path: `C:\projects\jcd`, VisitCount: 4, LastVisitedUtc: base.Add(4 * time.Hour), Favorite: true},
				{Path: `C:\users\jaime\documents\projects\doccompose\backend`, VisitCount: 1, LastVisitedUtc: base.Add(3 * time.Hour)},
				{Path: `D:\temp`, VisitCount: 7, LastVisitedUtc: base.Add(2 * time.Hour)},
				{Path: `C:\repos\client-x`, VisitCount: 2, LastVisitedUtc: base.Add(1 * time.Hour)},
			})
			m.openQuickAccess()
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := snapshotModel()
			tt.setup(&m)
			assertGolden(t, tt.name, normalizeSnapshot(m.render()))
		})
	}
}

func TestHelpTextFitsMinimumWidth(t *testing.T) {
	// The minimum terminal width is 80; the help viewport is 76 columns.
	for i, line := range strings.Split(helpText(), "\n") {
		if w := lipgloss.Width(line); w > 72 {
			t.Errorf("help line %d is %d columns wide, want <= 72: %q", i, w, line)
		}
	}
}

func normalizeSnapshot(s string) string {
	s = ansi.Strip(s)
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " ")
	}
	return strings.Join(lines, "\n")
}

func assertGolden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name+".golden")

	if os.Getenv("UPDATE_GOLDEN") != "" {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatalf("create testdata: %v", err)
		}
		if err := os.WriteFile(path, []byte(got+"\n"), 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		return
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v (run with UPDATE_GOLDEN=1 to create it)", path, err)
	}
	want := strings.ReplaceAll(string(data), "\r\n", "\n")
	if got+"\n" != want {
		t.Errorf("snapshot %q mismatch:\n--- got ---\n%s\n--- want ---\n%s", name, got, want)
	}
}
