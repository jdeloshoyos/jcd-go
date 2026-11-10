package tui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"jcd-go/internal/fsentry"
)

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.Mkdir(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func mustWrite(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// newTestModel builds a model over alpha/, beta/, and charlie.txt, sorted by
// name (directories first).
func newTestModel(t *testing.T) (Model, string) {
	t.Helper()
	dir := t.TempDir()
	mustMkdir(t, filepath.Join(dir, "alpha"))
	mustMkdir(t, filepath.Join(dir, "beta"))
	mustWrite(t, filepath.Join(dir, "charlie.txt"), []byte("12345"))

	m, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return m, dir
}

func key(code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: code}
}

func runeKey(r rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: r, Text: string(r)}
}

func ctrlKey(r rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: r, Mod: tea.ModCtrl}
}

func press(t *testing.T, m Model, keys ...tea.KeyPressMsg) Model {
	t.Helper()
	for _, k := range keys {
		next, _ := m.Update(k)
		updated, ok := next.(Model)
		if !ok {
			t.Fatalf("Update returned %T, want tui.Model", next)
		}
		m = updated
	}
	return m
}

func TestNew(t *testing.T) {
	m, dir := newTestModel(t)

	if m.Dir() != dir {
		t.Errorf("Dir() = %q, want %q", m.Dir(), dir)
	}
	if m.Cursor() != 0 {
		t.Errorf("Cursor() = %d, want 0", m.Cursor())
	}
	if m.SortMode() != fsentry.SortNameAsc {
		t.Errorf("SortMode() = %v, want %v", m.SortMode(), fsentry.SortNameAsc)
	}

	l := m.Listing()
	if l == nil {
		t.Fatal("Listing() is nil")
	}
	if len(l.Entries) != 3 {
		t.Errorf("entries = %d, want 3", len(l.Entries))
	}
	if l.NumDirs != 2 {
		t.Errorf("NumDirs = %d, want 2", l.NumDirs)
	}
	if l.NumFiles() != 1 {
		t.Errorf("NumFiles() = %d, want 1", l.NumFiles())
	}
}

func TestNavigation(t *testing.T) {
	m, _ := newTestModel(t)

	m = press(t, m, key(tea.KeyDown))
	if m.Cursor() != 1 {
		t.Errorf("after down: cursor = %d, want 1", m.Cursor())
	}

	m = press(t, m, key(tea.KeyDown), key(tea.KeyDown))
	if m.Cursor() != 2 {
		t.Errorf("after two downs: cursor = %d, want 2 (clamped)", m.Cursor())
	}

	m = press(t, m, key(tea.KeyUp))
	if m.Cursor() != 1 {
		t.Errorf("after up: cursor = %d, want 1", m.Cursor())
	}

	m = press(t, m, key(tea.KeyEnd))
	if m.Cursor() != 2 {
		t.Errorf("after end: cursor = %d, want 2", m.Cursor())
	}

	m = press(t, m, key(tea.KeyHome))
	if m.Cursor() != 0 {
		t.Errorf("after home: cursor = %d, want 0", m.Cursor())
	}
}

func TestEnterDirectoryAndUp(t *testing.T) {
	m, dir := newTestModel(t)

	m = press(t, m, key(tea.KeyEnter)) // alpha
	want := filepath.Join(dir, "alpha")
	if m.Dir() != want {
		t.Fatalf("after enter: dir = %q, want %q", m.Dir(), want)
	}
	if len(m.Listing().Entries) != 0 {
		t.Errorf("alpha should be empty, got %d entries", len(m.Listing().Entries))
	}

	m = press(t, m, key(tea.KeyBackspace))
	if m.Dir() != dir {
		t.Errorf("after backspace: dir = %q, want %q", m.Dir(), dir)
	}
	if m.Cursor() != 0 {
		t.Errorf("after backspace: cursor = %d, want 0", m.Cursor())
	}
}

