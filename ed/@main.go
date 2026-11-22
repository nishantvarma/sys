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
	"slices"
	"strconv"
	"strings"
	"unicode"

	"sys/outline"
	"sys/tty"
)

func main() {
	// +line is how an address arrives from outside: e passes it on, from
	// a caller that has one or from cur's own line. A line past the end
	// clamps.
	args, line, win := os.Args[1:], 0, false
	if len(args) > 0 && args[0] == "-w" { // in a window, not the terminal
		args, win = args[1:], true
	}
	if len(args) > 0 && strings.HasPrefix(args[0], "+") {
		line, _ = strconv.Atoi(args[0][1:])
		args = args[1:]
	}
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: ed [-w] [+line] file")
		os.Exit(1)
	}
	b, err := load(args[0])
	if err == nil {
		e := newEditor(b)
		if line > 0 {
			e.place(pos{line - 1, 0})
		}
		if win {
			err = e.window()
		} else {
			err = e.run()
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "ed:", err)
		os.Exit(1)
	}
}

<<others>>
