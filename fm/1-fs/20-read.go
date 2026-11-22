// read is what a file holds as an outline: a text file its items, read by
// their indent (outline.Text). ok is false when p is not text: text has no
// NUL byte in it, and is not over a megabyte, for it is read on every key.
func read(p string) (kids []*outline.Node, ok bool) {
	fi, err := os.Stat(p)
	if err != nil || !fi.Mode().IsRegular() || fi.Size() > 1<<20 {
		return nil, false
	}
	b, err := os.ReadFile(p)
	if err != nil || bytes.IndexByte(b, 0) >= 0 {
		return nil, false
	}
	return outline.Text(bytes.NewReader(b), p), true
}
