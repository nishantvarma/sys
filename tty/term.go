// Package tty is the shared terminal module for the sys commands: a raw-mode
// driver with buffered output and decoded keystrokes, plus the escape/colour
// primitives every command paints with. fm and ed are thin mains over it.
package tty

import (
	"bufio"
	"bytes"
	"os"
	"os/signal"
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
)

// Key is one decoded keystroke: a named special key, or a rune. A KPaste
// carries the pasted block in Text — one burst, not a run of keys.
type Key struct {
	Name string
	Rune rune
	Text string
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
	focus bool     // focus reporting asked for, so Suspend lends it out
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
	t.winch = r
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

// Suspend hands the terminal back to a foreground child; Resume reclaims it.
// The alt screen stays on, so the child scrolls there and not over the shell
// we were started from — leaving it would put the child's output on the
// primary buffer, still there when we exit. Focus reporting goes with the
// terminal: a child that never asked for CSI I must not be sent one.
func (t *Term) Suspend() {
	if t.focus {
		t.Write(FocusOff)
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
	t.Flush()
}

// Pending reports whether a decoded key is already waiting. A burst — a
// paste, a held key, a fast typist — queues faster than a screen is worth
// painting, so a caller's loop can skip the draw and come back when the
// queue has run dry.
func (t *Term) Pending() bool { return len(t.queue) > 0 }

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
// delivering short — only its closing marker went missing; a truncated
// escape is not. Either way the hold ends, so the next keystroke reads as one.
func (t *Term) flushRest() {
	b := t.rest
	t.rest = nil
	if !bytes.HasPrefix(b, []byte(pasteStart)) {
		return
	}
	t.queue = append(t.queue, pasteKey(b[len(pasteStart):]))
}

func parseKeys(b []byte) ([]Key, []byte) {
	var ks []Key
	for i := 0; i < len(b); {
		c := b[i]
		switch {
		case c == 0x1b:
			if i+1 >= len(b) {
				ks = append(ks, Key{Name: KEsc})
				i++
			} else if b[i+1] == '[' || b[i+1] == 'O' {
				j := i + 2
				for j < len(b) && !csiFinal(b[j]) {
					j++
				}
				if j >= len(b) {
					return ks, b[i:] // incomplete sequence
				}
				params := string(b[i+2 : j])
				if params == "200" && b[j] == '~' {
					k, n, ok := pasted(b[j+1:])
					if !ok {
						// hold it all: the closing
						// marker is still coming
						return ks, b[i:]
					}
					ks = append(ks, k)
					i = j + 1 + n
					continue
				}
				if k, ok := csiKey(params, b[j]); ok {
					ks = append(ks, k)
				}
				i = j + 1
			} else {
				ks = append(ks, Key{Name: KEsc})
				i++
			}
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
// and out are the bare I and O.
func csiKey(params string, final byte) (Key, bool) {
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
