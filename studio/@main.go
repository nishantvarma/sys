// studio is sys's frame: the lens view, the thing at fm's cursor seen its
// ways, its source in ed and the rest by the lens (`design/lens`); fm;
// and the terminals, each running as it is. The frame is a note, read by
// its indent (outline.Text): its slots in the order they take room, laid
// out as a golden spiral (layout), and under a slot its tabs, which take
// turns in it. The same note lays out as tmux panes in the terminal and
// as panes on a page. cur, which fm writes as its cursor moves, is all the
// panes share, and a slot's or a tab's name is how a program reaches it.
//
//	studio [-n] [frame]              the frame in a tmux window: this
//	                                 one, around the shell alone in it,
//	                                 else a new one; -n prints its tmux
//	                                 commands, runs none
//	studio page [frame]              the frame as a page, on stdout
//	studio picture note [page.html]  a note's picture as a page, on the
//	                                 skeleton given, or sys's own
//	studio tab [next | prev | number | name | close]
//	                                 turn the tabs of the slot it runs
//	                                 in, show one or end one; bare, say
//	                                 them
//	studio tab new [command...]      a terminal, a tab of the panel: a
//	                                 shell, or command
//	studio panel [hide | show]       hide the panel, the terminals, or
//	                                 show it again; bare, turn it
//	studio send slot text...         paste text into a slot, then enter
//	studio read slot                 a slot's screen, as text
//	studio show [lens] [-w | -s]     the lens: cur seen one way, the
//	                                 lens named, as nodes a view draws,
//	                                 a tab of the lens view; with none,
//	                                 every way, Tab turning them
//	studio close window pane         a pane of the frame ended: its hook
//	                                 runs it, and when the pane was the
//	                                 face, fm, the frame ends with it
//	studio slots [frame]             the frame's slots as JSON, for a
//	                                 host that lays them out itself
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

	"sys/outline"
	"sys/picture"
	"sys/state"
	"sys/tabs"
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