func TestEnterFileOpens(t *testing.T) {
	m, dir := newTestModel(t)

	var opened string
	m.openFile = func(path string) error {
		opened = path
		return nil
	}

	m = press(t, m, key(tea.KeyEnd)) // charlie.txt
	m = press(t, m, key(tea.KeyEnter))

	want := filepath.Join(dir, "charlie.txt")
	if opened != want {
		t.Errorf("opened %q, want %q", opened, want)
	}
	if m.Dir() != dir {
		t.Errorf("opening a file changed the directory to %q", m.Dir())
	}
}

func TestEscQuits(t *testing.T) {
	m, dir := newTestModel(t)

	next, cmd := m.Update(key(tea.KeyEscape))
	m = next.(Model)

	if !m.Quitting() {
		t.Error("Quitting() = false, want true")
	}
	if m.ChosenDir() != dir {
		t.Errorf("ChosenDir() = %q, want %q", m.ChosenDir(), dir)
	}
	if cmd == nil {
		t.Fatal("expected a quit command")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Errorf("command produced %T, want tea.QuitMsg", cmd())
	}
}

func TestCtrlCQuits(t *testing.T) {
	m, dir := newTestModel(t)

	next, cmd := m.Update(ctrlKey('c'))
	m = next.(Model)

	if !m.Quitting() || m.ChosenDir() != dir {
		t.Errorf("ctrl+c did not quit cleanly: quitting=%v chosen=%q", m.Quitting(), m.ChosenDir())
	}
	if cmd == nil {
		t.Error("expected a quit command")
	}
}

func TestChosenDirAfterNavigation(t *testing.T) {
	m, dir := newTestModel(t)

	m = press(t, m, key(tea.KeyEnter)) // alpha
	next, _ := m.Update(key(tea.KeyEscape))
	m = next.(Model)

	if want := filepath.Join(dir, "alpha"); m.ChosenDir() != want {
		t.Errorf("ChosenDir() = %q, want %q", m.ChosenDir(), want)
	}
}

func TestSearchJumpAndCancel(t *testing.T) {
	m, _ := newTestModel(t)

	m = press(t, m, runeKey('/'))
	if m.State() != StateSearching {
		t.Fatalf("state = %v, want searching", m.State())
	}

	m = press(t, m, runeKey('c'), runeKey('h'), runeKey('a'), runeKey('r'))
	if got := m.search.Value(); got != "char" {
		t.Fatalf("query = %q, want %q", got, "char")
	}

	m = press(t, m, key(tea.KeyEnter))
	if m.Cursor() != 2 {
		t.Errorf("after jump: cursor = %d, want 2 (charlie.txt)", m.Cursor())
	}

	m = press(t, m, key(tea.KeyEscape))
	if m.State() != StateBrowsing {
		t.Errorf("state = %v, want browsing", m.State())
	}
	if m.search.Value() != "" {
		t.Errorf("query = %q, want empty after cancel", m.search.Value())
	}
}

func TestSearchTabCyclesMatches(t *testing.T) {
	m, _ := newTestModel(t)

	// "a" matches alpha, beta and charlie.
	m = press(t, m, runeKey('/'), runeKey('a'))
	m = press(t, m, key(tea.KeyEnter))
	if m.Cursor() != 0 {
		t.Fatalf("first match: cursor = %d, want 0", m.Cursor())
	}

	m = press(t, m, key(tea.KeyTab))
	if m.Cursor() != 1 {
		t.Errorf("second match: cursor = %d, want 1", m.Cursor())
	}

	m = press(t, m, key(tea.KeyTab))
	if m.Cursor() != 2 {
		t.Errorf("third match: cursor = %d, want 2", m.Cursor())
	}

	m = press(t, m, key(tea.KeyTab))
	if m.Cursor() != 0 {
		t.Errorf("wrapped match: cursor = %d, want 0", m.Cursor())
	}
}

func TestSearchNoMatch(t *testing.T) {
	m, _ := newTestModel(t)
	before := m.Cursor()

	m = press(t, m, runeKey('/'), runeKey('z'), runeKey('z'), key(tea.KeyEnter))

	if m.Cursor() != before {
		t.Errorf("cursor changed to %d on no match, want %d", m.Cursor(), before)
	}
	if m.status != "no matches" {
		t.Errorf("status = %q, want %q", m.status, "no matches")
	}
}

