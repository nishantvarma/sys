// buttons are the bar's, tmux's control ranges, which a press on the
// border sends on as studio tab (`tmux/conf:/MouseDown1Control0/`): the
// first tabs' by their place, then the panel's new terminal and its
// close. tmux keeps 7 to 9 for its own floating panes.
const (
	tabButtons  = 5 // control 0 to 4: the tab at place 1 to 5
	newButton   = 5
	closeButton = 6
)
