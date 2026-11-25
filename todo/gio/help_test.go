package main

import (
	"image"
	"os"
	"path/filepath"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

// out is where a test's pictures go: $OUT to keep them
func out(t *testing.T) string {
	if d := os.Getenv("OUT"); d != "" {
		return d
	}
	return t.TempDir()
}

// programs is sys's bin, where ./build puts fm and ed
func programs(t *testing.T) string {
	bin, err := filepath.Abs("../../bin")
	if err != nil {
		t.Fatal(err)
	}
	return bin
}

func gtxFor(ops *op.Ops, w, h int) layout.Context {
	return layout.Context{
		Ops:         ops,
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Constraints: layout.Exact(image.Pt(w, h)),
	}
}
