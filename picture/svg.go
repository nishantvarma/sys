package picture

import (
	"fmt"
	"html"
	"math"
	"strconv"
	"strings"
)

// Style is a picture's look on its page, for the skeleton's local mark
// (`general/medium`), in the page's own tokens so that it follows the page
// into the dark: a box is a card, ground or raised as the page's cards
// are, and a message is in the accent. Its classes are the ones SVG
// writes.
const Style = `
    svg.picture {
      display: block;
      max-width: 100%;
      height: auto;
      margin: 0.25rem 0 0.5rem;
      font: 13px var(--mono);
    }
    .picture .ground {
      fill: var(--bg);
      stroke: var(--line);
    }
    .picture .raised {
      fill: var(--panel);
      stroke: var(--line);
    }
    .picture text {
      fill: var(--fg);
      dominant-baseline: central;
    }
    .picture .lane,
    .picture .label {
      font-weight: 600;
    }
    .picture .note {
      fill: var(--dim);
    }
    .picture .msg {
      fill: none;
      stroke: var(--accent);
      stroke-width: 1.25;
    }
    .picture .head,
    .picture .said {
      fill: var(--accent);
    }
    .picture .said {
      paint-order: stroke;
      stroke: var(--panel);
      stroke-width: 4px;
      stroke-linejoin: round;
    }`

// px is a length in px as an SVG attribute says it, to a tenth.
func px(v float64) string {
	return strconv.FormatFloat(math.Round(v*10)/10, 'f', -1, 64)
}

// text writes t at x, y, in cells, its anchor there, with the look its
// class gives it.
func text(s *strings.Builder, x, y float64, anchor, class, t string) {
	fmt.Fprintf(s, `<text class="%s" x="%s" y="%s" text-anchor="%s">%s`+
		`</text>`+"\n", class, px(x*cellW), px(y*cellH), anchor,
		html.EscapeString(t))
}

// svgSaid writes what m carries beside its line, when it carries a word,
// in a group that carries its item's address, as a box's does.
func (m *arrow) svgSaid(s *strings.Builder) {
	if m.label == "" {
		return
	}
	fmt.Fprintf(s, `<g data-id="%s">`+"\n", html.EscapeString(m.at))
	text(s, m.said[0], m.said[1], m.anchor, "said", m.label)
	s.WriteString("</g>\n")
}

// svg draws m: its line from its tail, and its head where it arrives.
func (m *arrow) svg(s *strings.Builder) {
	x0, y0 := m.tail[0]*cellW, m.tail[1]*cellH
	x1, y1 := m.head[0]*cellW, m.head[1]*cellH
	l := math.Hypot(x1-x0, y1-y0)
	ux, uy := (x1-x0)/l, (y1-y0)/l
	const long, half = 7.0, 3.5 // the head's length, and half its width
	bx, by := x1-ux*long, y1-uy*long
	fmt.Fprintf(s, `<path class="msg" d="M%s %sL%s %s"/>`+"\n",
		px(x0), px(y0), px(bx+ux), px(by+uy))
	fmt.Fprintf(s, `<path class="head" d="M%s %sL%s %sL%s %sZ"/>`+"\n",
		px(x1), px(y1), px(bx-uy*half), px(by+ux*half),
		px(bx+uy*half), px(by-ux*half))
}

// svg draws b and what it holds, as the page draws its cards: on the
// ground at an even depth, raised at an odd one. The card and all in it
// are a group that carries its item's address, so that a click anywhere
// on it, its words too, names the innermost box (At). A box that holds
// none has its label and note in it, centred, from its top; one that
// holds boxes has its title at its top, unless it is a lane, and its
// boxes under it.
func (b *box) svg(s *strings.Builder, depth int) {
	look := "ground"
	if depth%2 == 1 {
		look = "raised"
	}
	fmt.Fprintf(s, `<g data-id="%s">`+"\n", html.EscapeString(b.at))
	defer s.WriteString("</g>\n")
	fmt.Fprintf(s, `<rect class="%s" x="%s" y="%s" width="%s"`+
		` height="%s" rx="6"/>`+"\n", look,
		px((float64(b.x)+0.5)*cellW), px((float64(b.y)+0.5)*cellH),
		px(float64(b.w-1)*cellW), px(float64(b.h-1)*cellH))
	row := float64(b.y) + 1.5
	if len(b.kids) == 0 {
		mid := float64(b.x) + float64(b.w)/2
		if depth > 0 {
			text(s, mid, row, "middle", "label", b.label)
			row++
		}
		for i, t := range b.lines {
			text(s, mid, row+float64(i), "middle", "note", t)
		}
		return
	}
	if depth > 0 {
		x := float64(b.x) + 2
		text(s, x, row, "start", "label", b.label)
		x += float64(runes(b.label) + 2)
		text(s, x, row, "start", "note", b.note)
	}
	for _, k := range b.kids {
		k.svg(s, depth+1)
	}
}

// svgLane draws l: its label in the column at left, level with the first
// line in it, with its note under the label when it holds boxes; then the
// band itself.
func (f *Picture) svgLane(s *strings.Builder, l *box) {
	row := float64(l.y) + 1.5
	if len(l.kids) > 0 {
		row++
	}
	text(s, 0, row, "start", "lane", l.label)
	if len(l.kids) > 0 {
		for i, t := range l.lines {
			text(s, 0, row+1+float64(i), "start", "note", t)
		}
	}
	l.svg(s, 0)
}

// SVG is the picture on a page: the lanes and their boxes, the messages
// over them, and what the messages say over those. A cell is cellW by
// cellH px, and the colours are the page's, by class (see Style).
func (f *Picture) SVG() string {
	var s strings.Builder
	w, h := f.w*cellW, f.h*cellH
	fmt.Fprintf(&s, `<svg class="picture" viewBox="0 0 %d %d" width="%d"`+
		` height="%d" role="img" aria-label="%s">`+"\n",
		w, h, w, h, html.EscapeString(f.Title))
	for _, l := range f.lanes {
		f.svgLane(&s, l)
	}
	for _, m := range f.arrows {
		m.svg(&s)
	}
	for _, m := range f.arrows {
		m.svgSaid(&s)
	}
	s.WriteString("</svg>")
	return s.String()
}
