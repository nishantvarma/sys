// A spawned child shares our process group, so a Ctrl-C aimed at it reaches
// us too. We catch SIGINT rather than SIG_IGN it: an ignored disposition
// survives exec, so every descendant of a spawn would inherit a Ctrl-C that
// prints ^C and does nothing. A caught one the kernel resets to default.
var interrupt = make(chan os.Signal, 1)

func catchInterrupt() { signal.Notify(interrupt, syscall.SIGINT) }

// tookInterrupt reports a pending Ctrl-C, clearing it.
func tookInterrupt() bool {
	select {
	case <-interrupt:
		return true
	default:
		return false
	}
}
