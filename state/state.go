// Package state is the files a tmux window keeps in ~/.config/sys: its cur,
// the address the last verb, view or click was about, which `sh/c` keeps and
// fm, ed and the lens say; its lens, which studio's verbs set; and its
// unsaved, the file ed in the window's frame holds unsaved, which fm's q
// reads. Each file holds one line, written whole.
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
// is here.
func Pane() (window string, eye bool) {
	p := os.Getenv("TMUX_PANE")
	if p == "" {
		return "none", true
	}
	out, err := exec.Command("tmux", "display", "-t", p, "-p",
		"#{window_id} #{pane_active}").Output()
	f := strings.Fields(string(out))
	if err != nil || len(f) != 2 {
		return "none", true
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

// Write puts line in a state file: to a file beside it, then renamed onto
// it, so that a reader never meets half of one.
func Write(p, line string) error {
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	tmp := p + "." + strconv.Itoa(os.Getpid())
	if err := os.WriteFile(tmp, []byte(line+"\n"), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}
