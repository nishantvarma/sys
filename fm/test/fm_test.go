// Package test drives fm as a host does, through its stream (fm -s):
// keys go in as JSON lines and what it shows comes back, so a check reads
// what a person would see, and the folder it changed.
package test

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"sys/outline"
	"sys/vt"
)

// built is fm made from this tree, as ./build makes it: tangled, then
// compiled, into a folder of the test's own.
func built(t *testing.T) string {
	dir := t.TempDir()
	tangle := filepath.Join(dir, "tangle")
	run := func(in string, args ...string) {
		c := exec.Command(args[0], args[1:]...)
		c.Dir = in
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("%s: %v\n%s", strings.Join(args, " "), err,
				out)
		}
	}
	run("../../tangle", "go", "build", "-o", tangle, "main.go")
	run("..", tangle)
	run("..", "go", "build", "-o", filepath.Join(dir, "fm"), "main.go")
	return filepath.Join(dir, "fm")
}

// ended is fm's word once it has ended, and shows nothing more.
const ended = "(ended)"

// fm is fm -s on dir, env the last of its environment, and a way to press
// a key on a row, by its name, or on the cursor's: it answers with what
// fm shows next, its rows in order, the cursor's marked >, and its word. A
// key that starts with = answers a prompt with the rest, typed.
func fm(t *testing.T, dir string, env ...string) func(key,
	row string) (string, string) {
	return started(t, []string{"-s", dir}, dir, env...)
}

// started is fm run with args, its rows named from dir, as fm is. A cd
// tells zoxide, which runs on after fm and would write in a home the test
// is taking away: here it is a stand-in that writes nothing.
func started(t *testing.T, args []string, dir string, env ...string) func(
	key, row string) (string, string) {
	press, _ := begun(t, args, dir, false, env...)
	return press
}

// begun is started, and with cons fm's error and controlling terminal a
// pty, its console, as gio's frame gives one (vt.Start): the pty's other
// end comes back too, where what fm runs there is read and typed at. Its
// end hangs fm up first, for a child of fm's may wait there.
func begun(t *testing.T, args []string, dir string, cons bool,
	env ...string) (func(key, row string) (string, string), *os.File) {
	c := exec.Command(built(t), args...)
	bin := t.TempDir()
	stand := []byte("#!/bin/sh\n")
	os.WriteFile(filepath.Join(bin, "zoxide"), stand, 0o755)
	c.Env = append(os.Environ(), "HOME="+t.TempDir(), "TMUX_PANE=",
		"PATH="+bin+":"+os.Getenv("PATH"))
	c.Env = append(c.Env, env...)
	in, _ := c.StdinPipe()
	out, _ := c.StdoutPipe()
	var pty *os.File
	if cons {
		f, err := vt.Start(c, 80, 24)
		if err != nil {
			t.Fatal(err)
		}
		pty = f
	} else if err := c.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if pty != nil {
			pty.Close()
		}
		in.Close()
		c.Wait()
	})
	dec := json.NewDecoder(bufio.NewReader(out))
	enc := json.NewEncoder(in)
	shown := func() (string, string) {
		var m struct {
			Draw struct {
				Tree   outline.Node
				Status string
			}
		}
		switch err := dec.Decode(&m); {
		case err == io.EOF:
			return "", ended
		case err != nil:
			t.Fatal(err)
		}
		var rows []string
		for _, k := range m.Draw.Tree.Kids {
			mark := ""
			if k.Cur {
				mark = ">"
			}
			rows = append(rows, mark+k.Text)
		}
		return strings.Join(rows, " "), m.Draw.Status
	}
	shown()
	return func(key, row string) (string, string) {
		if typed, ok := strings.CutPrefix(key, "="); ok {
			enc.Encode(map[string]map[string]any{
				"answer": {"text": typed, "ok": true}})
			return shown()
		}
		k := outline.Key{Name: key}
		if row != "" {
			k.Node = filepath.Join(dir, row)
		}
		enc.Encode(map[string]outline.Key{"key": k})
		return shown()
	}, pty
}

