// the mouse in the window, the three buttons as in any pane
// (`design/mouse`): a left press picks (pressedAt), and a move with it
// held drags a text's selection; a right press looks at what it falls on,
// a middle one hands it to the shell, and the wheel moves
package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"sys/outline"
	"sys/tty"
)

// drag is a move with the left button held to cell x, y of a text: the
// selection's head goes there, as a drag in ed on the terminal
func (p *pane) drag(x, y int) {
	if id, at := p.spot(x, y); id != "" && p.isText() {
		p.send("key", outline.Key{Name: tty.KDrag, Node: id, At: at})
	}
}

// clicked is what a right or a middle press at cell x, y of a text names:
// the word there, or the selection it falls in (outline.Clicked)
func (p *pane) clicked(x, y int) string {
	i, at := p.lineAt(x, y)
	if i < 0 {
		return ""
	}
	return outline.Clicked(p.tree, i, at)
}

// look is a right press at cell x, y of p. On a row of the view it is
// enter, the program's own open, with the word it fell on in a doc (as
// sys/view's Word says it), and on the head's path a step there; on a
// text its word, or the selection, opens, as p opens it (lookat).
func (v *window) look(p *pane, x, y int) {
	if !p.isText() {
		id := p.v.At(x, y)
		switch {
		case id == "":
		case y == 0:
			p.key("", id)
		default:
			p.send("key", outline.Key{Name: tty.KEnter, Node: id,
				Text: p.v.Word(x, y)})
		}
		return
	}
	v.lookat(p.clicked(x, y), v.s.dir)
}

// lookat opens what text names, said in folder dir, as a right press on
// a text opens it. In the frame, whose ed and fm follow cur, saying a
// file is opening it: c says it, as the frame's programs hear it, and
// the lens view's source tab shows for a file, with the eye, and fm for
// a folder, as e and o take them in tmux (turn); what is no file p opens
// beside, as anywhere.
func (v *window) lookat(text, dir string) {
	if text == "" {
		return
	}
	if !v.frame {
		plumb(text)
		return
	}
	c := exec.Command("c", text)
	c.Dir, c.Env = dir, v.s.env
	out, _ := c.Output()
	fi, err := os.Stat(strings.TrimSpace(string(out)))
	switch {
	case err == nil && fi.IsDir():
		v.s.turn("", []string{"fm"})
	case err == nil:
		v.s.turn("", []string{"source"})
	default:
		plumb(text)
	}
}

// do is a middle press at cell x, y of p: a text's word or selection, or
// a row's address, handed to the shell at its prompt (handon)
func (v *window) do(p *pane, x, y int) {
	text := ""
	if p.isText() {
		text = p.clicked(x, y)
	} else if id := p.v.At(x, y); y > 0 {
		if path, _ := outline.Address(id); filepath.IsAbs(path) {
			text = id
		}
	}
	if v.frame {
		v.s.handon(text, v.s.dir)
	} else {
		plumb(text, "-a")
	}
}

// handon hands text, said in folder dir, to the shell at its prompt, as p
// -a does (`sh/p:/the shell takes it at its prompt/`): what c makes of
// it, a phrase as it is, after a space, in the panel's terminal, shown
// again if hidden, the eye going there, and waiting there for Return;
// but only while its shell waits at its prompt, for a program there
// would read it as its keys (prompting). With no terminal left, p -a's
// own way, to tmux's pane marked, or the last one
func (s *studio) handon(text, dir string) {
	t := s.panes[s.panel]
	switch {
	case text == "":
	case t == nil || t.term == nil || t.gone:
		plumb(text, "-a")
	case !s.prompting(t):
	default:
		if !strings.Contains(text, " ") {
			c := exec.Command("c", text)
			c.Dir, c.Env = dir, s.env
			if out, err := c.Output(); err == nil && len(out) > 0 {
				text = strings.TrimSpace(string(out))
			}
		}
		s.show(s.panel, t)
		t.term.Paste(" " + text)
	}
}

// say says text to cur from folder dir, as a left press on a word does in
// tmux (p -c): fm and ed, following cur, go where it names
func (s *studio) say(text, dir string) {
	if text == "" {
		return
	}
	c := exec.Command("c", text)
	c.Dir, c.Env = dir, s.env
	if c.Start() == nil {
		go c.Wait()
	}
}

// wheel is the wheel turned, dy past 0 down: a text's cursor three lines,
// as on the terminal, and the view's cursor a row
func (p *pane) wheel(dy float32) {
	name := tty.KDown
	if dy < 0 {
		name = tty.KUp
	}
	if p.isText() {
		for range 3 {
			p.key(name, "")
		}
		return
	}
	if ev, theirs := p.v.Handle(tty.Key{Name: name}); theirs {
		p.send("key", ev)
	}
}

// plumb hands text to p, as ed's view does (`ed/4-view/29-plumb.go`),
// and does not wait
func plumb(text string, mode ...string) {
	if text == "" {
		return
	}
	c := exec.Command("p", append(mode, "-")...)
	c.Stdin = strings.NewReader(text)
	if c.Start() == nil {
		go c.Wait()
	}
}
