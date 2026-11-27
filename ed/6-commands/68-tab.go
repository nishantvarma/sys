// tab turns the frame's tabs, as l does from a shell: in a frame ed is
// the source tab of the lens view (`design/lens`), and Tab shows the next,
// the lens's outline first. ed knows no tabs: it runs l, as a right click
// runs p (`ed/4-view/29-plumb.go`), telling it the slot ed runs in, which
// a host down a pipe, gio's window, has no pane to say. Outside a frame
// there are none.
func (e *editor) tab() {
	if e.unsaved == "" {
		e.msg = "no tabs outside a frame"
		return
	}
	c := exec.Command("l")
	c.Env = append(os.Environ(), "STUDIO="+e.slot)
	if c.Start() == nil {
		go c.Wait() // reaped, as fm's detach does
	}
}
