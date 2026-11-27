// sys's windows in gio: an extension beside the terminal, and the proof
// that sys's programs need none: fm, ed and the lens run here as they are,
// and a shell in a terminal of the window's own (term.go).
// With no argument it is `window`, the view at the other end of a
// program's pipe, as fm -w starts it. -studio dir lays studio/frame out, as
// `studio slots` says it, each slot's program at the end of a pipe. Only
// this file knows gio.
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"unicode"

	gioapp "gioui.org/app"
	"gioui.org/io/clipboard"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/system"
	"gioui.org/io/transfer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"

	"sys/tty"
	"sys/view"
	"sys/vt"
)

const (
	cellH = 20 // a terminal's line at 16 pixels
	size  = 16
)

func hex(c string) color.NRGBA {
	n, _ := strconv.ParseUint(c[1:], 16, 32)
	return color.NRGBA{
		R: uint8(n >> 16),
		G: uint8(n >> 8),
		B: uint8(n),
		A: 255,
	}
}

type window struct {
	s     *studio
	news  <-chan news // what the next frame takes
	frame bool        // the studio's frame, else one pane
	text  *typeface
	// the runs' pictures the last frame drew, and this one's, so that a
	// frame makes only what is new and keeps two frames' worth
	pictures, drawn map[look]paint.ImageOp
	tag             int // what events are addressed to

	focused bool
	full    bool
	hits    []hit           // where each pane was drawn last, for a click
	bars    []barHit        // where each slot's bar was, for a click
	held    *pane           // where the left button went down, for a drag
	buttons pointer.Buttons // held, as the last pointer event had them
}

type hit struct {
	area image.Rectangle
	p    *pane
}

// barHit is a slot's bar as it was drawn last: its line, the slot, and
// where each tab's name stands on it, the tab a press there shows
type barHit struct {
	area  image.Rectangle
	slot  string
	chips []view.Chip
}

func newWindow(s *studio, frame bool) (*window, error) {
	const dir = "/usr/share/fonts/truetype/dejavu/"
	t, err := newTypeface(
		size,
		dir+"DejaVuSansMono.ttf",
		dir+"DejaVuSansMono-Bold.ttf",
	)
	if err != nil {
		return nil, err
	}
	return &window{s: s, news: s.events, frame: frame, text: t,
		pictures: map[look]paint.ImageOp{},
		drawn:    map[look]paint.ImageOp{}}, nil
}

// look is a run's picture, kept so that a frame makes only what is new
type look struct {
	txt    string
	fg, bg color.NRGBA
	bold   bool
	high   int
}

// label draws txt at x, y on the ground bg
func (v *window) label(
	gtx layout.Context,
	x, y int,
	txt string,
	fg, bg color.NRGBA,
	bold bool,
) {
	k := look{txt, fg, bg, bold, gtx.Dp(cellH)}
	pic, ok := v.drawn[k]
	if !ok {
		if pic, ok = v.pictures[k]; !ok {
			pic = paint.NewImageOp(v.text.image(txt, fg, bg, bold,
				k.high))
			pic.Filter = paint.FilterNearest
		}
		v.drawn[k] = pic
	}
	defer op.Offset(image.Pt(x, y)).Push(gtx.Ops).Pop()
	defer clip.Rect{Max: pic.Size()}.Push(gtx.Ops).Pop()
	pic.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
}

func fill(gtx layout.Context, r image.Rectangle, c color.NRGBA) {
	paint.FillShape(gtx.Ops, c, clip.Rect(r).Op())
}

// line draws one strip with its corner at at, in columns cw wide
func (v *window) line(gtx layout.Context, k *strip, at image.Point, cw int) {
	x, h := 0, gtx.Dp(cellH)
	for _, r := range k.runs {
		n := len([]rune(r.text))
		left, right := at.X+x*cw, at.X+(x+n)*cw
		bg := hex(groundHex)
		if r.bg != "" {
			bg = hex(r.bg)
		}
		v.label(gtx, left, at.Y, r.text, hex(r.fg), bg, r.bold)
		if r.underline {
			fill(
				gtx,
				image.Rect(left, at.Y+h-2, right, at.Y+h-1),
				hex(r.fg),
			)
		}
		x += n
	}
}

