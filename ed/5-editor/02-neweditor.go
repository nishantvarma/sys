func newEditor(b *buffer) *editor {
	e := &editor{b: b, nor: normalMode{}, ins: insertMode{}}
	e.mode = e.nor
	return e
}
