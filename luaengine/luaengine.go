package luaengine

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"

	lua "github.com/yuin/gopher-lua"
)

type LuaExtender struct {
	mutex    sync.RWMutex
	luaState *lua.LState
}

func pwd() string {
	d, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	return d
}

func (le *LuaExtender) pwd(l *lua.LState) int {
	res := lua.LString(pwd())
	l.Push(res)
	return 1
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func (le *LuaExtender) write(w io.Writer, msg string) {
	le.mutex.Lock()
	_, err := w.Write([]byte(msg))
	le.mutex.Unlock()
	if err != nil {
		fmt.Println("write error:", err)
	}
}

func (le *LuaExtender) GetState() *lua.LState {
	return le.luaState
}

func (le *LuaExtender) Run(r io.Reader) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	err = le.luaState.DoString(string(b))
	return err
}

func New() *LuaExtender {
	le := &LuaExtender{}

	le.luaState = lua.NewState()
	le.luaState.SetGlobal("pwd", le.luaState.NewFunction(le.pwd))
	return le
}
