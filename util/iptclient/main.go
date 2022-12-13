package main

import (
	"fmt"
	"io"
	"net"
	"os"

	"golang.org/x/term"
)

func main() {
	// tcp socket client
	// dial connect to server

	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}

	defer conn.Close()

	// set terminal to raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

	// read from stdin, parse commands and keypresses then write to socket
	go func() {
		b := make([]byte, 1024)
		for {
			// read from stdin
			n, err := os.Stdin.Read(b)
			if err != nil {
				panic(err)
			}

			// parse commands and keypresses
			s := string(b[:n])
			fmt.Printf("read %q %x\r\n", s, s)

			switch s {
			case "q":
				// restore terminal
				_ = term.Restore(int(os.Stdin.Fd()), oldState)
				// close connection
				_ = conn.Close()
				os.Exit(0)
			case ":":
				// go to botton of terminal
				fmt.Print("\033[999B")
				// clear line
				fmt.Print("\033[2K")
				// read line
				term := term.NewTerminal(os.Stdin, ": ")
				line, err := term.ReadLine()
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				fmt.Printf("read %q %x\r\n", line, line)
				b = []byte(line)
				n = len(line)
			}

			// write to socket
			_, err = conn.Write(b[:n])
			if err != nil {
				panic(err)
			}
		}
	}()

	// read from socket and write to stdout
	_, _ = io.Copy(os.Stdout, conn)
}
