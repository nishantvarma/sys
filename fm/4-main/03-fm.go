type fm struct {
	t       *tty.Term
	ui      outline.View // what shows fm and takes its keys: the terminal
	cwd     string
	root    string           // the tree's top: where fm began, or above it
	open    map[string]bool  // what a view asked to see inside
	item    string           // the item the cursor is on, in a text file
	items   map[string]place // the items drawn last, by id: where each is
	last    string
	alt     string
	pat     string
	msg     string
	hit     string // the file a pending search line belongs to
	this    string // the cur file this window's pronoun is kept in
	said    string // what was last written there, so a still row is silent
	pos     map[string]string
	sel     map[string]bool
	clip    []string
	files   []entry
	idx     int
	line    int // a content search's line, waiting for the next edit
	cutting bool
	on      bool // this pane has the eye, so the row is worth saying
	hidden  bool
	done    bool
	keys    map[string]*binding
	order   []*binding
}
