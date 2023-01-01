//go:build !windows
// +build !windows

package term

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kr/pty"
)

func HandleResize(ptmx *os.File) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
				log.Printf("error resizing pty: %s\r\n", err)
			}
		}
	}()
	ch <- syscall.SIGWINCH
}
