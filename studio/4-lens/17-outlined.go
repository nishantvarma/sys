// outlined is the thing's outline: its note's parts as the reader for its
// kind reads them (readerOf), a tree a view folds, each part named by its
// line, so that a key on a row is a key on that line. What its reader
// cannot read, a line half typed, it reads as far as it can, and else by
// its indent, the floor; and the word says where the reader stopped.
func outlined(th thing) (*outline.Node, string, error) {
	if th.note == "" {
		return nil, "its parts", th.why
	}
	r, _ := readerOf(th)
	word := r.thing + ", by " + r.by
	kids, err := r.read(bytes.NewReader(th.text), th.note)
	if err != nil {
		word += ": " + err.Error()
	}
	if kids == nil {
		kids = outline.Text(bytes.NewReader(th.text), th.note)
	}
	return &outline.Node{Kind: "outline", Kids: kids}, word, nil
}
