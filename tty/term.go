// Package tty is the shared terminal module for the sys commands: a raw-mode
// driver with buffered output and decoded keystrokes, plus the escape/colour
// primitives every command paints with. fm and ed are thin mains over it.
package tty

import (
	"bufio"
	"bytes"
	"os"
	"os/signal"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"unicode/utf8"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

// Named special keys. A bare rune keystroke has Name == "".
const (
	KEsc   = "ESC"
	KEnter = "ENTER"
	KTab   = "TAB"
	KBack  = "BACKSPACE"
	KInt   = "INT"
	KUp    = "UP"
	KDown  = "DOWN"
	KLeft  = "LEFT"
	KRight = "RIGHT"
	KPgUp  = "PGUP"
	KPgDn  = "PGDN"
	// KResize is a synthetic key ReadKey returns when the terminal
	// resizes, so the caller's loop redraws to the new size on its own.
	KResize = "RESIZE"
	// KPaste is a whole bracketed paste, its text in Key.Text. It only
	// arrives once a caller has asked for PasteOn.
	KPaste = "PASTE"
	// KFocus and KBlur say the pane gained or lost the eye. They only
	// arrive once a caller has asked with Focus.
	KFocus = "FOCUS"
	KBlur  = "BLUR"
	// KMouse is a button pressed, or moved while held, and KMouseUp one
	// let go, at Key.X, Key.Y. They only arrive once a caller has asked
	// with Mouse; no program asks yet, for what a click means waits on
	// the keys (`todo/studio/outline:/the keys, decided last/`).
	KMouse   = "MOUSE"
	KMouseUp = "MOUSEUP"
)

// Key is one decoded keystroke: a named special key, or a rune. A KPaste
// carries the pasted block in Text — one burst, not a run of keys. A
// chord with Ctrl is named C- and its letter: C-r. A mouse key has its
// cell, counted from 1 as the terminal counts, and its button with any
// modifiers, as the terminal's report gives them.
type Key struct {
	Name         string
	Rune         rune
	Text         string
	X, Y, Button int
}

// Token is the map key a binding table looks up: the special-key name, else
// the rune as a string.
func (k Key) Token() string {
	if k.Name != "" {
		return k.Name
	}
	return string(k.Rune)
}

type Term struct {
	fd    int
	state *term.State
	out   *bufio.Writer
	queue []Key
	rest  []byte
	winch *os.File // read end of the self-pipe SIGWINCH is forwarded onto
	wake  *os.File // its write end, for Wake
	focus bool     // focus reporting asked for, so Suspend lends it out
	mouse bool     // mouse reporting asked for, lent out the same way
}

// New puts stdin in raw mode and returns the driver.
func New() (*Term, error) {
	fd := int(os.Stdin.Fd())
	state, err := term.MakeRaw(fd)
	if err != nil {
		return nil, err
	}
	t := &Term{
		fd:    fd,
		state: state,
		out:   bufio.NewWriter(os.Stdout),
	}
	t.watchResize()
	return t, nil
}

// watchResize forwards SIGWINCH onto a self-pipe so ReadKey can poll for a
// resize alongside stdin. The forwarder touches only the pipe, never stdin,
// so a suspended child keeps sole ownership of the terminal.
func (t *Term) watchResize() {
	r, w, err := os.Pipe()
	if err != nil {
		return // resize won't wake the loop; keystrokes still redraw
	}
	t.winch, t.wake = r, w
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			w.Write([]byte{0})
		}
	}()
}

func (t *Term) Close() { term.Restore(t.fd, t.state) }

func (t *Term) Size() (int, int) {
	w, h, err := term.GetSize(t.fd)
	if err != nil || w == 0 || h == 0 {
		return 80, 24
	}
	return w, h
}

func (t *Term) Write(s string) { t.out.WriteString(s) }
func (t *Term) Flush()         { t.out.Flush() }

// Focus asks the terminal to report focus, so KFocus and KBlur arrive as
// the pane gains and loses the eye — how a view tells whether it is the
// one being looked at.
func (t *Term) Focus() {
	t.focus = true
	t.Write(FocusOn)
	t.Flush()
}

// Mouse asks the terminal to report the mouse, so KMouse and KMouseUp
// arrive as keys, at the cell they fell on. Like focus, Suspend lends it
// to a child and Resume takes it back, and a program turns it off as it
// ends, with MouseOff.
func (t *Term) Mouse() {
	t.mouse = true
	t.Write(MouseOn)
	t.Flush()
}

// Suspend hands the terminal back to a foreground child; Resume reclaims it.
// The alt screen stays on, so the child scrolls there and not over the shell
// we were started from — leaving it would put the child's output on the
// primary buffer, still there when we exit. Focus reporting goes with the
// terminal: a child that never asked for CSI I must not be sent one.
func (t *Term) Suspend() {
	if t.focus {
		t.Write(FocusOff)
	}
	if t.mouse {
		t.Write(MouseOff)
	}
	t.Write(Cnorm + Home + Clear)
	t.Flush()
	term.Restore(t.fd, t.state)
}

