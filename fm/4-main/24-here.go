// here is where the pronoun points, asked of c rather than read: c owns
// where cur is kept, and answers the cwd when nothing is set. So an fm
// started with no argument opens on what the last verb was about.
func here() string {
	out, err := exec.Command("c").Output()
	if err != nil {
		return "."
	}
	if p := strings.TrimSpace(string(out)); p != "" {
		return p
	}
	return "."
}
