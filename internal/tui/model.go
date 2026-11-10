// Package tui implements the Bubble Tea interface of the directory navigator.
package tui

import (
	"os"
	"path/filepath"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"jcd-go/internal/fsentry"
	"jcd-go/internal/platform"
)

// Minimum terminal dimensions required by the interface.
const (
	MinWidth  = 80
	MinHeight = 20
)

const (
	defaultWidth  = 80
	defaultHeight = 24
)

// State is the active interaction mode.
type State int

const (
	// StateBrowsing is normal directory navigation.
	StateBrowsing State = iota
	// StateSearching is the explicit search-input mode.
	StateSearching
)

// Model is the Bubble Tea model for the navigator.
type Model struct {
	dir      string
	listing  *fsentry.Listing
	cursor   int
	offset   int
	sortMode fsentry.SortMode
	state    State
	search   textinput.Model

	width  int
	height int

	status   string
	quitting bool
	chosen   string

	openFile func(string) error
}

// New creates a model rooted at dir and performs the initial scan.
func New(dir string) (Model, error) {
	m := Model{
		dir:      dir,
		sortMode: fsentry.SortNameAsc,
		state:    StateBrowsing,
		width:    defaultWidth,
		height:   defaultHeight,
		search:   newSearchInput(),
		openFile: platform.OpenFile,
	}
	if err := m.loadDir(dir); err != nil {
		return Model{}, err
	}
	return m, nil
}

func newSearchInput() textinput.Model {
	ti := textinput.New()
	ti.Prompt = ""
	ti.Placeholder = "type to search"
	ti.CharLimit = 120
	ti.SetVirtualCursor(true)
	return ti
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Dir returns the directory currently being browsed.
func (m Model) Dir() string { return m.dir }

// ChosenDir returns the directory to hand back to the shell after quitting.
func (m Model) ChosenDir() string { return m.chosen }

// Quitting reports whether the user asked to quit.
func (m Model) Quitting() bool { return m.quitting }

// State returns the active interaction mode.
func (m Model) State() State { return m.state }

// Cursor returns the zero-based cursor index.
func (m Model) Cursor() int { return m.cursor }

// SortMode returns the active sort mode.
func (m Model) SortMode() fsentry.SortMode { return m.sortMode }

// Listing returns the current listing.
func (m Model) Listing() *fsentry.Listing { return m.listing }

func (m *Model) loadDir(dir string) error {
	listing, err := fsentry.ScanDir(dir, platform.Attrs, m.sortMode)
	if err != nil {
		return err
	}
	m.dir = listing.Dir
	m.listing = listing
	m.cursor = 0
	m.offset = 0
	return nil
}

func (m *Model) visibleRows() int {
	rows := m.height - 3
	if rows < 1 {
		rows = 1
	}
	return rows
}

func (m *Model) ensureVisible() {
	if m.listing == nil {
		return
	}
	rows := m.visibleRows()
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+rows {
		m.offset = m.cursor - rows + 1
	}
	maxOffset := len(m.listing.Entries) - rows
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.offset > maxOffset {
		m.offset = maxOffset
	}
	if m.offset < 0 {
		m.offset = 0
	}
}

func (m *Model) setCursor(i int) {
	if m.listing == nil || len(m.listing.Entries) == 0 {
		m.cursor = 0
		m.offset = 0
		return
	}
	if i < 0 {
		i = 0
	}
	if i >= len(m.listing.Entries) {
		i = len(m.listing.Entries) - 1
	}
	m.cursor = i
	m.ensureVisible()
}

func (m *Model) move(delta int) {
	m.setCursor(m.cursor + delta)
}

func (m *Model) enter() {
	if m.listing == nil || len(m.listing.Entries) == 0 {
		return
	}
	e := m.listing.Entries[m.cursor]
	switch e.Kind {
	case fsentry.KindDir, fsentry.KindDrive:
		if err := m.loadDir(e.Path); err != nil {
			m.status = err.Error()
		}
	case fsentry.KindSymlink:
		if info, err := os.Stat(e.Path); err == nil && info.IsDir() {
			if err := m.loadDir(e.Path); err != nil {
				m.status = err.Error()
			}
			return
		}
		m.openEntry(e)
	default:
		m.openEntry(e)
	}
}

func (m *Model) openEntry(e fsentry.Entry) {
	if err := m.openFile(e.Path); err != nil {
		m.status = "open: " + err.Error()
	}
}

func (m *Model) goUp() {
	parent := filepath.Dir(m.dir)
	if parent == m.dir {
		m.status = "already at the root"
		return
	}
	if err := m.loadDir(parent); err != nil {
		m.status = err.Error()
	}
}

func (m *Model) cycleSort() {
	m.sortMode = m.sortMode.Next()
	if m.listing != nil {
		m.listing.SortMode = m.sortMode
		fsentry.Sort(m.listing.Entries, m.sortMode)
	}
	m.setCursor(0)
	m.status = "sorted by " + m.sortMode.String()
}

func (m *Model) startSearch() {
	m.state = StateSearching
	m.search.Reset()
	m.search.SetWidth(m.searchInputWidth())
	m.search.Focus()
	m.status = ""
}

func (m Model) searchInputWidth() int {
	w := m.width - len("Search: ") - 44
	if w < 12 {
		w = 12
	}
	return w
}

func (m *Model) jumpFirstMatch() {
	if m.listing == nil {
		return
	}
	matches := fsentry.Matches(m.listing.Entries, m.search.Value())
	if len(matches) == 0 {
		m.status = "no matches"
		return
	}
	m.setCursor(matches[0] - 1)
	m.status = ""
}

func (m *Model) jumpNextMatch(direction int) {
	if m.listing == nil {
		return
	}
	matches := fsentry.Matches(m.listing.Entries, m.search.Value())
	if len(matches) == 0 {
		m.status = "no matches"
		return
	}
	current := -1
	for i, oneBased := range matches {
		if oneBased-1 == m.cursor {
			current = i
			break
		}
	}
	var next int
	if current < 0 {
		if direction >= 0 {
			next = 0
		} else {
			next = len(matches) - 1
		}
	} else {
		next = (current + direction + len(matches)) % len(matches)
	}
	m.setCursor(matches[next] - 1)
	m.status = ""
}
