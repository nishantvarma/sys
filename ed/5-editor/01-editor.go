type editor struct {
	t     *tty.Term
	b     *buffer
	caret sel
	ext   bool
	goal  int // sought column: verticals aim at it, head writes record it
	top   int
	off   int
	mode  mode
	nor   mode
	ins   mode
	reg   []rune
	pat   string
	msg   string
	count int  // a pending numeric prefix; 0 means none, i.e. once
	armed bool // a dirty-quit was requested once
	done  bool
	typed []tty.Key // keys read since normal mode was last idle
	again []tty.Key // the last typed that changed the text: what . types
	seen  int       // b.edits at that idle
}
