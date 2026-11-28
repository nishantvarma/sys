// readers are the ways the lens reads a thing into its parts, the right
// tool for each kind, picked by its name's suffix (readerOf): a note's
// and markdown's, its parts as its writer marks them; Go's, its
// declarations; Python's, its classes and defs, by page/py, a program, as
// pic is troff's. A reader joins as a line here.
var readers = []reader{
	{"", "a note", "its parts", outline.Parts},
	{".md", "a note", "its parts", outline.Parts},
	{".go", "Go", "its declarations", outline.Go},
	{".py", "Python", "its classes and defs", python},
}
