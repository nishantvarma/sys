// paint is row r as line i: its number, its hint in the tree, and its name
// in its kind's look; outside the tree, where no hint says so, … after one
// shut with more inside.
func paint(i int, r row) string {
	color, suf := look(r.n.Kind)
	num := tty.Yellow(fmt.Sprintf("%2d", i+1))
	name := color(r.n.Text + suf)
	lead := ""
	if r.lead != "" {
		lead = r.lead
	} else if !r.n.Open && len(r.n.Kids) > 0 {
		name += " …"
	}
	return fmt.Sprintf(" %s  %s%s%s", num, lead, hl(r.n), name)
}
