// plumb hands text to p, which opens what it names, or with -a hands it to
// the shell, and does not wait.
func plumb(text string, mode ...string) {
	if text == "" {
		return
	}
	c := exec.Command("p", append(mode, "-")...)
	c.Stdin = strings.NewReader(text)
	if c.Start() == nil {
		go c.Wait() // reap it, as fm's detach does
	}
}
