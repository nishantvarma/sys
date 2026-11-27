// tookInterrupt reports a pending Ctrl-C, clearing it.
func tookInterrupt() bool {
	select {
	case <-interrupt:
		return true
	default:
		return false
	}
}
