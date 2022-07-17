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
	"github.com/crgimenes/atomic/config"
	"github.com/crgimenes/atomic/luaengine"
	"golang.org/x/crypto/ssh"
)

type SSHServer struct {
	mux sync.Mutex
	cfg config.Config
}

func New(cfg config.Config) *SSHServer {
	return &SSHServer{
		cfg: cfg,
	}
}

func newServerConfig() (*ssh.ServerConfig, error) {
	b, err := ioutil.ReadFile("id_rsa")
	if err != nil {
		return nil, fmt.Errorf("failed to load private key, %v", err.Error())
	}

	pk, err := ssh.ParsePrivateKey(b)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key, %v", err.Error())
	}

	scfg := &ssh.ServerConfig{
		// TODO: improve authentication (allow key pair authentication, etc.)

		/*
			PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
				if c.User() == "foo" && string(pass) == "bar" {
					return nil, nil
				}
				return nil, fmt.Errorf("password rejected for %q", c.User())
			},
		*/

		NoClientAuth: true, // allow anonymous client authentication
		// The BBS system will authenticate the user in another step that is
		// more lime the old BBS systems and also allow guest access.
	}
	scfg.AddHostKey(pk)

	// SSH-2.0-Go
	scfg.ServerVersion = "SSH-2.0-ATOMIC"

	return scfg, nil

}

func (s *SSHServer) ListenAndServe() error {

	l, err := net.Listen("tcp", "0.0.0.0:2200")
	if err != nil {
		return fmt.Errorf("failed to listen on 2200, %v", err.Error())
	}

	scfg, err := newServerConfig()
	if err != nil {
		return err
	}

	log.Print("listening at 0.0.0.0:2200")

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("failed to accept incoming conn", err.Error())
			continue
		}

		fmt.Println("ip:", conn.RemoteAddr())

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
		go s.handleChannels(sshConn, chans)
	}
}

func (s *SSHServer) handleChannels(serverConn *ssh.ServerConn, chans <-chan ssh.NewChannel) {
	// Service the incoming Channel channel in go routine
	for newChannel := range chans {
		go s.handleChannel(serverConn, newChannel)
	}
}

func (s *SSHServer) handleChannel(serverConn *ssh.ServerConn, newChannel ssh.NewChannel) {
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

	l := luaengine.New(s.cfg)
	// defer l.Close() TODO: Verificar se é necessário fechar o engine
	ci := client.NewInstance(conn)

	// Sessions have out-of-band requests such as "shell", "pty-req" and "env"
	go func() {
		for req := range requests {
			log.Printf("%q %q", req.Type, req.Payload)
			switch req.Type {
			case "shell":
				// We only accept the default shell
				// (i.e. no command in the Payload)
				if len(req.Payload) == 0 {
					err = req.Reply(true, nil)
					if err != nil {
						log.Println(err.Error())
					}
				}
			case "pty-req":
				termLen := req.Payload[3]
				s.mux.Lock()
				ci.W, ci.H = parseDims(req.Payload[termLen+4:])
				s.mux.Unlock()
				err := req.Reply(true, nil)
				if err != nil {
					log.Println(err.Error())
				}
			case "window-change":
				s.mux.Lock()
				ci.W, ci.H = parseDims(req.Payload)
				s.mux.Unlock()
			}
		}
	}()
	go func() {
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
			ok, err := l.RunTrigger(k)
			if err != nil {
				log.Println("error RunTrigger", err.Error())
				break
			}
			if !ok {
				l.Input(string(b[:n]))
			}
		}
		ci.IsConnected = false
	}()

	if l.Proto == nil {
		proto, err := l.Compile("fixtures/init.lua")
		if err != nil {
			log.Println(err.Error())
			os.Exit(1)
		}
		l.Proto = proto
	}

	err = l.InitState(l.Proto, ci)
	if err != nil {
		log.Println("can't open init.lua file", err.Error())
		os.Exit(1)
	}
}

// parseDims extracts terminal dimensions (width x height) from the provided buffer.
func parseDims(b []byte) (uint32, uint32) {
	w := binary.BigEndian.Uint32(b)
	h := binary.BigEndian.Uint32(b[4:])
	return w, h
}
