// Package figure is a note drawn as its parts and the messages between
// them: a box an item, a message an item that starts with →. It lays them
// out in cells, as a terminal shows text, so that one layout draws as SVG
// on a page and as runes in a terminal; studio's figure lens draws it so,
// and studio figure makes its page. Each part keeps its item's address,
// so a click on one names it (At).
package figure

import "sys/outline"

// A figure is laid out in cells: a cell is a rune wide and a line high,
// so the layout that draws an SVG draws runes too. A box's border runs
// through the middle of its outer cells.
const (
	cellW = 8  // a cell's width on the page, in px: a rune of 13px mono
	cellH = 18 // a line's height, in px
	noteW = 18 // the runes a box's note wraps at
	laneW = 16 // the runes a lane's note wraps at, in the column at left
	gapW  = 2  // the cells between boxes side by side
	gapH  = 3  // the lines between lanes: room for a message's label
)

// box is a part: what it is called, what it is, the parts inside it, and
// where it is drawn, in cells. A box no box holds is a lane, a band
// across the whole figure, its label in the column at its left.
type box struct {
	label, note string
	at          string // its item's address, path:line
	kids        []*box
	lines       []string // the note, wrapped as it is drawn
	x, y, w, h  int
}

// arrow is a message from one box to another: what it carries, the line
// it is drawn on, from its tail to its head, and where its label goes and
// which end of the label is there (start, middle or end), in cells.
type arrow struct {
	from, to *box
	label    string
	at       string // its item's address, path:line
	tail     [2]float64
	head     [2]float64
	said     [2]float64
	anchor   string
}

// Figure is a note drawn: its title, the lead under it, its lanes from
// the top, the messages between their boxes, and its size in cells; left
// is the column the lanes' boxes start at, right of their labels.
type Figure struct {
	Title, Lead string
	lanes       []*box
	arrows      []*arrow
	left, w, h  int
}

// New is the figure a note's items make (see read), laid out, its
// messages routed, and room made for what they say. An error in an item
// starts with its path and line.
func New(items []*outline.Node) (*Figure, error) {
	f, err := read(items)
	if err != nil {
		return nil, err
	}
	f.lay()
	if err := f.route(); err != nil {
		return nil, err
	}
	f.fit()
	return f, nil
}
