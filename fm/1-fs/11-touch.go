func touch(p string) error {
	f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY, 0o644)
	if err == nil {
		f.Close()
	}
	return err
}
