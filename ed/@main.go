// ed is a single-buffer modal editor over the shared tty module. A selection
// is [anchor, head]; a collapsed one is the caret. Single caret by design —
// no multi-cursor.
package main

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"sys/outline"
	"sys/state"
	"sys/tty"
	"sys/watch"
)

func main() {
	// +line is how an address arrives from outside: e passes it on, from
	// a caller that has one or from cur's own line. A line past the end
	// clamps.
	// -w is a window, -s a stream a host draws; else the terminal. -f
	// chases cur (see chase), and with no file edits what cur names. A
	// chase needs a waking when cur changes, which the terminal gives and
	// a host down a pipe can, and a window cannot: -f is not for -w.
	args, line, view, chasing := os.Args[1:], 0, "", false
	for len(args) > 0 {
		if args[0] == "-f" {
			chasing = true
		} else if args[0] == "-w" || args[0] == "-s" {
			view = args[0]
		} else {
			break
		}
		args = args[1:]
	}
	if len(args) > 0 && strings.HasPrefix(args[0], "+") {
		line, _ = strconv.Atoi(args[0][1:])
		args = args[1:]
	}
	if len(args) == 0 && !chasing || chasing && view == "-w" {
		fmt.Fprintln(os.Stderr, "usage: ed [-w | -s] [+line] file\n"+
			"       ed [-s] -f [+line] [file]")
		os.Exit(1)
	}
	path, curf, at := "", "", ""
	if chasing {
		win, _ := state.Pane()
		curf = state.Path("cur", win)
		at = state.Read(curf)
	}
	if len(args) > 0 {
		path = args[0]
	} else if p, l, ok := edits(at); ok {
		path = p
		if line == 0 {
			line = l
		}
	}
	b, err := load(path)
	if err == nil {
		e := newEditor(b)
		e.curf, e.chased = curf, at
		if chasing && path == "" {
			e.msg = "cur names nothing ed edits"
		}
		if line > 0 {
			e.place(pos{line - 1, 0})
		}
		switch view {
		case "-w":
			err = e.window()
		case "-s":
			err = e.stream()
		default:
			err = e.run()
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "ed:", err)
		os.Exit(1)
	}
}

<<others>>
