// Package state is the files a tmux window keeps in ~/.config/sys: its cur,
// the address the last verb, view or click was about, which `sh/c` keeps and
// fm, ed and the lens say; its unsaved, the file ed in the window's frame
// holds unsaved, which fm's q reads; its text, what ed holds of that file,
// which the lens reads in place of the file; and, for a frame's host down
// a pipe, gio's window, which has no tmux to ask, its tab, what studio tab
// asks of a slot's tabs, and its panel, hidden or not. Each file is written
// whole, and each but text holds one line.
package state

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Pane is the pane a program runs in: its window, which state is kept per,
// and whether it is the pane being looked at, the eye. One ask, at start;
// a program keeps the eye after that by focus keys. $TMUX_PANE names a pane
// and not its window, and an untargeted tmux answers about the session's
// current window and not ours. Outside tmux the window is none and the eye
// is here. A frame's tab out of sight, parked in a window of its own, is
// still its frame's, as $STUDIO_WINDOW names it, and is no eye until it
// is shown there; that name too is taken out of the environment.
func Pane() (window string, eye bool) {
	p := os.Getenv("TMUX_PANE")
	frame := os.Getenv("STUDIO_WINDOW")
	os.Unsetenv("STUDIO_WINDOW")
	if p == "" {
		return "none", true
	}
	out, err := exec.Command("tmux", "display", "-t", p, "-p",
		"#{window_id} #{pane_active}").Output()
	f := strings.Fields(string(out))
	if err != nil || len(f) != 2 {
		return "none", true
	}
	if frame != "" && frame != f[0] {
		return frame, false
	}
	return f[0], f[1] == "1"
}

// Slot is the frame's slot a program runs in, as $STUDIO names it: a
// frame's host, studio in tmux or gio, sets it for each program it starts
// in a slot. Empty outside a frame. It is taken out of the environment, so
// that what the program starts is in no frame.
func Slot() string {
	s := os.Getenv("STUDIO")
	os.Unsetenv("STUDIO")
	return s
}

// Path is where a window's state of a kind is kept: Path("cur", "@3") is
// ~/.config/sys/cur/@3, the file `sh/c` keeps for window @3.
func Path(kind, window string) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "sys", kind, window)
}

// Read is a state file's line, empty when there is none yet.
func Read(p string) string {
	b, err := os.ReadFile(p)
	if err != nil {
		return ""
	}
	s, _, _ := strings.Cut(string(b), "\n")
	return strings.TrimSpace(s)
}

// Write puts line in a state file, written whole (Put).
func Write(p, line string) error {
	return Put(p, []byte(line+"\n"))
}

// Put puts b in a state file: to a file beside it, then renamed onto it,
// so that a reader never meets half of one. ed's text is put so, lines
// and all.
func Put(p string, b []byte) error {
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	tmp := p + "." + strconv.Itoa(os.Getpid())
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}
