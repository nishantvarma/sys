type fm struct {
	t       *tty.Term
	cwd     string
	last    string
	alt     string
	pat     string
	msg     string
	tags    string
	pos     map[string]int
	sel     map[string]bool
	clip    []string
	files   []entry
	idx     int
	cutting bool
	hidden  bool
	done    bool
	keys    map[string]*binding
	order   []*binding
}
