// frame is a note read as panes: its title, its lead, its slots, which
// stack down, its panel, the terminals (layout), and the note itself,
// where it is and as written.
type frame struct {
	title, lead string
	slots       []*slot
	panel       *slot
	note, src   string
}
