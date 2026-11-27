package outline

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// said is items as one paragraph: each one's text, then its kids'.
func said(items []*Node) string {
	var words []string
	for _, n := range items {
		words = append(words, n.Text)
		if len(n.Kids) > 0 {
			words = append(words, said(n.Kids))
		}
	}
	return strings.Join(words, " ")
}

// At is the address of a note's line, path:line: the id Text gives the
// item there, and the name every view and verb knows the line by.
func At(path string, line int) string {
	return path + ":" + strconv.Itoa(line)
}

// Address cuts an address back into its path and its line, as At made
// it: a/b:12 is line 12 of a/b, and :12 line 12 of a buffer with no file.
// What has no line above 0 after its last colon, a path alone, is line 0,
// none, and is all path.
func Address(id string) (string, int) {
	if i := strings.LastIndex(id, ":"); i >= 0 {
		if n, err := strconv.Atoi(id[i+1:]); err == nil && n > 0 {
			return id[:i], n
		}
	}
	return id, 0
}

// File is what the file at p holds, when it is text a person types and so
// a note: a regular file, of at most a megabyte, for a note is read on
// every key, with no zero byte in it. The error says why not, and stat
// comes first, so a fifo or a device is never opened.
func File(p string) ([]byte, error) {
	fi, err := os.Stat(p)
	if err != nil {
		return nil, err
	}
	switch {
	case !fi.Mode().IsRegular():
		return nil, fmt.Errorf("%s: not a regular file", p)
	case fi.Size() > 1<<20:
		return nil, fmt.Errorf("%s: over a megabyte", p)
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	if bytes.IndexByte(b, 0) >= 0 {
		return nil, fmt.Errorf("%s: binary, a zero byte in it", p)
	}
	return b, nil
}

// Note is the note an address names, as ed -f edits it and the lens
// reads it: a file, from its line; a folder's README, the folder's own
// note, from its top. Nothing for the rest: a folder with no README, a
// file gone, or one that is no note a person types (File).
func Note(at string) (string, int, bool) {
	path, line := Address(at)
	fi, err := os.Stat(path)
	if err != nil {
		return "", 0, false
	}
	if fi.IsDir() {
		notes := []string{"README.md", "README", "readme.md"}
		for _, r := range notes {
			p := filepath.Join(path, r)
			if _, err := File(p); err == nil {
				return p, 0, true
			}
		}
		return "", 0, false
	}
	_, err = File(path)
	return path, line, err == nil
}

// Head is a note's head: its first item's text, the title, and the items
// under that as one paragraph, the lead.
func Head(items []*Node) (title, lead string) {
	if len(items) == 0 {
		return "", ""
	}
	return items[0].Text, said(items[0].Kids)
}
