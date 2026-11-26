// lenses are the ways ed's view shows the text it holds, in the order Tab
// turns through them (`design/lens`): source, the text as written and the
// only one typed in; outline, its items by their indent, as rows or as
// thumbs; and render, what the source makes for a person to see, for a
// note its picture, as sys/picture lays it out. The two after source are
// read, never typed, and drawn from the lines as they are, unsaved ones
// too.
var lenses = []string{"source", "outline", "render"}

// turn shows the next lens, and source again after the last.
func (v *term) turn() {
	v.lens = (v.lens + 1) % len(lenses)
}
