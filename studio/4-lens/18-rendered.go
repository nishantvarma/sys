// rendered is the thing as it is made for the eye, its render: for a
// note, all of its items as they nest, which the view lays out as its
// parts and the messages between them (sys/picture) and draws. A renderer
// for another kind of thing joins here, by that kind.
func rendered(th thing) ([]*outline.Node, error) {
	if th.note == "" {
		return nil, th.why
	}
	return outline.Text(bytes.NewReader(th.text), th.note), nil
}
