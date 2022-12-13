package main

// go run main.go
// nc localhost 8080

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/kr/pty"
	"golang.org/x/term"
)

type output struct {
	clients map[net.Conn]struct{}
}

var (
	out     = output{}
	listner net.Listener
	err     error
)

func isClosedConnErr(err error) bool {
	return errors.Is(err, net.ErrClosed) ||
		errors.Is(err, io.EOF) ||
		errors.Is(err, syscall.EPIPE)
}

func (o output) Write(p []byte) (n int, err error) {
	n = len(p)

	fmt.Print(string(p))

	for c := range o.clients {
		_, err := c.Write(p)
		if err != nil {
			if !isClosedConnErr(err) {
				fmt.Println("Error writing:", err.Error())
			}
			c.Close()
			delete(o.clients, c)
		}
	}

	return n, nil
}

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
	_, _ = io.Copy(out, ptmx)

	return nil
}

// Handles incoming requests.
func handleRequest(conn net.Conn) {
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if !isClosedConnErr(err) {
				fmt.Println("Error reading:", err.Error())
			}
			conn.Close()
			delete(out.clients, conn)
			return
		}

		cmd := strings.TrimSpace(string(buf[:n]))

		fmt.Printf("cmd: %q\r\n", cmd)

		switch cmd {
		case "close", "exit":
			conn.Close()
			delete(out.clients, conn)
			return
		case "clear":
			conn.Write([]byte("\033[2J\033[0;0H"))
		}
	}
}

func main() {

	out.clients = make(map[net.Conn]struct{})

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
			if !isClosedConnErr(err) {
				fmt.Println("Error accepting: ", err.Error())
				os.Exit(1)
			}
			return
		}

		out.clients[conn] = struct{}{}
		go handleRequest(conn)
	}
}

func closeConn() {
	for c := range out.clients {
		c.Write([]byte("\r\ncloseing connection\r\n"))
		c.Close()
	}

	listner.Close()
}
