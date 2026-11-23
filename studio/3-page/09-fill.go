// fill is the skeleton with a page's own in it: its title, its look under
// local, the one place a page adds rules of its own (`general/medium`), and
// its main in place of the skeleton's; the skeleton's style and engine
// stay as they are. The skeleton has both places (skeleton).
func fill(skel, title, look, main string) string {
	head := regexp.MustCompile(`<title>[^<]*</title>`)
	out := head.ReplaceAllLiteralString(skel,
		"<title>"+html.EscapeString(title)+"</title>")
	out = strings.Replace(out, localMark, localMark+look, 1)
	i, j := strings.Index(out, "<main>"), strings.Index(out, "</main>")
	return out[:i] + main + out[j+len("</main>"):]
}
