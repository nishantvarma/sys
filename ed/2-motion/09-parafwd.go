// paraFwd/paraBack are wordFwd/wordBack lifted from runes to lines. Where a
// word motion skips the class-run the cursor sits in and stops at the next
// boundary, a paragraph motion partitions lines into two classes — blank and
// non-blank — skips the run the cursor is in, and stops at the next blank
// line. "Blank" means empty (len 0), not whitespace-only: simplest and
// predictable. Both land at col 0 of the boundary, falling back to the last
// or first line like toBottom/toTop when there is no further blank.
func paraFwd(b *buffer, p pos) pos {
	l := p.line
	for l < b.last() && len(b.lines[l]) == 0 {
		l++ // skip the blank run the cursor is in
	}
	for l < b.last() && len(b.lines[l]) != 0 {
		l++ // scan to the next blank line
	}
	return pos{l, 0}
}
