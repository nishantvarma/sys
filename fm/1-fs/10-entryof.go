func entryOf(dir string, e os.DirEntry) entry {
	link := e.Type()&os.ModeSymlink != 0
	isdir := e.IsDir()
	bad := link && !exists(filepath.Join(dir, e.Name()))
	f := entry{
		name: e.Name(),
		path: filepath.Join(dir, e.Name()),
		dir:  isdir,
		link: link,
		bad:  bad,
	}
	if !isdir && !link {
		if fi, err := e.Info(); err == nil {
			f.exec = fi.Mode()&0o111 != 0
			f.size, f.mod = fi.Size(), fi.ModTime()
		}
	}
	return f
}