// TestReorder: numbered rows stand first and in their order, a folder
// among files; shift-down on 01-a swaps its number with 02-b's, the
// cursor going with it, and again with 03-c's; past the last numbered
// row it says so and renames nothing; shift-up undoes it.
func TestReorder(t *testing.T) {
	dir := t.TempDir()
	for _, f := range []string{"01-a", "03-c", "README", "zz"} {
		os.WriteFile(filepath.Join(dir, f), nil, 0o644)
	}
	os.Mkdir(filepath.Join(dir, "02-b"), 0o755)
	os.Mkdir(filepath.Join(dir, "plain"), 0o755)
	press := fm(t, dir)
	for _, c := range []struct {
		key, row, want, said string
	}{
		{"", "01-a", ". .. >01-a 02-b 03-c plain README zz", ""},
		{"S-DOWN", "", ". .. 01-b >02-a 03-c plain README zz", ""},
		{"S-DOWN", "", ". .. 01-b 02-c >03-a plain README zz", ""},
		{"S-DOWN", "", ". .. 01-b 02-c >03-a plain README zz",
			"not numbered: plain"},
		{"S-UP", "", ". .. 01-b >02-a 03-c plain README zz", ""},
		{"S-UP", "", ". .. >01-a 02-b 03-c plain README zz", ""},
	} {
		got, said := press(c.key, c.row)
		if got != c.want || said != c.said {
			t.Errorf("%s on %q: %q %q, want %q %q", c.key, c.row,
				got, said, c.want, c.said)
		}
	}
	if fi, err := os.Stat(filepath.Join(dir, "02-b")); err != nil ||
		!fi.IsDir() {
		t.Errorf("02-b is not the folder it was: %v", err)
	}
	names, _ := filepath.Glob(filepath.Join(dir, ".*"))
	if len(names) > 0 {
		t.Errorf("left behind: %s", fmt.Sprint(names))
	}
}

// . and .. stand at the top, the folder's own, whether dotfiles show or
// not, and the cursor starts past them; . shows the dotfiles, among the
// rest by their rank and name.
func TestDots(t *testing.T) {
	dir := t.TempDir()
	for _, f := range []string{"a", ".b"} {
		os.WriteFile(filepath.Join(dir, f), nil, 0o644)
	}
	press := fm(t, dir)
	if got, _ := press("", ""); got != ". .. >a" {
		t.Errorf("dotfiles hidden: %q, not . and .. over a", got)
	}
	if got, _ := press(".", ""); got != ". .. .b >a" {
		t.Errorf("dotfiles shown: %q", got)
	}
}

// Two rows whose names differ only in their numbers swap through a third
// name, so neither is lost; a name either would take that another entry
// holds, a broken link too, stops the swap and renames nothing.
func TestReorderNames(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "01-x"), []byte("first"), 0o644)
	os.WriteFile(filepath.Join(dir, "02-x"), []byte("second"), 0o644)
	press := fm(t, dir)
	press("", "01-x")
	if got, _ := press("S-DOWN", ""); got != ". .. 01-x >02-x" {
		t.Errorf("same names: %q, want the cursor on 02-x", got)
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "02-x")); string(b) !=
		"first" {
		t.Errorf("02-x holds %q, want what 01-x held", b)
	}
	for _, link := range []bool{false, true} {
		dir := t.TempDir()
		for _, f := range []string{"01-z", "02-a"} {
			os.WriteFile(filepath.Join(dir, f), nil, 0o644)
		}
		taken := filepath.Join(dir, "02-z")
		if link {
			os.Symlink("nowhere", taken)
		} else {
			os.WriteFile(taken, nil, 0o644)
		}
		press := fm(t, dir)
		press("", "01-z")
		if _, said := press("S-DOWN", ""); said != "taken: 02-z" {
			t.Errorf("link %v: said %q, want taken: 02-z", link,
				said)
		}
		names, _ := filepath.Glob(filepath.Join(dir, "*"))
		if len(names) != 3 {
			t.Errorf("link %v: %v, want the three as they were",
				link, names)
		}
		if _, err := os.Lstat(taken); err != nil {
			t.Errorf("link %v: 02-z is gone", link)
		}
	}
}

