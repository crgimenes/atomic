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
	"time"

	"github.com/crgimenes/atomic/client"
	"github.com/crgimenes/atomic/config"
	"github.com/crgimenes/atomic/luaengine"
	lua "github.com/yuin/gopher-lua"
	"golang.org/x/crypto/ssh"
)

type SSHServer struct {
	mux   sync.Mutex
	proto *lua.FunctionProto
	cfg   config.Config
}

func New(cfg config.Config) *SSHServer {
	return &SSHServer{
		cfg: cfg,
	}
}

func (s *SSHServer) newServerConfig() (*ssh.ServerConfig, error) {

	authorizedKeysBytes, err := ioutil.ReadFile("authorized_keys")
	if err != nil {
		log.Fatalf("Failed to load authorized_keys, err: %v", err)
	}

	authorizedKeysMap := map[string]bool{}
	for len(authorizedKeysBytes) > 0 {
		pubKey, _, _, rest, err := ssh.ParseAuthorizedKey(authorizedKeysBytes)
		if err != nil {
			log.Fatal(err)
		}

		authorizedKeysMap[string(pubKey.Marshal())] = true
		authorizedKeysBytes = rest

		log.Printf("authorized key fingerprint: %s, %s", pubKey.Type(), ssh.FingerprintSHA256(pubKey))
	}

	b, err := os.ReadFile(s.cfg.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key, %v", err.Error())
	}

	pk, err := ssh.ParsePrivateKey(b)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key, %v", err.Error())
	}

	log.Printf("private key fingerprint: %s", ssh.FingerprintSHA256(pk.PublicKey()))

	scfg := &ssh.ServerConfig{
		NoClientAuth: false,
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			log.Println("password callback")
			if c.User() == "foo" && string(pass) == "bar" {
				log.Printf("user %v authenticated with password", c.User())
				return nil, nil
			}
			return nil, fmt.Errorf("password rejected for %q", c.User())
		},
		PublicKeyCallback: func(c ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			log.Printf("Public key for %q, %q, %q, %s",
				c.User(),
				c.RemoteAddr(),
				key.Type(),
				ssh.FingerprintSHA256(key))

			if authorizedKeysMap[string(key.Marshal())] {
				return &ssh.Permissions{
					// Record the public key used for authentication.
					Extensions: map[string]string{
						"pubkey-fp": ssh.FingerprintSHA256(key),
					},
				}, nil
			}
			return nil, fmt.Errorf("unknown public key for %q", c.User())
		},
	}
	scfg.AddHostKey(pk)
	// set ssh authentication method
	scfg.AuthLogCallback = func(conn ssh.ConnMetadata, method string, err error) {
		if err != nil {
			log.Printf("Failed authentication for %q from %v, error: %v", conn.User(), conn.RemoteAddr(), err)
			return
		}
		log.Printf("Successful authentication for %q from %v using %v", conn.User(), conn.RemoteAddr(), method)

	}
	scfg.KeyboardInteractiveCallback = func(conn ssh.ConnMetadata, client ssh.KeyboardInteractiveChallenge) (*ssh.Permissions, error) {
		log.Println("keyboard interactive callback")
		if conn.User() == "foo" {
			// We don't care about the provided instructions or echos.
			// We only accept one answer, and it must be "bar".
			answers, err := client("", "", []string{"Password:"}, []bool{true})
			if err != nil {
				return nil, err
			}
			if len(answers) != 1 || answers[0] != "bar" {
				return nil, fmt.Errorf("keyboard-interactive challenge failed")
			}
			return nil, nil
		}
		return nil, fmt.Errorf("keyboard-interactive challenge failed")
	}

	// SSH-2.0-Go
	scfg.ServerVersion = "SSH-2.0-ATOMIC"
	scfg.BannerCallback = func(conn ssh.ConnMetadata) string {
		return "Welcome to Atomic\n"
		// ssh -o PreferredAuthentications=password -o PubkeyAuthentication=no localhost -p 2200

		// supported key types
		// ssh-keygen -t ecdsa -b 521
		// ssh-keygen -t ed25519

	}

	scfg.MaxAuthTries = 3
	return scfg, nil
}

func (s *SSHServer) ListenAndServe() error {

	l, err := net.Listen("tcp", s.cfg.Listen)
	if err != nil {
		return fmt.Errorf("failed to listen on %v, %v\n", s.cfg.Listen, err.Error())
	}

	scfg, err := s.newServerConfig()
	if err != nil {
		return err
	}

	log.Printf("listening at %v\n", s.cfg.Listen)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("failed to accept incoming conn", err.Error())
			continue
		}

		log.Println("ip:", conn.RemoteAddr())

		// Before use, a handshake must be performed on the incoming net.Conn.
		sshConn, chans, reqs, err := ssh.NewServerConn(conn, scfg)
		if err != nil {
			log.Printf("failed to handshake, %v", err.Error())
			continue
		}

		log.Printf("new SSH connection from %s, %s",
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
	ci := client.NewInstance(conn)

	// Sessions have out-of-band requests such as "shell", "pty-req" and "env"
	go func() {
		for req := range requests {
			// log.Printf("%q %q", req.Type, req.Payload)
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
			default:
				log.Printf("unknown request: %s, %q, %v\n", req.Type, req.Payload, req.WantReply)
				err := req.Reply(false, nil)
				if err != nil {
					log.Println(err.Error())
				}
			}
		}
	}()

	go func() {
		b := make([]byte, 8)

		for {
			if l.ExternalExec {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			fmt.Println("reading from client")
			n, err := conn.Read(b)
			fmt.Printf("read: %s (%q)", string(b[:n]), b[:n])
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

	l.Proto = s.proto
	if s.proto == nil {
		//TODO: move to main
		proto, err := l.Compile(s.cfg.InitBBSFile)
		if err != nil {
			log.Println(err.Error())
			os.Exit(1)
		}
		l.Proto = proto
		s.proto = proto
	}
	/*
		if l.Proto == nil {
			proto, err := l.Compile(s.cfg.InitBBSFile)
			if err != nil {
				log.Println(err.Error())
				os.Exit(1)
			}
			l.Proto = proto
		}
	*/

	err = l.InitState(ci)
	if err != nil {
		log.Printf("can't open %v file, %v\n", s.cfg.InitBBSFile, err.Error())
		os.Exit(1)
	}
}

// parseDims extracts terminal dimensions (width x height) from the provided buffer.
func parseDims(b []byte) (uint32, uint32) {
	w := binary.BigEndian.Uint32(b)
	h := binary.BigEndian.Uint32(b[4:])
	return w, h
}
