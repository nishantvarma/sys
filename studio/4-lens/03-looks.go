// looks is how sys/view draws the lens's kinds, from tty's palette: a
// diff's added lines as additions are, its removed lines as removals are
// and its hunks as links, as git colours them in its way, and its header
// faint, as a placeholder's words are.
var looks = map[string]view.Look{
	"add":   {Color: tty.Plus},
	"del":   {Color: tty.Minus},
	"hunk":  {Color: tty.Link},
	"meta":  {Color: tty.Faint},
	"waits": {Color: tty.Faint},
}
