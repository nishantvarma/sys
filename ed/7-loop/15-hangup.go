// hangup takes ed's marks of work unsaved away as it is hung up or told
// to end, as the frame's end does to its pane, for the unmark the loop
// defers never runs then (mark): files, the window's unsaved file and its
// text, and the signal's own end, as ed would have had it.
func hangup(files ...string) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGTERM)
	go func() {
		s := <-c
		for _, f := range files {
			os.Remove(f)
		}
		signal.Reset(s)
		syscall.Kill(os.Getpid(), s.(syscall.Signal))
	}()
}
