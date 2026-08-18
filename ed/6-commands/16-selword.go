func (e *editor) selWord() {
	e.selectWith(func(b *buffer, p pos) pos {
		return back(b, wordFwd(b, fwd(b, p)))
	})
}
