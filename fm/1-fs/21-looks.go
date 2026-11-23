// looks is how sys/view draws fm's kinds (kind): each in its colour from
// tty's palette, with the mark ls -F gives it after its name, and the
// word a card says it by; a broken link's target is said in its own red.
var looks = map[string]view.Look{
	"bad": {Color: tty.Minus, Mark: "@", Word: "broken link",
		Said: tty.Gone},
	"link": {Color: tty.Link, Mark: "@", Word: "link"},
	"dir":  {Color: tty.Accent, Mark: "/", Word: "folder"},
	"exec": {Color: tty.Plus, Mark: "*", Word: "program"},
	"file": {Word: "file"},
}
