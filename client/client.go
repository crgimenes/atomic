package client

import (
	"golang.org/x/crypto/ssh"
)

type LuaEngine interface {
	InitState(source string, ci *Instance) error
	Input(c string)
	RunTrigger(name string) (bool, error)
}

type Instance struct {
	Le          LuaEngine
	Conn        ssh.Channel
	H, W        uint32
	IsConnected bool
}

func NewInstance(conn ssh.Channel) *Instance {
	return &Instance{
		Conn:        conn,
		H:           25,
		W:           80,
		IsConnected: true,
	}
}
