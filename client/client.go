package client

import (
	"golang.org/x/crypto/ssh"
)

type Instance struct {
	Conn ssh.Channel
	H, W uint32
}

func NewInstance(conn ssh.Channel) *Instance {
	return &Instance{
		Conn: conn,
		H:    25,
		W:    80,
	}
}
