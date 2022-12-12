package main

// go run main.go
// nc localhost 8080

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/kr/pty"
	"golang.org/x/term"
)

type out struct {
	connections map[net.Conn]struct{}
}

func (o out) Write(p []byte) (n int, err error) {
	n = len(p)

	fmt.Print(string(p))

	for c := range o.connections {
		_, err := c.Write(p)
		if err != nil {
			delete(o.connections, c)
		}
	}

	return n, nil
}

var (
	o       = out{}
	listner net.Listener
	err     error
)

func runCmd() error {
	c := exec.Command(os.Args[1], os.Args[2:]...)

	// Start the command with a pty.
	ptmx, err := pty.Start(c)
	if err != nil {
		return err
	}
	// Make sure to close the pty at the end.
	defer func() { _ = ptmx.Close() }() // Best effort.

	// Handle pty size.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
				log.Printf("error resizing pty: %s\r\n", err)
			}
		}
	}()
	ch <- syscall.SIGWINCH // Initial resize.

	// Set stdin in raw mode.
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }() // Best effort.

	// Copy stdin to the pty and the pty to stdout.
	go func() { _, _ = io.Copy(ptmx, os.Stdin) }()
	_, _ = io.Copy(o, ptmx)

	return nil
}

// Handles incoming requests.
func handleRequest(conn net.Conn) {
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
		conn.Close()
		return
	}
	switch string(buf[:n]) {
	case "close":
		conn.Close()
		delete(o.connections, conn)
	case "clear":
		conn.Write([]byte("\033[2J"))
		conn.Write([]byte("\033[0;0H"))
	}
}

func main() {

	o.connections = make(map[net.Conn]struct{})

	go func() {
		err := runCmd()
		if err != nil {
			fmt.Println(err)
			closeConn()
			os.Exit(1)
		}

		closeConn()
		os.Exit(0)
	}()

	// Listen for incoming connections.
	listner, err = net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	fmt.Printf("\r\nListening on :8080\r\n")

	for {
		conn, err := listner.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}

		o.connections[conn] = struct{}{}
		go handleRequest(conn)
	}
}

func closeConn() {
	for c := range o.connections {
		c.Write([]byte("\r\ncloseing connection\r\n"))
		c.Close()
	}

	listner.Close()
}
