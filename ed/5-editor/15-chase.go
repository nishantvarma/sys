// chase follows cur, as ed -f does in a studio's slot: source is ed, so
// the thing at fm's cursor is edited where it shows. When cur names
// another file, ed edits it from cur's line, unless its own file is
// unsaved: that holds it, saying so as cur moves, until a save lets it
// go. A buffer with no file holds nothing, for no save can let it go.
// Within its own file, by whatever path cur names it, the cursor goes to
// cur's line as cur moves.
func (e *editor) chase() {
	at := state.Read(e.curf)
	moved := at != e.chased
	if at == "" || !moved && (!e.held || e.b.dirty) {
		return
	}
	e.chased = at
	path, line, ok := outline.Note(at)
	if !ok {
		return
	}
	if same(path, e.where()) {
		e.held = false
		if moved && line > 0 && line-1 != e.head().line {
			e.place(pos{line - 1, 0})
		}
		return
	}
	if e.b.dirty && e.b.path != "" {
		e.held = true
		if moved {
			e.msg = "held, unsaved: cur is at " + tilde(at)
		}
		return
	}
	b, err := load(path)
	if err != nil {
		e.msg = err.Error()
		return
	}
	e.held = false
	e.swap(b, line)
}
