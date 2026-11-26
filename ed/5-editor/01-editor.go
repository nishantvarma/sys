type editor struct {
	t     *tty.Term
	ui    outline.View // what shows ed and takes its keys: the terminal
	b     *buffer
	caret sel
	ext   bool
	goal  int // sought column: verticals aim at it, head writes record it
	mode  mode
	nor   mode
	ins   mode
	reg   []rune
	pat   string
	msg   string
	count int  // a pending numeric prefix; 0 means none, i.e. once
	armed bool // a dirty-quit was requested once
	done  bool
	// the insert under way is no change . can type again: the mouse
	// dragged in it, or an undo or a redo was done in it
	unkept bool
	lens   int // the lens shown, of lenses (`design/lens`)
	// the insert under way owes a snapshot, which its first edit takes
	// (pay)
	owed bool
	typed []tty.Key // keys read since normal mode was last idle
	again []tty.Key // the last typed that changed the text: what . types
	fed   []tty.Key // what . put back, read ahead of the view
	seen  int       // b.edits at that idle

	// with -f, ed chases cur (see chase)
	curf   string // the cur file; empty, ed does not chase
	chased string // cur as ed last chased it
	held   bool   // cur names another file, and an unsaved buffer waits

	// in a frame's slot, the window's unsaved file, which names ed's
	// file while it is unsaved (mark); empty outside a frame
	unsaved string
	marked  bool // the unsaved file names ed's
}
