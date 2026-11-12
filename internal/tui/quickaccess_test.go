package tui

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"jcd-go/internal/history"
)

// newQuickModel builds a model with three real directories and a deterministic
// tracked-directory store:
//
//	alpha  favorite, last visited 1h ago, 3 visits
//	beta   not favorite, last visited 5h ago, 1 visit
//	gamma  not favorite, last visited 2h ago, 5 visits
func newQuickModel(t *testing.T) (Model, []string) {
	t.Helper()
	root := t.TempDir()
	paths := make([]string, 0, 3)
	for _, name := range []string{"alpha", "beta", "gamma"} {
		p := filepath.Join(root, name)
		mustMkdir(t, p)
		paths = append(paths, p)
	}

	m, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	m.history.Replace([]history.TrackedDirectory{
		{Path: paths[0], VisitCount: 3, LastVisitedUtc: base.Add(1 * time.Hour), Favorite: true},
		{Path: paths[1], VisitCount: 1, LastVisitedUtc: base.Add(5 * time.Hour)},
		{Path: paths[2], VisitCount: 5, LastVisitedUtc: base.Add(2 * time.Hour)},
	})
	return m, paths
}

func TestQuickAccessOpensOnF4(t *testing.T) {
	m, paths := newQuickModel(t)

	m = press(t, m, key(tea.KeyF4))
	if m.State() != StateQuickAccess {
		t.Fatalf("state = %v, want quick access", m.State())
	}
	if len(m.quickRows) == 0 {
		t.Fatal("quick rows are empty")
	}
	if m.quickCursor < 0 || m.quickRows[m.quickCursor].section {
		t.Fatalf("cursor is not on an entry: %d", m.quickCursor)
	}
	if got := m.quickRows[m.quickCursor].path; got != paths[0] {
		t.Errorf("initial selection = %q, want favorite %q", got, paths[0])
	}
}

func TestQuickAccessNavigationSkipsHeaders(t *testing.T) {
	m, _ := newQuickModel(t)
	m = press(t, m, key(tea.KeyF4))

	seen := map[string]bool{}
	for i := 0; i < 3*len(m.quickRows); i++ {
		if m.quickCursor < 0 || m.quickRows[m.quickCursor].section {
			t.Fatalf("cursor %d is not on an entry", m.quickCursor)
		}
		seen[m.quickRows[m.quickCursor].path] = true
		m = press(t, m, key(tea.KeyDown))
	}
	if len(seen) < 2 {
		t.Errorf("navigation only reached %d entries", len(seen))
	}
}

func TestQuickAccessEnterJumpsAndQuits(t *testing.T) {
	m, paths := newQuickModel(t)
	m = press(t, m, key(tea.KeyF4))

	next, cmd := m.Update(key(tea.KeyEnter))
	got := next.(Model)

	if !got.Quitting() {
		t.Error("Enter did not quit")
	}
	if got.ChosenDir() != paths[0] {
		t.Errorf("ChosenDir() = %q, want %q", got.ChosenDir(), paths[0])
	}
	if cmd == nil {
		t.Fatal("expected a quit command")
	}
}

func TestQuickAccessEscClosesWithoutQuitting(t *testing.T) {
	m, _ := newQuickModel(t)
	m = press(t, m, key(tea.KeyF4))

	m = press(t, m, key(tea.KeyEscape))
	if m.State() != StateBrowsing {
		t.Errorf("state = %v, want browsing", m.State())
	}
	if m.Quitting() {
		t.Error("Esc in Quick Access quit the app")
	}
}

func TestQuickAccessToggleFavoriteRefreshesSections(t *testing.T) {
	m, paths := newQuickModel(t)
	m = press(t, m, key(tea.KeyF4))
	m = press(t, m, key(tea.KeyDown)) // beta, most recent non-favorite

	if got := m.quickRows[m.quickCursor].path; got != paths[1] {
		t.Fatalf("selection = %q, want beta %q", got, paths[1])
	}

	// Favoriting adds it to the Favorites section, removes it from the other
	// sections, and keeps the cursor on it.
	m = press(t, m, runeKey('f'))
	favorites, recent, frequent := m.history.Sections()
	if !containsPath(favorites, paths[1]) {
		t.Error("beta was not added to favorites")
	}
	if containsPath(recent, paths[1]) || containsPath(frequent, paths[1]) {
		t.Error("beta still appears in recent/frequent after favoriting")
	}
	if got := m.quickRows[m.quickCursor].path; got != paths[1] {
		t.Errorf("cursor moved to %q after favoriting", got)
	}
	if !m.quickRows[m.quickCursor].favorite {
		t.Error("selected row is not marked as favorite")
	}

	// Un-favoriting removes it from Favorites and restores it to the other
	// sections where it qualifies.
	m = press(t, m, runeKey('f'))
	favorites, recent, frequent = m.history.Sections()
	if containsPath(favorites, paths[1]) {
		t.Error("beta is still a favorite after toggling off")
	}
	if !containsPath(recent, paths[1]) || !containsPath(frequent, paths[1]) {
		t.Error("beta did not return to recent and frequent after un-favoriting")
	}
	if got := m.quickRows[m.quickCursor].path; got != paths[1] {
		t.Errorf("cursor moved to %q after un-favoriting", got)
	}
}

