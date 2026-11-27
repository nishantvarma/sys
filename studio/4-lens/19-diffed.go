// diffed is git's diff of the thing as saved, against the last commit, a
// node a line, its kind what the line does: add, del or same in a hunk, a
// hunk's head, or meta, a file's header or a note such as \ No newline.
// In a file's diff, a line of the file as it is now is named by its
// address, so a key on one goes there; a removed line, and a folder's
// diff, name no place. git's own word says why there is no diff (git).
func diffed(th thing) ([]*outline.Node, error) {
	path := th.path
	dir, file := path, false
	if fi, err := os.Stat(path); err == nil && !fi.IsDir() {
		dir, file = filepath.Dir(path), true
	}
	// outside a repository git diff compares two paths, HEAD the first:
	// ask for the repository first, so that its absence is the word
	if _, err := git(dir, "rev-parse", "--git-dir"); err != nil {
		return nil, err
	}
	out, err := git(dir, "diff", "--no-color", "HEAD", "--", path)
	if err != nil {
		return nil, err
	}
	diff := strings.TrimRight(string(out), "\n")
	if diff == "" {
		return nil, errors.New("no change since the last commit," +
			" or not tracked")
	}
	var ns []*outline.Node
	in := false // in a hunk, where a line's first sign says what it does
	now := 0    // the line of the file as it is now
	for i, l := range strings.Split(diff, "\n") {
		text := strings.ReplaceAll(l, "\t", "        ")
		n := &outline.Node{ID: "diff#" + strconv.Itoa(i+1),
			Kind: "meta", Text: text}
		switch {
		case strings.HasPrefix(l, "diff "):
			in = false
		case strings.HasPrefix(l, "@@"):
			in, n.Kind, now = true, "hunk", hunk(l)
		case !in, strings.HasPrefix(l, `\`):
			// a file's header, or a note on the line above: meta
		case strings.HasPrefix(l, "-"):
			n.Kind = "del"
		default:
			n.Kind = "same"
			if strings.HasPrefix(l, "+") {
				n.Kind = "add"
			}
			if file && now > 0 {
				n.ID = outline.At(path, now)
				now++
			}
		}
		ns = append(ns, n)
	}
	return ns, nil
}
