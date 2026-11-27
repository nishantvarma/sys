// tmux runs tmux's commands; a dry one only prints them, and answers
// with ids it makes up, so that a dry run reads as a real one would, and
// says the back session made once it parks a tab there (park).
type tmux struct {
	dry    bool
	n      int
	backed bool
}
