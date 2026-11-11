//go:build windows

// Package integration contains end-to-end tests that exercise the shell
// wrapper contract.
package integration

import (
	"fmt"
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

// buildNativeMock compiles a real executable that writes a banner and the path
// given as its first argument. Unlike a .ps1 mock, a native process emits raw
// UTF-8 bytes on stdout, which is exactly what the wrapper must decode.
func buildNativeMock(t *testing.T, dir string) string {
	t.Helper()

	src := filepath.Join(dir, "mockjcd.go")
	const body = `package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("JCD banner")
	fmt.Println(os.Args[1])
}
`
	if err := os.WriteFile(src, []byte(body), 0o644); err != nil {
		t.Fatalf("write native mock source: %v", err)
	}

	exe := filepath.Join(dir, "mockjcd.exe")
	if out, err := exec.Command("go", "build", "-o", exe, src).CombinedOutput(); err != nil {
		t.Fatalf("build native mock: %v\n%s", err, out)
	}
	return exe
}

// runWrapper runs the wrapper in a fresh PowerShell session. start is the
// location before the wrapper runs (empty to leave it alone), target is the
// path passed to the wrapper, and expect is the location the wrapper must end
// in. The comparison is done inside PowerShell to avoid output-encoding issues
// with non-ASCII paths.
func runWrapper(t *testing.T, shell, wrapper, mock, start, target, expect string) {
	t.Helper()
	runWrapperWithPrologue(t, shell, wrapper, mock, start, target, expect, "")
}

// runWrapperWithPrologue is runWrapper with a snippet evaluated in the test
// session before the wrapper runs. Tests use it to force a specific output
// encoding and prove the wrapper handles jcd's UTF-8 stdout.
func runWrapperWithPrologue(t *testing.T, shell, wrapper, mock, start, target, expect, prologue string) {
	t.Helper()

	script := fmt.Sprintf(`& {
    %s
    if ($env:JCD_START) { Set-Location -LiteralPath $env:JCD_START }
    & $env:JCD_WRAPPER $env:JCD_TARGET
    if ((Get-Location).Path -eq $env:JCD_EXPECT) { 'LOCATION-OK' } else { 'LOCATION-BAD' }
}`, prologue)

	cmd := exec.Command(shell, "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", script)
	cmd.Env = append(os.Environ(),
		"JCD_BIN="+mock,
		"JCD_WRAPPER="+wrapper,
		"JCD_START="+start,
		"JCD_TARGET="+target,
		"JCD_EXPECT="+expect,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("powershell: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "LOCATION-OK") {
		t.Fatalf("wrapper ended in the wrong directory; output:\n%s", out)
	}
}

func TestPowerShellWrapperChangesDirectory(t *testing.T) {
	shell := findPowerShell(t)
	wrapper := wrapperPath(t)

	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatal(err)
	}
	mock := writeMock(t, dir, "Write-Output 'JCD banner'\nWrite-Output $args[0]\n")

	runWrapper(t, shell, wrapper, mock, "", target, target)
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

	// The mock prints a target but reports failure; the wrapper must ignore it.
	mock := writeMock(t, dir, "Write-Output 'JCD banner'\nWrite-Output $args[0]\nexit 1\n")

	runWrapper(t, shell, wrapper, mock, start, target, start)
}

func TestPowerShellWrapperRejectsInvalidTarget(t *testing.T) {
	shell := findPowerShell(t)
	wrapper := wrapperPath(t)

	dir := t.TempDir()
	start := filepath.Join(dir, "start")
	if err := os.Mkdir(start, 0o755); err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(dir, "missing")

	// The mock exits successfully but reports a path that does not exist; the
	// wrapper must reject it and stay put.
	mock := writeMock(t, dir, "Write-Output 'JCD banner'\nWrite-Output $args[0]\n")

	runWrapper(t, shell, wrapper, mock, start, missing, start)
}

func TestPowerShellWrapperHandlesSpecialPaths(t *testing.T) {
	shell := findPowerShell(t)
	wrapper := wrapperPath(t)

	names := []string{
		"target with spaces",
		"café-Δ-日本",
		"amp & (paren) 'quote';,=",
	}

	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			target := filepath.Join(dir, name)
			if err := os.Mkdir(target, 0o755); err != nil {
				t.Fatalf("mkdir: %v", err)
			}
			mock := writeMock(t, dir, "Write-Output 'JCD banner'\nWrite-Output $args[0]\n")

			runWrapper(t, shell, wrapper, mock, "", target, target)
		})
	}
}

// TestPowerShellWrapperDecodesUTF8Path reproduces the mojibake bug: a native
// jcd writes UTF-8, but Windows PowerShell 5.1 decodes native stdout with the
// OEM code page, turning "Imágenes" into "Im├ígenes". The wrapper must force
// UTF-8 while it reads the path.
func TestPowerShellWrapperDecodesUTF8Path(t *testing.T) {
	shell := findPowerShell(t)
	wrapper := wrapperPath(t)

	dir := t.TempDir()
	target := filepath.Join(dir, "Imágenes")
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	mock := buildNativeMock(t, dir)

	// Force a non-UTF-8 decoding so the test reproduces the bug even when the
	// test session is PowerShell 7, where UTF-8 is already the default.
	runWrapperWithPrologue(t, shell, wrapper, mock, "", target, target,
		"[Console]::OutputEncoding = [System.Text.Encoding]::ASCII")
}
