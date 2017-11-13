package log

import (
	"io"
	"os"
	"syscall"
	"unsafe"
)

type color string

const (
	colorBlue   = "\x1b[0;34m"
	colorGreen  = "\x1b[0;32m"
	colorYellow = "\x1b[0;33m"
	colorRed    = "\x1b[0;31m"

	colorRST = "\x1b[0;m"
)

func painter(c color, s string) string {
	return string(c) + s + colorRST
}

// IsTerminal returns true if the given writer supports colored output
func IsTerminal(w io.Writer) bool {
	var fd int
	switch w {
	case os.Stdout:
		fd = syscall.Stdout
	case os.Stderr:
		fd = syscall.Stderr
	case os.Stdin: // is this resonable?
		fd = syscall.Stdin
	default:
		return false
	}
	var termios syscall.Termios
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), ioctlReadTermios, uintptr(unsafe.Pointer(&termios)), 0, 0, 0)
	return err == 0
}
