package vt

// strsequence starts a string, st's tstrsequence: a DCS, an APC, a PM or
// an OSC, by its letter or its C1 code, or a k, the old title. Its runes
// are taken up to its end (put), for what it says is let go (strhandle).
func (t *Term) strsequence() {
	t.esc |= escstr
}

// strhandle ends a string, st's strhandle, and lets what it says go:
// the title an OSC 0 or 2 and a k give, the colours an OSC sets and asks
// for, and the clipboard it fills are the host's window's, which no host
// here reads; a DCS, an APC and a PM st lets go too.
func (t *Term) strhandle() {
	t.esc &^= escstrend | escstr
}
