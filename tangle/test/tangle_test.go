// Package test runs tangle, built from this tree as ./build builds it,
// on trees of the test's own.
package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// A ref that names no chunk stops the tangle, saying where it is, where
// it used to be left in main.go to surface as a compile error; a ref to
// what is no name, <<%s>> in a format, is text, and so is an escaped one.
func TestRefs(t *testing.T) {
	tangle := filepath.Join(t.TempDir(), "tangle")
	build := exec.Command("go", "build", "-o", tangle, "main.go")
	build.Dir = ".."
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	run := func(root string) (string, string, error) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "@main.go"), []byte(root),
			0o644)
		os.WriteFile(filepath.Join(dir, "01-x.go"),
			[]byte("var x = 1\n"), 0o644)
		c := exec.Command(tangle)
		c.Dir = dir
		out, err := c.CombinedOutput()
		made, _ := os.ReadFile(filepath.Join(dir, "main.go"))
		return string(out), string(made), err
	}
	out, _, err := run("package main\n\nfunc main() {\n\t<<missing>>\n}\n")
	if err == nil || !strings.Contains(out, "@main.go: <<missing>>") {
		t.Errorf("a ref to no chunk: %v, %q", err, out)
	}
	root := "package main\n\nimport \"fmt\"\n\n<<x>>\n\n" +
		"func main() { fmt.Printf(\"<<%s>> \\<<y>>\", \"z\") }\n"
	out, made, err := run(root)
	if err != nil || !strings.Contains(made, `"<<%s>> <<y>>"`) {
		t.Errorf("text that only looks like a ref: %v, %q\n%s", err,
			out, made)
	}
}