// In a frame's slot fm's q ends the frame: while ed names a file it holds
// unsaved, the first q says so and ends nothing, a key between takes it
// back, and a second q lets the work go, taking the name away, and ends
// fm; with nothing unsaved one q ends it, and outside a frame the file is
// neither read nor touched.
func TestQuit(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a"), nil, 0o644)
	home := t.TempDir()
	mark := filepath.Join(home, ".config", "sys", "unsaved", "none")
	os.MkdirAll(filepath.Dir(mark), 0o755)
	os.WriteFile(mark, []byte("/x/note.go\n"), 0o644)
	asks := "unsaved in ed: note.go — q again to discard"
	press := fm(t, dir, "HOME="+home, "STUDIO=fm")
	for i, k := range []string{"q", "j", "q", "q"} {
		want := []string{asks, "", asks, ended}[i]
		if _, said := press(k, ""); said != want {
			t.Errorf("%d: %s says %q, not %q", i+1, k, said, want)
		}
	}
	if _, err := os.Stat(mark); err == nil {
		t.Error("the second q let the work go, and its name stays")
	}
	os.WriteFile(mark, []byte("/x/note.go\n"), 0o644)
	press = fm(t, dir, "HOME="+home)
	if _, said := press("q", ""); said != ended {
		t.Errorf("outside a frame, q says %q and ends nothing", said)
	}
	if _, err := os.Stat(mark); err != nil {
		t.Error("outside a frame, q took the name of ed's work away")
	}
	os.Remove(mark)
	press = fm(t, dir, "HOME="+home, "STUDIO=fm")
	if _, said := press("q", ""); said != ended {
		t.Errorf("nothing unsaved, and q says %q", said)
	}
}

// Backspace goes up a folder, and a key on a folder above, as a click on
// a part of the head's path names one, goes up there; either lands on
// the folder it came up from, not on the first row.
func TestUp(t *testing.T) {
	dir := t.TempDir()
	deep := filepath.Join(dir, "a", "b")
	for _, d := range []string{deep, filepath.Join(dir, "0"),
		filepath.Join(dir, "a", "0")} {
		os.MkdirAll(d, 0o755)
	}
	press := fm(t, deep)
	if rows, _ := press("BACKSPACE", ""); rows != ". .. 0 >b" {
		t.Errorf("backspace in a/b shows %q, not a's", rows)
	}
	if rows, _ := press("", "../.."); rows != ". .. 0 >a" {
		t.Errorf("a key on the folder two up shows %q, not its",
			rows)
	}
}

// With -f fm goes where cur is said: cur is a row, so a file's folder
// with the cursor on it, and a folder's folder on the folder; cur naming
// the row already, by a link's path, moves nothing, nor does cur that
// names nothing. Given the eye, fm says its row only where cur does not
// name it, so cur keeps its line. Without -f, fm stays.
func TestFollow(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	os.MkdirAll(sub, 0o755)
	for _, f := range []string{"a", "sub/b", "sub/c"} {
		os.WriteFile(filepath.Join(dir, f), nil, 0o644)
	}
	os.Symlink("sub", filepath.Join(dir, "link"))
	home := t.TempDir()
	cur := filepath.Join(home, ".config", "sys", "cur", "none")
	os.MkdirAll(filepath.Dir(cur), 0o755)
	say := func(addr string) {
		os.WriteFile(cur, []byte(addr+"\n"), 0o644)
	}
	said := func() string {
		b, _ := os.ReadFile(cur)
		return strings.TrimSpace(string(b))
	}
	press := started(t, []string{"-s", "-f", dir}, dir, "HOME="+home)
	press("", "a")
	// a wake that moves fm draws; one that moves nothing draws nothing,
	// so a key, which fm draws after, shows where it stayed
	for _, c := range []struct{ cur, key, want string }{
		{filepath.Join(sub, "c") + ":3", "RESIZE", ". .. b >c"},
		{sub, "RESIZE", ". .. link >sub a"},
		{filepath.Join(dir, "link"), "", ". .. link >sub a"},
		{filepath.Join(dir, "gone"), "", ". .. link >sub a"},
	} {
		say(c.cur)
		if rows, _ := press(c.key, ""); rows != c.want {
			t.Errorf("cur %s: %q, want %q", c.cur, rows, c.want)
		}
	}
	press("FOCUS", "")
	if got := said(); got != sub {
		t.Errorf("the eye on sub, cur naming nothing: cur is %q", got)
	}
	say(filepath.Join(sub, "c") + ":3")
	press("BLUR", "")
	press("FOCUS", "")
	if got := said(); got != filepath.Join(sub, "c")+":3" {
		t.Errorf("the eye on c, cur naming c:3: cur is %q", got)
	}
	say(filepath.Join(sub, "b"))
	still := fm(t, dir, "HOME="+home)
	if rows, _ := still("", ""); strings.Contains(rows, ">b") {
		t.Errorf("without -f, fm went to cur: %q", rows)
	}
}

