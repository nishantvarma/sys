// same is whether a and b name one file, by whatever path: a link's, or
// one relative where the other is absolute.
func same(a, b string) bool {
	if a == b {
		return true
	}
	fa, err := os.Stat(a)
	if err != nil {
		return false
	}
	fb, err := os.Stat(b)
	return err == nil && os.SameFile(fa, fb)
}
