// block is the cursor's item, as a note holds one: the line it is on, and
// the lines after it that stand further in, its kids; a blank line ends
// it. The first and the last, counted from 0.
func block(n *outline.Node, line int) (int, int) {
	in := indent(n.Kids[line].Text)
	end := line
	for end+1 < len(n.Kids) && indent(n.Kids[end+1].Text) > in {
		end++
	}
	return line, end
}
