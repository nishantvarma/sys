// git runs git in dir: what it prints, or, when it fails, the first line
// of why, in git's own words.
func git(dir string, args ...string) ([]byte, error) {
	out, err := exec.Command("git", append([]string{"-C", dir},
		args...)...).Output()
	if ee, ok := err.(*exec.ExitError); ok {
		why, _, _ := strings.Cut(strings.TrimSpace(string(ee.Stderr)),
			"\n")
		return nil, errors.New(why)
	}
	return out, err
}
