// the host, as in the go-gui poc: one pane at the end of this program's own
// pipe, as window is, or studio's fm and ed -f, each a process speaking -s
package main

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"sys/state"
	"sys/tty"
	"sys/watch"
)

type pane struct {
	app        // the model main.go has: rows, cursor, moves
	in         io.WriteCloser
	cmd        *exec.Cmd
	name       string
	cols, high int
	first, off int // ed's scroll: the first line and column shown
	gone       bool
}

// news is one thing for the window's thread: a pane's message or its
// end, or, with no pane, a change to cur
type news struct {
	p     *pane
	m     *in
	ended bool
}

type studio struct {
	panes  []*pane // fm, ed
	eye    int
	events chan news
}

// start runs fm and ed on dir, and watches cur for ed
func start(dir string) (*studio, error) {
	s := &studio{events: make(chan news, 64)}
	// outside tmux, so that cur is the window none's and no pane's
	env := slices.DeleteFunc(os.Environ(), func(v string) bool {
		return strings.HasPrefix(v, "TMUX")
	})
	curf := state.Path("cur", "none")
	if err := os.MkdirAll(filepath.Dir(curf), 0o755); err != nil {
		return nil, err
	}
	wt, err := watch.New(filepath.Dir(curf))
	if err != nil {
		return nil, err
	}
	programs := [][]string{{"fm", "-s", dir}, {"ed", "-s", "-f"}}
	for _, args := range programs {
		c := exec.Command(args[0], args[1:]...)
		c.Env, c.Dir, c.Stderr = env, dir, os.Stderr
		to, err := c.StdinPipe()
		if err != nil {
			return nil, err
		}
		from, err := c.StdoutPipe()
		if err != nil {
			return nil, err
		}
		if err := c.Start(); err != nil {
			return nil, err
		}
		p := &pane{name: args[0], cmd: c, in: to}
		p.out = json.NewEncoder(to)
		s.panes = append(s.panes, p)
		go func() {
			feed(
				from,
				func(m in) { s.events <- news{p: p, m: &m} },
			)
			c.Wait()
			s.events <- news{p: p, ended: true}
		}()
	}
	go func() {
		for range wt.On {
			s.events <- news{}
		}
	}()
	s.focus(
		0,
	) // fm starts without the eye, and says no cur until it has it
	return s, nil
}

// focus gives pane i the eye. Only fm is told: ed asks for no focus keys
func (s *studio) focus(i int) {
	s.eye = i
	name := tty.KBlur
	if i == 0 {
		name = tty.KFocus
	}
	if len(s.panes) > 1 { // a lone view is its program's own, eye and all
		s.panes[0].key(name, "")
	}
}

// handle does one piece of news, on the window's thread: what to put on the
// clipboard, if anything, and whether the window is done
func (s *studio) handle(e news) (copied string, done bool) {
	switch {
	case e.p == nil: // cur changed: a size wakes ed, which then chases
		if ed := s.panes[1]; ed.cols > 0 {
			ed.send("size", object{"w": ed.cols, "h": ed.high})
		}
	case e.ended:
		// one q ends the frame, each slot ending itself, so that ed
		// keeps what is unsaved
		e.p.gone = true
		live := false
		for i, p := range s.panes {
			if !p.gone {
				live, s.eye = true, i
				p.key(tty.KEsc, "")
				p.key("q", "")
			}
		}
		return "", !live
	case e.m.Draw != nil:
		d := e.m.Draw
		e.p.tree, e.p.status, e.p.state = d.Tree, d.Status, d.State
		e.p.layout()
	case e.m.Ask != nil:
		ask := e.m.Ask
		e.p.asking, e.p.prompt, e.p.answer = true, ask.Prompt, ask.Def
	case e.m.Copy != nil:
		return e.m.Copy.Text, false
	}
	return "", false
}

// piped is the view at the other end of this program's own pipe
func piped(r io.Reader, w io.Writer) *studio {
	s := &studio{events: make(chan news, 64)}
	p := &pane{name: "window"}
	p.out = json.NewEncoder(w)
	s.panes = []*pane{p}
	go func() {
		feed(r, func(m in) { s.events <- news{p: p, m: &m} })
		s.events <- news{p: p, ended: true}
	}()
	return s
}
