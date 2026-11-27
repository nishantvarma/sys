// lensing is the lens's own state: the window's files it follows, cur and
// what ed holds unsaved, its name and its text; the lens shown, of lenses,
// and whether it is fixed, a tab of the frame's lens view, whose Tab turns
// the frame's tabs; the address it is on, cur as it last read it, whether
// its pane has the eye, so that it may say a move of its own to cur, a
// word to say once in place of the lens's own, and the frame's slot it
// runs in, if any, where q is fm's.
type lensing struct {
	curf, unsavedf, textf string
	lens                  int
	fixed                 bool
	at, said              string
	on                    bool
	msg                   string
	slot                  string
	framed                bool
}
