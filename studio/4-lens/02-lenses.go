// lenses are the ways the lens sees the thing at cur, in the order Tab
// turns through them, each read and never typed: its outline and its
// render, from the text ed holds of it while that is unsaved, else from
// the file (noted); and its diff against the last commit, of the file as
// saved. A way to see a thing joins as a line here.
var lenses = []lens{
	{name: "outline", by: "its items, by their indent",
		nodes: outlined, kind: "outline", typed: true},
	{name: "render", by: "what it makes for the eye: a note's picture",
		nodes: rendered, kind: "picture", typed: true},
	{name: "diff", by: "git diff, against the last commit",
		nodes: diffed},
}
