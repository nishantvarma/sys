package tty

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// X's two selections, which every program on the display shares: Primary,
// which selecting text sets and a middle click pastes, and Clipboard,
// which a copy sets and a paste reads. A program on the terminal reaches
// them through xclip, for the terminal passes neither on: st takes no
// OSC 52 unless built to, and tmux drops an application's unless told.
const (
	Primary   = "primary"
	Clipboard = "clipboard"
)

// Put sets selection sel to text. xclip keeps it, in the background,
// until another program takes the selection. Off X there is none to set.
func Put(sel, text string) error {
	if os.Getenv("DISPLAY") == "" {
		return errors.New("no X display")
	}
	c := exec.Command("xclip", "-selection", sel)
	c.Stdin = strings.NewReader(text)
	if err := c.Run(); err != nil {
		return fmt.Errorf("xclip: %w", err)
	}
	return nil
}

// Get is the text selection sel holds; empty when it holds none, or X is
// not there to ask.
func Get(sel string) string {
	out, err := exec.Command("xclip", "-o", "-selection", sel).Output()
	if err != nil {
		return ""
	}
	return string(out)
}
