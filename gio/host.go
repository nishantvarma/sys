// the host: studio's frame, its slots each a program at the other end of a
// pipe speaking -s, with a console behind its view for what it runs, or a
// shell in a terminal of the window's own (term.go); or one pane, at the
// end of this program's own pipe, as window is
package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"sys/state"
	"sys/tabs"
	"sys/tty"
	"sys/vt"
	"sys/watch"
)

// slot is a frame's slot as `studio slots` says it: its name, its weight
// among the slots beside it, what runs there, none for a shell, the slots
// it splits into, the other way, its tabs, when it has more than one, the
// first the one it shows, and whether it is the frame's panel, the
// terminals across its bottom (`design/lens`), or its face, fm's, whose
// end ends the frame
type slot struct {
	Name   string
	Weight int
	Run    string
	Kids   []slot
	Tabs   []slot
	Panel  bool
	Face   bool
}

type pane struct {
	app // the program as the window holds it
	cmd *exec.Cmd
	// the screen the pane shows: a shell's (term.go), or a program's
	// console while what the program runs writes there (handle)
	term       *vt.Term
	cons       *vt.Term // a program's console, behind its view (launch)
	pty        *os.File
	name       string
	tab        string // the tab it shows, by name; its slot's, for none
	cols, high int
	first, off int  // a text's scroll: the first line and column shown
	gone       bool // what ran in it ended
}

// news is one thing for the window's thread: a pane's message, what its
// shell wrote, or its end; or, with no pane, a change to cur, or what
// studio asked of the window (asks)
type news struct {
	p     *pane
	m     *in
	out   []byte
	ended bool
	asked bool
}

type studio struct {
	frame []slot
	panes map[string]*pane   // by slot name, the tab each shows
	tabs  map[string][]*pane // by slot name, its tabs, in their order
	order []*pane            // the frame's slots, as it reads
	// the frame's panel, by its slot's name, and whether it is hidden,
	// which the window alone keeps; and the window's ask file, what
	// studio asks of it, and how far into it the window has read
	panel  string
	hidden bool
	askf   string
	read   int64
	eye    int   // order's index of the one keys go to
	face   *pane // fm's, whose end ends the frame
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
// slot a shell in a terminal of the window's own (openterm); each is told
// the frame's window, none, $STUDIO_WINDOW, which is how studio knows a
// program in the window's frame (`studio/5-verbs/22-ask/02-ask.go`). cur
// is watched for the programs that follow it, and the ask file apart, for
// what studio asks
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
	// outside tmux, so that cur is the window none's and no pane's, in
	// the frame's folder, as $PWD says it
	env := slices.DeleteFunc(os.Environ(), func(v string) bool {
		return strings.HasPrefix(v, "TMUX") ||
			strings.HasPrefix(v, "STUDIO") ||
			strings.HasPrefix(v, "PWD=")
	})
	s.dir, s.env = dir, append(env, "PWD="+dir, "STUDIO_WINDOW=none")
	s.unsaved = state.Path("unsaved", "none")
	s.askf = state.Path("ask", "none")
	if fi, err := os.Stat(s.askf); err == nil {
		s.read = fi.Size() // what was asked before is done
	}
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
				tab: tb.Name}
			s.tabs[sl.Name] = append(s.tabs[sl.Name], p)
			if i == 0 {
				s.panes[sl.Name] = p
				s.order = append(s.order, p)
			}
			if i == 0 && sl.Face {
				s.face = p
			}
			if tb.Run == "" {
				err = s.openterm(p, nil)
			} else {
				err = s.launch(p, tb.Run)
			}
			if err != nil {
				return nil, err
			}
		}
	}
	for _, w := range []struct {
		file  string
		asked bool
	}{{state.Path("cur", "none"), false}, {s.askf, true}} {
		dir := filepath.Dir(w.file)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
		wt, err := watch.New(dir)
		if err != nil {
			return nil, err
		}
		go func(asked bool) {
			for range wt.On {
				s.events <- news{asked: asked}
			}
		}(w.asked)
	}
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

// focus gives the eye to the first pane from i on that can take it, one
// with a program, in sight, and tells each program whether it has it
// (eye). A lone pane is its program's own, eye and all
func (s *studio) focus(i int) {
	for range s.order {
		if p := s.order[i%len(s.order)]; !p.gone && !s.away(p) {
			break
		}
		i++
	}
	s.eye = i % len(s.order)
	if len(s.order) < 2 {
		return
	}
	for j, p := range s.order {
		p.eye(j == s.eye)
	}
}

