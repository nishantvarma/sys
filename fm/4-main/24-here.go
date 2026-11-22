// here is where the pronoun points, asked of c rather than read: c owns
// where cur is kept, and answers the cwd when nothing is set. So an fm
// started with no argument opens on what the last verb was about: cur is a
// row, so fm stands in the folder that holds it, on it; the cwd is a place.
func here() (path string, row bool) {
	out, err := exec.Command("c").Output()
	p := strings.TrimSpace(string(out))
	if err != nil || p == "" {
		return ".", false
	}
	at, _ := os.Stat(p)
	wd, _ := os.Stat(".")
	if os.SameFile(at, wd) { // nothing set: c answered the cwd
		return ".", false
	}
	return p, true
}
