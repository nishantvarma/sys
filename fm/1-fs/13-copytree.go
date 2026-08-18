func copyTree(src, dst string) error {
	// dst inside src would copy its own output forever
	if under(dst, src) {
		return fmt.Errorf("refusing to copy %s into itself", src)
	}
	return filepath.WalkDir(
		src,
		func(p string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			rel, _ := filepath.Rel(src, p)
			target := filepath.Join(dst, rel)
			switch {
			case d.IsDir():
				mode := os.FileMode(0o755)
				if fi, e := d.Info(); e == nil {
					mode = fi.Mode().Perm()
				}
				return os.MkdirAll(target, mode)
			case d.Type()&os.ModeSymlink != 0:
				if lp, e := os.Readlink(p); e == nil {
					return os.Symlink(lp, target)
				}
				return nil
			default:
				return copyFile(p, target)
			}
		},
	)
}
