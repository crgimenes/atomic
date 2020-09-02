package server

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"sync"

	"github.com/crgimenes/atomic/client"
	"golang.org/x/crypto/ssh"
)

type LuaEngine interface {
	InitState(r io.Reader, ci *client.Instance) error
	Input(c string)
	RunTriggrer(name string) (bool, error)
}

type SSHServer struct {
	mux sync.Mutex
	le  LuaEngine
}

func New(le LuaEngine) *SSHServer {
	return &SSHServer{
		le: le,
	}
}

func (s *SSHServer) ListenAndServe() error {

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

	l, err := net.Listen("tcp", "0.0.0.0:2200")
	if err != nil {
		return fmt.Errorf("failed to listen on 2200, %v", err.Error())
	}

	log.Print("listening at 0.0.0.0:2200")
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("failed to accept incoming conn", err.Error())
			continue
		}
		// Before use, a handshake must be performed on the incoming net.Conn.
		sshConn, chans, reqs, err := ssh.NewServerConn(conn, scfg)
		if err != nil {
			log.Println("failed to handshake", err.Error())
			continue
		}

		log.Printf("new SSH conn from %s (%s)",
			sshConn.RemoteAddr(),
			sshConn.ClientVersion())
		// Discard all global out-of-band Requests
		go ssh.DiscardRequests(reqs)
		// Accept all channels
		go s.handleChannels(chans)
	}
}

func (s *SSHServer) handleChannels(chans <-chan ssh.NewChannel) {
	// Service the incoming Channel channel in go routine
	for newChannel := range chans {
		go s.handleChannel(newChannel)
	}
}

func (s *SSHServer) handleChannel(newChannel ssh.NewChannel) {
	// Since we're handling a shell, we expect a
	// channel type of "session". The also describes
	// "x11", "direct-tcpip" and "forwarded-tcpip"
	// channel types.
	t := newChannel.ChannelType()
	if t != "session" {
		err := newChannel.Reject(ssh.UnknownChannelType,
			fmt.Sprintf("unknown channel type: %s", t))
		if err != nil {
			log.Printf("error unknow channel type (%s)", err)
		}
		return
	}

	// At this point, we have the opportunity to reject the client's
	// request for another logical conn
	conn, requests, err := newChannel.Accept()
	if err != nil {
		log.Printf("Could not accept channel (%s)", err)
		return
	}

	file, err := os.Open("fixtures/init.lua")
	if err != nil {
		log.Println("can't open init.lua file", err.Error())
		os.Exit(1)
	}

	ci := client.NewInstance(conn)

	err = s.le.InitState(file, ci)
	if err != nil {
		log.Println("can't open init.lua file", err.Error())
		os.Exit(1)
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
					err = req.Reply(true, nil)
					if err != nil {
						log.Println(err.Error())
					}
				}
			case "pty-req":
				fmt.Println("pty-req")
				termLen := req.Payload[3]
				s.mux.Lock()
				ci.W, ci.H = parseDims(req.Payload[termLen+4:])
				s.mux.Unlock()
				err := req.Reply(true, nil)
				if err != nil {
					log.Println(err.Error())
				}
			case "window-change":
				fmt.Println("window-change")
				s.mux.Lock()
				ci.W, ci.H = parseDims(req.Payload)
				s.mux.Unlock()
			}
		}
	}()

	b := make([]byte, 8)
	for {
		n, err := conn.Read(b)
		if err != nil {
			if err != io.EOF {
				log.Println(err.Error())
			}
			break
		}

		k := string(b[:n])
		ok, err := s.le.RunTriggrer(k)
		if err != nil {
			log.Println("error RunTrigger", err.Error())
			break
		}
		if ok {
			continue
		}

		s.le.Input(string(b[:n]))
	}
}

// parseDims extracts terminal dimensions (width x height) from the provided buffer.
func parseDims(b []byte) (uint32, uint32) {
	w := binary.BigEndian.Uint32(b)
	h := binary.BigEndian.Uint32(b[4:])
	return w, h
}
