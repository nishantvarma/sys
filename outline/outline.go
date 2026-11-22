// Package outline is the tree a view draws. A program says what to show as
// nodes; a view, on the terminal, in a window or on a page, says how, and
// hands back what was pressed. All of it is plain data, so a view can be
// at the other end of a pipe or a socket. The idea: todo/studio/outline.
package outline

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
)

// Node is one thing to show: what kind it is, what it says, and the nodes
// under it. id is how its program knows it, so a key can name it.
type Node struct {
	ID   string   `json:"id,omitempty"`
	Kind string   `json:"kind,omitempty"` // what a view picks its look by
	Text string   `json:"text,omitempty"`
	Tags []string `json:"tags,omitempty"` // short words beside it: [2]
	Cur  bool     `json:"cur,omitempty"`  // the cursor is on it
	Sel  bool     `json:"sel,omitempty"`  // it is selected
	Open bool     `json:"open,omitempty"` // its kids are here
	In   bool     `json:"in,omitempty"`   // the cursor is among its kids
	Kids []*Node  `json:"kids,omitempty"`
}

// Key is what the person did: a key's name, a letter or a name as tty
// spells it; the id of the node it was said on, as a click names one, else
// empty for the cursor's; and the text that came with it, as a paste brings.
type Key struct {
	Name string `json:"name"`
	Node string `json:"node,omitempty"`
	Text string `json:"text,omitempty"`
}

// View draws a tree and a line of status, and hands back what the person
// does: each key, and a line typed at a prompt. false: the view is gone,
// or the line was cancelled. The status is the program's word, what just
// happened or what it asks, and its state, a few short words about itself
// ([2], cut:3); the view adds its own, its layout and where the cursor is,
// and shows the line at the bottom.
type View interface {
	Draw(n *Node, status string, state []string)
	Key() (Key, bool)
	Ask(prompt, def string) (string, bool) // def: the line it starts from
	Size() (w, h int)                      // how much it shows at once
	Copy(text string)                      // text onto the clipboard
}

// msg is one line on the wire, a JSON object with one part set: draw, ask
// and copy go to the view; key, answer and size come back.
type msg struct {
	Draw   *draw   `json:"draw,omitempty"`
	Ask    *ask    `json:"ask,omitempty"`
	Copy   *clip   `json:"copy,omitempty"`
	Key    *Key    `json:"key,omitempty"`
	Answer *answer `json:"answer,omitempty"`
	Size   *size   `json:"size,omitempty"`
}

type draw struct {
	Tree   *Node    `json:"tree"`
	Status string   `json:"status"`
	State  []string `json:"state,omitempty"`
}

type ask struct {
	Prompt string `json:"prompt"`
	Def    string `json:"def"`
}

type clip struct {
	Text string `json:"text"`
}

type answer struct {
	Text string `json:"text"`
	OK   bool   `json:"ok"`
}

type size struct {
	W int `json:"w"`
	H int `json:"h"`
}

// Stream is a view at the other end of a byte stream, a pipe or a socket:
// a message a line, as JSON.
func Stream(r io.Reader, w io.Writer) View {
	return &stream{json.NewDecoder(r), json.NewEncoder(w), 80, 24}
}

type stream struct {
	in   *json.Decoder
	out  *json.Encoder
	w, h int
}

func (s *stream) Draw(n *Node, status string, state []string) {
	s.out.Encode(msg{Draw: &draw{n, status, state}})
}

func (s *stream) Copy(text string) {
	s.out.Encode(msg{Copy: &clip{text}})
}

// Window starts the program window and is the view at its other end: what
// to show goes down its input and keys come back up its output, a JSON
// line each. done closes its input, which ends it, and waits for it.
func Window() (v View, done func() error, err error) {
	c := exec.Command("window")
	in, err := c.StdinPipe()
	if err != nil {
		return nil, nil, err
	}
	out, err := c.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}
	c.Stderr = os.Stderr
	if err := c.Start(); err != nil {
		return nil, nil, err
	}
	done = func() error {
		in.Close()
		return c.Wait()
	}
	return Stream(out, in), done, nil
}

// Key reads the next key. A new size is a key too, RESIZE as tty names
// it, so the program draws again.
func (s *stream) Key() (Key, bool) {
	for {
		var m msg
		if s.in.Decode(&m) != nil {
			return Key{}, false
		}
		switch {
		case m.Key != nil:
			return *m.Key, true
		case m.Size != nil:
			s.w, s.h = m.Size.W, m.Size.H
			return Key{Name: "RESIZE"}, true
		}
	}
}

// Ask sends a prompt and waits for its answer; keys meanwhile are the
// view's own, typing the line.
func (s *stream) Ask(prompt, def string) (string, bool) {
	if s.out.Encode(msg{Ask: &ask{prompt, def}}) != nil {
		return "", false
	}
	for {
		var m msg
		if s.in.Decode(&m) != nil {
			return "", false
		}
		if m.Answer != nil {
			return m.Answer.Text, m.Answer.OK
		}
	}
}

func (s *stream) Size() (int, int) { return s.w, s.h }
