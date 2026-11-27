// back is the session a frame's window parks its tabs in while they are
// out of sight (park), one window a tab: a session of its own, so that
// its windows stay out of the window list a person turns through.
func back(win string) string {
	return "studio-" + strings.TrimPrefix(win, "@")
}
