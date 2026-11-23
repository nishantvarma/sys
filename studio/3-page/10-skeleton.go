// skeleton is the page a verb fills: the one at p, or sys's own,
// `medium/page.html`. One without a main to fill, or without local's mark,
// is no skeleton, an error at its path.
func skeleton(p string) (string, error) {
	if p == "" {
		p = filepath.Join(root(), "medium", "page.html")
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return "", err
	}
	s := string(b)
	i, j := strings.Index(s, "<main>"), strings.Index(s, "</main>")
	switch {
	case i < 0 || j < i:
		return "", fmt.Errorf("%s: no <main> to fill", p)
	case !strings.Contains(s, localMark):
		return "", fmt.Errorf("%s: no %s to put a look under", p,
			localMark)
	}
	return s, nil
}
