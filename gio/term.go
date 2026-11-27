// a terminal in the window: a frame's shell slot runs a shell on a pty,
// its screen st's, as sys/vt keeps it, a row a strip; main.go draws it
// and hands it the keys and the mouse, as st hands them
package main

import (
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strconv"
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

// openterm starts a shell in pane p on a pty, as st starts one: in the
// frame's folder, in the world its programs have, TERM st's (vt.Start).
// What the shell writes comes to the window's thread as news, which puts
// it on p's screen (handle); what p's terminal sends the shell goes by a
// typist, for the window never waits on a shell slow to read.
func (s *studio) openterm(p *pane) error {
	sh := os.Getenv("SHELL")
	if sh == "" {
		sh = "/bin/sh"
	}
	c := exec.Command(sh)
	c.Dir, c.Env = s.dir, s.env
	f, err := vt.Start(c, 80, 24)
	if err != nil {
		return err
	}
	ty := &typist{keys: make(chan []byte, 64), gone: make(chan struct{})}
	p.cmd, p.pty, p.term = c, f, vt.New(80, 24, ty)
	go ty.run(f)
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
		c.Wait()
		close(ty.gone)
		f.Close()
		s.events <- news{p: p, ended: true}
	}()
	return nil
}

// typist is what a terminal sends its shell, written to the pty by a
// goroutine of its own, until the shell is gone
type typist struct {
	keys chan []byte
	gone chan struct{}
}

func (t *typist) Write(b []byte) (int, error) {
	select {
	case t.keys <- slices.Clone(b):
	case <-t.gone:
	}
	return len(b), nil
}

func (t *typist) run(f *os.File) {
	for {
		select {
		case b := <-t.keys:
			f.Write(b)
		case <-t.gone:
			return
		}
	}
}

// newterm opens a terminal as a tab of the panel, named by the next
// number free among its tabs, and shows it with the eye on it, the panel
// too when it was hidden, as l new does in tmux
// (`studio/2-term/20-newtab.go`); a last tab that ended goes for it
func (s *studio) newterm() {
	if s.panel == "" {
		return
	}
	n := 2
	for slices.ContainsFunc(s.tabs[s.panel], func(p *pane) bool {
		return p.tab == strconv.Itoa(n)
	}) {
		n++
	}
	p := &pane{app: newApp(nil), name: s.panel, tab: strconv.Itoa(n)}
	if s.openterm(p) != nil {
		return
	}
	s.tabs[s.panel] = append(s.tabs[s.panel], p)
	s.showTab(s.panel, p)
	s.tabs[s.panel] = slices.DeleteFunc(s.tabs[s.panel],
		func(p *pane) bool { return p.gone })
	if s.hidden {
		s.togglePanel()
	}
	s.focus(slices.Index(s.order, p))
}

// closetab ends the tab a slot shows, as l close asks: its program hung
// up, as st hangs up its shell, and the next tab takes its place as it
// ends (drop). A slot's last tab stays, for it is the slot.
func (s *studio) closetab(slot string) {
	p := s.panes[slot]
	if p == nil || p.gone || p.cmd == nil || len(s.tabs[slot]) < 2 {
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

// codes are the colours code named, kept, for a screen shows few
var codes = map[[3]uint8]string{}

// code is a colour as a strip names it, #rrggbb
func code(c [3]uint8) string {
	s, ok := codes[c]
	if !ok {
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
