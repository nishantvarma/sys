func entryOf(dir string, e os.DirEntry) entry {
	link := e.Type()&os.ModeSymlink != 0
	isdir := e.IsDir()
	bad := link && !exists(filepath.Join(dir, e.Name()))
	exec := false
	if !isdir && !link {
		if fi, err := e.Info(); err == nil {
			exec = fi.Mode()&0o111 != 0
		}
	}
	return entry{
		name: e.Name(),
		path: filepath.Join(dir, e.Name()),
		dir:  isdir,
		link: link,
		bad:  bad,
		exec: exec,
	}
}