// slot draws the pane of the slot named in area, telling its program the
// size when that changes; a shell's terminal as st draws one (term), or,
// its shell ended, says so. A slot of tabs has its bar over the tab it
// shows, each tab's name a button (tabBar)
func (v *window) slot(gtx layout.Context, name string, area image.Rectangle,
	cw int) {
	p, h := v.s.panes[name], gtx.Dp(cellH)
	defer clip.Rect(area).Push(gtx.Ops).Pop()
	if k, chips := v.s.tabBar(name, area.Dx()/cw); k != nil {
		row := image.Rect(area.Min.X, area.Min.Y, area.Max.X,
			area.Min.Y+h)
		v.line(gtx, k, area.Min, cw)
		v.bars = append(v.bars, barHit{row, name, chips})
		area.Min.Y += h
	}
	if p.term != nil && p.gone {
		said := " " + name + ": the shell ended"
		if name == v.s.panel {
			said += "; + opens another"
		}
		k := (&strip{}).add(said, tty.Faint, "", false)
		v.line(gtx, k, area.Min, cw)
		return
	}
	if p.term != nil {
		v.term(gtx, p, area, cw)
		v.hits = append(v.hits, hit{area, p})
		return
	}
	cols, high := area.Dx()/cw, area.Dy()/h
	if cols != p.cols || high != p.high {
		p.cols, p.high = cols, high
		p.v.In(0, 0, cols, high)
		p.send("size", object{"w": cols, "h": high})
	}
	var lines []*strip
	if p.isText() {
		lines = p.source()
	} else if p.tree != nil {
		lines = p.list()
	}
	if p.asking && len(lines) > 0 {
		lines[len(lines)-1] = (&strip{}).add(
			" "+p.prompt+p.answer+"▏", "", "", false)
	}
	for y, l := range lines {
		v.line(gtx, l, area.Min.Add(image.Pt(0, y*h)), cw)
	}
	v.hits = append(v.hits, hit{area, p})
}

// lay puts slots in r, stacked down or side by side, each its weight's
// part, a rule between; a pane in each that holds one. A row is a whole
// line high, so that text stands on the grid
func (v *window) lay(gtx layout.Context, slots []slot, r image.Rectangle,
	down bool, cw int) {
	total, h := 0, gtx.Dp(cellH)
	for _, s := range slots {
		total += s.Weight
	}
	at, given := r.Min, 0
	for i, s := range slots {
		given += s.Weight
		area := r
		if down {
			end := r.Min.Y + (r.Dy()*given/max(1, total))/h*h
			if i == len(slots)-1 {
				end = r.Max.Y
			}
			area.Min.Y, area.Max.Y = at.Y, end
			at.Y = end
		} else {
			end := r.Min.X + r.Dx()*given/max(1, total)
			if i == len(slots)-1 {
				end = r.Max.X
			}
			area.Min.X, area.Max.X = at.X, end
			at.X = end
		}
		edge := area // the rule before the slot, in the frame's colour
		if down {
			edge.Max.Y = edge.Min.Y + 1
		} else {
			edge.Max.X = edge.Min.X + 1
			area.Min.X++
		}
		if len(s.Kids) > 0 {
			v.lay(gtx, s.Kids, area, !down, cw)
		} else {
			v.slot(gtx, s.Name, area, cw)
		}
		// across, the rule takes a column of pixels; down, it is
		// drawn over the slot's first, a row's blank top, so that rows
		// keep the grid
		if i > 0 {
			fill(gtx, edge, hex(tty.Rule))
		}
	}
}

// under is the pane at a point, and the cell there, counted from its
// corner
func (v *window) under(at image.Point, cw, h int) (*pane, int, int, bool) {
	for _, x := range v.hits {
		if at.In(x.area) {
			return x.p, (at.X - x.area.Min.X) / cw,
				(at.Y - x.area.Min.Y) / h, true
		}
	}
	return nil, 0, 0, false
}

// cellOf is the cell of pane p a point is at, counted from its corner,
// though it be past p's edge, as a drag that left it is
func (v *window) cellOf(p *pane, at image.Point, cw, h int) (int, int) {
	for _, x := range v.hits {
		if x.p == p {
			return (at.X - x.area.Min.X) / cw,
				(at.Y - x.area.Min.Y) / h
		}
	}
	return 0, 0
}

