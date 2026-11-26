// verbs are studio's words after its name, each with what does it;
// anything else is a frame to lay out (openVerb).
var verbs = map[string]func(args []string) error{
	"page":    pageVerb,
	"picture": pictureVerb,
	"send":    sendVerb,
	"read":    readVerb,
	"close":   closeVerb,
	"slots":   slotsVerb,
}
