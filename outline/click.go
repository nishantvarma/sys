package outline

// Clicked is what a right or a middle click on line i of a text names,
// at rune at counted from 1 (`design/mouse`): the selection, when the
// click falls in it, as acme's buttons take it, else the word there. A
// text is a tree of lines as ed draws one, each line's runs marking the
// selection and the cursor, and a line holds runs only where one of them
// is. ed's terminal view and gio's window ask it alike.
func Clicked(text *Node, i, at int) string {
	if i < 0 || i >= len(text.Kids) {
		return ""
	}
	ln := text.Kids[i]
	r := 0
	for _, run := range ln.Kids {
		n := len([]rune(run.Text))
		if at-1 >= r && at-1 < r+n {
			if run.Sel {
				return Selected(text)
			}
			break
		}
		r += n
	}
	return wordAt([]rune(ln.Text), at-1)
}

// Selected is the text under the selection in the lines text holds: the
// runs selected, a selected line break a newline.
func Selected(text *Node) string {
	var b []rune
	for _, ln := range text.Kids {
		for _, r := range ln.Kids {
			if r.Sel {
				b = append(b, []rune(r.Text)...)
			}
		}
	}
	return string(b)
}

// wordAt is the word around rune i of line, a run of runes between
// blanks, as tmux takes the word under a click
// (`tmux/conf:/word-separators/`); empty on a blank.
func wordAt(line []rune, i int) string {
	space := func(j int) bool { return line[j] == ' ' || line[j] == '\t' }
	if i < 0 || i >= len(line) || space(i) {
		return ""
	}
	a, z := i, i
	for a > 0 && !space(a-1) {
		a--
	}
	for z+1 < len(line) && !space(z+1) {
		z++
	}
	return string(line[a : z+1])
}
