// tabPane is a tab as tmux holds it: its pane, the window the pane is in,
// the frame's when it is the tab shown, its name, whether what ran in it
// has ended, and what runs in it now, as a terminal's menu says it.
type tabPane struct {
	pane, window, name string
	dead               bool
	cmd                string
}
