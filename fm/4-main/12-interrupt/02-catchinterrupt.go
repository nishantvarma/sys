func catchInterrupt() { signal.Notify(interrupt, syscall.SIGINT) }
