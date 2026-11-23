// tmux runs tmux's commands; a dry one only prints them, and answers
// with ids it makes up, so that a dry run reads as a real one would.
type tmux struct {
	dry bool
	n   int
}
