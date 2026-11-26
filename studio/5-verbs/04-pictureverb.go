// pictureVerb writes a note's picture as a page, `design/arch`'s say, on the
// skeleton, `medium/page.html`, or the one given: ed's picture lens on a
// page. An error in the note says its line.
func pictureVerb(args []string) error {
	if len(args) < 1 || len(args) > 2 {
		return errors.New("usage: studio picture note [page.html]")
	}
	note, skel := args[0], ""
	if len(args) > 1 {
		skel = args[1]
	}
	src, err := outline.File(note)
	if err != nil {
		return err
	}
	tpl, err := skeleton(skel)
	if err != nil {
		return err
	}
	items := outline.Text(bytes.NewReader(src), note)
	if len(items) == 0 {
		return fmt.Errorf("%s: nothing to draw", note)
	}
	f, err := picture.New(items)
	if err != nil {
		return err
	}
	fmt.Print(picturePage(f, tpl, note, string(src)))
	return nil
}