func TestQuickAccessRendersSections(t *testing.T) {
	m, _ := newQuickModel(t)
	next, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m = next.(Model)
	m = press(t, m, key(tea.KeyF4))

	out := m.render()
	for _, want := range []string{"Quick Access", "Favorites", "Recent", "Frequent", "Enter"} {
		if !strings.Contains(out, want) {
			t.Errorf("render missing %q:\n%s", want, out)
		}
	}
}

func TestQuickAccessSectionHeaderFillsWidth(t *testing.T) {
	m, _ := newQuickModel(t)

	for _, title := range []string{"★ Favorites", "🕒 Recent", "🔥 Frequent"} {
		line := m.renderQuickSection(title, 40)
		if w := lipgloss.Width(line); w != 40 {
			t.Errorf("%q section width = %d, want 40", title, w)
		}
		wantPrefix := "- " + title + " "
		if plain := ansi.Strip(line); !strings.HasPrefix(plain, wantPrefix) {
			t.Errorf("%q section = %q, want prefix %q", title, plain, wantPrefix)
		}
		// The leading dash uses the same dim style as the trailing rule.
		if want := m.theme.Meta.Render("- "); !strings.HasPrefix(line, want) {
			t.Errorf("%q leading dash is not dim-styled", title)
		}
	}
}

func TestQuickAccessRenderFitsScreen(t *testing.T) {
	m, _ := newQuickModel(t)
	next, _ := m.Update(tea.WindowSizeMsg{Width: 90, Height: 28})
	m = next.(Model)
	m = press(t, m, key(tea.KeyF4))

	out := m.render()
	lines := strings.Split(out, "\n")
	if len(lines) != 28 {
		t.Fatalf("render height = %d lines, want 28", len(lines))
	}
	for i, line := range lines {
		if w := lipgloss.Width(line); w != 90 {
			t.Errorf("line %d width = %d, want 90", i, w)
		}
	}
}

func containsPath(dirs []history.TrackedDirectory, path string) bool {
	for _, d := range dirs {
		if d.Path == path {
			return true
		}
	}
	return false
}

func TestQuickAccessShowsListingBehind(t *testing.T) {
	m, _ := newQuickModel(t)
	next, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	m = next.(Model)
	m = press(t, m, key(tea.KeyF4))

	out := m.render()
	// "directories," only appears in the browser's info row, never in the popup.
	if !strings.Contains(out, "directories,") {
		t.Errorf("listing is not visible behind the popup:\n%s", out)
	}
	if !strings.Contains(out, "Quick Access") {
		t.Error("popup is missing")
	}
}

func TestOverlayBlock(t *testing.T) {
	base := []string{
		"0123456789",
		"abcdefghij",
		"klmnopqrst",
	}
	block := []string{"XX"}

	got := overlayBlock(base, block, 3, 1)
	want := []string{
		"0123456789",
		"abcXXfghij",
		"klmnopqrst",
	}
	if len(got) != len(want) {
		t.Fatalf("overlay returned %d lines, want %d", len(got), len(want))
	}
	for i := range want {
		if plain := ansi.Strip(got[i]); plain != want[i] {
			t.Errorf("line %d = %q, want %q", i, plain, want[i])
		}
	}
}

func TestSpliceLinePreservesWidthAndText(t *testing.T) {
	base := lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render("0123456789")
	overlay := lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true).Render("XX")

	got := spliceLine(base, overlay, 3)
	if w := lipgloss.Width(got); w != 10 {
		t.Errorf("width = %d, want 10", w)
	}
	if plain := ansi.Strip(got); plain != "012XX56789" {
		t.Errorf("plain text = %q, want %q", plain, "012XX56789")
	}
}

func TestSpliceLineEdgeCases(t *testing.T) {
	base := "abcdefghij"

	if got := ansi.Strip(spliceLine(base, "XX", 0)); got != "XXcdefghij" {
		t.Errorf("left edge = %q, want %q", got, "XXcdefghij")
	}
	if got := ansi.Strip(spliceLine(base, "XX", 8)); got != "abcdefghXX" {
		t.Errorf("right edge = %q, want %q", got, "abcdefghXX")
	}
	if got := spliceLine(base, "XX", 20); got != base {
		t.Errorf("out of range = %q, want %q", got, base)
	}
}
