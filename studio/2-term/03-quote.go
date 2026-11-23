// quote is a command's words as a shell reads them back: a word with a
// space, or a sign a shell takes for its own, in single quotes, which
// keep all but a single quote, said as '\''.
func quote(args []string) string {
	const signs = " \t\n!\"#$&'()*;<>?[\\]`{|}~"
	out := make([]string, len(args))
	for i, a := range args {
		out[i] = a
		if a == "" || strings.ContainsAny(a, signs) {
			out[i] = "'" + strings.ReplaceAll(a, "'", `'\''`) + "'"
		}
	}
	return strings.Join(out, " ")
}
