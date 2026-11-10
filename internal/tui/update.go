package tui

import tea "charm.land/bubbletea/v2"

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.search.SetWidth(m.searchInputWidth())
		if m.state == StateHelp {
			m.help.SetWidth(m.helpViewportWidth())
			m.help.SetHeight(m.helpViewportHeight())
		}
		m.ensureVisible()
		return m, nil

	case tea.KeyPressMsg:
		switch m.state {
		case StateSearching:
			return m.updateSearching(msg)
		case StateHelp:
			return m.updateHelp(msg)
		case StateDriveSelect:
			return m.updateDriveSelect(msg)
		default:
			return m.updateBrowsing(msg)
		}

	case tea.PasteMsg:
		if m.state == StateSearching {
			var cmd tea.Cmd
			m.search, cmd = m.search.Update(msg)
			m.followFirstMatch()
			return m, cmd
		}

	case tea.MouseWheelMsg:
		return m.updateMouseWheel(msg)

	case tea.MouseClickMsg:
		return m.updateMouseClick(msg)
	}
	return m, nil
}

func (m Model) updateMouseWheel(msg tea.MouseWheelMsg) (tea.Model, tea.Cmd) {
	if !m.mouse {
		return m, nil
	}

	switch m.state {
	case StateBrowsing, StateDriveSelect:
		switch msg.Button {
		case tea.MouseWheelUp:
			m.move(-3)
		case tea.MouseWheelDown:
			m.move(3)
		}
	case StateHelp:
		var cmd tea.Cmd
		m.help, cmd = m.help.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) updateMouseClick(msg tea.MouseClickMsg) (tea.Model, tea.Cmd) {
	if !m.mouse || msg.Button != tea.MouseLeft {
		return m, nil
	}
	if m.state != StateBrowsing && m.state != StateDriveSelect {
		return m, nil
	}

	// The first two rows are the header; entries start on row 2 (zero-based).
	row := msg.Y - 2
	if row < 0 || row >= m.visibleRows() {
		return m, nil
	}
	idx := m.offset + row
	if m.listing == nil || idx >= len(m.listing.Entries) {
		return m, nil
	}

	m.setCursor(idx)
	m.status = ""
	return m, nil
}

func (m Model) updateBrowsing(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if key := msg.String(); key != "ctrl+c" {
		m.status = ""
	}

	switch msg.String() {
	case "ctrl+c", "esc":
		return m.quit()
	case "up":
		m.move(-1)
	case "down":
		m.move(1)
	case "pgup":
		m.move(-m.visibleRows())
	case "pgdown":
		m.move(m.visibleRows())
	case "home":
		m.setCursor(0)
	case "end":
		if m.listing != nil {
			m.setCursor(len(m.listing.Entries) - 1)
		}
	case "enter":
		m.enter()
	case "backspace":
		m.goUp()
	case "f1", "?":
		m.startHelp(StateBrowsing)
	case "f2":
		m.cycleSort()
	case "/":
		m.startSearch()
	}
	return m, nil
}

func (m Model) updateDriveSelect(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if key := msg.String(); key != "ctrl+c" {
		m.status = ""
	}

	switch msg.String() {
	case "ctrl+c", "esc":
		return m.quit()
	case "up":
		m.move(-1)
	case "down":
		m.move(1)
	case "pgup":
		m.move(-m.visibleRows())
	case "pgdown":
		m.move(m.visibleRows())
	case "home":
		m.setCursor(0)
	case "end":
		if m.listing != nil {
			m.setCursor(len(m.listing.Entries) - 1)
		}
	case "enter":
		m.openSelected()
	case "f1", "?":
		m.startHelp(StateDriveSelect)
	}
	return m, nil
}

func (m Model) updateSearching(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m.quit()
	case "esc":
		m.exitSearch()
		return m, nil
	case "enter":
		m.exitSearch()
		m.enter()
		return m, nil
	case "tab", "down":
		m.jumpNextMatch(1)
		return m, nil
	case "shift+tab", "up":
		m.jumpNextMatch(-1)
		return m, nil
	default:
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)
		m.followFirstMatch()
		return m, cmd
	}
}

func (m Model) updateHelp(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m.quit()
	case "esc", "f1", "?":
		m.closeHelp()
		return m, nil
	}

	var cmd tea.Cmd
	m.help, cmd = m.help.Update(msg)
	return m, cmd
}

func (m Model) quit() (tea.Model, tea.Cmd) {
	m.quitting = true
	m.chosen = m.dir
	return m, tea.Quit
}
