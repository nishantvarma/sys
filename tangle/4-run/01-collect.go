err := filepath.WalkDir(
	".",
	func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		name := d.Name()
		if !strings.HasSuffix(name, ".go") ||
			name[0] < '0' || name[0] > '9' {
			return nil
		}
		// the key is the path, digit prefixes stripped from
		// every segment: 4-run/01-collect.go -> run/collect
		segs := strings.Split(
			strings.TrimSuffix(path, ".go"), "/")
		for i, s := range segs {
			segs[i] = strings.TrimPrefix(
				strings.TrimLeft(s, "0123456789"), "-")
		}
		key := strings.Join(segs, "/")
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if _, dup := chunks[key]; dup {
			return fmt.Errorf(
				"duplicate chunk %q at %s", key, path)
		}
		chunks[key] = chunk{path: path, body: string(body)}
		return nil
	})
check(err)
