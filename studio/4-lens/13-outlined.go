// outlined is the thing read by its indent, as fm reads it (outline.Text):
// its items, each under the one it hangs from, all of them open. A
// folder's outline is fm's own tree, beside this.
func outlined(path string) ([]*outline.Node, error) {
	if fi, err := os.Stat(path); err == nil && fi.IsDir() {
		return nil, errors.New("a folder's outline is fm's tree")
	}
	b, err := outline.File(path)
	if err != nil {
		return nil, err
	}
	items := outline.Text(bytes.NewReader(b), path)
	unfold(items)
	return items, nil
}
