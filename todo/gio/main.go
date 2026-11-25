// sys's windows in gio: a poc. With no argument it is `window`, the view at
// the other end of a program's pipe, as sh/window is. -studio dir hosts
// studio/frame's top row, fm beside ed -f. Only this file knows gio.
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
	"time"
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

type view struct {
	s        *studio
	news     <-chan news // what the next frame takes
	frame    bool        // the studio's frame, else one pane
	text     *typeface
	pictures map[look]paint.ImageOp
	tag      int // what events are addressed to

	focused   bool
	full      bool
	copied    string
	clicked   string
	clickedAt time.Time
	hits      []hit // where each pane was drawn last, for a click
}

type hit struct {
	area  image.Rectangle
	i     int
	lines []spot
}

func newView(s *studio, frame bool) (*view, error) {
	const dir = "/usr/share/fonts/truetype/dejavu/"
	t, err := newTypeface(
		size,
		dir+"DejaVuSansMono.ttf",
		dir+"DejaVuSansMono-Bold.ttf",
	)
	if err != nil {
		return nil, err
	}
	return &view{s: s, news: s.events, frame: frame, text: t,
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
func (v *view) label(
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
func (v *view) line(gtx layout.Context, k *strip, at image.Point, cw int) {
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

// slot draws pane i in area, telling its program the size when that changes
func (v *view) slot(gtx layout.Context, i int, area image.Rectangle, cw int) {
	p := v.s.panes[i]
	cols, high := area.Dx()/cw, area.Dy()/gtx.Dp(cellH)
	if cols != p.cols || high != p.high {
		p.cols, p.high = cols, high
		p.send("size", object{"w": cols, "h": high})
	}
	var lines []spot
	if p.isText() {
		lines = p.source()
	} else {
		lines = p.list()
	}
	if p.asking && len(lines) > 0 {
		lines[len(lines)-1] = spot{
			strip: (&strip{}).add(
				" "+p.prompt+p.answer+"▏",
				"",
				"",
				false,
			),
		}
	}
	defer clip.Rect(area).Push(gtx.Ops).Pop()
	for y, l := range lines {
		v.line(
			gtx,
			l.strip,
			area.Min.Add(image.Pt(0, y*gtx.Dp(cellH))),
			cw,
		)
	}
	v.hits = append(v.hits, hit{area, i, lines})
}

func (v *view) pressedAt(at image.Point, h int) {
	for _, x := range v.hits {
		if !at.In(x.area) {
			continue
		}
		v.s.focus(x.i)
		y := (at.Y - x.area.Min.Y) / h
		if y >= len(x.lines) || x.lines[y].id == "" {
			return
		}
		v.s.panes[x.i].click(x.lines[y].id)
	}
}

var named = map[key.Name]string{
	key.NameReturn: "ENTER", key.NameEnter: "ENTER", key.NameEscape: "ESC",
	key.NameTab: "TAB", key.NameDeleteBackward: "BACKSPACE",
	key.NameUpArrow: "UP", key.NameDownArrow: "DOWN",
	key.NameLeftArrow: "LEFT", key.NameRightArrow: "RIGHT",
	key.NamePageUp: "PGUP", key.NamePageDown: "PGDN",
}

// input does what came in since the last frame; it is the window's mode it
// returns when f11 asks for the other, else nil
func (v *view) input(gtx layout.Context) (toggle bool) {
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
		p := v.s.panes[v.s.eye]
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
				v.s.focus(1 - v.s.eye)
			case ctrl && e.Name == "C":
				p.pressed("INT")
			case ctrl && e.Modifiers.Contain(key.ModShift) &&
				e.Name == "V":
				gtx.Execute(clipboard.ReadCmd{Tag: &v.tag})
			default:
				p.pressed(named[e.Name])
			}
		case transfer.DataEvent:
			paste(p, e)
		case pointer.Event:
			v.pressedAt(e.Position.Round(), gtx.Dp(cellH))
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
		p.send("key", object{"name": "PASTE", "text": string(b)})
	}
}

// draw is a frame: the news since the last one, the input, then the picture
func (v *view) draw(gtx layout.Context) (toggle, done bool) {
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
	if !v.frame {
		v.slot(gtx, 0, all, cw)
		return
	}
	// the frame's shares: the top 70%, in it fm 40%
	h := gtx.Dp(cellH)
	top := int(float32(all.Dy())*0.7) / h * h
	left := int(float32(all.Dx()) * 0.4)
	v.slot(gtx, 0, image.Rect(0, 0, left, top), cw)
	v.slot(gtx, 1, image.Rect(left+1, 0, all.Dx(), top), cw)
	fill(gtx, image.Rect(left, 0, left+1, top), hex("#262b28"))
	fill(gtx, image.Rect(0, top, all.Dx(), top+1), hex("#262b28"))
	eye := v.s.panes[v.s.eye].name
	for y, said := range []string{
		" shell: not in this poc, it needs a terminal",
		" keys go to " + eye + " · ctrl-w the other · click picks"} {
		v.line(
			gtx,
			(&strip{}).add(said, "#6f7873", "", false),
			image.Pt(0, top+1+y*h),
			cw,
		)
	}
	return
}

// listen has wake called for each piece of news, which comes from the
// programs' own threads; a frame then takes it. The senders keep their
// channel: only what the frames read changes
func (v *view) listen(wake func()) {
	waiting := make(chan news, 64)
	v.news = waiting
	go func() {
		for e := range v.s.events {
			waiting <- e
			wake()
		}
	}()
}

func (v *view) run(w *gioapp.Window) error {
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
		"host fm and ed -f on the directory given",
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
	v, err := newView(s, *frame)
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
