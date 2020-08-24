package ssh

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"

	"golang.org/x/crypto/ssh"
)

func ListenAndServe() error {

	scfg := &ssh.ServerConfig{
		// TODO: improve authentication (allow key pair authentication, etc.)
		NoClientAuth: true,
		// The BBS system will authenticate the user in another step that is
		// more lime the old BBS systems and also allow guest access.
	}

	b, err := ioutil.ReadFile("id_rsa")
	if err != nil {
		return fmt.Errorf("failed to load private key, %v", err.Error())
	}

	pk, err := ssh.ParsePrivateKey(b)
	if err != nil {
		return fmt.Errorf("failed to parse private key, %v", err.Error())
	}

	scfg.AddHostKey(pk)

	listener, err := net.Listen("tcp", "0.0.0.0:2200")
	if err != nil {
		return fmt.Errorf("failed to listen on 2200, %v", err.Error())
	}

	log.Print("Listening on 2200...")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("failed to accept incoming connection", err.Error())
			continue
		}
		// Before use, a handshake must be performed on the incoming net.Conn.
		sshConn, chans, reqs, err := ssh.NewServerConn(conn, scfg)
		if err != nil {
			log.Println("failed to handshake", err.Error())
			continue
		}

		log.Printf("New SSH connection from %s (%s)", sshConn.RemoteAddr(), sshConn.ClientVersion())
		// Discard all global out-of-band Requests
		go ssh.DiscardRequests(reqs)
		// Accept all channels
		go handleChannels(chans)
	}
}

func handleChannels(chans <-chan ssh.NewChannel) {
	// Service the incoming Channel channel in go routine
	for newChannel := range chans {
		go handleChannel(newChannel)
	}
}

func handleChannel(newChannel ssh.NewChannel) {
	// Since we're handling a shell, we expect a
	// channel type of "session". The also describes
	// "x11", "direct-tcpip" and "forwarded-tcpip"
	// channel types.
	t := newChannel.ChannelType()
	if t != "session" {
		newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
		return
	}

	// At this point, we have the opportunity to reject the client's
	// request for another logical connection
	connection, requests, err := newChannel.Accept()
	if err != nil {
		log.Printf("Could not accept channel (%s)", err)
		return
	}

	// Sessions have out-of-band requests such as "shell", "pty-req" and "env"
	go func() {
		for req := range requests {
			switch req.Type {
			case "shell":
				fmt.Println("shell")
				// We only accept the default shell
				// (i.e. no command in the Payload)
				if len(req.Payload) == 0 {
					req.Reply(true, nil)
				}
			case "pty-req":
				fmt.Println("pty-req")
				termLen := req.Payload[3]
				w, h := parseDims(req.Payload[termLen+4:])
				fmt.Println(w, h)
				req.Reply(true, nil)
			case "window-change":
				fmt.Println("window-change")
				w, h := parseDims(req.Payload)
				fmt.Println(w, h)
			}
		}
	}()

	io.WriteString(connection, "Welcome banner\n")
	b := make([]byte, 8)
	for {
		n, err := connection.Read(b)
		fmt.Printf("n = %v err = %v b = %v\n", n, err, b)
		fmt.Printf("b[:n] = %q\n", b[:n])

		if b[0] == 'q' {
			io.WriteString(connection, "*** Bye! ***\n")
			connection.Close()
			break
		}

		nb := []byte{}
		for _, c := range b[:n] {
			nb = append(nb, c)
			if c == '\r' {
				nb = append(nb, '\n')
			}
		}

		io.WriteString(connection, string(nb))
		if err == io.EOF {
			break
		}
	}
}

// parseDims extracts terminal dimensions (width x height) from the provided buffer.
func parseDims(b []byte) (uint32, uint32) {
	w := binary.BigEndian.Uint32(b)
	h := binary.BigEndian.Uint32(b[4:])
	return w, h
}
