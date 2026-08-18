func exists(p string) bool { _, err := os.Stat(p); return err == nil }
