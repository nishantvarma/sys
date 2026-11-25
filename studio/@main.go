// studio is sys's frame: fm, the thing at fm's cursor in ed, and a
// shell, side by side, each running as it is; a slot that runs studio
// show adds a lens, the thing seen another way. The frame is a note,
// read by its indent (outline.Text): its lines stack down, the lines
// under one go across, and so on, turn about. The
// same note lays out as tmux panes in the terminal and as panes on a
// page. cur, which fm writes as its cursor moves, is all the panes
// share, and a slot's name is how a program reaches it.
//
//	studio [-n] [frame]             the frame in a tmux window: this
//	                                one, around the shell alone in it,
//	                                else a new one; -n prints its tmux
//	                                commands, runs none
//	studio page [frame]             the frame as a page, on stdout
//	studio figure note [page.html]  a note's figure as a page, on the
//	                                skeleton given, or sys's own
//	studio lens [name]              the window's lens, or set it
//	studio send slot text...        paste text into a slot, then enter
//	studio read slot                a slot's screen, as text
//	studio show [-w | -s]           the lens: cur through the window's
//	                                lens, as nodes a view draws; a slot
//	                                runs it
//	studio close window pane        a pane of the frame ended: its hook
//	                                runs it, and when the pane was the
//	                                face, fm, the frame ends with it
//	studio slots [frame]            the frame's slots as JSON, for a
//	                                host that lays them out itself
//	studio pane slot                the pane of a slot of this window's
//	                                frame: where e, o and p -a land
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"sys/figure"
	"sys/outline"
	"sys/state"
	"sys/tty"
	"sys/view"
	"sys/watch"
)

func main() {
	args, do := os.Args[1:], openVerb
	if len(args) > 0 {
		if v, ok := verbs[args[0]]; ok {
			args, do = args[1:], v
		}
	}
	check(do(args))
}

<<others>>
