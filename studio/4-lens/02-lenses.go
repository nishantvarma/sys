// lenses are the frame's, in order: what is derived from the thing at cur,
// in motion, and so read, never typed. The thing as written is ed's, in a
// slot of its own: source is ed (`design/todo:/typed at rest/`), and a
// note's outline is its indent there. diff and figure show; run is a
// placeholder until what it waits on is built.
var lenses = []lens{
	{name: "diff", by: "git diff, against the last commit", nodes: diffed},
	{name: "figure", by: "its parts and messages, laid out by sys/figure",
		nodes: figured, kind: "figure"},
	{name: "run", by: "not built yet", waits: []string{
		"run shows the last run of the tests that cover the thing.",
		"It waits on tests that write what they found to a file.",
	}},
}
