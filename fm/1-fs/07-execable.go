func execable(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.Mode().IsRegular() &&
		fi.Mode().Perm()&0o111 != 0
}
