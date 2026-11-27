// Package watch says when files change in the folders it watches, through
// inotify in golang.org/x/sys, which sys already builds on. The programs
// that follow cur, studio's lens and ed -f, watch the folder cur is kept
// in, and wake when anyone says it.
package watch

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

// mask is what a watcher hears of: a file written and closed, renamed in,
// as fm writes cur, made, or removed.
const mask = unix.IN_CLOSE_WRITE | unix.IN_MOVED_TO | unix.IN_CREATE |
	unix.IN_DELETE

// Watcher says on On when a file changes in the folders it watches: those
// it was made on, fixed, and one it follows, which a program moves as it
// goes. One change can be several words on On, a burst of them one word.
type Watcher struct {
	On    <-chan struct{}
	fd    int
	fixed map[int]bool // the fixed folders' watches, by descriptor
	dir   string       // the folder followed
	wd    int          // its watch; -1 when it has none of its own
}

// New is a watcher on dirs, which it watches for as long as it lives. A
// folder it cannot watch is an error that names it.
func New(dirs ...string) (*Watcher, error) {
	fd, err := unix.InotifyInit1(unix.IN_CLOEXEC)
	if err != nil {
		return nil, os.NewSyscallError("inotify_init1", err)
	}
	on := make(chan struct{}, 1)
	w := &Watcher{On: on, fd: fd, fixed: map[int]bool{}, wd: -1}
	for _, d := range dirs {
		wd, err := unix.InotifyAddWatch(fd, d, mask)
		if err != nil {
			unix.Close(fd)
			return nil, fmt.Errorf("watch %s: %w", d, err)
		}
		w.fixed[wd] = true
	}
	go func() {
		buf := make([]byte, 4096)
		for {
			if n, err := unix.Read(fd, buf); err != nil || n <= 0 {
				return
			}
			select {
			case on <- struct{}{}:
			default:
			}
		}
	}()
	return w, nil
}

// Follow watches dir as well, in place of the folder it followed last, so
// that a save there, anyone's, is heard. A folder it cannot watch goes
// unfollowed: only a save there goes unheard, for the fixed folders still
// are. A fixed folder, by whatever name it is given, is watched already,
// and following it never lets its watch go.
func (w *Watcher) Follow(dir string) {
	if dir == w.dir {
		return
	}
	if w.wd >= 0 {
		unix.InotifyRmWatch(w.fd, uint32(w.wd))
	}
	w.dir, w.wd = dir, -1
	wd, err := unix.InotifyAddWatch(w.fd, dir, mask)
	if err == nil && !w.fixed[wd] {
		w.wd = wd
	}
}

// Wake calls wake as a file in dir changes, for as long as the program
// lives, dir made when it is not there yet: how a program that follows a
// state file, cur, hears it change while it waits on a key, its
// terminal's read woken (tty's Wake). A folder it cannot watch is an
// error, and the program follows only as keys come.
func Wake(dir string, wake func()) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	w, err := New(dir)
	if err != nil {
		return err
	}
	go func() {
		for range w.On {
			wake()
		}
	}()
	return nil
}
