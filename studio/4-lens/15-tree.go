// tree is what the lens shows of the thing, as nodes for a view to draw,
// and how it saw it: the thing, the tree the lens makes of it, its parts
// under it, the one at cur's line the cursor's (mark). A typed lens makes
// its tree once for each text it reads (made), so a move only marks the
// cursor again. A lens with nothing to show says why, dim, in a thing of
// no kind.
func (l *lensing) tree(th thing, ln lens) (*outline.Node, string) {
	m := l.made
	if !ln.typed || th.note == "" || m.lens != l.lens ||
		m.path != th.path || m.note != th.note ||
		!bytes.Equal(m.text, th.text) {
		root, word, err := ln.nodes(th)
		if err != nil {
			why := &outline.Node{ID: ln.name + "#1", Kind: "waits",
				Text: err.Error()}
			root = &outline.Node{Kids: []*outline.Node{why}}
		}
		root.ID, root.Text = th.path, th.path
		m = made{l.lens, th.path, th.note, th.text, root, word}
		l.made = m
	}
	mark(m.tree, th.at)
	return m.tree, m.word
}
