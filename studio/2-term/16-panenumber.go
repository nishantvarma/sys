// paneNumber is the number in a pane's id, %12's 12: the order tmux made
// the panes in.
func paneNumber(id string) int {
	n, _ := strconv.Atoi(strings.TrimPrefix(id, "%"))
	return n
}
