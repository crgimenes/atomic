package client

import (
	"crg.eti.br/go/atomic/database"
	"golang.org/x/crypto/ssh"
)

type KeyValue struct {
	Key   string
	Value string
}

type Instance struct {
	Conn        ssh.Channel
	IsConnected bool
	Environment map[string]string
	User        *database.User
	ServerConn  *ssh.ServerConn
}

func NewInstance(conn ssh.Channel, serverConn *ssh.ServerConn) *Instance {
	return &Instance{
		Conn:        conn,
		IsConnected: true,
		Environment: make(map[string]string),
		User:        &database.User{},
		ServerConn:  serverConn,
	}
}
