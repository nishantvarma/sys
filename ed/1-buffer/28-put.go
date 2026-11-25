// put writes data over the file at path whole or not at all: to a file
// beside it, then renamed onto it, as state.Write puts cur, so a save
// cut short leaves the file as it was. A link is followed, so the link
// stays and the file it names changes, and the file keeps its mode. A new
// file has nothing to lose, and is written as one; so is a file with a
// second name, a hard link, which a rename would leave on the old text,
// and one in a folder ed cannot write in.
func put(path string, data []byte) error {
	if p, err := filepath.EvalSymlinks(path); err == nil {
		path = p
	}
	fi, err := os.Stat(path)
	if err != nil {
		return os.WriteFile(path, data, 0o644)
	}
	mode := fi.Mode().Perm()
	if st, ok := fi.Sys().(*syscall.Stat_t); ok && st.Nlink > 1 {
		return os.WriteFile(path, data, mode)
	}
	f, err := os.CreateTemp(filepath.Dir(path),
		"."+filepath.Base(path)+".*")
	if err != nil {
		return os.WriteFile(path, data, mode)
	}
	_, err = f.Write(data)
	if err == nil {
		err = f.Sync()
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Chmod(f.Name(), mode)
	}
	if err == nil {
		err = os.Rename(f.Name(), path)
	}
	if err != nil {
		os.Remove(f.Name())
	}
	return err
}
