// slot is one line of the frame, a pane or a split into slots: its name,
// which a program reaches it by; its share of its parent, a percent, 0
// when it takes an even part of what the others leave; what runs in it,
// empty for a shell; where the frame says it, path:line; and the slots
// it is split into, if any.
type slot struct {
	name  string
	share int
	cmd   string
	at    string
	kids  []*slot
}
