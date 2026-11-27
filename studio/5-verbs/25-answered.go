// answered is what a word from a binding gets back, one that names its
// pane, -t, as a press on a tab line or alt-t does: it has no terminal to
// read an error on, and tmux would show one over the pane, as a job's
// output, so the error goes to tmux's message line instead, and the job
// ends well.
func answered(pane string, err error) error {
	if err != nil {
		(&tmux{}).run("display-message", "-t", pane,
			"studio: "+err.Error())
	}
	return nil
}
