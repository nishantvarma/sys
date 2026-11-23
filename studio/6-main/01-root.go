// root is sys's own folder, the one above bin/ where studio is, so the
// frame and the page's skeleton are found from wherever it runs.
func root() string {
	p, err := os.Executable()
	if err == nil {
		p, err = filepath.EvalSymlinks(p)
	}
	if err != nil {
		return "."
	}
	return filepath.Dir(filepath.Dir(p))
}
