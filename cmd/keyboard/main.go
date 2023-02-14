package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
)

func formatHex(b []byte) string {
	sb := strings.Builder{}

	for _, v := range b {
		if sb.Len() > 0 {
			sb.WriteString(" ")
		}

		sb.WriteString(fmt.Sprintf("%02x", v))
	}

	return sb.String()
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
