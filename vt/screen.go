package vt

// limit is v kept from lo to hi, st's LIMIT.
func limit(v, lo, hi int) int {
	return min(max(v, lo), hi)
}

// moveto moves the cursor to x, y, kept on the screen, or in origin mode
// in the scroll region, st's tmoveto; the next rune no longer wraps
// first.
func (t *Term) moveto(x, y int) {
	miny, maxy := 0, t.rows-1
	if t.c.state&cursororigin != 0 {
		miny, maxy = t.top, t.bot
	}
	t.c.state &^= cursorwrapnext
	t.c.x = limit(x, 0, t.cols-1)
	t.c.y = limit(y, miny, maxy)
}

// moveato is a move to a place the program names, st's tmoveato: in
// origin mode its lines count from the scroll region's top.
func (t *Term) moveato(x, y int) {
	if t.c.state&cursororigin != 0 {
		y += t.top
	}
	t.moveto(x, y)
}

// graphic0 is DEC's special graphics from 0x41 to 0x7e, as st has it
// from rxvt: 0 where a rune is itself.
var graphic0 = [62]rune{
	'↑', '↓', '→', '←', '█', '▚', '☃', // A to G
	0, 0, 0, 0, 0, 0, 0, 0, // H to O
	0, 0, 0, 0, 0, 0, 0, 0, // P to W
	0, 0, 0, 0, 0, 0, 0, ' ', // X to _
	'◆', '▒', '␉', '␌', '␍', '␊', '°', '±', // ` to g
	'␤', '␋', '┘', '┐', '┌', '└', '┼', '⎺', // h to o
	'⎻', '─', '⎼', '⎽', '├', '┤', '┴', '┬', // p to w
	'│', '≤', '≥', 'π', '≠', '£', '·', // x to ~
}

// setchar puts r at x, y with attr's attributes, st's tsetchar: in DEC's
// graphics its line or corner, and a wide rune it lands on half of made
// blank whole.
func (t *Term) setchar(r rune, attr Cell, x, y int) {
	if t.trantbl[t.charset] == csgraphic0 && r >= 0x41 && r <= 0x7e &&
		graphic0[r-0x41] != 0 {
		r = graphic0[r-0x41]
	}
	line := t.lines[y]
	if line[x].Attr&Wide != 0 {
		if x+1 < t.cols {
			line[x+1].Rune = ' '
			line[x+1].Attr &^= Wdummy
		}
	} else if line[x].Attr&Wdummy != 0 {
		line[x-1].Rune = ' '
		line[x-1].Attr &^= Wide
	}
	attr.Rune = r
	line[x] = attr
}

// clearregion blanks the cells of the rectangle from x1, y1 to x2, y2,
// in the cursor's colours, st's tclearregion.
func (t *Term) clearregion(x1, y1, x2, y2 int) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	x1, x2 = limit(x1, 0, t.cols-1), limit(x2, 0, t.cols-1)
	y1, y2 = limit(y1, 0, t.rows-1), limit(y2, 0, t.rows-1)
	blank := Cell{Rune: ' ', Fg: t.c.attr.Fg, Bg: t.c.attr.Bg}
	for y := y1; y <= y2; y++ {
		for x := x1; x <= x2; x++ {
			t.lines[y][x] = blank
		}
	}
}

// deletechar deletes n cells at the cursor, those after them moving
// left, st's tdeletechar.
func (t *Term) deletechar(n int) {
	n = limit(n, 0, t.cols-t.c.x)
	line := t.lines[t.c.y]
	copy(line[t.c.x:], line[t.c.x+n:])
	t.clearregion(t.cols-n, t.c.y, t.cols-1, t.c.y)
}

// insertblank inserts n blank cells at the cursor, those after them
// moving right and off the line's end, st's tinsertblank.
func (t *Term) insertblank(n int) {
	n = limit(n, 0, t.cols-t.c.x)
	line := t.lines[t.c.y]
	copy(line[t.c.x+n:], line[t.c.x:t.cols-n])
	t.clearregion(t.c.x, t.c.y, t.c.x+n-1, t.c.y)
}

// insertblankline inserts n blank lines at the cursor's, in the scroll
// region, st's tinsertblankline.
func (t *Term) insertblankline(n int) {
	if t.c.y >= t.top && t.c.y <= t.bot {
		t.scrolldown(t.c.y, n)
	}
}

// deleteline deletes n lines at the cursor's, in the scroll region, st's
// tdeleteline.
func (t *Term) deleteline(n int) {
	if t.c.y >= t.top && t.c.y <= t.bot {
		t.scrollup(t.c.y, n)
	}
}

// scrollup moves the lines from orig to the scroll region's bottom up
// by n, n blank ones coming in at the bottom, st's tscrollup.
func (t *Term) scrollup(orig, n int) {
	n = limit(n, 0, t.bot-orig+1)
	t.clearregion(0, orig, t.cols-1, orig+n-1)
	for i := orig; i <= t.bot-n; i++ {
		t.lines[i], t.lines[i+n] = t.lines[i+n], t.lines[i]
	}
}

