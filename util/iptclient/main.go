package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"golang.org/x/term"
)

func main() {
	var (
		connEcho = true
		conn     net.Conn
		err      error
	)

	// set terminal to raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}

	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

	// handle SIGINT
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		<-c

		_ = term.Restore(int(os.Stdin.Fd()), oldState)
		_ = conn.Close()

		os.Exit(0)
	}()

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
			//fmt.Printf("read %q %x\r\n", s, s)

			switch s {
			case "q", "Q", "\x03", "\x04", "\x1b": // q, Q, ctrl-c, ctrl-d, esc
				// restore terminal
				_ = term.Restore(int(os.Stdin.Fd()), oldState)
				// close connection
				if conn != nil {
					_ = conn.Close()
				}
				os.Exit(0)
			case ":":
				connEcho = false
				_, _ = os.Stdout.Write([]byte("\033[999B\033[2K"))
				// read line
				nterm := term.NewTerminal(os.Stdin, ":")
				line, err := nterm.ReadLine()
				if err != nil {
					log.Println(err)
					os.Exit(1)
				}
				b = []byte(line)
				n = len(line)
				connEcho = true
				switch line {
				case "q", "Q", "quit", "exit":
					// restore terminal
					_ = term.Restore(int(os.Stdin.Fd()), oldState)
					// close connection
					if conn != nil {
						_ = conn.Close()
					}
					os.Exit(0)
				default:
					_, _ = os.Stdout.Write([]byte("\033[999B\033[2Kunknown command: " + line + "\r\n"))
				}
			}

			if conn == nil {
				_, err = conn.Write(b[:n])
				if err != nil {
					panic(err)
				}
			}
		}
	}()

	for {
		conn, err = net.Dial("tcp", "localhost:8080")
		if err != nil {
			fmt.Printf("error dialing: %s\r\n", err.Error())
			fmt.Printf("retrying in 5 seconds\r\n")
			time.Sleep(3 * time.Second)

			continue
		}

		break
	}

	defer func() {
		if conn != nil {
			_ = conn.Close()
		}
	}()

	// read from socket and write to stdout
	// _, _ = io.Copy(os.Stdout, conn)
	b := make([]byte, 1024)
	tempBuf := ""
	for {
		n, err := conn.Read(b)
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		if !connEcho {
			tempBuf += string(b[:n])
			continue
		}

		if tempBuf != "" {
			fmt.Print("\033[999B")
			fmt.Print("\033[2K")
			fmt.Print(tempBuf)
			tempBuf = ""
		}

		_, _ = os.Stdout.Write(b[:n])
	}
}
