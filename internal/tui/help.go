package tui

import "jcd-go/internal/version"

// helpText returns the full help document. Lines are kept within 72 columns so
// they fit the minimum supported terminal width without wrapping.
func helpText() string {
	return helpBody
}

const helpBody = `by Jaime de los Hoyos M.

Commands:
  Up / Down / PgUp / PgDn   Navigate the listing
  Home / End                First / last entry
  Enter                     Open a directory or file
  Backspace                 Go up one level
                            (at a drive root: show the drive list)
  Esc                       Quit and report the current directory
  F1 or ?                   Toggle this help
  F2                        Cycle the sort mode
  F3                        Start a search (or start typing)

Search:
  Matching is case-insensitive and matches anywhere in the entry name.
  Typing moves the selection to the first match.
    Enter       Open the selected entry and leave search
    Tab         Next match (wraps)
    Shift+Tab   Previous match
    Esc         Cancel the search

Sort modes (F2):
  Name / Date / Size, ascending and descending. Directories always
  come first; ties are broken by name.

Entry attributes:
  d = Directory
  l = Reparse point / symbolic link
  r = Read-only
  h = Hidden
  s = System
  a = Archive
  c = Compressed
  i = Not content indexed
  e = Encrypted

Drives:
  Backspace at a drive root shows the system drives. Select one and
  press Enter to open its root. Esc quits and keeps the last browsed
  directory.

Configuration:
  An optional config.json can select the theme ("default" or "mono"),
  the initial sort, and mouse support. See the project README.

Sizes use IEC units: B, KiB, MiB, GiB, TiB, ... (1 KiB = 1024 B).

JCD is a Go port of the PowerBASIC JVA Directory Navigator (v1.2, 2010).`

// renderHelp draws the help panel. The panel occupies exactly the screen: a
// one-line title, a scrollable viewport, and a one-line footer inside a border.
func (m Model) renderHelp() string {
	innerWidth := m.helpViewportWidth()

	title := padRight(truncate("JCD Help — "+version.Version, innerWidth, "…"), innerWidth)
	footer := padRight(truncate("[Up/Down/PgUp/PgDn]=Scroll  [F1]/[Esc]=Close", innerWidth, "…"), innerWidth)

	content := m.theme.HeaderPath.Render(title) + "\n" +
		m.help.View() + "\n" +
		m.theme.Footer.Render(footer)

	return m.theme.HelpBox.Render(content)
}
