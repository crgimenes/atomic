package server

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

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
	Users map[string]*database.User
}

const (
	MaxAuthTries = 5
)

func New(cfg config.Config) *SSHServer {
	return &SSHServer{
		cfg:   cfg,
		Users: make(map[string]*database.User),
	}
}

func (s *SSHServer) authLogCallback(c ssh.ConnMetadata, method string, err error) {
	if err != nil {
		log.Printf("failed authentication for %q from %v, method %v, error: %v", c.User(), c.RemoteAddr(), method, err)
		return
	}
	log.Println("auth log callback", s.Users[c.User()].Nickname)
	log.Printf("successful authentication for %q from %v, method %v", c.User(), c.RemoteAddr(), method)
}

func (s *SSHServer) passwordCallback(c ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
	log.Println("password callback")

	return s.validateLogin(c.User(), string(password))
}

func (s *SSHServer) keyboardInteractiveCallback(c ssh.ConnMetadata, client ssh.KeyboardInteractiveChallenge) (*ssh.Permissions, error) {
	log.Println("keyboard interactive callback")

	answers, err := client("", "", []string{"Password:"}, []bool{true})
	if err != nil {
		return nil, err
	}

	password := ""
	if len(answers) != 1 {
		password = answers[0]
	}

	return s.validateLogin(c.User(), password)
}

func (s *SSHServer) validateLogin(nickname, password string) (*ssh.Permissions, error) {
	db, err := database.New()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	user, err := db.CheckAndReturnUser(nickname, password)
	if err != nil {
		return nil, err
	}

	s.mux.Lock()
	s.Users[nickname] = &user
	s.mux.Unlock()

	return nil, nil
}

func (s *SSHServer) publicKeyCallback(c ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
	log.Printf("public key for %q, %q, %q, %s",
		c.User(),
		c.RemoteAddr(),
		key.Type(),
		ssh.FingerprintSHA256(key))

	db, err := database.New()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	user, err := db.GetUserByNickname(c.User())
	if err != nil {
		return nil, err
	}

	log.Printf("user %q, %q\n", user.Nickname, user.Email)

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

	s.mux.Lock()
	s.Users[c.User()] = &user // TODO: add last login date time
	s.mux.Unlock()

	log.Printf("stu user %q, %q\n", user.Nickname, user.Email)
	log.Printf("map user %q, %q\n", s.Users[c.User()].Nickname, s.Users[c.User()].Email)

	// list all users
	for k, u := range s.Users {
		log.Printf("%v, user %q, %q\n", k, u.Nickname, u.Email)
	}

	return &ssh.Permissions{
		// Record the public key used for authentication.
		Extensions: map[string]string{
			"pubkey-fp": ssh.FingerprintSHA256(key),
		},
	}, nil
}

func (s *SSHServer) bannerCallback(conn ssh.ConnMetadata) string {
	dafaltBanner := []byte("Welcome to Atomic\n")

	f := filepath.Join(s.cfg.BaseBBSDir, "banner")

	_, err := os.Stat(f)
	if err != nil {
		log.Printf("failed to stat banner file, %v", err.Error())

		return string(dafaltBanner)
	}

	b, err := os.ReadFile(f)
	if err != nil {
		log.Printf("failed to read banner file, %v", err.Error())

		return string(dafaltBanner)
	}

	return string(b)
}

func (s *SSHServer) newServerConfig() (*ssh.ServerConfig, error) {
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
		NoClientAuth:                false,
		AuthLogCallback:             s.authLogCallback,
		PasswordCallback:            s.passwordCallback,
		PublicKeyCallback:           s.publicKeyCallback,
		KeyboardInteractiveCallback: s.keyboardInteractiveCallback,
		BannerCallback:              s.bannerCallback,
		ServerVersion:               "SSH-2.0-ATOMIC",
		MaxAuthTries:                MaxAuthTries,
	}
	scfg.AddHostKey(pk)

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
	conn, requests, err := newChannel.Accept()
	if err != nil {
		log.Printf("could not accept channel, %v", err.Error())
		return
	}

	log.Printf("session id: %v", serverConn.SessionID())
	// TODO: close serverConn.Conn.Close()

	term := term.Term{
		C:              conn,
		InputTrigger:   make(chan struct{}),
		OutputMode:     term.UTF8,
		MaxInputLength: 80,
	}

	le := luaengine.New(
		s.cfg,
		&s.Users,
		s.Users[serverConn.User()],
		&term,
		serverConn,
		conn,
	)

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
				log.Println("shell request")
				// We only accept the default shell
				// (i.e. no command in the Payload)
				if len(req.Payload) == 0 {
					err = req.Reply(true, nil)
					if err != nil {
						log.Println(err.Error())
						return
					}
				}

				//////////////////////////////

				_, ok := s.Users[serverConn.User()]
				if !ok {
					log.Printf("user %v not found, recreating\n", serverConn.User())
					// TODO: user is reconnecting, but not found in the list
					return
				}

				// list users
				log.Println("users:")
				for _, u := range s.Users {
					log.Printf("  %v\n", u.Nickname)
				}

				//////////////////////////////

				go func() {
					b := make([]byte, 1024)

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
							le.Input(k)
						}
					}
					le.ClearTriggers(nil)
					le.IsConnected = false
					le.Conn.Close()
					serverConn.Conn.Close() // TODO: detect multiple connections
					delete(s.Users, serverConn.User())
				}()

				err = le.InitState()
				if err != nil {
					log.Printf("error %v\n", err.Error())
					os.Exit(1)
				}

				return

			case "pty-req":
				log.Println("pty-req request")
				termLen := req.Payload[3]
				s.mux.Lock()
				term.Width, term.Height = parseDims(req.Payload[termLen+4:])
				s.mux.Unlock()
				err := req.Reply(true, nil)
				if err != nil {
					log.Println(err.Error())
					return
				}
			case "window-change":
				log.Println("window-change request")
				s.mux.Lock()
				term.Width, term.Height = parseDims(req.Payload)
				s.mux.Unlock()
			case "env":
				err := req.Reply(true, nil)
				if err != nil {
					req.Reply(false, nil)
					log.Println(err.Error())
					return
				}

				if len(le.Environment) > 1000 {
					log.Println("too many env vars")
					return
				}

				var p luaengine.KeyValue
				ssh.Unmarshal(req.Payload, &p)
				log.Printf("env: %s = %s", p.Key, p.Value)
				le.Environment[p.Key] = p.Value
				req.Reply(true, nil)

			case "subsystem":
				log.Println("subsystem request")

				var subsystem string
				if len(req.Payload) > 4 {
					subsystem = string(req.Payload[4:])
				}

				switch subsystem {
				case "sftp":
					log.Println("sftp request unimplemented")
					req.Reply(false, nil)
					return
				default:
					log.Printf("unknown subsystem request: %q", subsystem)
					req.Reply(false, nil)
					return
				}
			default:
				log.Println("default request")
				log.Printf("unknown request: %s, %q, %v\n", req.Type, req.Payload, req.WantReply)
				err := req.Reply(false, nil)
				if err != nil {
					log.Println(err.Error())
				}
			}
		}
	}()

}

// parseDims extracts terminal dimensions (width x height) from the provided buffer.
func parseDims(b []byte) (int, int) {
	w := int(binary.BigEndian.Uint32(b))
	h := int(binary.BigEndian.Uint32(b[4:]))
	return w, h
}
