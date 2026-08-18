t := &tangler{
	chunks:   chunks,
	placed:   placed,
	swept:    map[string]bool{},
	expanded: map[string]bool{},
}
out := t.expand(string(root), ".", "", nil)

var unplaced []string
for name := range chunks {
	// placed by a mention that never expanded
	if !t.swept[name] && !t.expanded[name] {
		unplaced = append(unplaced, chunks[name].path)
	}
}
sort.Strings(unplaced)
for _, p := range unplaced {
	fmt.Fprintln(os.Stderr,
		"warning: unplaced chunk dropped from output:", p)
}

pretty, err := format.Source([]byte(out))
if err != nil {
	fmt.Fprintln(os.Stderr,
		"warning: tangled output is not valid Go:", err)
	pretty = []byte(out)
}
check(os.WriteFile("main.go", pretty, 0644))
fmt.Printf("tangled %d chunks -> main.go\n", len(chunks))
