// looks is how sys/view draws the lens's kinds: a diff's added lines
// green, its removed lines red and its hunks cyan, as git colours them,
// and its header dim, as a placeholder's words are.
var looks = map[string]view.Look{
	"add":   {Color: tty.Green},
	"del":   {Color: tty.Red},
	"hunk":  {Color: tty.Cyan},
	"meta":  {Color: tty.Dim},
	"waits": {Color: tty.Dim},
}
