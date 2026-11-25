// clicked is what a right or a middle click on line i, at rune at counted
// from 1, hands on: the selection, when the click falls in it, as acme's
// buttons take it, else the word there (wordAt). A line holds runs only
// where the selection or the cursor is (`ed/7-loop/02-line.go`).
func (v *term) clicked(i, at int) string {
	ln := v.last.Kids[i]
	r := 0
	for _, run := range ln.Kids {
		n := len([]rune(run.Text))
		if at-1 >= r && at-1 < r+n {
			if run.Sel {
				return selected(v.last)
			}
			break
		}
		r += n
	}
	return wordAt([]rune(ln.Text), at-1)
}
