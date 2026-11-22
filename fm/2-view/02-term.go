// term is fm's terminal view: the outline drawn on tty, and the keys read
// from it. The layout is its own: the cursor's folder as a list, the tree,
// or the folder as tiles; and it moves the cursor itself among what it
// shows, telling fm only where it landed.
type term struct {
	t      *tty.Term
	last   *outline.Node // the tree drawn last
	status string
	state  []string // fm's state, drawn last
	lay    string   // "" the folder, "tree" or "tiles"
}
