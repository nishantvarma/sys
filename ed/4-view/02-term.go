// term is ed's terminal view: the outline drawn on tty, and keys read from
// it. The window is the view's own: top and off are the first line and
// visual column it shows, moved only to keep the cursor in sight.
type term struct {
	t        *tty.Term
	top, off int
}
