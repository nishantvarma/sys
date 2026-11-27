// a terminal in the window: a frame's shell slot runs a shell on a pty,
// its screen st's, as sys/vt keeps it, a row a strip; main.go draws it
// and hands it the keys and the mouse, as st hands them
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"sys/outline"
	"sys/vt"
)

// palette is the first 16 colours as the st here has them, its
// config.h's colorname, in red, green and blue; its own are defaultfg on
// defaultbg, and its cursor defaultcs
var palette = [16][3]uint8{
	{0x12, 0x12, 0x12}, {0xfd, 0x5e, 0x53}, {0x50, 0xc8, 0x78},
	{0xe6, 0xdb, 0x74}, {0x31, 0x8c, 0xe7}, {0xfd, 0x3f, 0x92},
	{0x50, 0xd8, 0xd8}, {0xf8, 0xf8, 0xf2},
	{0x26, 0x26, 0x26}, {0xfd, 0x5e, 0x53}, {0x50, 0xc8, 0x78},
	{0xe6, 0xdb, 0x74}, {0x31, 0x8c, 0xe7}, {0xfd, 0x3f, 0x92},
	{0x50, 0xd8, 0xd8}, {0xf8, 0xf8, 0xf2},
}

const defaultfg, defaultbg, defaultcs = 15, 0, 15

// groundHex and caretHex are the window's ground and its text's cursor,
// the terminal's own two, so that a terminal sits in the window seamless
var groundHex, caretHex = code(palette[defaultbg]), code(palette[defaultcs])

// openterm starts a shell in pane p on a pty, as st starts one: in the
// frame's folder, in the world its programs have, TERM st's (vt.Start);
// or what run says, one word as a shell reads it, more as a command and
// its arguments, as tmux's new-window runs them
// (`studio/2-term/20-newtab.go`). What it writes comes to the window's
// thread as news, which puts it on p's screen (handle), until the shell
// ends, which ends the terminal, as st ends with its shell, though a job
// it left holds the pty; what p's terminal sends the shell goes by a
// typist, for the window never waits on a shell slow to read.
func (s *studio) openterm(p *pane, run []string) error {
	sh := shell()
	c := exec.Command(sh)
	switch {
	case len(run) == 1:
		c = exec.Command(sh, "-c", run[0])
	case len(run) > 1:
		c = exec.Command(run[0], run[1:]...)
	}
	c.Dir, c.Env = s.dir, s.env
	f, err := vt.Start(c, 80, 24)
	if err != nil {
		return err
	}
	ty := &typist{wake: make(chan struct{}, 1), gone: make(chan struct{})}
	p.cmd, p.pty, p.term = c, f, vt.New(80, 24, ty)
	go ty.run(f)
	go func() {
		c.Wait()
		f.Close() // which ends the Read under way
	}()
	go func() {
		for b := make([]byte, 32<<10); ; {
			n, err := f.Read(b)
			if n > 0 {
				out := slices.Clone(b[:n])
				s.events <- news{p: p, out: out}
			}
			if err != nil {
				break
			}
		}
		close(ty.gone)
		s.events <- news{p: p, ended: true}
	}()
	return nil
}

// typist is what a terminal sends its shell, queued without a bound and
// written to the pty by a goroutine of its own until the shell is gone,
// so that a key, a paste or an answer vt makes never waits on a shell
// slow to read, nor the window with it (vt.New)
type typist struct {
	mu   sync.Mutex
	q    [][]byte
	wake chan struct{} // holds one: something is queued
	gone chan struct{}
}

func (t *typist) Write(b []byte) (int, error) {
	t.mu.Lock()
	t.q = append(t.q, slices.Clone(b))
	t.mu.Unlock()
	select {
	case t.wake <- struct{}{}:
	default:
	}
	return len(b), nil
}

func (t *typist) run(f *os.File) {
	for {
		select {
		case <-t.wake:
		case <-t.gone:
			return
		}
		t.mu.Lock()
		q := t.q
		t.q = nil
		t.mu.Unlock()
		for _, b := range q {
			f.Write(b)
		}
	}
}

// newterm opens a terminal as a tab of the panel, running run, a shell
// for none (openterm), named by the next number free among the panel's
// tabs, as l new does in tmux (`studio/2-term/20-newtab.go`), and shows
// it with the eye on it, the panel with it when it was hidden (show); a
// last tab that ended goes for it
func (s *studio) newterm(run []string) {
	if s.panel == "" {
		return
	}
	tabs := s.tabs[s.panel]
	n := len(tabs) + 1
	for slices.ContainsFunc(tabs, func(p *pane) bool {
		return p.tab == strconv.Itoa(n)
	}) {
		n++
	}
	p := &pane{app: newApp(nil), name: s.panel, tab: strconv.Itoa(n)}
	if err := s.openterm(p, run); err != nil {
		fmt.Fprintln(os.Stderr, "window: a terminal:", err)
		return
	}
	s.tabs[s.panel] = append(tabs, p)
	s.show(s.panel, p)
	s.tabs[s.panel] = slices.DeleteFunc(s.tabs[s.panel],
		func(p *pane) bool { return p.gone })
}

