// deleteToEnd deletes through the last rune of the line, not lineEnd's
// col == len(line): that position rests on the line break (cut treats it
// as inclusive, per cut's doc comment), which would join the next line on.
func (e *editor) deleteToEnd() {
	if e.b.lineLen(e.head().line) == 0 {
		return
	}
	e.selectWith(func(b *buffer, p pos) pos {
		return pos{p.line, b.lineLen(p.line) - 1}
	})
	e.del()
}