func (t *Term) Resume() {
	term.MakeRaw(t.fd)
	t.queue, t.rest = nil, nil
	t.Write(AltOn + Civis + Home + Clear)
	if t.focus {
		t.Write(FocusOn)
	}
	if t.mouse {
		t.Write(MouseOn)
	}
	t.Flush()
}

// Pending reports whether a decoded key is already waiting. A burst — a
// paste, a held key, a fast typist — queues faster than a screen is worth
// painting, so a caller's loop can skip the draw and come back when the
// queue has run dry.
func (t *Term) Pending() bool { return len(t.queue) > 0 }

// Feed queues ks to be read next, ahead of anything already waiting, as
// though typed again — a replay runs through the very reads its keys did.
func (t *Term) Feed(ks []Key) { t.queue = slices.Concat(ks, t.queue) }

// Wake has ReadKey return as a resize does, from any goroutine: a program
// blocked on keys hears that something it watches changed, and draws
// again. Like the resize forwarder, it touches only the self-pipe.
func (t *Term) Wake() {
	if t.wake != nil {
		t.wake.Write([]byte{0})
	}
}

// ReadKey blocks for the next keystroke. A single terminal read delivers a
// whole escape sequence, so decoding stays synchronous.
func (t *Term) ReadKey() (Key, bool) {
	for len(t.queue) == 0 {
		// A held block is an unfinished one. Wait for the rest of it,
		// but not for ever: give up once it stops arriving.
		ms := -1
		if len(t.rest) > 0 {
			ms = holdMs
		}
		resize, quiet, ok := t.waitInput(ms)
		if !ok {
			return Key{}, false
		}
		if resize {
			return Key{Name: KResize}, true
		}
		if quiet {
			t.flushRest()
			continue
		}
		// Big enough that a paste is a couple of reads, not a
		// hundred: parseKeys hands back any partial tail anyway.
		var buf [4096]byte
		n, err := os.Stdin.Read(buf[:])
		if n > 0 {
			data := make([]byte, len(t.rest)+n)
			copy(data, t.rest)
			copy(data[len(t.rest):], buf[:n])
			keys, rest := parseKeys(data)
			t.rest, t.queue = rest, append(t.queue, keys...)
		}
		if err != nil && len(t.queue) == 0 {
			// A read with both data and an error delivers the
			// keys first.
			return Key{}, false
		}
	}
	k := t.queue[0]
	t.queue = t.queue[1:]
	return k, true
}

// waitInput blocks until stdin has bytes to read, a resize arrived, or ms
// elapsed (ms < 0 waits for ever). It returns resize=true (already drained)
// for the second, quiet=true for the third, ok=false if the terminal is gone.
func (t *Term) waitInput(ms int) (resize, quiet, ok bool) {
	fds := []unix.PollFd{{Fd: int32(t.fd), Events: unix.POLLIN}}
	if t.winch != nil {
		w := unix.PollFd{Fd: int32(t.winch.Fd()), Events: unix.POLLIN}
		fds = append(fds, w)
	}
	for {
		n, err := unix.Poll(fds, ms)
		if err == unix.EINTR { // SIGURG and friends; retry
			continue
		}
		if err != nil {
			return false, false, false
		}
		if n == 0 {
			return false, true, true
		}
		if len(fds) > 1 && fds[1].Revents&unix.POLLIN != 0 {
			var b [64]byte
			t.winch.Read(b[:]) // drain, coalescing a burst
			return true, false, true
		}
		if fds[0].Revents != 0 {
			return false, false, true
		}
	}
}

// flushRest gives up on a block that stopped mid-arrival. A paste is worth
// delivering short — only its closing marker went missing; Esc then O,
// with nothing after it yet, is the two keys vi means by it; any other
// truncated escape is not worth delivering. Either way the hold ends, so
// the next keystroke reads as one.
func (t *Term) flushRest() {
	b := t.rest
	t.rest = nil
	switch {
	case bytes.HasPrefix(b, []byte(pasteStart)):
		t.queue = append(t.queue, pasteKey(b[len(pasteStart):]))
	case string(b) == "\x1bO":
		t.queue = append(t.queue, Key{Name: KEsc}, Key{Rune: 'O'})
	}
}

// mouseKey is a mouse report as a key: a press, or a move with a button
// held, or a release, at a cell counted from 1.
func mouseKey(button, x, y int, up bool) Key {
	name := KMouse
	if up {
		name = KMouseUp
	}
	return Key{Name: name, X: x, Y: y, Button: button}
}

func parseKeys(b []byte) ([]Key, []byte) {
	var ks []Key
	for i := 0; i < len(b); {
		c := b[i]
		switch {
		case c == 0x1b:
			k, n := escKeys(b[i:])
			if n == 0 {
				return ks, b[i:] // held: not all here yet
			}
			ks = append(ks, k...)
			i += n
		case c == '\r' || c == '\n':
			ks = append(ks, Key{Name: KEnter})
			i++
		case c == '\t':
			ks = append(ks, Key{Name: KTab})
			i++
		case c == 0x7f || c == 0x08:
			ks = append(ks, Key{Name: KBack})
			i++
		case c == 0x03:
			ks = append(ks, Key{Name: KInt})
			i++
		case c >= 0x01 && c <= 0x1a:
			// Ctrl and a letter, a chord, by name: C-r
			name := "C-" + string(rune('a'+c-1))
			ks = append(ks, Key{Name: name})
			i++
		case c < 0x20:
			i++ // ignore other control bytes
		default:
			if c >= 0x80 && !utf8.FullRune(b[i:]) {
				return ks, b[i:] // truncated rune
			}
			r, sz := utf8.DecodeRune(b[i:])
			ks = append(ks, Key{Rune: r})
			i += sz
		}
	}
	return ks, nil
}