// focusOn gives the eye to pane p, when the frame shows it and it has it
// not already
func (s *studio) focusOn(p *pane) {
	if i := slices.Index(s.order, p); i >= 0 && i != s.eye {
		s.focus(i)
	}
}

// away is whether pane p is out of sight, the panel's, hidden
func (s *studio) away(p *pane) bool {
	return s.hidden && p.name == s.panel
}

// eye tells what runs in a pane whether it has the eye: fm says its row
// to cur only then, and a terminal tells a program that asked (vt.Focus)
func (p *pane) eye(has bool) {
	switch {
	case p.gone:
	case p.term != nil:
		p.term.Focus(has)
	case p.out != nil:
		name := tty.KBlur
		if has {
			name = tty.KFocus
		}
		p.key(name, "")
	}
}

// handle does one piece of news, on the window's thread: what to put on the
// clipboard, if anything, and whether the window is done
func (s *studio) handle(e news) (copied string, done bool) {
	switch {
	case e.asked:
		s.asks()
	case e.p == nil:
		// cur changed: a size wakes a program to chase it, a
		// terminal's shell none
		for _, p := range s.order {
			if !p.gone && p.out != nil && p.cols > 0 {
				size := object{"w": p.cols, "h": p.high}
				p.send("size", size)
			}
		}
	case e.out != nil:
		// what writes on a program's console while the program acts on
		// a key, not drawn since, is what it runs: the console comes
		// to the front, as a child takes fm's own pane in tmux, until
		// the program draws again (lower); written at rest it waits
		// behind
		t := e.p.term
		if e.p.cons != nil {
			t = e.p.cons
			if e.p.acting {
				e.p.term = t
			}
		}
		t.Write(e.out)
	case e.ended:
		// the face's end, fm's, ends the frame: the window goes, and
		// with it each program's view, so each ends. But while ed
		// names work it holds unsaved, which fm's second q takes
		// away, fm ended some other way, as a crash ends it, and the
		// frame stays for ed to save. Another's end is its own: a
		// tab gives its place to the next (drop), and the eye moves
		// on from it
		e.p.gone = true
		if s.drop(e.p) {
			break
		}
		if e.p == s.face && state.Read(s.unsaved) == "" {
			return "", true
		}
		if s.order[s.eye] == e.p {
			s.focus(s.eye)
		}
	case e.m.Draw != nil:
		d := e.m.Draw
		e.p.tree, e.p.status, e.p.state = d.Tree, d.Status, d.State
		e.p.acting = false
		if e.p.term != nil && e.p.term == e.p.cons {
			s.lower(e.p)
		}
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
// speaking -s, its view this window, told its slot ($STUDIO), and on a
// pty of the pane's own, as tmux starts a pane's program on one: its
// error and its controlling terminal, its console, where what it runs
// runs (`design/view:/console/`). The console waits behind the
// program's view, and comes to the front while what it runs writes
// there (handle)
func (s *studio) launch(p *pane, run string) error {
	args := append(strings.Fields(run), "-s")
	c := exec.Command(args[0], args[1:]...)
	c.Env = append(slices.Clip(s.env), "STUDIO="+p.name)
	c.Dir = s.dir
	to, err := c.StdinPipe()
	if err != nil {
		return err
	}
	from, err := c.StdoutPipe()
	if err != nil {
		return err
	}
	cons, _, err := s.terminal(p, c)
	if err != nil {
		return err
	}
	p.cmd, p.out, p.cons = c, json.NewEncoder(to), cons
	go func() {
		feed(from, func(m in) { s.events <- news{p: p, m: &m} })
		c.Wait()
		p.pty.Close() // which ends the console's Read under way
		s.events <- news{p: p, ended: true}
	}()
	return nil
}

// lower puts a program's console behind its view again, as the program
// draws once what it ran is done: its size is said again as it is drawn,
// for the console took the slot's changes meanwhile (slot), and it hears
// whether it has the eye, which its console heard instead (eye)
func (s *studio) lower(p *pane) {
	p.term, p.cols, p.high = nil, 0, 0
	if len(s.order) > 1 {
		p.eye(s.order[s.eye] == p)
	}
}

// turn does to the frame's tabs what studio tab asks of a slot's, or a
// press on a tab line's button names: what each word means is tabs.Turn's
// (`tabs/tabs.go:/func Turn/`), as in tmux, carried out here, the eye
// going to the tab shown (show); an error has no line to be said on here
func (s *studio) turn(slot string, words []string) {
	act, err := tabs.Turn(s.slots(), slot, words)
	if err != nil {
		return
	}
	switch act.Do {
	case tabs.New:
		s.newterm(act.Run)
	case tabs.Close:
		s.closetab(act.Slot)
	case tabs.Show:
		s.show(act.Slot, s.tabs[act.Slot][act.Tab])
	}
}

// slots is the frame's slots as the tab words see them (tabs.Turn): each
// one's tabs by name, the one in its place, and whether it is hidden
func (s *studio) slots() []tabs.Slot {
	var out []tabs.Slot
	for _, sl := range leaves(s.frame) {
		in, shown := s.tabs[sl.Name], s.panes[sl.Name]
		t := tabs.Slot{Name: sl.Name, Shown: slices.Index(in, shown),
			Hidden: s.away(shown)}
		for _, p := range in {
			t.Tabs = append(t.Tabs, p.tab)
		}
		out = append(out, t)
	}
	return out
}

// show puts a slot's tab p in its place with the eye on it, as studio
// shows one in tmux, the panel with it when it was hidden
func (s *studio) show(slot string, p *pane) {
	if slot == s.panel {
		s.hidden = false
	}
	s.showTab(slot, p)
	s.focusOn(p)
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
	old.eye(false)
	if s.eye == i {
		s.focus(i)
	} else {
		p.eye(false)
	}
}

// drop takes a tab that ended out of its slot's tabs, the next taking its
// place when it was the one shown, as a tab in tmux gives its place
// (`studio/2-term/22-died.go`); false for a slot's last, which stays,
// ended
func (s *studio) drop(p *pane) bool {
	tabs := s.tabs[p.name]
	i := slices.Index(tabs, p)
	if i < 0 || len(tabs) < 2 {
		return false
	}
	tabs = slices.Delete(tabs, i, i+1)
	s.tabs[p.name] = tabs
	if s.panes[p.name] == p {
		s.showTab(p.name, tabs[i%len(tabs)])
	}
	return true
}

// asked is a line of the window's ask file, as studio adds it
// (`studio/5-verbs/22-ask/01-asked.go`): the verb, tab or panel, the slot
// asked of, "." for the one with the eye, and the verb's words
type asked struct {
	Verb  string
	Slot  string
	Words []string
}

// asks does what studio asked of the window since it last read the ask
// file, a line an ask, in order (asked): a slot's tabs turned (turn), the
// panel hidden, shown, or, with no word, turned (hide). A line not yet
// whole waits for the rest
func (s *studio) asks() {
	b, err := os.ReadFile(s.askf)
	if err != nil {
		return
	}
	if int64(len(b)) < s.read {
		s.read = 0 // the file was made again
	}
	for b = b[s.read:]; ; {
		i := bytes.IndexByte(b, '\n')
		if i < 0 {
			return
		}
		var a asked
		err := json.Unmarshal(b[:i], &a)
		b, s.read = b[i+1:], s.read+int64(i+1)
		if err != nil {
			continue
		}
		if a.Slot == "." {
			a.Slot = s.order[s.eye].name
		}
		switch a.Verb {
		case "tab":
			s.turn(a.Slot, a.Words)
		case "panel":
			hidden := !s.hidden
			if len(a.Words) == 1 {
				hidden = a.Words[0] == "hide"
			}
			s.hide(hidden)
		}
	}
}

// hide hides the frame's panel, its room going to the slots above, or
// shows it again, as alt-t and studio panel ask: the eye leaves it as it
// hides, and comes to it as it shows, as in tmux
func (s *studio) hide(hidden bool) {
	if s.panel == "" || s.hidden == hidden {
		return
	}
	s.hidden = hidden
	p := s.panes[s.panel]
	switch {
	case !hidden:
		s.focusOn(p)
	case s.order[s.eye] == p:
		s.focus(s.eye + 1)
	}
}