func TestSortCycle(t *testing.T) {
	m, _ := newTestModel(t)

	m = press(t, m, key(tea.KeyF2))

	if m.SortMode() != fsentry.SortDateAsc {
		t.Errorf("SortMode() = %v, want %v", m.SortMode(), fsentry.SortDateAsc)
	}
	if m.Listing().SortMode != fsentry.SortDateAsc {
		t.Errorf("listing sort mode = %v, want %v", m.Listing().SortMode, fsentry.SortDateAsc)
	}
}

func TestWindowSize(t *testing.T) {
	m, _ := newTestModel(t)

	next, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = next.(Model)

	if m.width != 120 || m.height != 40 {
		t.Errorf("size = %dx%d, want 120x40", m.width, m.height)
	}
	if got := m.visibleRows(); got != 37 {
		t.Errorf("visibleRows() = %d, want 37", got)
	}
}

func TestPageDown(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 10; i++ {
		mustWrite(t, filepath.Join(dir, fmt.Sprintf("f%02d.txt", i)), []byte("x"))
	}
	m, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	next, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 5}) // 2 visible rows
	m = next.(Model)

	m = press(t, m, key(tea.KeyPgDown))
	if m.Cursor() != 2 {
		t.Errorf("cursor = %d, want 2", m.Cursor())
	}
	if m.offset != 1 {
		t.Errorf("offset = %d, want 1", m.offset)
	}
}

func TestRenderContainsEntries(t *testing.T) {
	m, dir := newTestModel(t)
	next, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	m = next.(Model)

	out := ansi.Strip(m.render())
	for _, want := range []string{dir, "alpha", "beta", "charlie.txt", "2 directories, 1 files"} {
		if !strings.Contains(out, want) {
			t.Errorf("render missing %q:\n%s", want, out)
		}
	}
}

func TestRenderTooSmall(t *testing.T) {
	m, _ := newTestModel(t)
	next, _ := m.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
	m = next.(Model)

	if out := ansi.Strip(m.render()); !strings.Contains(out, "Terminal too small") {
		t.Errorf("render = %q, want a 'Terminal too small' message", out)
	}
}

func TestHelpToggle(t *testing.T) {
	m, _ := newTestModel(t)

	m = press(t, m, key(tea.KeyF1))
	if m.State() != StateHelp {
		t.Fatalf("after F1: state = %v, want help", m.State())
	}

	m = press(t, m, key(tea.KeyEscape))
	if m.State() != StateBrowsing {
		t.Errorf("after Esc: state = %v, want browsing", m.State())
	}

	m = press(t, m, runeKey('?'))
	if m.State() != StateHelp {
		t.Errorf("after ?: state = %v, want help", m.State())
	}

	m = press(t, m, runeKey('?'))
	if m.State() != StateBrowsing {
		t.Errorf("after second ?: state = %v, want browsing", m.State())
	}
}

func TestHelpScrolls(t *testing.T) {
	m, _ := newTestModel(t)
	next, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = next.(Model)

	m = press(t, m, key(tea.KeyF1))
	before := m.help.View()

	m = press(t, m, key(tea.KeyPgDown))
	after := m.help.View()

	if before == after {
		t.Error("help did not scroll after PgDn")
	}
}

func TestHelpFromDriveSelectReturns(t *testing.T) {
	m, _ := newTestModel(t)
	m.state = StateDriveSelect

	m = press(t, m, key(tea.KeyF1))
	if m.State() != StateHelp {
		t.Fatalf("state = %v, want help", m.State())
	}

	m = press(t, m, key(tea.KeyEscape))
	if m.State() != StateDriveSelect {
		t.Errorf("state = %v, want drive-select", m.State())
	}
}