// folder is where a terminal's program runs, the one in the foreground of
// its pty, as tmux finds a pane's (pane_current_path): the folder that
// printed its words; the frame's when it cannot say
func (s *studio) folder(p *pane) string {
	if g, err := vt.Foreground(p.pty); err == nil {
		at, err := os.Readlink(fmt.Sprintf("/proc/%d/cwd", g))
		if err == nil {
			return at
		}
	}
	return s.dir
}

// prompting is whether a terminal's shell waits at its prompt, as p -a
// asks tmux: the program in the foreground of its pty is the shell
func (s *studio) prompting(p *pane) bool {
	g, err := vt.Foreground(p.pty)
	if err != nil {
		return false
	}
	name, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", g))
	return err == nil &&
		strings.TrimSpace(string(name)) == filepath.Base(shell())
}

// shell is the shell a terminal runs, the person's, st's way
func shell() string {
	if sh := os.Getenv("SHELL"); sh != "" {
		return sh
	}
	return "/bin/sh"
}

// closetab ends the tab a slot shows, as l close asks (tabs.Turn): its
// program hung up, as st hangs up its shell, and the next tab takes its
// place as it ends (drop).
func (s *studio) closetab(slot string) {
	p := s.panes[slot]
	if p == nil || p.gone || p.cmd == nil {
		return
	}
	p.cmd.Process.Signal(syscall.SIGHUP)
}

// row is a terminal's line as a strip, its cells in st's colours, an
// underlined one a run of its own
func row(cells []vt.Cell, rev bool) *strip {
	k := &strip{}
	for _, c := range cells {
		fg, bg := colours(c, rev)
		bold := c.Attr&vt.Bold != 0
		if c.Attr&vt.Underline == 0 {
			k.add(string(glyph(c)), fg, bg, bold)
			continue
		}
		r := run{string(glyph(c)), fg, bg, bold, true}
		k.runs, k.w = append(k.runs, r), k.w+1
	}
	return k
}

// glyph is the rune a cell shows: a blank for one a wide rune covers, for
// a strip holds a rune a column
func glyph(c vt.Cell) rune {
	if c.Attr&vt.Wdummy != 0 || c.Rune == 0 {
		return ' '
	}
	return c.Rune
}

// colours is a cell's foreground and background as st's xdrawglyph
// makes them: bold brightens the first eight; the screen reversed inverts
// a colour, and trades the terminal's own two; faint halves; a cell
// reversed trades the two; and an invisible one hides its rune.
func colours(c vt.Cell, rev bool) (fg, bg string) {
	f, b := c.Fg, c.Bg
	if f == vt.Fore {
		f = defaultfg
	}
	if b == vt.Back {
		b = defaultbg
	}
	if c.Attr&(vt.Bold|vt.Faint) == vt.Bold && f <= 7 {
		f += 8
	}
	fc, bc := rgb(f), rgb(b)
	if rev {
		fc = flip(f, defaultfg, defaultbg, fc)
		bc = flip(b, defaultbg, defaultfg, bc)
	}
	if c.Attr&(vt.Bold|vt.Faint) == vt.Faint {
		fc = [3]uint8{fc[0] / 2, fc[1] / 2, fc[2] / 2}
	}
	if c.Attr&vt.Reverse != 0 {
		fc, bc = bc, fc
	}
	if c.Attr&vt.Invisible != 0 {
		fc = bc
	}
	return code(fc), code(bc)
}

// flip is colour x, c in red, green and blue, as the screen reversed
// shows it: the terminal's own, own, the other of its two, and any other
// its every part inverted, as st inverts it
func flip(x, own, other vt.Color, c [3]uint8) [3]uint8 {
	if x == own {
		return rgb(other)
	}
	return [3]uint8{^c[0], ^c[1], ^c[2]}
}

// rgb is a colour's red, green and blue: the palette's for its first
// 16, else its own (vt.Color.Rgb)
func rgb(c vt.Color) [3]uint8 {
	if c < 16 {
		return palette[c]
	}
	r, g, b, _ := c.Rgb()
	return [3]uint8{r, g, b}
}

// codes are the colours code named, kept, for a screen shows few; a
// program that paints in many fills it, and it starts again
var codes = map[[3]uint8]string{}

// code is a colour as a strip names it, #rrggbb
func code(c [3]uint8) string {
	s, ok := codes[c]
	if !ok {
		if len(codes) > 4096 {
			clear(codes)
		}
		s = fmt.Sprintf("#%02x%02x%02x", c[0], c[1], c[2])
		codes[c] = s
	}
	return s
}

// termword is the word at cell x of a terminal's line, as tmux takes the
// word under a click (outline.Word): a cell a wide rune covers is the
// rune's own
func termword(cells []vt.Cell, x int) string {
	var line []rune
	at := -1
	for i, c := range cells {
		if c.Attr&vt.Wdummy != 0 {
			if i == x {
				at = len(line) - 1
			}
			continue
		}
		if i == x {
			at = len(line)
		}
		line = append(line, glyph(c))
	}
	return outline.Word(line, at)
}
