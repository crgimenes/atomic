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
	Users map[string]database.User
}

func New(cfg config.Config) *SSHServer {
	return &SSHServer{
		cfg:   cfg,
		Users: make(map[string]database.User),
	}
}

func (s *SSHServer) authLogCallback(c ssh.ConnMetadata, method string, err error) {
	if err != nil {
		log.Printf("Failed authentication for %q from %v, method %v, error: %v", c.User(), c.RemoteAddr(), method, err)
		return
	}
	log.Printf("Successful authentication for %q from %v, method %v", c.User(), c.RemoteAddr(), method)
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
	s.Users[nickname] = user
	s.mux.Unlock()

	return nil, nil
}

func (s *SSHServer) publicKeyCallback(c ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
	log.Printf("Public key for %q, %q, %q, %s",
		c.User(),
		c.RemoteAddr(),
		key.Type(),
		ssh.FingerprintSHA256(key))

	//////////////////////////////////////////////
	// sysop user
	//////////////////////////////////////////////

	if c.User() == "sysop" {
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

		if !authorizedKeysMap[string(key.Marshal())] {
			return nil, fmt.Errorf("error validating public key for sysop")
		}

		// TODO: verificar se o sysop existe no banco de dados
		// se existir carregar os dados do banco de dados.
		// se não existir, criar o sysop no banco de dados.

		s.Users[c.User()] = database.User{ // TODO: get user from database
			Nickname: c.User(),
		}

		return &ssh.Permissions{
			Extensions: map[string]string{
				"pubkey-fp": ssh.FingerprintSHA256(key),
			},
		}, nil

	}

	//////////////////////////////////////////////
	// normal user
	//////////////////////////////////////////////

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

	s.Users[c.User()] = user // TODO: add last login date time

	return &ssh.Permissions{
		// Record the public key used for authentication.
		Extensions: map[string]string{
			"pubkey-fp": ssh.FingerprintSHA256(key),
		},
	}, nil
}

func (s *SSHServer) bannerCallback(conn ssh.ConnMetadata) string {
	return "Welcome to Atomic\n"
	// ssh -o PreferredAuthentications=password -o PubkeyAuthentication=no localhost -p 2200

	// supported key types
	// ssh-keygen -t ecdsa -b 521
	// ssh-keygen -t ed25519

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
		MaxAuthTries:                3,
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

	term := term.Term{
		C:              conn,
		InputTrigger:   make(chan struct{}),
		OutputMode:     term.UTF8,
		MaxInputLength: 80,
	}
	le := luaengine.New(s.cfg)
	le.Users = &s.Users
	ci := client.NewInstance(conn, term)
	le.Ci = ci
	ci.User = s.Users[serverConn.User()]
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
					log.Printf("user %v not found\n", serverConn.User())
					// add user
					s.Users[serverConn.User()] = database.User{
						Nickname: serverConn.User(),
					}
				}

				// list users
				log.Println("users:")
				for _, u := range s.Users {
					log.Printf("  %v\n", u.Nickname)
				}

				//////////////////////////////

				err = le.InitState()
				if err != nil {
					log.Printf("can't open %v file, %v\n", filepath.Join(s.cfg.BaseBBSDir, "init.lua"), err.Error())
					os.Exit(1)
				}

				return

			case "pty-req":
				log.Println("pty-req request")
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
				log.Println("window-change request")
				s.mux.Lock()
				ci.Term.W, ci.Term.H = parseDims(req.Payload)
				s.mux.Unlock()
			case "env":
				log.Println("env request")
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
				ci.Environment[p.Key] = p.Value
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
		ci.IsConnected = false
		delete(s.Users, serverConn.User())
	}()
}

// parseDims extracts terminal dimensions (width x height) from the provided buffer.
func parseDims(b []byte) (int, int) {
	w := int(binary.BigEndian.Uint32(b))
	h := int(binary.BigEndian.Uint32(b[4:]))
	return w, h
}
