// chordKeys are the keys every editor has, ctrl and a letter, the same in
// insert as in normal, so that a person on the mouse need not leave
// insert. ctrl-v is the view's, which makes it a paste.
func chordKeys() []binding {
	return []binding{
		{tty.KInt, "copy +clipboard", (*editor).copySel},
		{"C-x", "cut +clipboard", (*editor).cutSel},
		{"C-v", "paste the clipboard", nil},
		{"C-z", "undo", (*editor).undo},
		{"C-y", "redo", (*editor).redo},
		{"C-s", "save", (*editor).doSave},
	}
}
