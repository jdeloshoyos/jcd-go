package tui

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// TestProgramQuitsOnEscape drives the real Bubble Tea program loop with an
// injected input reader and output buffer.
func TestProgramQuitsOnEscape(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "hello.txt"), []byte("x"))

	model, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	var out bytes.Buffer
	program := tea.NewProgram(
		model,
		tea.WithInput(strings.NewReader("\x1b")), // Esc
		tea.WithOutput(&out),
		tea.WithWindowSize(100, 30),
	)

	final, err := program.Run()
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	m, ok := final.(Model)
	if !ok {
		t.Fatalf("final model is %T, want tui.Model", final)
	}
	if !m.Quitting() {
		t.Error("program did not quit")
	}
	if m.ChosenDir() != dir {
		t.Errorf("ChosenDir() = %q, want %q", m.ChosenDir(), dir)
	}
	if out.Len() == 0 {
		t.Error("renderer produced no output")
	}
}
