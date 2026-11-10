// Command jcd is a terminal directory navigator.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	tea "charm.land/bubbletea/v2"

	"jcd-go/internal/cli"
	"jcd-go/internal/tui"
	"jcd-go/internal/version"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) > 0 {
		switch args[0] {
		case "--version", "-v":
			fmt.Printf("jcd %s\n", version.Version)
			return cli.ExitOK
		case "--help", "-h":
			usage(os.Stdout)
			return cli.ExitOK
		}
	}
	if len(args) > 1 {
		usage(os.Stderr)
		return cli.ExitUsage
	}

	var arg string
	if len(args) == 1 {
		arg = args[0]
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "jcd: cannot determine the current directory: %v\n", err)
		return cli.ExitError
	}

	dir, err := cli.ResolveStartDir(arg, cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "jcd: %v\n", err)
		return cli.ExitUsage
	}

	if w, h, err := cli.TerminalSize(); err == nil && (w < tui.MinWidth || h < tui.MinHeight) {
		fmt.Fprintf(os.Stderr, "jcd: terminal too small (%dx%d); at least %dx%d required\n",
			w, h, tui.MinWidth, tui.MinHeight)
		return cli.ExitTooSmall
	}

	model, err := tui.New(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "jcd: %v\n", err)
		return cli.ExitError
	}

	// The interface renders to stderr so stdout stays clean for the shell
	// wrapper, which captures the final directory from stdout.
	program := tea.NewProgram(model, tea.WithOutput(os.Stderr))
	final, err := program.Run()
	if err != nil {
		if errors.Is(err, tea.ErrInterrupted) {
			return cli.ExitOK
		}
		fmt.Fprintf(os.Stderr, "jcd: %v\n", err)
		return cli.ExitError
	}

	m, ok := final.(tui.Model)
	if !ok || m.ChosenDir() == "" {
		return cli.ExitError
	}

	fmt.Println(m.ChosenDir())
	return cli.ExitOK
}

func usage(w io.Writer) {
	fmt.Fprintf(w, `jcd %s - terminal directory navigator

Usage:
  jcd [directory]

Navigate with the arrow keys, Enter, and Backspace. Press / to search and
Esc to quit. On exit, jcd prints the final directory to stdout so a shell
wrapper (for example scripts/jcd.ps1) can change into it.
`, version.Version)
}
