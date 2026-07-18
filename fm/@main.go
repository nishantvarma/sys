// fm is a single-pane terminal file manager: vim keys, per-directory cursor
// memory, copy/cut/paste/link/clone/tag, search, and zoxide, delegating opens
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

	"sys/tty"
)

func main() {
	path := "."
	if len(os.Args) > 1 {
		path = os.Args[1]
	}
	m, err := newFM(path)
	if err == nil {
		err = m.run()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "fm:", err)
		os.Exit(1)
	}
}

<<others>>
