// lenses are the ways ed shows the text it holds, in the order Tab turns
// through them (`design/lens`): source, the text as written and the only
// one typed in; outline, its items by their indent, as rows or as thumbs;
// and render, what the source makes for a person to see, for a note its
// picture. ed draws the one shown, and names them all on what it draws,
// so that every view of it, the terminal's and gio's, shows the same.
var lenses = []string{"source", "outline", "render"}
