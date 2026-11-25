// Package test drives fm as a host does, through its stream (fm -s):
// keys go in as JSON lines and what it shows comes back, so a check reads
// what a person would see, and the folder it changed.
package test

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"sys/outline"
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

// fm is fm -s on dir, and a way to press a key on a row, by its name, or
// on the cursor's: it answers with what fm shows next, its rows in order,
// the cursor's marked >, and its word.
func fm(t *testing.T, dir string) func(key, row string) (string, string) {
	c := exec.Command(built(t), "-s", dir)
	c.Env = append(os.Environ(), "HOME="+t.TempDir(), "TMUX_PANE=")
	in, _ := c.StdinPipe()
	out, _ := c.StdoutPipe()
	if err := c.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { in.Close(); c.Wait() })
	dec := json.NewDecoder(bufio.NewReader(out))
	enc := json.NewEncoder(in)
	shown := func() (string, string) {
		var m struct {
			Draw struct {
				Tree   outline.Node
				Status string
			}
		}
		if err := dec.Decode(&m); err != nil {
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
		k := outline.Key{Name: key}
		if row != "" {
			k.Node = filepath.Join(dir, row)
		}
		enc.Encode(map[string]outline.Key{"key": k})
		return shown()
	}
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
		{"", "01-a", ">01-a 02-b 03-c plain README zz", ""},
		{"S-DOWN", "", "01-b >02-a 03-c plain README zz", ""},
		{"S-DOWN", "", "01-b 02-c >03-a plain README zz", ""},
		{"S-DOWN", "", "01-b 02-c >03-a plain README zz",
			"not numbered: plain"},
		{"S-UP", "", "01-b >02-a 03-c plain README zz", ""},
		{"S-UP", "", ">01-a 02-b 03-c plain README zz", ""},
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

// Two rows whose names differ only in their numbers swap through a third
// name, so neither is lost; a name either would take that another entry
// holds, a broken link too, stops the swap and renames nothing.
func TestReorderNames(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "01-x"), []byte("first"), 0o644)
	os.WriteFile(filepath.Join(dir, "02-x"), []byte("second"), 0o644)
	press := fm(t, dir)
	press("", "01-x")
	if got, _ := press("S-DOWN", ""); got != "01-x >02-x" {
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