// escKeys decodes what starts at an escape, b[0]: the keys it makes and
// the bytes they take; 0 bytes when it is not all here yet, to be held.
// Esc alone, or before a byte no sequence starts with, is Esc. ESC O is
// SS3 only before a terminal's own letter, an arrow, home, end or F1-F4
// in application mode; before any other it is Esc then O, as vi types
// them. ESC [ M is an X10 mouse report, three bytes after it each 32 past
// its value, a button of 3 a release; any other ESC [ is a CSI sequence,
// a bracketed paste among them.
func escKeys(b []byte) ([]Key, int) {
	switch {
	case len(b) < 2:
		return []Key{{Name: KEsc}}, 1
	case b[1] == 'O':
		if len(b) < 3 {
			return nil, 0 // flushRest hands it back as two keys
		}
		if strings.IndexByte("ABCDHFPQRS", b[2]) < 0 {
			return []Key{{Name: KEsc}}, 1
		}
		if k, ok := csiKey("", b[2]); ok {
			return []Key{k}, 3
		}
		return nil, 3
	case b[1] == '[' && len(b) > 2 && b[2] == 'M':
		if len(b) < 6 {
			return nil, 0
		}
		bt := int(b[3]) - 32
		x, y := int(b[4])-32, int(b[5])-32
		return []Key{mouseKey(bt, x, y, bt&3 == 3)}, 6
	case b[1] == '[':
		j := 2
		for j < len(b) && !csiFinal(b[j]) {
			j++
		}
		if j >= len(b) {
			return nil, 0
		}
		params := string(b[2:j])
		if params == "200" && b[j] == '~' {
			k, n, ok := pasted(b[j+1:])
			if !ok {
				return nil, 0 // its closing marker is to come
			}
			return []Key{k}, j + 1 + n
		}
		if k, ok := csiKey(params, b[j]); ok {
			return []Key{k}, j + 1
		}
		return nil, j + 1
	}
	return []Key{{Name: KEsc}}, 1
}

func csiFinal(c byte) bool { return c >= 0x40 && c <= 0x7e }

const (
	pasteStart = "\x1b[200~"
	pasteEnd   = "\x1b[201~"
	// holdMs bounds the wait on a closing marker. Unbounded, a dropped
	// 201~ — a multiplexer detached mid-paste, a link cut — swallows
	// every later keystroke, Ctrl-C included, for good.
	holdMs = 250
)

// pasted cuts a bracketed paste's body off at its closing marker and
// returns it as one key, with the bytes it consumed. ok is false while
// that marker has yet to arrive, so the caller holds the bytes and reads
// on.
func pasted(b []byte) (Key, int, bool) {
	e := bytes.Index(b, []byte(pasteEnd))
	if e < 0 {
		return Key{}, 0, false
	}
	return pasteKey(b[:e]), e + len(pasteEnd), true
}

// pasteKey is a paste body as a key. A clipboard's CR line endings become
// the '\n' the buffer speaks.
func pasteKey(b []byte) Key {
	s := strings.ReplaceAll(string(b), "\r\n", "\n")
	return Key{Name: KPaste, Text: strings.ReplaceAll(s, "\r", "\n")}
}

// csiKey maps a CSI sequence's parameters and final byte to a named key.
// Arrows end in A-D; Page Up/Down are the parametered "5~"/"6~"; focus in
// and out are the bare I and O; an SGR mouse report is <button;x;y, M for
// a press or a move and m for a release.
func csiKey(params string, final byte) (Key, bool) {
	if p, ok := strings.CutPrefix(params, "<"); ok {
		f := strings.Split(p, ";")
		if len(f) != 3 || final != 'M' && final != 'm' {
			return Key{}, false
		}
		var n [3]int
		for i, s := range f {
			v, err := strconv.Atoi(s)
			if err != nil {
				return Key{}, false
			}
			n[i] = v
		}
		return mouseKey(n[0], n[1], n[2], final == 'm'), true
	}
	switch final {
	case 'I':
		return Key{Name: KFocus}, true
	case 'O':
		return Key{Name: KBlur}, true
	case 'A':
		return Key{Name: KUp}, true
	case 'B':
		return Key{Name: KDown}, true
	case 'C':
		return Key{Name: KRight}, true
	case 'D':
		return Key{Name: KLeft}, true
	case '~':
		switch params {
		case "5":
			return Key{Name: KPgUp}, true
		case "6":
			return Key{Name: KPgDn}, true
		}
	}
	return Key{}, false
}
