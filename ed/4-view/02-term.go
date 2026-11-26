// term is ed's terminal view: the outline drawn on tty, and keys read from
// it, the mouse's among them. The window is the view's own: top and off
// are the first line and visual column it shows, moved only to keep the
// cursor in sight. It keeps the lines it drew last, which a click lands
// in, the mode it said last, whether the left button is dragging, and
// whether a drag has ended, for the next draw to put its selection on X's
// primary; and the last press, which a second at the same cell soon after
// makes a double tap. The lens it shows is its own too (lenses), and so
// is sys/view, which draws the lenses after source, with what it drew
// last.
type term struct {
	t        *tty.Term
	top, off int
	last     *outline.Node
	mode     string
	dragged  bool
	ended    bool
	press    time.Time
	px, py   int
	lens     int           // the lens shown, of lenses
	seen     *view.Term    // the view that draws the lenses after source
	items    *outline.Node // what it drew last (itemsOf)
}
