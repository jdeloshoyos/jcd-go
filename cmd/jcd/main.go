package main

import (
	"errors"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"jcd-go/internal/cli"
	"jcd-go/internal/config"
	"jcd-go/internal/tui"
	"jcd-go/internal/version"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	opts, err := cli.ParseArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "jcd: %v\n", err)
		cli.Usage(os.Stderr, version.Version)
		return cli.ExitUsage
	}

	switch opts.Action {
	case cli.ActionVersion:
		fmt.Printf("jcd %s\n", version.Version)
		return cli.ExitOK
	case cli.ActionHelp:
		cli.Usage(os.Stdout, version.Version)
		return cli.ExitOK
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "jcd: cannot determine the current directory: %v\n", err)
		return cli.ExitError
	}

	dir, err := cli.ResolveStartDir(opts.Dir, cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "jcd: %v\n", err)
		return cli.ExitUsage
	}

	cfgPath := opts.ConfigPath
	if cfgPath == "" {
		cfgPath = os.Getenv("JCD_CONFIG")
	}
	if cfgPath == "" {
		cfgPath = config.DefaultPath()
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "jcd: %v\n", err)
		return cli.ExitUsage
	}
	sortMode, err := cfg.SortMode()
	if err != nil {
		fmt.Fprintf(os.Stderr, "jcd: %v\n", err)
		return cli.ExitUsage
	}
	if _, ok := tui.ThemeByName(cfg.Theme); !ok {
		fmt.Fprintf(os.Stderr, "jcd: unknown theme %q\n", cfg.Theme)
		return cli.ExitUsage
	}

	if w, h, err := cli.TerminalSize(); err == nil && (w < tui.MinWidth || h < tui.MinHeight) {
		fmt.Fprintf(os.Stderr, "jcd: terminal too small (%dx%d); at least %dx%d required\n",
			w, h, tui.MinWidth, tui.MinHeight)
		return cli.ExitTooSmall
	}

	model, err := tui.NewWithOptions(dir, tui.Options{
		Theme: cfg.Theme,
		Sort:  sortMode,
		Mouse: cfg.Mouse,
	})
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

	// Print a banner and the target path. The shell wrapper turns these into a
	// human-readable confirmation and applies the path with Set-Location.
	fmt.Printf("%s v%s\n", version.Name, version.Version)
	fmt.Println(m.ChosenDir())
	return cli.ExitOK
}
