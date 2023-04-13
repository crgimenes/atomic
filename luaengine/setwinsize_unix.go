//go:build !windows

package luaengine

import (
	"os/exec"
	"syscall"
	"unsafe"
)

func setCtrlTerm(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setctty: true,
		Setsid:  true,
	}
}

func SetWinsize(fd uintptr, w, h int) {
	ws := &Winsize{Width: uint16(w), Height: uint16(h)}
	syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCSWINSZ), uintptr(unsafe.Pointer(ws)))
}
