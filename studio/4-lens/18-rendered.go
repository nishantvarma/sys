// rendered is the thing as it is made for the eye: a note written as its
// parts and the messages between them its picture, which the view lays
// out as sys/picture does; else its parts as its reader reads them,
// typeset, a doc, as far as the reader can read them, the word saying
// where it stopped. A text whose kind no reader knows is its own render,
// and its source shows it.
func rendered(th thing) (*outline.Node, string, error) {
	word := "what it makes for the eye"
	if th.note == "" {
		return nil, word, th.why
	}
	r, known := readerOf(th)
	if !known {
		return nil, word, errors.New("a text of no kind a reader" +
			" knows: its source is its render")
	}
	if r.suffix == "" {
		items := outline.Text(bytes.NewReader(th.text), th.note)
		if picture.Pictured(items) {
			return &outline.Node{Kind: "picture", Kids: items},
				word + ": a note's picture", nil
		}
	}
	word += ": " + r.thing + ", typeset"
	kids, err := r.read(bytes.NewReader(th.text), th.note)
	if err != nil {
		if kids == nil {
			return nil, word, err
		}
		word += ": " + err.Error()
	}
	return &outline.Node{Kind: "doc", Kids: kids}, word, nil
}
