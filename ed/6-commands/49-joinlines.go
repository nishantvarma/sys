// joinLines pulls the next line onto this one by cutting the break the dd
// model already treats as the position at col == len.
func (e *editor) joinLines() {
	l := e.head().line
	if l >= e.b.last() {
		return
	}
	e.b.snapshot(e.caret)
	end := pos{l, e.b.lineLen(l)}
	e.b.cut(end, end)
	e.place(end)
}
