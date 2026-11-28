// python is a Python file's parts as page/py reads them, by Python's own
// ast (`medium/page/py:/def tree/`): a program run on the text, which
// prints them as JSON, and says why when it cannot, in its last line.
func python(r io.Reader, id string) ([]*outline.Node, error) {
	py := filepath.Join(root(), "medium", "page", "py")
	c := exec.Command(py, "-t", id)
	c.Stdin = r
	out, err := c.Output()
	if ee, ok := err.(*exec.ExitError); ok {
		why := strings.TrimSpace(string(ee.Stderr))
		return nil, errors.New(why[strings.LastIndex(why, "\n")+1:])
	}
	if err != nil {
		return nil, err
	}
	var parts []*outline.Node
	return parts, json.Unmarshal(out, &parts)
}
