// paint is row r as line i: its number, its depth in the tree, and its
// name in its kind's look, with … after one shut with more inside.
func paint(i int, r row) string {
	color, suf := look(r.n.Kind)
	num := tty.Yellow(fmt.Sprintf("%2d", i+1))
	name := color(r.n.Text + suf)
	if !r.n.Open && len(r.n.Kids) > 0 {
		name += tty.Dim(" …")
	}
	in := strings.Repeat("  ", r.depth)
	return fmt.Sprintf(" %s  %s%s%s", num, in, hl(r.n), name)
}
