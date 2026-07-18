// Package tty is the shared terminal module for the sys commands: a raw-mode
// driver with buffered output and decoded keystrokes, plus the escape/colour
// primitives every command paints with. fm and ed are thin mains over it.
package tty

import (
	"bufio"
	"os"
	"os/signal"
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
)

// Key is one decoded keystroke: a named special key, or a rune.
type Key struct {
	Name string
	Rune rune
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

// Suspend hands the terminal back to a foreground child; Resume reclaims it.
func (t *Term) Suspend() {
	t.Write(AltOff + Cnorm + Home + Clear)
	t.Flush()
	term.Restore(t.fd, t.state)
}

func (t *Term) Resume() {
	term.MakeRaw(t.fd)
	t.queue, t.rest = nil, nil
	t.Write(AltOn + Civis + Home + Clear)
	t.Flush()
}

// ReadKey blocks for the next keystroke. A single terminal read delivers a
// whole escape sequence, so decoding stays synchronous.
func (t *Term) ReadKey() (Key, bool) {
	for len(t.queue) == 0 {
		resize, ok := t.waitInput()
		if !ok {
			return Key{}, false
		}
		if resize {
			return Key{Name: KResize}, true
		}
		var buf [256]byte
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

// waitInput blocks until stdin has bytes to read or a resize arrived. It
// returns resize=true (already drained) for the latter, ok=false if the
// terminal is gone. With no resize pipe it falls straight through to a read.
func (t *Term) waitInput() (resize, ok bool) {
	if t.winch == nil {
		return false, true
	}
	fds := []unix.PollFd{
		{Fd: int32(t.fd), Events: unix.POLLIN},
		{Fd: int32(t.winch.Fd()), Events: unix.POLLIN},
	}
	for {
		_, err := unix.Poll(fds, -1)
		if err == unix.EINTR { // SIGURG and friends; retry
			continue
		}
		if err != nil {
			return false, false
		}
		if fds[1].Revents&unix.POLLIN != 0 {
			var b [64]byte
			t.winch.Read(b[:]) // drain, coalescing a burst
			return true, true
		}
		if fds[0].Revents != 0 {
			return false, true
		}
	}
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

// csiKey maps a CSI sequence's parameters and final byte to a named key.
// Arrows end in A-D; Page Up/Down are the parametered "5~"/"6~".
func csiKey(params string, final byte) (Key, bool) {
	switch final {
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
