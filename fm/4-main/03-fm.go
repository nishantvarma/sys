type fm struct {
	t       *tty.Term
	ui      outline.View // shows fm, takes its keys: tty, a window, a host
	cwd     string
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
	armed   bool   // a q said ed holds work unsaved: the next ends anyway
	unsaved string // in a frame's slot, where ed names what it holds
	follows bool   // -f: fm goes where cur is said (follow)
	heard   string // cur as fm last read it, following
	keys    map[string]*binding
	order   []*binding
}