// pointed is a mouse event, the three buttons as in any pane
// (`design/mouse`): a left press picks (pressedAt), on a slot's bar the
// tab named there too, and a move with it held drags a text's selection; a
// right press looks (look), a middle one hands on (do), and the wheel
// moves (wheel), each on the pane under it; on a terminal, the mouse is
// as st has it (mouse), and one pressed keeps it, as st's window does,
// until its release, though the pointer leave it. In the frame the eye
// follows the pointer: a pane moved over takes it
func (v *window) pointed(e pointer.Event, cw, h int) {
	at := e.Position.Round()
	was := v.buttons
	v.buttons = e.Buttons
	p, x, y, ok := v.under(at, cw, h)
	if t := v.held; t != nil && t.term != nil &&
		e.Kind&(pointer.Drag|pointer.Release) != 0 {
		p, ok = t, true
		x, y = v.cellOf(t, at, cw, h)
	}
	onBar := slices.ContainsFunc(v.bars, func(b barHit) bool {
		return at.In(b.area)
	})
	if e.Kind == pointer.Release {
		v.held = nil
	}
	if ok && v.frame && e.Kind == pointer.Move {
		v.s.focusOn(p)
	}
	switch {
	case onBar && e.Kind == pointer.Press &&
		e.Buttons.Contain(pointer.ButtonPrimary):
		v.pressedAt(at, cw, h) // a slot's bar, above its pane
	case ok && p.term != nil:
		if e.Kind == pointer.Press {
			v.held = p
		}
		v.mouse(p, e, was, x, y)
	case !ok, e.Kind == pointer.Release:
	case e.Kind == pointer.Press &&
		e.Buttons.Contain(pointer.ButtonPrimary):
		v.pressedAt(at, cw, h)
		v.held = p
	case e.Kind == pointer.Drag && p == v.held:
		p.drag(x, y)
	case e.Kind == pointer.Press &&
		e.Buttons.Contain(pointer.ButtonSecondary):
		v.look(p, x, y)
	case e.Kind == pointer.Press &&
		e.Buttons.Contain(pointer.ButtonTertiary):
		v.do(p, x, y)
	case e.Kind == pointer.Scroll && e.Scroll.Y != 0:
		p.wheel(e.Scroll.Y)
	}
}

// pressedAt is a press at a point: the eye goes to the pane there, and
// the press to its row, or to its text's line and rune
func (v *window) pressedAt(at image.Point, cw, h int) {
	for _, b := range v.bars {
		if !at.In(b.area) {
			continue
		}
		x := (at.X - b.area.Min.X) / cw
		for _, c := range b.chips {
			if c.From <= x && x < c.To {
				v.s.turn(b.slot, []string{c.Name})
			}
		}
		return
	}
	for _, x := range v.hits {
		if !at.In(x.area) {
			continue
		}
		v.s.focusOn(x.p)
		cx, cy := (at.X-x.area.Min.X)/cw, (at.Y-x.area.Min.Y)/h
		switch {
		case x.p.isText() && cy == x.p.high-1:
			if name := x.p.modeWord(cx); name != "" {
				x.p.key(name, "")
			}
		case x.p.isText():
			x.p.click(x.p.spot(cx, cy))
		default:
			x.p.click(x.p.v.At(cx, cy), 0)
		}
	}
}

// named is gio's keys as tty names them
var named = map[key.Name]string{
	key.NameReturn: tty.KEnter, key.NameEnter: tty.KEnter,
	key.NameEscape: tty.KEsc, key.NameTab: tty.KTab,
	key.NameDeleteBackward: tty.KBack,
	key.NameUpArrow:        tty.KUp, key.NameDownArrow: tty.KDown,
	key.NameLeftArrow: tty.KLeft, key.NameRightArrow: tty.KRight,
	key.NamePageUp: tty.KPgUp, key.NamePageDown: tty.KPgDn,
}

