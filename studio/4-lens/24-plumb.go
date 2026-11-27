// plumb hands an address to p, which opens what it names, as a right
// click does in any pane (`design/mouse`), and does not wait: a line of a
// diff opens in ed, at its line.
func plumb(addr string) {
	c := exec.Command("p", addr)
	if c.Start() == nil {
		go c.Wait() // reaped, as fm's detach does
	}
}
