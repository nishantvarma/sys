// swap puts another file in ed's buffer, as chasing cur does: the cursor
// at the line, normal mode, nothing pending and nothing for . to repeat,
// and no word about the file before. The register and the last search
// stay, as they do across a reload.
func (e *editor) swap(b *buffer, line int) {
	e.b, e.caret, e.ext, e.goal = b, sel{}, false, 0
	e.mode, e.count, e.armed, e.msg = e.nor, 0, false, ""
	e.unkept, e.owed = false, false
	e.typed, e.again, e.fed, e.seen = nil, nil, nil, b.edits
	if line > 0 {
		e.place(pos{line - 1, 0})
	}
}
