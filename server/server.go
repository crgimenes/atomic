package server

import (
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/crgimenes/atomic/con/adapters/telnet"
	"github.com/crgimenes/atomic/config"
	log "github.com/nuveo/logSys"
)

type Prot int

const (
	Raw Prot = iota
	Telnet
)

type Client struct {
	Protocol Prot
	Conn     *net.TCPConn
}

func handleRequest(client Client) {
	cChar := make(chan byte)
	cErr := make(chan error)

	telnet.SendSetup(client.Conn)

	defer client.Conn.Close()

	go func() {
		for {
			buf := make([]byte, 1024)

			rLen, err := client.Conn.Read(buf)
			if err != nil {
				if err == io.EOF {
					log.Println("close: ",
						client.Conn.LocalAddr(), " - ",
						client.Conn.RemoteAddr())
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
				log.Println("close: ",
					client.Conn.LocalAddr(), " - ",
					client.Conn.RemoteAddr())
				return
			}

			log.Errorln("Error reading:", err.Error())
			return
		case c := <-cChar:
			fmt.Printf("%q\t%v\t0x%0X\n", c, c, c)
			if c == 'q' {
				err := client.Conn.CloseRead()
				if err != nil {
					log.Errorln(err.Error())
				}
				time.Sleep(time.Second)
				return
			}
			var buf []byte
			buf = append(buf, c)
			client.Conn.Write(buf)

		case <-time.After(1 * time.Second):
			println(".")
			client.Conn.Write([]byte("."))
		}
	}
}

func Run() {
	host := fmt.Sprintf("%s:%d", config.Get.Host, config.Get.Port)
	rAddr, err := net.ResolveTCPAddr("tcp", host)
	if err != nil {
		panic(err)
	}

	l, err := net.ListenTCP("tcp", rAddr)
	if err != nil {
		log.Errorln("Error listening:", err.Error())
		os.Exit(1)
	}
	log.Warningln("Server listen at ", rAddr)

	// Close the listener when the application closes.
	defer func() {
		cErr := l.Close()
		if cErr != nil {
			log.Errorln(cErr)
		}
	}()

	for {
		// Listen for an incoming connection.
		conn, err := l.AcceptTCP()
		if err != nil {
			log.Warningln("Error accepting: ", err.Error())
			continue
		}

		client := Client{
			Conn:     conn,
			Protocol: Telnet,
		}

		// Handle connections in a new goroutine.
		go handleRequest(client)
	}
}
