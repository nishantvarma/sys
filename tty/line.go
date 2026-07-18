package tty

import (
	"strings"
	"unicode"
)

// ReadLine edits a line at the status bar: msg as the prompt, def as
// the initial text, Tab through complete when given. ok is false when
// cancelled.
func ReadLine(
	t *Term,
	msg string,
	complete func(string) string,
	def string,
) (string, bool) {
	buf := []rune(def)
	pos := len(buf)
	_, h := t.Size()
	show := func() {
		t.Write(Status(h, msg+string(buf), true))
		t.Write(strings.Repeat(MoveL, len(buf)-pos))
		t.Flush()
	}
	show()
	for {
		k, ok := t.ReadKey()
		if !ok {
			return "", false
		}
		switch {
		case k.Name == KEsc || k.Name == KInt:
			t.Write(Civis)
			t.Flush()
			return "", false
		case k.Name == KEnter:
			t.Write(Civis)
			t.Flush()
			return string(buf), true
		case k.Name == KLeft:
			if pos > 0 {
				pos--
			}
		case k.Name == KRight:
			if pos < len(buf) {
				pos++
			}
		case k.Name == KBack:
			if pos > 0 {
				buf = append(buf[:pos-1], buf[pos:]...)
				pos--
			}
		case k.Name == KTab && complete != nil:
			buf = []rune(complete(string(buf)))
			pos = len(buf)
		case k.Name == "" && unicode.IsPrint(k.Rune):
			buf = append(
				buf[:pos],
				append([]rune{k.Rune}, buf[pos:]...)...)
			pos++
		default:
			continue
		}
		show()
	}
}
