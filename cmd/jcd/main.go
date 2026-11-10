// Command jcd is a terminal directory navigator.
//
// The interactive TUI is implemented in a later milestone. For now this
// command provides the version flag and a placeholder so the module builds and
// tests cleanly.
package main

import (
	"fmt"
	"os"
)

// version is the semantic version of the application. It stays at the 0.1.0
// placeholder until the first release candidate.
const version = "0.1.0"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v":
			fmt.Printf("jcd %s\n", version)
			return
		}
	}

	fmt.Fprintln(os.Stderr, "jcd: the interactive TUI is not implemented yet")
	os.Exit(2)
}
