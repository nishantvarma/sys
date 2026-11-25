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
	"sys/watch"
)

func main() {
	// -w is a window, -s a stream a host draws; else the terminal. -f
	// follows cur (follow), as ed -f does, which needs a waking when cur
	// changes: the terminal gives one and a host down a pipe can, and a
	// window cannot, so -f is not for -w.
	args, mode, follows := os.Args[1:], "", false
	for len(args) > 0 {
		if args[0] == "-f" {
			follows = true
		} else if args[0] == "-w" || args[0] == "-s" {
			mode = args[0]
		} else {
			break
		}
		args = args[1:]
	}
	if follows && mode == "-w" {
		fmt.Fprintln(os.Stderr, "usage: fm [-w | -s] [path]\n"+
			"       fm [-s] -f [path]")
		os.Exit(1)
	}
	path, row := here() // with no argument, fm opens on cur
	if len(args) > 0 {
		path, row = args[0], false
	}
	m, err := newFM(path, row, follows)
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