// input does what came in since the last frame; toggle is whether f11
// asked for the other mode, full screen or not. A key goes as tty names
// it, ctrl and a letter as C- and the letter, ctrl-c as tty's INT; ctrl-v
// pastes, and the mouse is as in any pane (pointed). To a terminal a key
// goes as vt names it (typed), but the window's own, and ctrl-shift-v
// pastes, as in st
func (v *window) input(gtx layout.Context) (toggle bool) {
	all := key.ModCtrl | key.ModShift | key.ModAlt
	filters := []event.Filter{
		key.FocusFilter{Target: &v.tag},
		transfer.TargetFilter{
			Target: &v.tag,
			Type:   "application/text",
		},
		pointer.Filter{Target: &v.tag,
			Kinds: pointer.Press | pointer.Drag | pointer.Release |
				pointer.Move | pointer.Scroll,
			ScrollY: pointer.ScrollRange{Min: -1 << 30,
				Max: 1 << 30}},
		key.Filter{Focus: &v.tag, Optional: all},
		key.Filter{Focus: &v.tag, Name: key.NameTab, Optional: all},
	}
	for {
		e, ok := gtx.Event(filters...)
		if !ok {
			return toggle
		}
		p := v.s.order[v.s.eye]
		switch e := e.(type) {
		case key.EditEvent:
			for _, r := range e.Text {
				switch {
				case p.term != nil:
					p.term.Type(r, 0)
				case unicode.IsPrint(r):
					p.pressed(string(r))
				}
			}
		case key.Event:
			ctrl := e.Modifiers.Contain(key.ModCtrl)
			switch {
			case e.State != key.Press:
			case e.Name == key.NameF11:
				toggle = true
			case ctrl && e.Name == "W" && v.frame:
				v.s.focus(v.s.eye + 1)
			case e.Modifiers.Contain(key.ModAlt) &&
				e.Name == "T" && v.frame:
				v.s.hide(!v.s.hidden)
			case p.term != nil && ctrl && e.Name == "V" &&
				e.Modifiers.Contain(key.ModShift):
				gtx.Execute(clipboard.ReadCmd{Tag: &v.tag})
			case p.term != nil:
				typed(p.term, e)
			case ctrl && e.Name == "C":
				p.pressed(tty.KInt)
			case ctrl && e.Name == "V":
				gtx.Execute(clipboard.ReadCmd{Tag: &v.tag})
			case ctrl && len(e.Name) == 1:
				letter := strings.ToLower(string(e.Name))
				p.pressed("C-" + letter)
			case e.Modifiers.Contain(key.ModShift) &&
				e.Name == key.NameUpArrow:
				p.pressed(tty.KShiftUp)
			case e.Modifiers.Contain(key.ModShift) &&
				e.Name == key.NameDownArrow:
				p.pressed(tty.KShiftDown)
			default:
				p.pressed(named[e.Name])
			}
		case transfer.DataEvent:
			paste(p, e)
		case pointer.Event:
			v.pointed(e, v.text.cell, gtx.Dp(cellH))
		}
	}
}

// paste hands what the clipboard held to the pane with the eye: ed as
// one key, a terminal as st pastes (vt.Paste)
func paste(p *pane, e transfer.DataEvent) {
	if !p.isText() && p.term == nil {
		return
	}
	r := e.Open()
	defer r.Close()
	b, err := io.ReadAll(r)
	switch {
	case err != nil:
	case p.term != nil:
		p.term.Paste(string(b))
	default:
		p.send("key", object{"name": tty.KPaste, "text": string(b)})
	}
}

// draw is a frame: the news since the last one, as much as had come when
// it began, for a shell that writes without end must not starve the
// frames, and each piece wakes another; then the input, then the picture
func (v *window) draw(gtx layout.Context) (toggle, done bool) {
	for range len(v.news) {
		copied, over := v.s.handle(<-v.news)
		if copied != "" {
			gtx.Execute(clipboard.WriteCmd{
				Type: "application/text",
				Data: io.NopCloser(strings.NewReader(copied)),
			})
		}
		done = done || over
	}
	toggle = v.input(gtx)

	all := image.Rectangle{Max: gtx.Constraints.Max}
	fill(gtx, all, hex(groundHex))
	event.Op(gtx.Ops, &v.tag)
	if !v.focused {
		gtx.Execute(key.FocusCmd{Tag: &v.tag})
		v.focused = true
	}
	cw := v.text.cell

	v.hits, v.bars = v.hits[:0], v.bars[:0]
	slots := v.s.frame
	if v.s.hidden {
		panel := func(s slot) bool { return s.Name == v.s.panel }
		slots = slices.DeleteFunc(slices.Clone(slots), panel)
	}
	v.lay(gtx, slots, all, true, cw)
	v.pictures, v.drawn = v.drawn, v.pictures
	clear(v.drawn)
	return
}

