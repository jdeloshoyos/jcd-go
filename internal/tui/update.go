package tui

import tea "charm.land/bubbletea/v2"

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.search.SetWidth(m.searchInputWidth())
		m.ensureVisible()
		return m, nil

	case tea.KeyPressMsg:
		if m.state == StateSearching {
			return m.updateSearching(msg)
		}
		return m.updateBrowsing(msg)

	case tea.PasteMsg:
		if m.state == StateSearching {
			var cmd tea.Cmd
			m.search, cmd = m.search.Update(msg)
			return m, cmd
		}
	}
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
	case "f2":
		m.cycleSort()
	case "/":
		m.startSearch()
	}
	return m, nil
}

func (m Model) updateSearching(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m.quit()
	case "esc":
		m.state = StateBrowsing
		m.search.Blur()
		m.search.Reset()
		m.status = ""
		return m, nil
	case "enter":
		m.jumpFirstMatch()
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
		return m, cmd
	}
}

func (m Model) quit() (tea.Model, tea.Cmd) {
	m.quitting = true
	m.chosen = m.dir
	return m, tea.Quit
}
