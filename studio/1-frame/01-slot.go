// slot is a place in the frame, a pane or a split into slots: its name,
// which a program reaches it by; its share of its parent, a percent, which
// the layout gives it; what runs in it, its first tab's, empty for a shell;
// where the frame says it, path:line; the slots it is split into, if any;
// and its tabs, the programs that take turns in it, the first shown.
type slot struct {
	name  string
	share int
	cmd   string
	at    string
	kids  []*slot
	tabs  []*tab
}