// listen has wake called for each piece of news, which comes from the
// programs' own threads; a frame then takes it. The senders keep their
// channel: only what the frames read changes
func (v *window) listen(wake func()) {
	waiting := make(chan news, 64)
	v.news = waiting
	go func() {
		for e := range v.s.events {
			waiting <- e
			wake()
		}
	}()
}

func (v *window) run(w *gioapp.Window) error {
	var ops op.Ops
	v.listen(w.Invalidate)
	for {
		switch e := w.Event().(type) {
		case gioapp.DestroyEvent:
			return e.Err
		case gioapp.ConfigEvent:
			v.full = e.Config.Mode == gioapp.Fullscreen
		case gioapp.FrameEvent:
			gtx := gioapp.NewContext(&ops, e)
			toggle, done := v.draw(gtx)
			e.Frame(gtx.Ops)
			if toggle {
				mode := gioapp.Fullscreen
				if v.full {
					mode = gioapp.Windowed
				}
				w.Option(mode.Option())
			}
			if done {
				w.Perform(system.ActionClose)
			}
		}
	}
}

func main() {
	frame := flag.Bool(
		"studio",
		false,
		"lay studio's frame out on the folder given",
	)
	flag.Parse()
	var s *studio
	title, wide, high := "window", unit.Dp(760), unit.Dp(560)
	if *frame {
		dir := flag.Arg(0)
		if dir == "" {
			dir = "."
		}
		dir, err := filepath.Abs(dir)
		if err != nil {
			log.Fatal(err)
		}
		if s, err = start(dir); err != nil {
			log.Fatal(err)
		}
		title, wide, high = "studio (gio)", 1500, 800
	} else {
		s = piped(os.Stdin, os.Stdout)
	}
	v, err := newWindow(s, *frame)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		w := new(gioapp.Window)
		w.Option(gioapp.Title(title), gioapp.Size(wide, high))
		if err := v.run(w); err != nil {
			fmt.Fprintln(os.Stderr, "window:", err)
			os.Exit(1)
		}
		os.Exit(0)
	}()
	gioapp.Main()
}

// -- the terminal, its gio side (term.go)

// term draws a terminal in area: its program told the size when it
// changes (vt.Setsize), each row a strip in st's colours, and the cursor
// as st draws it, unless the program hid it: in the shape it asked for
// while the terminal has the eye, else a box.
func (v *window) term(gtx layout.Context, p *pane, area image.Rectangle,
	cw int) {
	h := gtx.Dp(cellH)
	cols, rows := max(1, area.Dx()/cw), max(1, area.Dy()/h)
	if cols != p.cols || rows != p.high {
		p.cols, p.high = cols, rows
		p.term.Resize(cols, rows)
		vt.Setsize(p.pty, cols, rows)
	}
	rev := p.term.Reversed()
	for y := range rows {
		at := area.Min.Add(image.Pt(0, y*h))
		v.line(gtx, row(p.term.Line(y), rev), at, cw)
	}
	if p.term.Hidden() {
		return
	}
	x, y := p.term.Cursor()
	at := area.Min.Add(image.Pt(x*cw, y*h))
	r := image.Rectangle{Min: at, Max: at.Add(image.Pt(cw, h))}
	ink := hex(caretHex)
	switch shape := p.term.Shape(); {
	case v.s.order[v.s.eye] != p:
		for _, e := range []image.Rectangle{
			{r.Min, image.Pt(r.Max.X, r.Min.Y+1)},
			{image.Pt(r.Min.X, r.Max.Y-1), r.Max},
			{r.Min, image.Pt(r.Min.X+1, r.Max.Y)},
			{image.Pt(r.Max.X-1, r.Min.Y), r.Max},
		} {
			fill(gtx, e, ink)
		}
	case shape == 3 || shape == 4:
		r.Min.Y = r.Max.Y - 2
		fill(gtx, r, ink)
	case shape == 5 || shape == 6:
		r.Max.X = r.Min.X + 2
		fill(gtx, r, ink)
	default:
		c := p.term.Line(y)[x]
		k := (&strip{}).add(string(glyph(c)),
			code(palette[defaultbg]), code(palette[defaultcs]),
			c.Attr&vt.Bold != 0)
		v.line(gtx, k, at, cw)
	}
}

