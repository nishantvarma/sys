// paneOf is the pane the studio window keeps for a slot, by its name.
func paneOf(name string) (string, error) {
	win, err := target()
	if err != nil {
		return "", err
	}
	return slotPane(&tmux{}, win, name)
}
