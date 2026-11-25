// text as the terminal draws it: freetype, hinted slightly, smoothed by the
// screen's red, green and blue parts (hintslight, rgba rgb, as this system
// asks). Gio draws only grayscale, so a run is made here as an opaque image,
// its ground blended in, and gio shows it pixel for pixel. This holds for an
// unrotated RGB panel at scale 1. One font, no fallback and no shaping: a
// grid of one rune a column
package main

/*
#cgo pkg-config: freetype2
#include <stdlib.h>
#include <ft2build.h>
#include FT_FREETYPE_H
#include FT_LCD_FILTER_H

// FT_HAS and the load flags are macros, which cgo cannot call
static int glyph(FT_Face f, FT_ULong c) {
	if (FT_Load_Char(f, c, FT_LOAD_TARGET_LIGHT)) return 1;
	return FT_Render_Glyph(f->glyph, FT_RENDER_MODE_LCD);
}
*/
import "C"

import (
	"fmt"
	"image"
	"image/color"
	"unsafe"
)

type bitmap struct {
	left, top  int
	wide, high int     // in pixels
	cover      []uint8 // three a pixel: red, green, blue
}

type typeface struct {
	faces  [2]C.FT_Face // regular, bold
	ascent int
	cell   int // a column's width, whole pixels
	cache  [2]map[rune]*bitmap
}

func newTypeface(px int, files ...string) (*typeface, error) {
	var lib C.FT_Library
	if C.FT_Init_FreeType(&lib) != 0 {
		return nil, fmt.Errorf("freetype: cannot start")
	}
	C.FT_Library_SetLcdFilter(lib, C.FT_LCD_FILTER_DEFAULT)
	t := &typeface{}
	for i, file := range files[:2] {
		name := C.CString(file)
		bad := C.FT_New_Face(lib, name, 0, &t.faces[i])
		C.free(unsafe.Pointer(name))
		if bad != 0 ||
			C.FT_Set_Pixel_Sizes(
				t.faces[i],
				0,
				C.FT_UInt(px),
			) != 0 {
			return nil, fmt.Errorf(
				"freetype: cannot load %s",
				file,
			)
		}
		t.cache[i] = map[rune]*bitmap{}
	}
	m := t.faces[0].size.metrics
	t.ascent = int(m.ascender+63) >> 6
	if C.glyph(t.faces[0], '0') != 0 {
		return nil, fmt.Errorf("freetype: no 0 in %s", files[0])
	}
	t.cell = int(t.faces[0].glyph.advance.x+32) >> 6
	return t, nil
}

func (t *typeface) glyph(r rune, bold int) *bitmap {
	if g, ok := t.cache[bold][r]; ok {
		return g
	}
	g := &bitmap{}
	f := t.faces[bold]
	if C.glyph(f, C.FT_ULong(r)) == 0 {
		s := f.glyph
		b := s.bitmap
		g.left, g.top = int(s.bitmap_left), int(s.bitmap_top)
		g.wide, g.high = int(b.width)/3, int(b.rows)
		g.cover = make([]uint8, 3*g.wide*g.high)
		for y := 0; y < g.high; y++ {
			row := unsafe.Add(
				unsafe.Pointer(b.buffer),
				y*int(b.pitch),
			)
			copy(
				g.cover[3*g.wide*y:],
				unsafe.Slice((*uint8)(row), 3*g.wide),
			)
		}
	}
	t.cache[bold][r] = g
	return g
}

// blend lays ink to, covering a of 255, over what is there; a neighbour's ink
// may be, so it only ever goes further from the ground, from
func blend(have, to, from, a uint8) uint8 {
	if a == 0 {
		return have
	}
	v := uint8(int(have) + (int(to)-int(have))*int(a)/255)
	if to >= from {
		return max(have, v)
	}
	return min(have, v)
}

// stamp inks g with its corner at at, in to on the ground from
func stamp(img *image.RGBA, g *bitmap, at image.Point, to, from [3]uint8) {
	for y := range g.high {
		for x := range g.wide {
			p := at.Add(image.Pt(x, y))
			if !p.In(img.Rect) {
				continue
			}
			o := img.PixOffset(p.X, p.Y)
			ink := g.cover[3*(g.wide*y+x):]
			px := img.Pix[o : o+3]
			for c := range px {
				px[c] = blend(px[c], to[c], from[c], ink[c])
			}
		}
	}
}

// image is txt in fg on bg, high pixels tall, a rune a column
func (t *typeface) image(
	txt string,
	fg, bg color.NRGBA,
	bold bool,
	high int,
) *image.RGBA {
	runes := []rune(txt)
	img := image.NewRGBA(image.Rect(0, 0, max(1, len(runes)*t.cell), high))
	for i := 0; i < len(img.Pix); i += 4 {
		p := img.Pix[i : i+4]
		p[0], p[1], p[2], p[3] = bg.R, bg.G, bg.B, 255
	}
	weight := 0
	if bold {
		weight = 1
	}
	to := [3]uint8{fg.R, fg.G, fg.B}
	from := [3]uint8{bg.R, bg.G, bg.B}
	base := t.ascent + 1
	for i, r := range runes {
		g := t.glyph(r, weight)
		stamp(img, g, image.Pt(i*t.cell+g.left, base-g.top), to, from)
	}
	return img
}
