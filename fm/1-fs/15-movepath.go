func movePath(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	if isRealDir(src) {
		if err := copyTree(src, dst); err != nil {
			return err
		}
		return os.RemoveAll(src)
	}
	if err := copyFile(src, dst); err != nil {
		return err
	}
	return os.Remove(src)
}
