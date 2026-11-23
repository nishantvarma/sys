// figureVerb writes a note's figure as a page, `design/arch`'s say, on the
// skeleton, `medium/page.html`, or the one given: the figure lens's page,
// until the page hosts the lens itself. An error in the note says its
// line.
func figureVerb(args []string) error {
	if len(args) < 1 || len(args) > 2 {
		return errors.New("usage: studio figure note [page.html]")
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
	f, err := figure.New(items)
	if err != nil {
		return err
	}
	fmt.Print(figurePage(f, tpl, note, string(src)))
	return nil
}
