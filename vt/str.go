package vt

import "bytes"

// str is a string sequence read, st's STREscape: DCS, OSC, PM or APC, by
// its letter, or k, st's old title, and its bytes up to its end.
type str struct {
	kind byte
	buf  []byte
}

// strargsiz is how many parts of a string st reads, its STR_ARG_SIZ.
const strargsiz = 16

// strhandle does what a string says, st's strhandle: an OSC 0 or 2, or
// a k, names the window. The colours an OSC sets and asks for, and the
// clipboard it fills, are the host's window's, and are let go; so are a
// DCS, an APC and a PM, as st lets them go.
func (t *Term) strhandle() {
	t.esc &^= escstrend | escstr
	args := strparse(t.str.buf)
	par := 0
	if len(args) > 0 {
		par, _ = strtol([]byte(args[0]))
	}
	switch t.str.kind {
	case ']':
		if (par == 0 || par == 2) && len(args) > 1 {
			t.title = args[1]
		}
	case 'k':
		t.title = ""
		if len(args) > 0 {
			t.title = args[0]
		}
	}
}

// strparse is a string's bytes in parts, split at ;, st's strparse: no
// more than strargsiz, and none past a NUL, where st's C string ends.
func strparse(b []byte) []string {
	if i := bytes.IndexByte(b, 0); i >= 0 {
		b = b[:i]
	}
	if len(b) == 0 {
		return nil
	}
	var args []string
	for len(args) < strargsiz {
		i := bytes.IndexByte(b, ';')
		if i < 0 {
			return append(args, string(b))
		}
		args = append(args, string(b[:i]))
		b = b[i+1:]
	}
	return args
}
