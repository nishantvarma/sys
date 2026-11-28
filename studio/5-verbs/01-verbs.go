// verbs are studio's words after its name, by name, each with what does
// it; anything else is a frame to lay out (openVerb).
var verbs = map[string]func(args []string) error{
	"close":   closeVerb,
	"hand":    handVerb,
	"page":    pageVerb,
	"panel":   panelVerb,
	"picture": pictureVerb,
	"read":    readVerb,
	"send":    sendVerb,
	"show":    showVerb,
	"slots":   slotsVerb,
	"tab":     tabVerb,
}
