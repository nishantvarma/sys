// tmux runs tmux's commands; a dry one only prints them, and answers
// with ids it makes up, so that a dry run reads as a real one would, and
// says the back session made the first time a tab waits there (backed).
type tmux struct {
	dry  bool
	n    int
	made bool
}
