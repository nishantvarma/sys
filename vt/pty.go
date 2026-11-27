//go:build linux

package vt

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

// Start runs cmd on a new pty cols by rows, as st's ttynew does: in a
// session of its own, the pty its controlling terminal and its input,
// output and error. It is the pty's other end: what cmd writes is read
// from it, and what is written to it cmd reads. Its pty is Linux's,
// /dev/ptmx and its ioctls, where st's openpty is the C library's.
//
// cmd's environment, as os/exec makes it, the host's and PWD when it
// has none, loses what would size it otherwise, COLUMNS, LINES and
// TERMCAP, and its TERM is st's, as st's execsh has it. execsh also sets
// LOGNAME, USER, SHELL and HOME from the password file, which the host's
// environment says already, and gives back the signals st took; a
// signal the host ignores stays ignored, for Go cannot undo it between
// the fork and the exec.
func Start(cmd *exec.Cmd, cols, rows int) (*os.File, error) {
	ptm, err := os.OpenFile("/dev/ptmx", os.O_RDWR|syscall.O_NOCTTY, 0)
	if err != nil {
		return nil, err
	}
	var n uint32
	err = ioctl(ptm, func(fd int) error {
		if err := unix.IoctlSetPointerInt(fd, unix.TIOCSPTLCK,
			0); err != nil {
			return err
		}
		n, err = unix.IoctlGetUint32(fd, unix.TIOCGPTN)
		return err
	})
	if err == nil {
		err = Setsize(ptm, cols, rows)
	}
	var pts *os.File
	if err == nil {
		pts, err = os.OpenFile(fmt.Sprintf("/dev/pts/%d", n),
			os.O_RDWR|syscall.O_NOCTTY, 0)
	}
	if err != nil {
		ptm.Close()
		return nil, err
	}
	defer pts.Close()
	cmd.Stdin, cmd.Stdout, cmd.Stderr = pts, pts, pts
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true, Setctty: true}
	cmd.Env = environ(cmd.Environ())
	if err := cmd.Start(); err != nil {
		ptm.Close()
		return nil, err
	}
	return ptm, nil
}

// Setsize tells the program on pty f its terminal is cols by rows, as
// st's ttyresize does, in cells alone: the pixels st gives too are the
// host's window's. The program hears SIGWINCH.
func Setsize(f *os.File, cols, rows int) error {
	return ioctl(f, func(fd int) error {
		return unix.IoctlSetWinsize(fd, unix.TIOCSWINSZ,
			&unix.Winsize{Row: uint16(rows), Col: uint16(cols)})
	})
}

// Foreground is the process group in the foreground of pty f, as tmux
// asks which program runs in a pane: the shell's own while it waits at
// its prompt, else the command it runs.
func Foreground(f *os.File) (int, error) {
	var g int
	err := ioctl(f, func(fd int) (err error) {
		g, err = unix.IoctlGetInt(fd, unix.TIOCGPGRP)
		return err
	})
	return g, err
}

// ioctl runs do on f's descriptor, as Go's poller lets it be used: f
// stays its, so that a Close ends a Read under way.
func ioctl(f *os.File, do func(fd int) error) error {
	sc, err := f.SyscallConn()
	if err != nil {
		return err
	}
	var derr error
	if err := sc.Control(func(fd uintptr) {
		derr = do(int(fd))
	}); err != nil {
		return err
	}
	return derr
}

// environ is env without COLUMNS, LINES, TERMCAP and TERM, and TERM
// st's.
func environ(env []string) []string {
	out := []string{"TERM=" + termname}
	for _, kv := range env {
		switch k, _, _ := strings.Cut(kv, "="); k {
		case "COLUMNS", "LINES", "TERMCAP", "TERM":
		default:
			out = append(out, kv)
		}
	}
	return out
}
