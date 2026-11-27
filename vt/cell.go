package vt

// A Cell is a place on the screen: the rune a program wrote there, and
// how it is drawn.
type Cell struct {
	Rune   rune
	Attr   Attr
	Fg, Bg Color
}

// Attr is how a cell is drawn, as SGR sets it, and two marks st keeps:
// Wide on a rune two cells wide, and Wdummy on the cell after it, which
// the rune covers. st's third, a line that ran on to the next, is for a
// selection, and a selection is the host's.
type Attr uint16

const (
	Bold Attr = 1 << iota
	Faint
	Italic
	Underline
	Blink
	Reverse
	Invisible
	Struck
	Wide
	Wdummy
)

// Color is a cell's colour as the program named it: one of the 256 a
// palette holds, by its index, red, green and blue past them, or the
// terminal's own foreground or background, Fore and Back, which the host
// picks.
type Color uint32

const (
	Fore Color = 256 + iota
	Back
)

// truecolor marks a colour named by its red, green and blue, as st's
// TRUECOLOR does.
const truecolor Color = 1 << 24

// Rgb is the colour's red, green and blue: its own, or, from 16 to 255,
// xterm's, a cube of 6 by 6 by 6 and 24 greys, which st loads too. The
// first 16 and the terminal's own are the host's to pick: ok is false.
func (c Color) Rgb() (r, g, b uint8, ok bool) {
	switch {
	case c&truecolor != 0:
		return uint8(c >> 16), uint8(c >> 8), uint8(c), true
	case c < 16 || c > 255:
		return 0, 0, 0, false
	case c < 232:
		i := c - 16
		return sixd(i / 36 % 6), sixd(i / 6 % 6), sixd(i % 6), true
	}
	v := uint8(0x08 + 0x0a*(c-232))
	return v, v, v, true
}

// sixd is a step of the cube as a byte, st's sixd_to_16bit.
func sixd(x Color) uint8 {
	if x == 0 {
		return 0
	}
	return uint8(0x37 + 0x28*x)
}