func TestHelpViewFitsScreen(t *testing.T) {
	m, _ := newTestModel(t)
	next, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m = next.(Model)

	m = press(t, m, key(tea.KeyF1))
	out := m.renderHelp()

	if got := lipgloss.Height(out); got != 30 {
		t.Errorf("help height = %d, want 30", got)
	}
	for i, line := range strings.Split(out, "\n") {
		if w := lipgloss.Width(line); w != 100 {
			t.Errorf("help line %d width = %d, want 100", i, w)
		}
	}
}

func TestBackspaceAtRootOpensDriveSelect(t *testing.T) {
	m, _ := newTestModel(t)

	vol := filepath.VolumeName(t.TempDir())
	m.dir = vol + string(filepath.Separator)

	m = press(t, m, key(tea.KeyBackspace))
	if m.State() != StateDriveSelect {
		t.Fatalf("state = %v, want drive-select", m.State())
	}
	if m.Listing() == nil || len(m.Listing().Entries) == 0 {
		t.Error("drive list is empty")
	}
}

func TestDriveSelectEnterNavigates(t *testing.T) {
	m, dir := newTestModel(t)

	m.state = StateDriveSelect
	m.listing = &fsentry.Listing{
		Entries:  []fsentry.Entry{{Name: "X:", Path: dir, Kind: fsentry.KindDrive}},
		NumDirs:  1,
		SortMode: fsentry.SortNameAsc,
	}
	m.cursor = 0

	m = press(t, m, key(tea.KeyEnter))

	if m.State() != StateBrowsing {
		t.Errorf("state = %v, want browsing", m.State())
	}
	if m.Dir() != dir {
		t.Errorf("Dir() = %q, want %q", m.Dir(), dir)
	}
}

func TestDriveSelectEscQuits(t *testing.T) {
	m, dir := newTestModel(t)
	m.state = StateDriveSelect

	next, cmd := m.Update(key(tea.KeyEscape))
	m = next.(Model)

	if !m.Quitting() || m.ChosenDir() != dir {
		t.Errorf("drive-select Esc: quitting=%v chosen=%q", m.Quitting(), m.ChosenDir())
	}
	if cmd == nil {
		t.Error("expected a quit command")
	}
}

func TestSearchShiftTab(t *testing.T) {
	m, _ := newTestModel(t)

	m = press(t, m, runeKey('/'), runeKey('a'))
	m = press(t, m, key(tea.KeyEnter))
	if m.Cursor() != 0 {
		t.Fatalf("first match: cursor = %d, want 0", m.Cursor())
	}

	m = press(t, m, tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	if m.Cursor() != 2 {
		t.Errorf("after shift+tab: cursor = %d, want 2 (wrapped)", m.Cursor())
	}
}

func TestEnterMissingDirectoryShowsStatus(t *testing.T) {
	m, dir := newTestModel(t)

	if len(m.Listing().Entries) == 0 || m.Listing().Entries[0].Kind != fsentry.KindDir {
		t.Fatal("test setup: first entry is not a directory")
	}
	m.listing.Entries[0].Path = filepath.Join(dir, "does-not-exist")

	m = press(t, m, key(tea.KeyEnter))

	if m.Dir() != dir {
		t.Errorf("Dir() = %q, want %q (navigation should fail)", m.Dir(), dir)
	}
	if m.status == "" {
		t.Error("expected a status message for an unreadable directory")
	}
}

func TestOpenFileFailureShowsStatus(t *testing.T) {
	m, _ := newTestModel(t)
	m.openFile = func(string) error { return errors.New("boom") }

	m = press(t, m, key(tea.KeyEnd)) // charlie.txt
	m = press(t, m, key(tea.KeyEnter))

	if !strings.Contains(m.status, "open:") {
		t.Errorf("status = %q, want an open failure", m.status)
	}
}

func TestBrokenSymlinkShowsStatus(t *testing.T) {
	m, dir := newTestModel(t)
	m.listing = &fsentry.Listing{
		Entries: []fsentry.Entry{{
			Name: "broken",
			Path: filepath.Join(dir, "broken-link"),
			Kind: fsentry.KindSymlink,
		}},
	}
	m.cursor = 0
	m.openFile = func(string) error { return errors.New("no target") }

	m = press(t, m, key(tea.KeyEnter))

	if !strings.Contains(m.status, "open:") {
		t.Errorf("status = %q, want an open failure for the broken link", m.status)
	}
}

func TestThemeByName(t *testing.T) {
	for _, name := range []string{"", "default", "mono", "monochrome", "MONO"} {
		if _, ok := ThemeByName(name); !ok {
			t.Errorf("ThemeByName(%q) not found", name)
		}
	}
	if _, ok := ThemeByName("rainbow"); ok {
		t.Error("ThemeByName(rainbow) unexpectedly found")
	}
}

func TestNewWithOptions(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "a.txt"), []byte("1"))

	m, err := NewWithOptions(dir, Options{Theme: "mono", Sort: fsentry.SortSizeDesc, Mouse: true})
	if err != nil {
		t.Fatalf("NewWithOptions: %v", err)
	}
	if m.SortMode() != fsentry.SortSizeDesc {
		t.Errorf("SortMode() = %v, want %v", m.SortMode(), fsentry.SortSizeDesc)
	}
	if !m.mouse {
		t.Error("mouse = false, want true")
	}
	if !m.theme.Selected.GetReverse() {
		t.Error("mono theme selection does not use reverse video")
	}
}

