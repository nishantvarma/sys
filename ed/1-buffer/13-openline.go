// openLine inserts a blank line after row and returns its position.
func (b *buffer) openLine(row int) pos {
	b.lines = slices.Insert(b.lines, row+1, []rune{})
	b.dirty = true
	return pos{row + 1, 0}
}
