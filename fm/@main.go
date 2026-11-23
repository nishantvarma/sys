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
	"sys/state"
	"sys/tty"
	"sys/view"
)

func main() {
	args := os.Args[1:]
	// -w is a window, -s a stream a host draws; else the terminal
	mode := ""
	if len(args) > 0 && (args[0] == "-w" || args[0] == "-s") {
		mode, args = args[0], args[1:]
	}
	path, row := here() // with no argument, fm opens on cur
	if len(args) > 0 {
		path, row = args[0], false
	}
	m, err := newFM(path, row)
	switch {
	case err == nil && mode == "-w":
		err = m.window()
	case err == nil && mode == "-s":
		err = m.stream()
	case err == nil:
		err = m.run()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "fm:", err)
		os.Exit(1)
	}
}

<<others>>
