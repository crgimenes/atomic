package luaengine

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"

	"github.com/crgimenes/atomic/client"
	"github.com/crgimenes/atomic/term"
	lua "github.com/yuin/gopher-lua"
)

// LuaExtender holds an instance of the moon interpreter and the state variables of the extensions we made.
type LuaExtender struct {
	Term        term.Term
	mutex       sync.RWMutex
	luaState    *lua.LState
	triggerList map[string]*lua.LFunction
	ci          *client.Instance
}

// New creates a new instance of LuaExtender
func New() *LuaExtender {
	le := &LuaExtender{}
	le.triggerList = make(map[string]*lua.LFunction)
	le.luaState = lua.NewState()
	le.luaState.SetGlobal("pwd", le.luaState.NewFunction(le.pwd))
	le.luaState.SetGlobal("trigger", le.luaState.NewFunction(le.trigger))
	le.luaState.SetGlobal("quit", le.luaState.NewFunction(le.quit))
	le.luaState.SetGlobal("cls", le.luaState.NewFunction(le.cls))
	le.luaState.SetGlobal("setANSI", le.luaState.NewFunction(le.setANSI))
	le.luaState.SetGlobal("setEcho", le.luaState.NewFunction(le.setEcho))
	le.luaState.SetGlobal("getField", le.luaState.NewFunction(le.getField))
	le.luaState.SetGlobal("getPassword", le.luaState.NewFunction(le.getPassword))
	le.luaState.SetGlobal("write", le.luaState.NewFunction(le.write))
	le.luaState.SetGlobal("writeFromASCII", le.luaState.NewFunction(le.writeFromASCII))

	return le
}

func (le *LuaExtender) Input(s string) {
	le.Term.Input(s)
}

// GetState returns the state of the moon interpreter
func (le *LuaExtender) GetState() *lua.LState {
	return le.luaState
}

func (le *LuaExtender) cls(l *lua.LState) int {
	le.Term.Clear()
	return 0
}

func (le *LuaExtender) writeFromASCII(l *lua.LState) int {
	return 0
}

func (le *LuaExtender) write(l *lua.LState) int {
	return 0
}

func (le *LuaExtender) setANSI(l *lua.LState) int {
	return 0
}

// InitState starts the lua interpreter with a script
func (le *LuaExtender) InitState(r io.Reader, ci *client.Instance) error {
	le.ci = ci
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	le.Term = term.Term{
		C: le.ci.Conn,
	}
	le.Term.Init()
	err = le.luaState.DoString(string(b))
	return err
}

func (le *LuaExtender) getField(l *lua.LState) int {
	res := lua.LString(le.Term.GetField())
	l.Push(res)
	return 1
}

func (le *LuaExtender) getPassword(l *lua.LState) int {
	res := lua.LString(le.Term.GetField())
	l.Push(res)
	return 1
}

// RunTrigger executes a pre-configured trigger
func (le *LuaExtender) RunTrigger(name string) (bool, error) {
	le.mutex.Lock()
	defer le.mutex.Unlock()

	f, ok := le.triggerList[name]
	if !ok {
		return false, nil
	}

	err := le.luaState.CallByParam(lua.P{
		Fn:      f,     // Lua function
		NRet:    0,     // number of returned values
		Protect: false, // return err or panic
	})
	return true, err
}

func (le *LuaExtender) setEcho(l *lua.LState) int {
	b := l.ToBool(1)
	le.Term.SetEcho(b)
	return 0
}

func (le *LuaExtender) trigger(l *lua.LState) int {
	a := l.ToString(1)
	f := l.ToFunction(2)

	le.mutex.Lock()
	le.triggerList[a] = f
	le.mutex.Unlock()

	res := lua.LString(a)
	l.Push(res)
	return 1
}

func (le *LuaExtender) quit(l *lua.LState) int {
	le.ci.Conn.Close()
	return 0
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
