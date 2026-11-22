// fm is a single-pane terminal file manager: vim keys, per-directory cursor
// memory, copy/cut/paste/link/clone, search, and zoxide, delegating opens
// and edits to the user's own tools.
package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"

	"sys/outline"
	"sys/tty"
)

func main() {
	args := os.Args[1:]
	win := len(args) > 0 && args[0] == "-w" // a window, not the terminal
	if win {
		args = args[1:]
	}
	path := here() // with no argument, cur is the argument
	if len(args) > 0 {
		path = args[0]
	}
	m, err := newFM(path)
	switch {
	case err == nil && win:
		err = m.window()
	case err == nil:
		err = m.run()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "fm:", err)
		os.Exit(1)
	}
}

<<others>>
