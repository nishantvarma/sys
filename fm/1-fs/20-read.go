// read is what a file holds as an outline: a text file its items, read by
// their indent (outline.Text). ok is false when p is no note: not a
// regular file of text, or over a megabyte, for it is read on every key
// (outline.File).
func read(p string) (kids []*outline.Node, ok bool) {
	b, err := outline.File(p)
	if err != nil {
		return nil, false
	}
	return outline.Text(bytes.NewReader(b), p), true
}
