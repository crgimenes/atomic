package client

import (
	"io"

	"golang.org/x/crypto/ssh"
)

type LuaEngine interface {
	InitState(r io.Reader, ci *Instance) error
	Input(c string)
	RunTrigger(name string) (bool, error)
}

type Instance struct {
	Le   LuaEngine
	Conn ssh.Channel
	H, W uint32
}

func NewInstance(conn ssh.Channel, le LuaEngine) *Instance {
	return &Instance{
		Le:   le,
		Conn: conn,
		H:    25,
		W:    80,
	}
}