// A paste onto a name taken asks for another: nothing typed leaves the
// clip pending and the taken file as it was, a new name is pasted to, and
// only the taken name typed again writes over it.
func TestClash(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	os.Mkdir(sub, 0o755)
	os.WriteFile(filepath.Join(dir, "a"), []byte("old"), 0o644)
	os.WriteFile(filepath.Join(sub, "a"), []byte("new"), 0o644)
	holds := func(name string) string {
		b, _ := os.ReadFile(filepath.Join(dir, name))
		return string(b)
	}
	press := fm(t, sub)
	press("", "a")
	press("c", "")
	press("BACKSPACE", "")
	for _, c := range []struct{ typed, name, want string }{
		{"", "a", "old"},
		{"b", "b", "new"},
		{"a", "a", "new"},
	} {
		press("p", "")
		press("="+c.typed, "")
		if got := holds(c.name); got != c.want {
			t.Errorf("typed %q: %s holds %q, not %q", c.typed,
				c.name, got, c.want)
		}
	}
}

// A prompt asked when the pane is resized is painted again on its new
// bottom line, where it used to stay on the old one: fm on a terminal of
// tmux's own, its / asking, the window then made taller, for a shorter
// one tmux scrolls to keep the prompt's line in sight.
func TestPromptResize(t *testing.T) {
	if _, err := exec.LookPath("tmux"); err != nil {
		t.Skip("no tmux")
	}
	fm, dir := built(t), t.TempDir()
	os.WriteFile(filepath.Join(dir, "a"), nil, 0o644)
	bin := t.TempDir()
	os.WriteFile(filepath.Join(bin, "zoxide"), []byte("#!/bin/sh\n"),
		0o755)
	sock := fmt.Sprintf("sys-fm-%d", os.Getpid())
	tm := func(args ...string) string {
		t.Helper()
		c := exec.Command("tmux", append([]string{"-L", sock, "-f",
			"/dev/null"}, args...)...)
		c.Env = append(os.Environ(), "TMUX=", "TMUX_PANE=",
			"HOME="+t.TempDir(), "PATH="+bin+":"+os.Getenv("PATH"))
		out, err := c.CombinedOutput()
		if err != nil {
			t.Fatalf("tmux %s: %v: %s", strings.Join(args, " "),
				err, out)
		}
		return strings.TrimRight(string(out), "\n ")
	}
	tm("new-session", "-d", "-x", "80", "-y", "12", fm+" "+dir)
	t.Cleanup(func() { tm("kill-server") })
	bottom := func() string { // the pane's last row, blank or not
		h, _ := strconv.Atoi(tm("display", "-p", "#{pane_height}"))
		at := strconv.Itoa(h - 1)
		return strings.TrimSpace(tm("capture-pane", "-p", "-S", at,
			"-E", at))
	}
	asks := func() bool { return strings.HasPrefix(bottom(), "/") }
	until := func(ok func() bool) bool {
		for end := time.Now().Add(5 * time.Second); time.Now().
			Before(end); time.Sleep(20 * time.Millisecond) {
			if ok() {
				return true
			}
		}
		return false
	}
	if !until(func() bool {
		return strings.Contains(tm("capture-pane", "-p"), " a")
	}) {
		t.Fatal("fm never drew its folder")
	}
	tm("send-keys", "/")
	if !until(asks) {
		t.Fatalf("/ asks nothing: the bottom line is %q", bottom())
	}
	tm("resize-window", "-y", "24")
	if !until(asks) {
		t.Errorf("resized, the bottom line is %q, not the prompt",
			bottom())
	}
}