// scrolldown moves the lines from orig to the scroll region's bottom
// down by n, n blank ones coming in at orig, st's tscrolldown.
func (t *Term) scrolldown(orig, n int) {
	n = limit(n, 0, t.bot-orig+1)
	t.clearregion(0, t.bot-n+1, t.cols-1, t.bot)
	for i := t.bot; i >= orig+n; i-- {
		t.lines[i], t.lines[i-n] = t.lines[i-n], t.lines[i]
	}
}

// newline moves the cursor down a line, the scroll region scrolling at
// its bottom, and to the first column with first, st's tnewline.
func (t *Term) newline(first bool) {
	y := t.c.y
	if y == t.bot {
		t.scrollup(t.top, 1)
	} else {
		y++
	}
	x := t.c.x
	if first {
		x = 0
	}
	t.moveto(x, y)
}

// setscroll makes lines top to bot the scroll region, kept on the
// screen and in order, st's tsetscroll.
func (t *Term) setscroll(top, bot int) {
	top, bot = limit(top, 0, t.rows-1), limit(bot, 0, t.rows-1)
	if top > bot {
		top, bot = bot, top
	}
	t.top, t.bot = top, bot
}

// swapscreen shows the other screen, st's tswapscreen.
func (t *Term) swapscreen() {
	t.lines, t.alt = t.alt, t.lines
	t.mode ^= modealtscreen
}

// screen is which screen is shown, as saved holds a cursor for each: 1
// the other one.
func (t *Term) screen() int {
	if t.mode&modealtscreen != 0 {
		return 1
	}
	return 0
}

// savecursor keeps the cursor, the screen's own, for loadcursor, st's
// tcursor(CURSOR_SAVE).
func (t *Term) savecursor() {
	t.saved[t.screen()] = t.c
}

// loadcursor brings back the cursor the screen kept, st's
// tcursor(CURSOR_LOAD).
func (t *Term) loadcursor() {
	c := t.saved[t.screen()]
	t.c = c
	t.moveto(c.x, c.y)
}

// puttab moves the cursor n tab stops right, or left for n less than 0,
// st's tputtab: to the line's end or its start where none is left.
func (t *Term) puttab(n int) {
	x := t.c.x
	if n > 0 {
		for x < t.cols && n > 0 {
			n--
			for x++; x < t.cols && !t.tabs[x]; x++ {
			}
		}
	} else if n < 0 {
		for x > 0 && n < 0 {
			n++
			for x--; x > 0 && !t.tabs[x]; x-- {
			}
		}
	}
	t.c.x = limit(x, 0, t.cols-1)
}

// reset brings the terminal back to its first state, st's treset: the
// cursor home in the terminal's colours, a tab stop each tabspaces, the
// whole screen the scroll region, the sets the USA's, wrap and UTF-8
// on, and both screens blank.
func (t *Term) reset() {
	t.c = cursor{attr: Cell{Fg: Fore, Bg: Back}}
	for x := range t.tabs {
		t.tabs[x] = x > 0 && x%tabspaces == 0
	}
	t.top, t.bot = 0, t.rows-1
	t.mode = t.mode&^termmodes | modewrap | modeutf8
	t.trantbl = [4]charset{}
	t.charset = 0
	for range 2 {
		t.moveto(0, 0)
		t.savecursor()
		t.clearregion(0, 0, t.cols-1, t.rows-1)
		t.swapscreen()
	}
}

// Resize makes the terminal cols by rows, st's tresize: the lines the
// cursor would fall off the bottom of slide up out of sight, each line
// is cut or grown at its end, the room new is blank, tab stops go on
// past the old end, and the whole screen is the scroll region. A size
// short of 1 by 1 is let go, as st lets it go.
func (t *Term) Resize(cols, rows int) {
	if cols < 1 || rows < 1 {
		return
	}
	minrow, mincol := min(rows, t.rows), min(cols, t.cols)
	if n := t.c.y - rows + 1; n > 0 {
		t.lines, t.alt = t.lines[n:], t.alt[n:]
	}
	t.lines, t.alt = fit(t.lines, cols, rows), fit(t.alt, cols, rows)
	tabs := make([]bool, cols)
	copy(tabs, t.tabs)
	if cols > t.cols {
		x := t.cols - 1
		for x > 0 && !tabs[x] {
			x--
		}
		for x += tabspaces; x < cols; x += tabspaces {
			tabs[x] = true
		}
	}
	t.tabs = tabs
	t.cols, t.rows = cols, rows
	t.setscroll(0, rows-1)
	t.moveto(t.c.x, t.c.y)
	c := t.c
	for range 2 {
		if mincol < cols && minrow > 0 {
			t.clearregion(mincol, 0, cols-1, minrow-1)
		}
		if minrow < rows {
			t.clearregion(0, minrow, cols-1, rows-1)
		}
		t.swapscreen()
		t.loadcursor()
	}
	t.c = c
}

// fit is lines cut or grown to rows lines of cols cells each, the cells
// new zero, for Resize to blank.
func fit(lines [][]Cell, cols, rows int) [][]Cell {
	out := make([][]Cell, rows)
	for y := range out {
		out[y] = make([]Cell, cols)
		if y < len(lines) {
			copy(out[y], lines[y])
		}
	}
	return out
}
