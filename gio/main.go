// sys's windows in gio: an extension beside the terminal, and the proof
// that sys's programs need none: fm and ed run here as they are.
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
)

const (
	cellH = 20 // a terminal's line at 16 pixels
	size  = 16
	// the terminal's ground and cursor, alacritty's
	groundHex = "#121212"
	caretHex  = "#f8f8f2"
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
	s        *studio
	news     <-chan news // what the next frame takes
	frame    bool        // the studio's frame, else one pane
	text     *typeface
	pictures map[look]paint.ImageOp
	tag      int // what events are addressed to

	focused bool
	full    bool
	hits    []hit // where each pane was drawn last, for a click
}

type hit struct {
	area image.Rectangle
	p    *pane
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
		pictures: map[look]paint.ImageOp{}}, nil
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
	pic, ok := v.pictures[k]
	if !ok {
		if len(v.pictures) > 4000 {
			clear(v.pictures)
		}
		pic = paint.NewImageOp(v.text.image(txt, fg, bg, bold, k.high))
		pic.Filter = paint.FilterNearest
		v.pictures[k] = pic
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
// size when that changes; a shell, which needs a terminal, says so
func (v *window) slot(gtx layout.Context, name string, area image.Rectangle,
	cw int) {
	p, h := v.s.panes[name], gtx.Dp(cellH)
	defer clip.Rect(area).Push(gtx.Ops).Pop()
	if p.out == nil {
		for y, said := range []string{
			" " + name + ": a shell needs a terminal; studio in" +
				" tmux has one",
			" ctrl-w moves the eye · a click picks"} {
			k := (&strip{}).add(said, tty.Faint, "", false)
			v.line(gtx, k, area.Min.Add(image.Pt(0, y*h)), cw)
		}
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

// pressedAt is a press at a point: the eye goes to the pane there, and
// the press to its row, or to its text's line and rune
func (v *window) pressedAt(at image.Point, cw, h int) {
	for _, x := range v.hits {
		if !at.In(x.area) {
			continue
		}
		for i, p := range v.s.order {
			if p == x.p {
				v.s.focus(i)
			}
		}
		cx, cy := (at.X-x.area.Min.X)/cw, (at.Y-x.area.Min.Y)/h
		switch {
		case cy == 0 && x.p.tops() > 0:
			if name := x.p.barAt(cx); name != "" {
				lens := object{"name": tty.KLens, "text": name}
				x.p.send("key", lens)
			}
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
// pastes, and a press is the left button's alone
func (v *window) input(gtx layout.Context) (toggle bool) {
	all := key.ModCtrl | key.ModShift | key.ModAlt
	filters := []event.Filter{
		key.FocusFilter{Target: &v.tag},
		transfer.TargetFilter{
			Target: &v.tag,
			Type:   "application/text",
		},
		pointer.Filter{Target: &v.tag, Kinds: pointer.Press},
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
				if unicode.IsPrint(r) {
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
			if e.Buttons.Contain(pointer.ButtonPrimary) {
				at := e.Position.Round()
				v.pressedAt(at, v.text.cell, gtx.Dp(cellH))
			}
		}
	}
}

// paste hands ed what the clipboard held, as one key
func paste(p *pane, e transfer.DataEvent) {
	if !p.isText() {
		return
	}
	r := e.Open()
	defer r.Close()
	if b, err := io.ReadAll(r); err == nil {
		p.send("key", object{"name": tty.KPaste, "text": string(b)})
	}
}

// draw is a frame: the news since the last one, the input, then the picture
func (v *window) draw(gtx layout.Context) (toggle, done bool) {
	for more := true; more; {
		select {
		case e := <-v.news:
			copied, over := v.s.handle(e)
			if copied != "" {
				gtx.Execute(
					clipboard.WriteCmd{
						Type: "application/text",
						Data: io.NopCloser(
							strings.NewReader(
								copied,
							),
						),
					},
				)
			}
			done = done || over
		default:
			more = false
		}
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

	v.hits = v.hits[:0]
	v.lay(gtx, v.s.frame, all, true, cw)
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
