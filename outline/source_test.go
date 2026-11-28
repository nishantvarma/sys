package outline

import (
	"strings"
	"testing"
)

// TestGo: one file carried through: its package clause and doc; a struct,
// a field's tag left out and its comment under it, and a method of it
// found later; an interface and its method; a group of consts, a part
// holding its specs; a var whose value runs on, its own line alone; and a
// func whose params wrap, as one line, its doc of two lines one paragraph.
func TestGo(t *testing.T) {
	in := `// Package p is a sample.
package p

import "io"

// Node is a thing, as ` + "`design/view`" + ` says.
type Node struct {
	ID   string ` + "`json:\"id\"`" + ` // its address
	Kids []*Node
}

// Size is how much.
func (n *Node) Size() int {
	return len(n.Kids)
}

// Reader reads.
type Reader interface {
	Read(p []byte) (int, error) // reads p
}

const (
	a = 1 // one
	b = 2
)

var big = map[string]int{
	"x": 1,
}

// Stream is a view at the other
// end of a stream.
func Stream(r io.Reader,
	w io.Writer) *Node {
	return nil
}
`
	want := "title package p   :1\n" +
		"  para Package p is a sample.   :1\n" +
		"type type *Node* struct   :6\n" +
		"  para Node is a thing, as `design/view` says.   :6\n" +
		"  value *ID* string   :8\n" +
		"    para its address   :8\n" +
		"  value *Kids* []*Node   :9\n" +
		"  func func (n *Node) *Size*() int   :12\n" +
		"    para Size is how much.   :12\n" +
		"type type *Reader* interface   :17\n" +
		"  para Reader reads.   :17\n" +
		"  func *Read*(p []byte) (int, error)   :19\n" +
		"    para reads p   :19\n" +
		"value const   :22\n" +
		"  value *a* = 1   :23\n" +
		"    para one   :23\n" +
		"  value *b* = 2   :24\n" +
		"value var *big* = map[string]int   :27\n" +
		"func func *Stream*(r io.Reader, w io.Writer) *Node   :31\n" +
		"  para Stream is a view at the other end of a stream.   :31\n"
	ps, err := Go(strings.NewReader(in), "p.go")
	if err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	parted(&b, ps, 0)
	if b.String() != want {
		t.Errorf("go:\n%s\nwant:\n%s", b.String(), want)
	}
}

// TestChunk: a tangle chunk, a declaration with no package clause, reads
// as one, each line its own number still; and what is no Go says why.
func TestChunk(t *testing.T) {
	in := "// lens is one way to see a thing.\n" +
		"type lens struct {\n" +
		"\tname string\n" +
		"}\n"
	want := "type type *lens* struct   :1\n" +
		"  para lens is one way to see a thing.   :1\n" +
		"  value *name* string   :3\n"
	ps, err := Go(strings.NewReader(in), "01-lens.go")
	if err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	parted(&b, ps, 0)
	if b.String() != want {
		t.Errorf("a chunk:\n%s\nwant:\n%s", b.String(), want)
	}
	if _, err := Go(strings.NewReader("func {"), "x.go"); err == nil {
		t.Error("what is no Go read as Go")
	}
}

// TestBroken: source that does not parse reads as far as it does, a
// method of a generic type under it, and says where it stops, at the
// line the file numbers, a chunk's too.
func TestBroken(t *testing.T) {
	in := "package p\n\ntype M[K comparable, V any] struct{}\n\n" +
		"func (m *M[K, V]) Get() {}\n\nfunc F() {\n"
	ps, err := Go(strings.NewReader(in), "p.go")
	var b strings.Builder
	parted(&b, ps, 0)
	want := "title package p   :1\n" +
		"type type *M*[K comparable, V any] struct   :3\n" +
		"  func func (m *M[K, V]) *Get*()   :5\n" +
		"func func *F*()   :7\n"
	if b.String() != want {
		t.Errorf("broken:\n%s\nwant:\n%s", b.String(), want)
	}
	if err == nil || !strings.HasPrefix(err.Error(), "p.go:7:") {
		t.Errorf("the error says %v, not where it stops, line 7", err)
	}
	_, err = Go(strings.NewReader("func f() {\n\tx :=\n}\n"), "c.go")
	if err == nil || !strings.HasPrefix(err.Error(), "c.go:3:") {
		t.Errorf("a chunk's error says %v, not its line 3", err)
	}
}
