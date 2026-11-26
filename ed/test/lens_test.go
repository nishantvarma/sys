package test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"sys/outline"
)

// TestLens: on the terminal ed's bar names its lenses, and Tab turns from
// source to the outline, the note's items as rows under the bar; there j
// moves to the next item, and enter goes back to source on its line. A
// press on the bar's render turns there, the note drawn as its parts and
// messages, and Tab from the last lens is source again. The lenses are
// the terminal view's, so tmux holds ed, on a server of the test's own,
// and a press is the bytes a terminal sends for one.
func TestLens(t *testing.T) {
	if _, err := exec.LookPath("tmux"); err != nil {
		t.Skip("no tmux")
	}
	ed := built(t)
	note := filepath.Join(t.TempDir(), "note")
	os.WriteFile(note, []byte("t\nup\n  a\n    → b  go\nlow\n  b\n"),
		0o644)
	sock := fmt.Sprintf("sys-ed-%d", os.Getpid())
	tm := func(args ...string) string {
		t.Helper()
		c := exec.Command("tmux", append([]string{"-L", sock, "-f",
			"/dev/null"}, args...)...)
		c.Env = append(os.Environ(), "TMUX=", "HOME="+t.TempDir())
		out, err := c.CombinedOutput()
		if err != nil {
			t.Fatalf("tmux %s: %v: %s", strings.Join(args, " "),
				err, out)
		}
		return strings.TrimRight(string(out), "\n")
	}
	tm("new-session", "-d", "-x", "80", "-y", "24", ed+" "+note)
	t.Cleanup(func() { tm("kill-server") })
	screen := func() []string {
		return strings.Split(tm("capture-pane", "-p"), "\n")
	}
	// until is whether the screen comes to pass ok within a few seconds
	until := func(why string, ok func(s []string) bool) {
		t.Helper()
		for end := time.Now().Add(5 * time.Second); ; {
			s := screen()
			if len(s) > 0 && ok(s) {
				return
			}
			if time.Now().After(end) {
				t.Fatalf("%s:\n%s", why, strings.Join(s, "\n"))
			}
			time.Sleep(20 * time.Millisecond)
		}
	}
	has := func(s []string, text string) bool {
		return strings.Contains(strings.Join(s, "\n"), text)
	}
	bar := func(s []string) bool {
		return strings.Contains(s[0], "source  outline  render")
	}
	// ruled is whether the line under the bar is a rule, as under fm's
	// head, in every lens
	ruled := func(s []string) bool {
		return len(s) > 1 && s[1] != "" &&
			strings.Trim(s[1], "─") == ""
	}
	at := func(s []string, where string) bool {
		return strings.HasSuffix(strings.TrimSpace(s[len(s)-1]), where)
	}

	until("no bar over the source", func(s []string) bool {
		return bar(s) && ruled(s) && has(s, "up") && at(s, "1:1")
	})
	tm("send-keys", "Tab")
	until("Tab showed no outline", func(s []string) bool {
		return bar(s) && ruled(s) && has(s, "rows  thumbs") &&
			has(s, "    → b  go")
	})
	tm("send-keys", "j", "j")
	tm("send-keys", "Enter")
	until("enter on the third item left no source on its line",
		func(s []string) bool {
			return bar(s) && !has(s, "rows  thumbs") &&
				at(s, "3:1")
		})
	x := strings.Index(screen()[0], "render") + 1 // its column, from 1
	tm("send-keys", "-l", fmt.Sprintf("\x1b[<0;%d;1M\x1b[<0;%d;1m", x, x))
	until("a press on the bar's render drew no picture",
		func(s []string) bool { return bar(s) && has(s, "▼") })
	tm("send-keys", "Tab")
	until("Tab from the render left no source", func(s []string) bool {
		return bar(s) && !has(s, "▼") && at(s, "3:1")
	})
}

// TestLensStream: the lens is what ed draws, so every view of it shows
// the same: the text names its lenses, Tab draws the note's items as the
// outline, its name in ed's state, and a key on an item puts the cursor
// on its line; there letters type nothing, enter is source on the line,
// a press on a lens's name shows it, and esc is source again.
func TestLensStream(t *testing.T) {
	file, press := ed(t, "t\nup\n  a\n    → b  go\nlow\n  b\n")
	s := keys(press, "TAB")
	if strings.Join(s.tags, " ") != "source outline render" ||
		s.kind != "outline" || !slices.Contains(s.state, "outline") {
		t.Fatalf("Tab drew %q %v %v, not the outline", s.kind, s.state,
			s.tags)
	}
	if got := strings.Join(s.lines, "|"); got !=
		"t|up|  a|    → b  go|low|  b" {
		t.Errorf("the outline's rows: %q", got)
	}
	press(outline.Key{Node: file + ":5"})
	if s = keys(press, "x", "d"); s.kind != "outline" ||
		strings.Join(s.lines, "|") != "t|up|  a|    → b  go|low|  b" {
		t.Errorf("x d in the outline did %q %q", s.kind, s.lines)
	}
	if s = keys(press, "ENTER"); s.kind != "text" || s.lines[4] != "low" {
		t.Fatalf("enter drew %q, not source", s.kind)
	}
	if s = keys(press, "ESC", "d"); s.lines[4] != "ow" {
		t.Errorf("source after enter was not on line 5: %q",
			s.lines[4])
	}
	s = press(outline.Key{Name: "LENS", Text: "render"})
	if s.kind != "picture" || !slices.Contains(s.state, "render") {
		t.Errorf("a press on render drew %q %v", s.kind, s.state)
	}
	if s = keys(press, "ESC"); s.kind != "text" {
		t.Errorf("esc in the render drew %q, not source", s.kind)
	}
}
