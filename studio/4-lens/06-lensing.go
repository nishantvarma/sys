// lensing is the lens's own state: the window's cur and lens files it
// follows, the address it is on, cur as it last read it, whether its pane
// has the eye, so that it may say a move of its own to cur, a word to say
// once in place of the lens's own, and whether it runs in a frame's slot,
// where q is fm's.
type lensing struct {
	curf, lensf string
	at, said    string
	on          bool
	msg         string
	framed      bool
}
