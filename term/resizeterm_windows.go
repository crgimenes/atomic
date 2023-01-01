//go:build windows
// +build windows

package term

import (
	"os"
)

func HandleResize(ptmx *os.File) {
	// TODO: add Windows support
}
