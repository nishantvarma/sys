func zoxideAdd(p string) {
	c := exec.Command("zoxide", "add", p)
	if c.Start() == nil {
		go c.Wait() // reap it; don't leak a zombie per cd
	}
}
