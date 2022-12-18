package server

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"crg.eti.br/go/atomic/client"
	"crg.eti.br/go/atomic/config"
	"crg.eti.br/go/atomic/database"
	"crg.eti.br/go/atomic/luaengine"
	"crg.eti.br/go/atomic/term"
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

	//TODO: authorized_keys file only to sysop user
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

			//////////////////////////////////////////////

			db, err := database.New(s.cfg)
			if err != nil {
				return nil, err
			}
			defer db.Close()

			user, err := db.GetUserByName(c.User())
			if err != nil {
				return nil, err
			}

			fmt.Printf("user %q, %q\n", user.Nickname, user.Email)

			if user.SSHPublicKey == "" {
				return nil, fmt.Errorf("user %q has no public key", c.User())
			}

			// TODO: support to multiple keys
			pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(user.SSHPublicKey))
			if err != nil {
				return nil, err
			}

			if string(pubKey.Marshal()) != string(key.Marshal()) {
				return nil, fmt.Errorf("error validating public key for %q", c.User())
			}

			return &ssh.Permissions{
				// Record the public key used for authentication.
				Extensions: map[string]string{
					"pubkey-fp": ssh.FingerprintSHA256(key),
				},
			}, nil
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

	term := term.Term{
		C:              conn,
		InputTrigger:   make(chan struct{}),
		OutputMode:     term.UTF8,
		MaxInputLength: 80,
	}
	le := luaengine.New(s.cfg)
	ci := client.NewInstance(conn, term)
	le.Ci = ci
	if s.proto == nil {
		log.Printf("compiling init BBS code\n")
		s.proto, err = le.Compile("init.lua")
		if err != nil {
			log.Println(err.Error())
			os.Exit(1)
		}
	}
	le.Proto = s.proto

	go func() {
		for req := range requests {
			switch req.Type {
			case "shell":
				fmt.Println("shell request")
				// We only accept the default shell
				// (i.e. no command in the Payload)
				if len(req.Payload) == 0 {
					err = req.Reply(true, nil)
					if err != nil {
						log.Println(err.Error())
						return
					}
				}

				err = le.InitState()
				if err != nil {
					log.Printf("can't open %v file, %v\n", filepath.Join(s.cfg.BaseBBSDir, "init.lua"), err.Error())
					os.Exit(1)
				}

			case "pty-req":
				fmt.Println("pty-req request")
				termLen := req.Payload[3]
				s.mux.Lock()
				ci.Term.W, ci.Term.H = parseDims(req.Payload[termLen+4:])
				s.mux.Unlock()
				err := req.Reply(true, nil)
				if err != nil {
					log.Println(err.Error())
					return
				}
			case "window-change":
				fmt.Println("window-change request")
				s.mux.Lock()
				ci.Term.W, ci.Term.H = parseDims(req.Payload)
				s.mux.Unlock()
			case "env":
				fmt.Println("env request")
				err := req.Reply(true, nil)
				if err != nil {
					req.Reply(false, nil)
					log.Println(err.Error())
					return
				}

				if len(ci.Environment) > 1000 {
					log.Println("too many env vars")
					return
				}

				var p client.KeyValue
				ssh.Unmarshal(req.Payload, &p)
				log.Printf("env: %s = %s", p.Key, p.Value)
				environ := ci.Environment
				environ[p.Key] = p.Value
				ci.Environment = environ
				req.Reply(true, nil)

			case "subsystem":
				fmt.Println("subsystem request")

				var subsystem string
				if len(req.Payload) > 4 {
					subsystem = string(req.Payload[4:])
				}

				switch subsystem {
				case "sftp":
					fmt.Println("sftp request unimplemented")
					req.Reply(false, nil)
					return
				default:
					log.Printf("unknown subsystem request: %q", subsystem)
					req.Reply(false, nil)
					return
				}
			default:
				fmt.Println("default request")
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
			if le.ExternalExec {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			n, err := conn.Read(b)
			if err != nil {
				if err != io.EOF {
					log.Println(err.Error())
				}
				break
			}
			k := string(b[:n])
			ok, err := le.RunTrigger(k)
			if err != nil {
				log.Println("error RunTrigger", err.Error())
				break
			}
			if !ok {
				le.Input(string(b[:n]))
			}
		}
		ci.IsConnected = false
	}()
}

// parseDims extracts terminal dimensions (width x height) from the provided buffer.
func parseDims(b []byte) (int, int) {
	w := int(binary.BigEndian.Uint32(b))
	h := int(binary.BigEndian.Uint32(b[4:]))
	return w, h
}
