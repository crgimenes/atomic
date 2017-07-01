package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/crgimenes/atomic/con/adapters/telnet"
	"github.com/crgimenes/goConfig"
	log "github.com/nuveo/logSys"
)

type config struct {
	Debug bool
}

func handleRequest(conn *net.TCPConn) {
	cChar := make(chan byte)
	cErr := make(chan error)

	telnet.SendSetup(conn)

	defer conn.Close()

	go func() {
		for {
			buf := make([]byte, 1024)

			rLen, err := conn.Read(buf)
			if err != nil {
				if err == io.EOF {
					log.Println("close: ", conn.LocalAddr(), " - ", conn.RemoteAddr())
					return
				}

				log.Errorln("Error reading:", err.Error())
				cErr <- err
				return
			}

			for i := 0; i < rLen; i++ {
				cChar <- buf[i]
			}
		}
	}()

	for {
		select {
		case err := <-cErr:
			if err == io.EOF {
				log.Println("close: ", conn.LocalAddr(), " - ", conn.RemoteAddr())
				return
			}

			log.Errorln("Error reading:", err.Error())
			return
		case c := <-cChar:
			fmt.Printf("%q\t%v\t0x%0X\n", c, c, c)
			if c == 'q' {
				err := conn.CloseRead()
				if err != nil {
					log.Errorln(err.Error())
				}
				time.Sleep(time.Second)
				return
			}
			var buf []byte
			buf = append(buf, c)
			conn.Write(buf)

		case <-time.After(1 * time.Second):
			println(".")
			conn.Write([]byte("."))
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

	rAddr, err := net.ResolveTCPAddr("tcp", ":8888")
	if err != nil {
		panic(err)
	}

	l, err := net.ListenTCP("tcp", rAddr)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()

	for {
		// Listen for an incoming connection.
		conn, err := l.AcceptTCP()
		if err != nil {
			log.Fatal("Error accepting: ", err.Error())
			os.Exit(1)
		}

		// Handle connections in a new goroutine.
		go handleRequest(conn)
	}

}
