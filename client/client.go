package client

import (
	"crg.eti.br/go/atomic/term"
	"golang.org/x/crypto/ssh"
)

type KeyValue struct {
	Key   string
	Value string
}

type Instance struct {
	Term        term.Term
	Conn        ssh.Channel
	IsConnected bool
	Environment map[string]string
}

func NewInstance(conn ssh.Channel, term term.Term) Instance {
	return Instance{
		Term:        term,
		Conn:        conn,
		IsConnected: true,
		Environment: make(map[string]string),
	}
}
