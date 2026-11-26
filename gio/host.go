// the host: studio's frame, its slots each a program at the other end of a
// pipe speaking -s, or one pane at the end of this program's own pipe, as
// window is
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

// slot is a frame's slot as `studio slots` says it: its name, its weight
// among the slots beside it, what runs there, none for a shell, and the
// slots it splits into, the other way
type slot struct {
	Name   string
	Weight int
	Run    string
	Kids   []slot
}

type pane struct {
	app        // the program as the window holds it
	cmd        *exec.Cmd
	name       string
	cols, high int
	first, off int  // a text's scroll: the first line and column shown
	gone       bool // its program ended, or a shell, which has none
}

// news is one thing for the window's thread: a pane's message or its
// end, or, with no pane, a change to cur
type news struct {
	p     *pane
	m     *in
	ended bool
}

type studio struct {
	frame []slot
	panes map[string]*pane // by slot name
	order []*pane          // the frame's slots, as it reads
	eye   int              // order's index of the one keys go to
	face  *pane            // fm's, whose end ends the frame
	// the window's unsaved file, where ed names work it holds unsaved
	unsaved string
	events  chan news
	// the frame's folder, and its programs' world, outside tmux: what a
	// look says cur in, as they hear it (look)
	dir string
	env []string
}

// start lays studio/frame out on dir: each slot that runs something runs
// it with -s, its view this window, told its slot ($STUDIO), and a shell
// slot waits, for a shell needs a terminal. cur is watched for the
// programs that follow it
func start(dir string) (*studio, error) {
	out, err := exec.Command("studio", "slots").Output()
	if err != nil {
		return nil, err
	}
	s := &studio{panes: map[string]*pane{}, events: make(chan news, 64)}
	if err := json.Unmarshal(out, &s.frame); err != nil {
		return nil, err
	}
	// outside tmux, so that cur is the window none's and no pane's
	env := slices.DeleteFunc(os.Environ(), func(v string) bool {
		return strings.HasPrefix(v, "TMUX")
	})
	s.dir, s.env = dir, env
	curf := state.Path("cur", "none")
	s.unsaved = state.Path("unsaved", "none")
	if err := os.MkdirAll(filepath.Dir(curf), 0o755); err != nil {
		return nil, err
	}
	wt, err := watch.New(filepath.Dir(curf))
	if err != nil {
		return nil, err
	}
	for _, sl := range leaves(s.frame) {
		p := &pane{app: newApp(nil), name: sl.Name, gone: sl.Run == ""}
		s.panes[sl.Name], s.order = p, append(s.order, p)
		if p.gone {
			continue
		}
		args := append(strings.Fields(sl.Run), "-s")
		if args[0] == "fm" {
			s.face = p
		}
		c := exec.Command(args[0], args[1:]...)
		c.Env = append(slices.Clip(env), "STUDIO="+sl.Name)
		c.Dir, c.Stderr = dir, os.Stderr
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
		p.cmd, p.out = c, json.NewEncoder(to)
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
	s.focus(0) // a program starts without the eye, saying no cur
	return s, nil
}

// leaves are the frame's slots that hold a pane, as the frame reads
func leaves(slots []slot) []slot {
	var out []slot
	for _, s := range slots {
		if len(s.Kids) == 0 {
			out = append(out, s)
		}
		out = append(out, leaves(s.Kids)...)
	}
	return out
}

// focus gives the eye to the first pane from i on with a program, and
// tells each program whether it has it; fm says its row to cur only
// then. A lone pane is its program's own, eye and all
func (s *studio) focus(i int) {
	for range s.order {
		if !s.order[i%len(s.order)].gone {
			break
		}
		i++
	}
	s.eye = i % len(s.order)
	if len(s.order) < 2 {
		return
	}
	for j, p := range s.order {
		if !p.gone && p.out != nil {
			name := tty.KBlur
			if j == s.eye {
				name = tty.KFocus
			}
			p.key(name, "")
		}
	}
}

// handle does one piece of news, on the window's thread: what to put on the
// clipboard, if anything, and whether the window is done
func (s *studio) handle(e news) (copied string, done bool) {
	switch {
	case e.p == nil: // cur changed: a size wakes a program to chase it
		for _, p := range s.order {
			if !p.gone && p.cols > 0 {
				size := object{"w": p.cols, "h": p.high}
				p.send("size", size)
			}
		}
	case e.ended:
		// the face's end, fm's, ends the frame: the window goes, and
		// with it each program's view, so each ends. But while ed
		// names work it holds unsaved, which fm's second q takes
		// away, fm ended some other way, as a crash ends it, and the
		// frame stays for ed to save. Another's end is its own, the
		// eye moving on from it
		e.p.gone = true
		if e.p == s.face && state.Read(s.unsaved) == "" {
			return "", true
		}
		if s.order[s.eye] == e.p {
			s.focus(s.eye)
		}
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

// piped is the view at the other end of this program's own pipe: a
// frame of one pane
func piped(r io.Reader, w io.Writer) *studio {
	p := &pane{app: newApp(json.NewEncoder(w)), name: "window"}
	s := &studio{frame: []slot{{Name: "window", Weight: 100}},
		panes: map[string]*pane{"window": p}, order: []*pane{p},
		face: p, events: make(chan news, 64)}
	go func() {
		feed(r, func(m in) { s.events <- news{p: p, m: &m} })
		s.events <- news{p: p, ended: true}
	}()
	return s
}