// termkeys is gio's keys as vt names them; f11 is the window's own
var termkeys = map[key.Name]vt.Key{
	key.NameUpArrow: vt.Up, key.NameDownArrow: vt.Down,
	key.NameLeftArrow: vt.Left, key.NameRightArrow: vt.Right,
	key.NameHome: vt.Home, key.NameEnd: vt.End,
	key.NamePageUp: vt.PageUp, key.NamePageDown: vt.PageDown,
	key.NameDeleteForward: vt.Delete, key.NameDeleteBackward: vt.Backspace,
	key.NameReturn: vt.Enter, key.NameEnter: vt.Enter,
	key.NameTab: vt.Tab, key.NameEscape: vt.Esc,
	key.NameF1: vt.F1, key.NameF2: vt.F2, key.NameF3: vt.F3,
	key.NameF4: vt.F4, key.NameF5: vt.F5, key.NameF6: vt.F6,
	key.NameF7: vt.F7, key.NameF8: vt.F8, key.NameF9: vt.F9,
	key.NameF10: vt.F10, key.NameF12: vt.F12,
}

// typed hands a terminal a key pressed: a named one as vt names it, with
// the modifiers held, or, with ctrl or alt held, its rune, which gio then
// gives as a key and not as text; plain text comes as text (input).
func typed(t *vt.Term, e key.Event) {
	m := mods(e.Modifiers)
	if k, ok := termkeys[e.Name]; ok {
		t.Key(k, m)
		return
	}
	r := []rune(string(e.Name))
	switch {
	case m&(vt.Ctrl|vt.Alt) == 0:
	case e.Name == key.NameSpace:
		t.Type(' ', m)
	case len(r) == 1 && m&vt.Shift == 0:
		t.Type(unicode.ToLower(r[0]), m)
	case len(r) == 1:
		t.Type(r[0], m)
	}
}

// mods is gio's modifiers as vt names them
func mods(m key.Modifiers) vt.Mod {
	var out vt.Mod
	if m.Contain(key.ModShift) {
		out |= vt.Shift
	}
	if m.Contain(key.ModCtrl) {
		out |= vt.Ctrl
	}
	if m.Contain(key.ModAlt) {
		out |= vt.Alt
	}
	if m.Contain(key.ModSuper) {
		out |= vt.Super
	}
	return out
}

// mouse is the mouse on a terminal at cell x, y, kept on its screen, as
// st keeps a press on its own, was the buttons held before: a program
// that took the mouse has it, but with shift held, as st keeps it then
// (vt.Mousing); else a left press gives the eye and says the word there,
// a right one looks at it and a middle one hands it to the shell, as on
// any pane (`design/mouse`), each in the folder that printed it
// (folder); and the wheel is st's (vt.Wheel).
func (v *window) mouse(p *pane, e pointer.Event, was pointer.Buttons, x,
	y int) {
	t, m := p.term, mods(e.Modifiers)
	cols, rows := t.Size()
	x, y = max(0, min(x, cols-1)), max(0, min(y, rows-1))
	switch {
	case e.Kind == pointer.Scroll && e.Scroll.Y != 0:
		t.Wheel(e.Scroll.Y < 0, x, y, m)
		return
	case e.Kind == pointer.Scroll:
		return
	}
	if t.Mousing(m) {
		switch e.Kind {
		case pointer.Press:
			v.s.focusOn(p)
			t.Press(button(e.Buttons&^was), x, y, m)
		case pointer.Release:
			t.Release(button(was&^e.Buttons), x, y, m)
		case pointer.Move, pointer.Drag:
			t.Move(x, y, m)
		}
		return
	}
	if e.Kind != pointer.Press {
		return
	}
	word, dir := termword(t.Line(y), x), v.s.folder(p)
	switch button(e.Buttons &^ was) {
	case 1:
		v.s.focusOn(p)
		v.s.say(word, dir)
	case 2:
		v.s.handon(word, dir)
	case 3:
		v.lookat(word, dir)
	}
}

// button is a button as st numbers them: 1 left, 2 middle, 3 right
func button(b pointer.Buttons) int {
	switch {
	case b.Contain(pointer.ButtonPrimary):
		return 1
	case b.Contain(pointer.ButtonTertiary):
		return 2
	case b.Contain(pointer.ButtonSecondary):
		return 3
	}
	return 0
}
