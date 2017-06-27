package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/crgimenes/goConfig"
	log "github.com/nuveo/logSys"
)

type config struct {
	Debug bool
}

func telnet(c byte) (r bool) {
	switch c {
	case 0x00:
		fmt.Println("BINARY")
	case 0x01:
		fmt.Println("ECHO")
	case 0x03:
		fmt.Println("SGA")
	case 0x18:
		fmt.Println("TTYPE")
	case 0x1F:
		fmt.Println("NAWS")
	case 0x22:
		fmt.Println("LINEMODE")
	case 0xF0:
		fmt.Println("SE")
	case 0xF1:
		fmt.Println("NOP")
	case 0xF2:
		fmt.Println("DM")
	case 0xf3:
		fmt.Println("BRK")
	case 0xF4:
		fmt.Println("IP")
	case 0xF5:
		fmt.Println("AO")
	case 0xF6:
		fmt.Println("AYT")
	case 0xF7:
		fmt.Println("EC")
	case 0xF8:
		fmt.Println("EL")
	case 0xF9:
		fmt.Println("GA")
	case 0xFA:
		fmt.Println("SB")
	case 0xFB:
		fmt.Println("WILL")
	case 0xFC:
		fmt.Println("WONT")
	case 0xFD:
		fmt.Println("DO")
	case 0xFE:
		fmt.Println("DONT")
	case 0xFF:
		fmt.Println("IAC")
	default:
		r = true
	}
	return
}

func sendSetup(conn net.Conn) (err error) {
	conn.Write([]byte{255, 251, 3})
	conn.Write([]byte{255, 251, 1})

	var buf []byte
	buf, err = read(conn, 6)
	if err != nil {
		return
	}
	if !bytes.Equal(buf[0:6], []byte{255, 253, 3, 255, 253, 1}) {
		log.Errorln("setup failed ", buf)
	}
	return
}

func read(conn net.Conn, size int) (rbuf []byte, err error) {
	buf := make([]byte, size+1024)
	var rLen int
	var totLen int
	for {
		rLen, err = conn.Read(buf)
		if err != nil {
			return
		}
		totLen += rLen
		for i := 0; i < rLen; i++ {
			rbuf = append(rbuf, buf[i])
		}
		if totLen >= size {
			break
		}
	}
	return
}

func handleRequest(conn net.Conn) {

	sendSetup(conn)
	cChar := make(chan byte)
	cErr := make(chan error)

	defer conn.Close()

	//conn.SetReadDeadline(time.Now().Add(timeoutDuration))
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
