// ed is a single-buffer modal editor over the shared tty module. A selection
// is [anchor, head]; a collapsed one is the caret. Single caret by design —
// no multi-cursor.
package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"sys/tty"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: ed file")
		os.Exit(1)
	}
	b, err := load(os.Args[1])
	if err == nil {
		err = newEditor(b).run()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "ed:", err)
		os.Exit(1)
	}
}

<<others>>