func TestViewEnablesMouse(t *testing.T) {
	m, _ := newTestModel(t)
	if got := m.View().MouseMode; got != tea.MouseModeNone {
		t.Errorf("MouseMode = %v, want none by default", got)
	}

	m.mouse = true
	if got := m.View().MouseMode; got != tea.MouseModeCellMotion {
		t.Errorf("MouseMode = %v, want cell motion", got)
	}
}

func TestMouseClickSelectsEntry(t *testing.T) {
	m, _ := newTestModel(t)
	m.mouse = true
	next, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = next.(Model)

	// Row 3 (zero-based) is the second entry (index 1).
	next, _ = m.Update(tea.MouseClickMsg{X: 0, Y: 3, Button: tea.MouseLeft})
	m = next.(Model)
	if m.Cursor() != 1 {
		t.Errorf("after click: cursor = %d, want 1", m.Cursor())
	}

	// Clicking the header must not move the cursor.
	next, _ = m.Update(tea.MouseClickMsg{X: 0, Y: 1, Button: tea.MouseLeft})
	m = next.(Model)
	if m.Cursor() != 1 {
		t.Errorf("after header click: cursor = %d, want 1", m.Cursor())
	}
}

func TestMouseWheelMovesCursor(t *testing.T) {
	m, _ := newTestModel(t)
	m.mouse = true
	next, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = next.(Model)

	next, _ = m.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	m = next.(Model)
	if m.Cursor() != 2 {
		t.Errorf("after wheel down: cursor = %d, want 2", m.Cursor())
	}

	next, _ = m.Update(tea.MouseWheelMsg{Button: tea.MouseWheelUp})
	m = next.(Model)
	if m.Cursor() != 0 {
		t.Errorf("after wheel up: cursor = %d, want 0", m.Cursor())
	}
}

func TestMouseDisabled(t *testing.T) {
	m, _ := newTestModel(t)
	next, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = next.(Model)

	next, _ = m.Update(tea.MouseClickMsg{X: 0, Y: 4, Button: tea.MouseLeft})
	m = next.(Model)
	if m.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0 with mouse disabled", m.Cursor())
	}

	next, _ = m.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	m = next.(Model)
	if m.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0 with mouse disabled", m.Cursor())
	}
}

func TestEnterOnEmptyDirectoryIsNoop(t *testing.T) {
	m, dir := newTestModel(t)

	m = press(t, m, key(tea.KeyEnter)) // alpha (empty)
	alpha := filepath.Join(dir, "alpha")
	if m.Dir() != alpha {
		t.Fatalf("Dir() = %q, want %q", m.Dir(), alpha)
	}

	before := m.Dir()
	m = press(t, m, key(tea.KeyEnter)) // nothing to open
	if m.Dir() != before {
		t.Errorf("Dir() = %q after Enter on empty directory, want %q", m.Dir(), before)
	}
}
