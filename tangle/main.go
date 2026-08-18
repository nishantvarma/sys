package main

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

func main() {
	chunks := map[string]chunk{}
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
	root, err := os.ReadFile("@main.go")
	check(err)
	placed := map[string]bool{}
	note := func(src string) {
		for _, m := range ref.FindAllStringSubmatch(src, -1) {
			if esc, n := m[1], m[2]; esc == "" && n != "others" {
				if key, ok := resolve(chunks, n); ok {
					placed[key] = true
				}
			}
		}
	}
	note(string(root))
	for _, c := range chunks {
		note(c.body)
	}
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
}

var ref = regexp.MustCompile(`(\\?)<<([^<>]+)>>`)

type chunk struct {
	path string
	body string
}

type tangler struct {
	chunks   map[string]chunk
	placed   map[string]bool
	swept    map[string]bool
	expanded map[string]bool
}

// resolve maps a ref to a chunk key: the exact key when one matches,
// else the one key having name as a path suffix — <<paste>> finds
// buffer/paste. Two candidates are an error, not a guess.
func resolve(chunks map[string]chunk, name string) (string, bool) {
	if _, ok := chunks[name]; ok {
		return name, true
	}
	var hits []string
	for key := range chunks {
		if strings.HasSuffix(key, "/"+name) {
			hits = append(hits, key)
		}
	}
	if len(hits) > 1 {
		sort.Strings(hits)
		check(fmt.Errorf("ambiguous <<%s>>: %s",
			name, strings.Join(hits, ", ")))
	}
	if len(hits) == 0 {
		return "", false
	}
	return hits[0], true
}

func (t *tangler) expand(src, dir, self string, stack []string) string {
	return ref.ReplaceAllStringFunc(src, func(m string) string {
		sub := ref.FindStringSubmatch(m)
		esc, name := sub[1], sub[2]
		if esc != "" {
			return "<<" + name + ">>"
		}
		if name == "others" {
			return t.others(dir, self, stack)
		}
		key, ok := resolve(t.chunks, name)
		if !ok {
			return m
		}
		c := t.chunks[key]
		for _, s := range stack {
			if s == key {
				check(fmt.Errorf(
					"cycle: <<%s>> "+
						"ultimately refers to itself",
					key))
			}
		}
		t.expanded[key] = true
		return strings.TrimRight(t.expand(
			c.body, filepath.Dir(c.path), key,
			append(stack, key)), "\n")
	})
}

func (t *tangler) others(dir, self string, stack []string) string {
	var names []string
	for name, c := range t.chunks {
		if name == self || t.placed[name] || t.swept[name] {
			continue
		}
		if dir == "." || strings.HasPrefix(c.path, dir+"/") {
			names = append(names, name)
		}
	}
	sort.Slice(names, func(i, j int) bool {
		return t.chunks[names[i]].path < t.chunks[names[j]].path
	})

	var parts []string
	for _, name := range names {
		t.swept[name] = true
		c := t.chunks[name]
		parts = append(parts, strings.TrimRight(t.expand(
			c.body, filepath.Dir(c.path), name,
			append(stack, name)), "\n"))
	}
	return strings.Join(parts, "\n\n")
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
