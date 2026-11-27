// hooks are a frame's window's own, each with its command, given what a
// pane's end runs (open); plain takes them away again. A pane ends, and
// a tab's dies, staying dead for closeVerb to show the next in its place.
func hooks(end string) [][2]string {
	return [][2]string{{"window-layout-changed", ""}, {"pane-exited", end},
		{"pane-died", end}}
}
