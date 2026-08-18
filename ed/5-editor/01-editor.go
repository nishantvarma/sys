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
}
