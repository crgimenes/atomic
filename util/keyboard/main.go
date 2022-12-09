package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/ssh/terminal"
)

func formatHex(b []byte) string {
	var s string
	for _, v := range b {
		if s != "" {
			s += " "
		}

		s += fmt.Sprintf("%02x", v)
	}

	return s
}

func main() {
	fmt.Println("Press 'q' to quit")

	if !terminal.IsTerminal(0) || !terminal.IsTerminal(1) {
		fmt.Println("stdin/stdout should be terminal")
		os.Exit(1)
	}
	oldState, err := terminal.MakeRaw(0)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer terminal.Restore(0, oldState)

	b := make([]byte, 1024)

	for {
		n, err := os.Stdin.Read(b)
		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Printf("len(%d) %q %x [%s]\r\n", n, b[:n], b[:n], formatHex(b[:n]))

		if b[0] == 'q' {
			break
		}
	}
}
