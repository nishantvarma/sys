// what freetype made must reach the screen pixel for pixel: gio blends in
// linear light and may filter, and either would undo the sharpness
package main

import (
	"image"
	"image/png"
	"os"
	"testing"

	"gioui.org/gpu/headless"
	"gioui.org/op"
)

func TestSharp(t *testing.T) {
	v, err := newWindow(&studio{}, false)
	if err != nil {
		t.Fatal(err)
	}
	const wide, high = 120, 100
	gpu, err := headless.NewWindow(wide, high)
	if err != nil {
		t.Fatal(err)
	}
	defer gpu.Release()
	var ops op.Ops
	gtx := gtxFor(&ops, wide, high)
	fill(gtx, image.Rect(0, 0, wide, high), hex(groundHex))
	words := []string{"bin/", "design/", "ed/", "picture/", "fm/"}
	for i, w := range words {
		v.line(
			gtx,
			(&strip{}).add(w, caretHex, "", false),
			image.Pt(0, i*cellH),
			v.text.cell,
		)
	}
	if err := gpu.Frame(&ops); err != nil {
		t.Fatal(err)
	}
	got := image.NewRGBA(image.Rect(0, 0, wide, high))
	if err := gpu.Screenshot(got); err != nil {
		t.Fatal(err)
	}
	f, _ := os.Create(out(t) + "/ft.png")
	png.Encode(f, got)
	f.Close()
	worst, fringed, ink := 0, 0, 0
	for i, w := range words {
		want := v.text.image(
			w,
			hex(caretHex),
			hex(groundHex),
			false,
			cellH,
		)
		for y := 0; y < cellH; y++ {
			for x := 0; x < want.Rect.Dx(); x++ {
				a, b := want.RGBAAt(
					x,
					y,
				), got.RGBAAt(
					x,
					i*cellH+y,
				)
				for _, d := range []int{
					int(a.R) - int(b.R),
					int(a.G) - int(b.G),
					int(a.B) - int(b.B),
				} {
					worst = max(worst, d, -d)
				}
				if b.R != 0x12 || b.B != 0x12 {
					ink++
					if int(b.R)-int(b.B) > 40 ||
						int(b.B)-int(b.R) > 40 {
						fringed++
					}
				}
			}
		}
	}
	t.Logf(
		"column %d px; screen vs freetype, worst channel "+
			"difference %d/255; %d of %d inked pixels carry "+
			"colour fringes",
		v.text.cell,
		worst,
		fringed,
		ink,
	)
	if worst > 2 {
		t.Errorf("gio changed the pixels")
	}
}
