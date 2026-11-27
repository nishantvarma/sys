// tabPane is a tab as tmux holds it: its pane, the window the pane is in,
// the frame's when it is the tab shown, its name, and whether what ran in
// it has ended.
type tabPane struct {
	pane, window, name string
	dead               bool
}
