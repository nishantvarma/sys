// pageVerb writes a frame as a page, on stdout: studio's own frame, or
// the note given, on the skeleton, `medium/page.html`.
func pageVerb(args []string) error {
	if len(args) > 1 {
		return errors.New("usage: studio page [frame]")
	}
	f, err := load(args)
	if err != nil {
		return err
	}
	skel, err := skeleton("")
	if err != nil {
		return err
	}
	fmt.Print(f.page(skel))
	return nil
}
