// plumb hands a word to p, which opens what it names, and does not wait.
func plumb(word string) {
	if word == "" {
		return
	}
	c := exec.Command("p", "-")
	c.Stdin = strings.NewReader(word)
	if c.Start() == nil {
		go c.Wait() // reap it, as fm's detach does
	}
}
