package test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
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
