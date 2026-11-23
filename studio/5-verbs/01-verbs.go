// verbs are studio's words after its name, each with what does it;
// anything else is a frame to lay out (openVerb).
var verbs = map[string]func(args []string) error{
	"page":   pageVerb,
	"figure": figureVerb,
	"lens":   lensVerb,
	"send":   sendVerb,
	"read":   readVerb,
	"show":   showVerb,
	"close":  closeVerb,
}
