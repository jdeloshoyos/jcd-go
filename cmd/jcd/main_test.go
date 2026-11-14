package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"jcd-go/internal/cli"
	"jcd-go/internal/version"
)

func TestRunVersion(t *testing.T) {
	stdout, _, code := captureRun(t, []string{"--version"})

	if code != cli.ExitOK {
		t.Errorf("exit code = %d, want %d", code, cli.ExitOK)
	}
	if !strings.HasPrefix(stdout, cli.ModeInfo+"\n") {
		t.Errorf("stdout = %q, want it to start with the %q marker", stdout, cli.ModeInfo)
	}
	if !strings.Contains(stdout, version.Version) {
		t.Errorf("stdout = %q, want the version", stdout)
	}
}

func TestRunHelp(t *testing.T) {
	stdout, _, code := captureRun(t, []string{"--help"})

	if code != cli.ExitOK {
		t.Errorf("exit code = %d, want %d", code, cli.ExitOK)
	}
	if !strings.HasPrefix(stdout, cli.ModeInfo+"\n") {
		t.Errorf("stdout = %q, want it to start with the %q marker", stdout, cli.ModeInfo)
	}
	if !strings.Contains(stdout, "Usage:") {
		t.Errorf("stdout = %q, want usage text", stdout)
	}
}

func TestRunUnknownOption(t *testing.T) {
	_, stderr, code := captureRun(t, []string{"--bogus"})

	if code != cli.ExitUsage {
		t.Errorf("exit code = %d, want %d", code, cli.ExitUsage)
	}
	if !strings.Contains(stderr, "unknown option") {
		t.Errorf("stderr = %q, want an unknown-option error", stderr)
	}
}

func TestRunInvalidDirectory(t *testing.T) {
	_, stderr, code := captureRun(t, []string{"definitely-not-a-real-directory-xyz"})

	if code != cli.ExitUsage {
		t.Errorf("exit code = %d, want %d", code, cli.ExitUsage)
	}
	if !strings.Contains(stderr, "not a valid directory") {
		t.Errorf("stderr = %q, want a directory error", stderr)
	}
}

func TestRunTooManyArgs(t *testing.T) {
	_, stderr, code := captureRun(t, []string{"a", "b"})

	if code != cli.ExitUsage {
		t.Errorf("exit code = %d, want %d", code, cli.ExitUsage)
	}
	if !strings.Contains(stderr, "too many arguments") {
		t.Errorf("stderr = %q, want a too-many-arguments error", stderr)
	}
}

func TestRunInvalidConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"theme":`), 0o644); err != nil {
		t.Fatal(err)
	}

	_, stderr, code := captureRun(t, []string{"--config", path, "."})

	if code != cli.ExitUsage {
		t.Errorf("exit code = %d, want %d", code, cli.ExitUsage)
	}
	if !strings.Contains(stderr, "config") {
		t.Errorf("stderr = %q, want a config error", stderr)
	}
}

func TestRunUnknownTheme(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"theme":"rainbow"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	_, stderr, code := captureRun(t, []string{"--config", path, "."})

	if code != cli.ExitUsage {
		t.Errorf("exit code = %d, want %d", code, cli.ExitUsage)
	}
	if !strings.Contains(stderr, "unknown theme") {
		t.Errorf("stderr = %q, want an unknown-theme error", stderr)
	}
}

func TestRunUnknownSort(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"sort":"sideways"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	_, stderr, code := captureRun(t, []string{"--config", path, "."})

	if code != cli.ExitUsage {
		t.Errorf("exit code = %d, want %d", code, cli.ExitUsage)
	}
	if !strings.Contains(stderr, "unknown sort") {
		t.Errorf("stderr = %q, want an unknown-sort error", stderr)
	}
}

func captureRun(t *testing.T, args []string) (stdout, stderr string, code int) {
	t.Helper()

	oldOut, oldErr := os.Stdout, os.Stderr
	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	rErr, wErr, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout, os.Stderr = wOut, wErr
	defer func() {
		os.Stdout, os.Stderr = oldOut, oldErr
	}()

	code = run(args)

	if err := wOut.Close(); err != nil {
		t.Fatal(err)
	}
	if err := wErr.Close(); err != nil {
		t.Fatal(err)
	}
	outBytes, _ := io.ReadAll(rOut)
	errBytes, _ := io.ReadAll(rErr)
	return string(outBytes), string(errBytes), code
}
