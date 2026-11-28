// readerOf is the reader for the thing's kind, by its name's suffix, a
// script, which starts #!, no note though its name has no dot; and for a
// kind none reads, the floor, its lines by their indent (outline.Text),
// and false.
func readerOf(th thing) (reader, bool) {
	suffix := filepath.Ext(th.note)
	if suffix == "" && bytes.HasPrefix(th.text, []byte("#!")) {
		suffix = "#!"
	}
	for _, r := range readers {
		if r.suffix == suffix {
			return r, true
		}
	}
	return reader{thing: "a text", by: "its indent",
		read: func(r io.Reader, id string) ([]*outline.Node, error) {
			return outline.Text(r, id), nil
		}}, false
}
