//go:build windows

// Package integration contains end-to-end tests that exercise the shell
// wrapper contract.
package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate the test file")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func findPowerShell(t *testing.T) string {
	t.Helper()
	for _, name := range []string{"pwsh", "powershell"} {
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
	}
	t.Skip("PowerShell is not available")
	return ""
}

func wrapperPath(t *testing.T) string {
	t.Helper()
	wrapper := filepath.Join(repoRoot(t), "scripts", "jcd.ps1")
	if _, err := os.Stat(wrapper); err != nil {
		t.Fatalf("wrapper not found: %v", err)
	}
	return wrapper
}

func writeMock(t *testing.T, dir, body string) string {
	t.Helper()
	mock := filepath.Join(dir, "mockjcd.ps1")
	if err := os.WriteFile(mock, []byte(body), 0o644); err != nil {
		t.Fatalf("write mock: %v", err)
	}
	return mock
}

func TestPowerShellWrapperChangesDirectory(t *testing.T) {
	shell := findPowerShell(t)
	wrapper := wrapperPath(t)

	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatal(err)
	}

	mock := writeMock(t, dir, "Write-Output $args[0]\n")

	cmd := exec.Command(shell, "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass",
		"-Command", `& { & $env:JCD_WRAPPER $env:JCD_TARGET; Write-Output "LOC=$(Get-Location)" }`)
	cmd.Env = append(os.Environ(),
		"JCD_BIN="+mock,
		"JCD_WRAPPER="+wrapper,
		"JCD_TARGET="+target,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("powershell: %v\n%s", err, out)
	}
	if !containsFold(string(out), "LOC="+target) {
		t.Fatalf("wrapper did not change directory; output:\n%s", out)
	}
}

func TestPowerShellWrapperIgnoresFailure(t *testing.T) {
	shell := findPowerShell(t)
	wrapper := wrapperPath(t)

	dir := t.TempDir()
	start := filepath.Join(dir, "start")
	target := filepath.Join(dir, "target")
	for _, p := range []string{start, target} {
		if err := os.Mkdir(p, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	// The mock prints a path but reports failure; the wrapper must ignore it.
	mock := writeMock(t, dir, "Write-Output $args[0]\nexit 1\n")

	cmd := exec.Command(shell, "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass",
		"-Command", `& { Set-Location $env:JCD_START; & $env:JCD_WRAPPER $env:JCD_TARGET; Write-Output "LOC=$(Get-Location)" }`)
	cmd.Env = append(os.Environ(),
		"JCD_BIN="+mock,
		"JCD_WRAPPER="+wrapper,
		"JCD_START="+start,
		"JCD_TARGET="+target,
	)

	out, _ := cmd.CombinedOutput()
	if containsFold(string(out), "LOC="+target) {
		t.Fatalf("wrapper changed directory despite failure; output:\n%s", out)
	}
	if !containsFold(string(out), "LOC="+start) {
		t.Fatalf("wrapper left the wrong directory; output:\n%s", out)
	}
}

func containsFold(haystack, needle string) bool {
	return strings.Contains(strings.ToLower(haystack), strings.ToLower(needle))
}
