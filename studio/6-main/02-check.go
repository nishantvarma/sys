// check ends studio on an error, said as it is: one about a frame starts
// with its path and line, so a click on it opens the line.
func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