// TestWindowed: off the terminal, in a window, ? shows fm's keys as rows,
// each a key and what it does, which esc closes; and, given no terminal,
// a finder's pick, whose fzy needs one, says so and moves nothing.
// Stand-ins play the finder and zoxide.
func TestWindowed(t *testing.T) {
	dir := t.TempDir()
	for _, f := range []string{"alpha", "beta"} {
		os.WriteFile(filepath.Join(dir, f), nil, 0o644)
	}
	stand := t.TempDir()
	for name, body := range map[string]string{
		"f":      "echo beta", // the finder: one hit
		"zoxide": "",
	} {
		os.WriteFile(filepath.Join(stand, name),
			[]byte("#!/bin/sh\n"+body+"\n"), 0o755)
	}
	press := fm(t, dir, "PATH="+stand+":"+os.Getenv("PATH"))
	if got, _ := press("?", ""); !regexp.MustCompile(`\?\s+help`).
		MatchString(got) {
		t.Errorf("? showed %q, not fm's keys", got)
	}
	if got, _ := press("ESC", ""); got != ". .. >alpha beta" {
		t.Errorf("esc left %q, not fm's rows", got)
	}
	if got, said := press("f", ""); got != ". .. >alpha beta" ||
		said != "pick: needs the terminal" {
		t.Errorf("a pick off the terminal: %q %q", got, said)
	}
}

// TestConsole: given a console, a pty for its error and controlling
// terminal, as gio's frame gives one, fm runs what it runs there as on a
// terminal of its own, and draws on neither: a pick's fzy shows the
// finder's lines there and takes its choice from there, and fm lands on
// the line chosen; F's lint writes there and pause holds it, and fm draws
// again only once a key is typed there; and ! hands the row to run.
// Stand-ins play the finder, fzy, lint, pause, run and zoxide.
func TestConsole(t *testing.T) {
	dir := t.TempDir()
	for _, f := range []string{"alpha", "beta"} {
		os.WriteFile(filepath.Join(dir, f), nil, 0o644)
	}
	stand := t.TempDir()
	for name, body := range map[string]string{
		"f": "echo beta",
		"fzy": `list=$(cat); echo "fzy has $list" >/dev/tty; ` +
			`read pick </dev/tty; echo "$pick"`,
		"lint":   `echo "lint $*"`,
		"pause":  "echo continue; read x",
		"run":    `echo "run $*"`,
		"zoxide": "",
	} {
		os.WriteFile(filepath.Join(stand, name),
			[]byte("#!/bin/sh\n"+body+"\n"), 0o755)
	}
	press, pty := begun(t, []string{"-s", dir}, dir, true,
		"PATH="+stand+":"+os.Getenv("PATH"))
	seen := screen(pty)
	// later presses a key whose act waits on the console, and is fm's
	// rows once it draws again
	later := func(key string) <-chan string {
		drew := make(chan string, 1)
		go func() {
			rows, _ := press(key, "")
			drew <- rows
		}()
		return drew
	}
	drawn := func(drew <-chan string) string {
		select {
		case rows := <-drew:
			return rows
		case <-time.After(5 * time.Second):
			t.Fatal("fm never drew again")
		}
		return ""
	}
	picked := later("f")
	if !seen("fzy has beta") {
		t.Fatal("f's fzy showed nothing on the console")
	}
	pty.Write([]byte("beta\n"))
	if got := drawn(picked); got != ". .. alpha >beta" {
		t.Errorf("beta picked on the console, and fm shows %q", got)
	}
	held := later("F")
	if !seen("lint -f "+filepath.Join(dir, "beta")) || !seen("continue") {
		t.Fatal("F's lint and pause showed nothing on the console")
	}
	select {
	case rows := <-held:
		t.Fatalf("pause held nothing: fm drew %q", rows)
	case <-time.After(200 * time.Millisecond):
	}
	pty.Write([]byte("\n"))
	if got := drawn(held); got != ". .. alpha >beta" {
		t.Errorf("a key on the console, and fm shows %q", got)
	}
	handed := later("!")
	if !seen("run " + filepath.Join(dir, "beta")) {
		t.Error("! handed run nothing on the console")
	}
	if got := drawn(handed); got != ". .. alpha >beta" {
		t.Errorf("! on beta, and fm shows %q", got)
	}
}

