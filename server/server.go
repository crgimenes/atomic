package server

import (
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/crgimenes/atomic/con/adapters/telnet"
	"github.com/crgimenes/atomic/config"
	"github.com/nuveo/log"
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

var clientList []Client

func (c *Client) Close() (err error) {
	err = c.Conn.CloseRead()
	if err != nil {
		return
	}
	time.Sleep(300 * time.Millisecond)
	err = c.Conn.Close()
	return
}

func (c *Client) Write(msg []byte) (err error) {
	_, err = c.Conn.Write(msg)
	return
}

func handleRequest(client Client) {
	cChar := make(chan byte)
	cErr := make(chan error)

	telnet.SendSetup(client.Conn)

	defer func() {
		err := client.Close()
		if err != nil {
			log.Errorln(err)
		}
	}()

	go func() {
		for {
			// loop que pega todo que vier do usuário
			buf := make([]byte, 1024)

			rLen, err := client.Conn.Read(buf) // Read precisa ser fechado explicitamente
			if err != nil {
				if err == io.EOF {
					log.Println("close: ",
						client.Conn.LocalAddr(), " - ",
						client.Conn.RemoteAddr())
					return
				}

				log.Errorln("Error reading:", err.Error())
				cErr <- err // envia erro para o supervisor
				// (poderia mandar mais infos erro é uma interface)
				return
			}

			for i := 0; i < rLen; i++ {
				cChar <- buf[i] // envia o char recebido para o supervisor
			}
		}
	}()

	for {
		// Esse loop representa a rotina supervisor
		// e faz tratamento dos canais inclusive erro
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
				return
			}
			var buf []byte
			buf = append(buf, c)
			//client.Conn.Write(buf)

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

		clientList = append(clientList, client)

		// Handle connections in a new goroutine.
		go handleRequest(client)
	}
}
