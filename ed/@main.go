// ed is a single-buffer modal editor over the shared tty module. A selection
// is [anchor, head]; a collapsed one is the caret. Single caret by design —
// no multi-cursor.
package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"sys/tty"
)

func main() {
	// +line is how an address arrives from outside: e passes it on, from
	// a caller that has one or from cur's own line. A line past the end
	// clamps.
	args, line := os.Args[1:], 0
	if len(args) > 0 && strings.HasPrefix(args[0], "+") {
		line, _ = strconv.Atoi(args[0][1:])
		args = args[1:]
	}
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: ed [+line] file")
		os.Exit(1)
	}
	b, err := load(args[0])
	if err == nil {
		e := newEditor(b)
		if line > 0 {
			e.place(pos{line - 1, 0})
		}
		err = e.run()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "ed:", err)
		os.Exit(1)
	}
}

<<others>>
