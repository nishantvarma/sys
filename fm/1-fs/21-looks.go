// looks is how sys/view draws fm's kinds (kind): each in its colour, and
// with the mark ls -F gives it after its name; a file is plain.
var looks = map[string]view.Look{
	"bad":  {Color: tty.Red, Mark: "@"},
	"link": {Color: tty.Cyan, Mark: "@"},
	"dir":  {Color: tty.Blue, Mark: "/"},
	"exec": {Color: tty.Green, Mark: "*"},
}
