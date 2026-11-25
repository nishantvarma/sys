// hooks are a frame's window's own, each with its command, given what
// ends the frame (open); plain takes them away again.
func hooks(end string) [][2]string {
	return [][2]string{{"window-layout-changed", ""}, {"pane-exited", end}}
}
