package main

import (
	"fmt"
	"io"
	"net"
	"os"

	"github.com/crgimenes/goConfig"
	log "github.com/nuveo/logSys"
)

type config struct {
	Debug bool
}

func handleRequest(conn io.ReadWriteCloser) {

	for {
		buf := make([]byte, 1024)

		reqLen, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading:", err.Error())
		}
		fmt.Println("reqLen:", reqLen)
		fmt.Println("buf:", string(buf))

		conn.Write([]byte("Message received.\n"))

		if buf[0] == 'q' {
			conn.Close()
		}
	}
}

func main() {
	cfg := config{}

	err := goConfig.Parse(&cfg)
	if err != nil {
		log.Errorln(err.Error())
		return
	}

	if cfg.Debug {
		log.DebugMode = cfg.Debug
		log.Warningln("debug mode on")
	}

	l, err := net.Listen("tcp", ":8888")
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()

	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			log.Fatal("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		go handleRequest(conn)
	}

}
