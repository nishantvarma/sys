// run is one tmux command, its output trimmed: an id, when it asks for
// one. A failure says what tmux said.
func (t *tmux) run(args ...string) (string, error) {
	if t.dry {
		fmt.Println("tmux " + quote(args))
		t.n++
		if slices.Contains(args, "#{window_id}") {
			return "@" + strconv.Itoa(t.n), nil
		}
		return "%" + strconv.Itoa(t.n), nil
	}
	out, err := exec.Command("tmux", args...).Output()
	if ee, ok := err.(*exec.ExitError); ok {
		why := strings.TrimSpace(string(ee.Stderr))
		return "", fmt.Errorf("tmux %s: %s", args[0], why)
	}
	return strings.TrimSpace(string(out)), err
}
