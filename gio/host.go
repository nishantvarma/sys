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
	"strconv"
	"strings"

	"sys/state"
	"sys/tty"
	"sys/watch"
)

// slot is a frame's slot as `studio slots` says it: its name, its weight
// among the slots beside it, what runs there, none for a shell, the slots
// it splits into, the other way, its tabs, when it has more than one, the
// first the one it shows, and whether it is the frame's panel, the
// terminals across its bottom (`design/lens`)
type slot struct {
	Name   string
	Weight int
	Run    string
	Kids   []slot
	Tabs   []slot
	Panel  bool
}

type pane struct {
	app        // the program as the window holds it
	cmd        *exec.Cmd
	name       string
	tab        string // the tab it shows, by name; its slot's, for none
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
	panes map[string]*pane   // by slot name, the tab each shows
	tabs  map[string][]*pane // by slot name, its tabs, in their order
	order []*pane            // the frame's slots, as it reads
	// the frame's panel, by its slot's name; whether it is hidden; and
	// the window's files that say so, and what studio tab asks, which
	// the window follows as it follows cur, and the ask done last
	panel, panelf string
	hidden        bool
	tabf, lastAsk string
	eye           int   // order's index of the one keys go to
	face          *pane // fm's, whose end ends the frame
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
	s := &studio{panes: map[string]*pane{}, tabs: map[string][]*pane{},
		events: make(chan news, 64)}
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
	s.tabf = state.Path("tab", "none")
	s.panelf = state.Path("panel", "none")
	var dirs []string
	for _, f := range []string{curf, s.tabf, s.panelf} {
		if err := os.MkdirAll(filepath.Dir(f), 0o755); err != nil {
			return nil, err
		}
		dirs = append(dirs, filepath.Dir(f))
	}
	wt, err := watch.New(dirs...)
	if err != nil {
		return nil, err
	}
	s.lastAsk = state.Read(s.tabf) // what was asked before is done
	s.hidden = state.Read(s.panelf) == "hidden"
	for _, sl := range leaves(s.frame) {
		if sl.Panel {
			s.panel = sl.Name
		}
		tabs := sl.Tabs
		if len(tabs) == 0 {
			tabs = []slot{{Name: sl.Name, Run: sl.Run}}
		}
		for i, tb := range tabs {
			p := &pane{app: newApp(nil), name: sl.Name,
				tab: tb.Name, gone: tb.Run == ""}
			s.tabs[sl.Name] = append(s.tabs[sl.Name], p)
			if i == 0 {
				s.panes[sl.Name] = p
				s.order = append(s.order, p)
			}
			if !p.gone {
				if err := s.launch(p, tb.Run); err != nil {
					return nil, err
				}
			}
		}
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
	case e.p == nil:
		// a window's file changed: a size wakes a program to chase
		// cur; a tab asked shows; the panel hides or shows
		for _, p := range s.order {
			if !p.gone && p.cols > 0 {
				size := object{"w": p.cols, "h": p.high}
				p.send("size", size)
			}
		}
		s.asked()
		s.hidden = state.Read(s.panelf) == "hidden"
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

// launch starts what runs in a pane, run, at the other end of a pipe
// speaking -s, its view this window, told its slot ($STUDIO); fm's pane is
// the frame's face
func (s *studio) launch(p *pane, run string) error {
	args := append(strings.Fields(run), "-s")
	if args[0] == "fm" {
		s.face = p
	}
	c := exec.Command(args[0], args[1:]...)
	c.Env = append(slices.Clip(s.env), "STUDIO="+p.name)
	c.Dir, c.Stderr = s.dir, os.Stderr
	to, err := c.StdinPipe()
	if err != nil {
		return err
	}
	from, err := c.StdoutPipe()
	if err != nil {
		return err
	}
	if err := c.Start(); err != nil {
		return err
	}
	p.cmd, p.out = c, json.NewEncoder(to)
	go func() {
		feed(from, func(m in) { s.events <- news{p: p, m: &m} })
		c.Wait()
		s.events <- news{p: p, ended: true}
	}()
	return nil
}

// turn shows a slot's tab as studio tab asks it
// (`studio/5-verbs/22-ask.go`), or a press on the slot's bar names it: the
// next or the one before, round, one by its place from 1, or one by its
// name
func (s *studio) turn(slot, word string) {
	tabs := s.tabs[slot]
	if len(tabs) < 2 {
		return
	}
	i, j := slices.Index(tabs, s.panes[slot]), -1
	switch word {
	case "next":
		j = (i + 1) % len(tabs)
	case "prev":
		j = (i + len(tabs) - 1) % len(tabs)
	default:
		if n, err := strconv.Atoi(word); err == nil {
			j = n - 1
		}
		for k, p := range tabs {
			if p.tab == word {
				j = k
			}
		}
	}
	if j >= 0 && j < len(tabs) {
		s.showTab(slot, tabs[j])
	}
}

// showTab puts a slot's tab p in its place: the order holds it where the
// tab shown was, the eye going with it when that one had it, and the one
// shown before hears it lost the eye; its size is said again as it is
// drawn (slot)
func (s *studio) showTab(slot string, p *pane) {
	old := s.panes[slot]
	if old == p {
		return
	}
	s.panes[slot] = p
	i := slices.Index(s.order, old)
	s.order[i] = p
	p.cols, p.high = 0, 0
	if !old.gone && old.out != nil {
		old.key(tty.KBlur, "")
	}
	if s.eye == i {
		s.focus(i)
	} else if !p.gone && p.out != nil {
		p.key(tty.KBlur, "")
	}
}

// asked does what studio tab asked last, once: the window's tab file says
// the slot, the ask and when, so that one asked twice is done twice
func (s *studio) asked() {
	line := state.Read(s.tabf)
	if line == "" || line == s.lastAsk {
		return
	}
	s.lastAsk = line
	if f := strings.Fields(line); len(f) >= 2 {
		s.turn(f[0], f[1])
	}
}

// togglePanel hides the frame's panel, its room going to the slots above,
// or shows it again, alt-t's; the window's panel file keeps which, for it
// is the window's state, and studio panel says it there too
// (`studio/5-verbs/24-askpanel.go`)
func (s *studio) togglePanel() {
	if s.panel == "" {
		return
	}
	s.hidden = !s.hidden
	said := ""
	if s.hidden {
		said = "hidden"
	}
	state.Write(s.panelf, said)
}