// screen reads what is written on a console, from its pty's other end:
// seen waits, five seconds at most, for text to be among it
func screen(pty *os.File) (seen func(text string) bool) {
	var mu sync.Mutex
	var all strings.Builder
	go func() {
		for b := make([]byte, 4096); ; {
			n, err := pty.Read(b)
			mu.Lock()
			all.Write(b[:n])
			mu.Unlock()
			if err != nil {
				return
			}
		}
	}()
	return func(text string) bool {
		for end := time.Now().Add(5 * time.Second); time.Now().
			Before(end); time.Sleep(20 * time.Millisecond) {
			mu.Lock()
			ok := strings.Contains(all.String(), text)
			mu.Unlock()
			if ok {
				return true
			}
		}
		return false
	}
}

// TestOwnWord: fm -f hears its own word to cur as its own. Each move is
// said to cur, and comes back as cur once the cursor may have gone on:
// three moves in a row land on the third, no word taking it back.
func TestOwnWord(t *testing.T) {
	dir, home := t.TempDir(), t.TempDir()
	for _, f := range []string{"a", "b", "c"} {
		os.WriteFile(filepath.Join(dir, f), nil, 0o644)
	}
	os.MkdirAll(filepath.Join(home, ".config", "sys", "cur"), 0o755)
	press := started(t, []string{"-s", "-f", dir}, dir, "HOME="+home)
	press("FOCUS", "")
	press("", "a")
	press("", "b")
	if rows, _ := press("", "c"); !strings.Contains(rows, ">c") {
		t.Errorf("moved to a, b and c, fm is on %q", rows)
	}
}

// TestChanged: fm reads its folder again when it changes, not at each
// key: a file another program makes there shows by the next key.
func TestChanged(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a"), nil, 0o644)
	press := fm(t, dir)
	press("", "a")
	os.WriteFile(filepath.Join(dir, "b"), nil, 0o644)
	rows, end := "", time.Now().Add(3*time.Second)
	for !strings.Contains(rows, " b") && time.Now().Before(end) {
		rows, _ = press("", "a")
		time.Sleep(20 * time.Millisecond)
	}
	if !strings.Contains(rows, " b") {
		t.Errorf("b made beside a, fm shows %q", rows)
	}
}

// TestStill: a wake that changes nothing draws nothing. fm -f, woken by
// its own word to cur coming back, as a host wakes it with its size,
// draws once a move, not twice; a wake with a new size draws.
func TestStill(t *testing.T) {
	dir, home := t.TempDir(), t.TempDir()
	os.WriteFile(filepath.Join(dir, "a"), nil, 0o644)
	os.MkdirAll(filepath.Join(home, ".config", "sys", "cur"), 0o755)
	c := exec.Command(built(t), "-s", "-f", dir)
	c.Env = append(os.Environ(), "HOME="+home, "TMUX_PANE=")
	in, _ := c.StdinPipe()
	out, _ := c.StdoutPipe()
	if err := c.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { in.Close(); c.Wait() })
	lines := make(chan string)
	go func() {
		sc := bufio.NewScanner(out)
		sc.Buffer(nil, 1<<20)
		for sc.Scan() {
			lines <- sc.Text()
		}
		close(lines)
	}()
	draws := func() int { // the draws fm makes till it is quiet
		n := 0
		for {
			select {
			case l, ok := <-lines:
				if !ok {
					return n
				}
				if strings.HasPrefix(l, `{"draw"`) {
					n++
				}
			case <-time.After(300 * time.Millisecond):
				return n
			}
		}
	}
	draws()
	send := func(m string) { io.WriteString(in, m+"\n") }
	a := filepath.Join(dir, "a")
	for _, c := range []struct {
		msg  string
		want int
	}{
		{`{"key":{"name":"FOCUS"}}`, 1},
		{`{"key":{"name":"","node":"` + a + `"}}`, 1},
		{`{"size":{"w":80,"h":24}}`, 0}, // the size it has
		{`{"size":{"w":100,"h":30}}`, 1},
	} {
		send(c.msg)
		if n := draws(); n != c.want {
			t.Errorf("%s: %d draws, want %d", c.msg, n, c.want)
		}
	}
}
