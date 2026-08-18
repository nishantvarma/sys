type mode interface {
	key(e *editor, k tty.Key)
	name() string
	// pastEnd reports whether the cursor may rest one column past the last
	// rune (the insertion point). Normal mode sits on a rune; insert does
	// not.
	pastEnd() bool
}
