package luaengine

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sync"

	"github.com/crgimenes/atomic/client"
	lua "github.com/yuin/gopher-lua"
)

type LuaExtender struct {
	mutex       sync.RWMutex
	luaState    *lua.LState
	triggerList map[string]*lua.LFunction
	ci          *client.Instance
}

func New() *LuaExtender {
	return &LuaExtender{}
}

func (le *LuaExtender) GetState() *lua.LState {
	return le.luaState
}

func (le *LuaExtender) InitState(r io.Reader, ci *client.Instance) error {
	le.ci = ci
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	le.triggerList = make(map[string]*lua.LFunction)

	le.luaState = lua.NewState()
	le.luaState.SetGlobal("pwd", le.luaState.NewFunction(le.pwd))
	le.luaState.SetGlobal("trigger", le.luaState.NewFunction(le.trigger))
	le.luaState.SetGlobal("quit", le.luaState.NewFunction(le.quit))
	le.luaState.SetGlobal("write", le.luaState.NewFunction(le.write))
	le.luaState.SetGlobal("cls", le.luaState.NewFunction(le.cls))
	le.luaState.SetGlobal("setANSI", le.luaState.NewFunction(le.setANSI))
	le.luaState.SetGlobal("resetScreen", le.luaState.NewFunction(le.resetScreen))
	le.luaState.SetGlobal("setEcho", le.luaState.NewFunction(le.setEcho))

	err = le.luaState.DoString(string(b))
	return err
}

func (le *LuaExtender) RunTriggrer(name string) (bool, error) {
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
	le.ci.Echo = b
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

func (le *LuaExtender) setANSI(l *lua.LState) int {

	s := "\u001b["
	for i := 1; i <= l.GetTop(); i++ {
		v := l.Get(i).String()
		if i > 1 {
			s = s + ";"
		}
		s = s + v
	}
	s = s + "m"
	_, err := io.WriteString(le.ci.Conn, s)
	if err != nil {
		log.Println(err.Error())
	}
	return 0
}

func (le *LuaExtender) resetScreen(l *lua.LState) int {
	_, err := io.WriteString(le.ci.Conn, "\u001bc")
	if err != nil {
		log.Println(err.Error())
	}
	return 0
}

func (le *LuaExtender) cls(l *lua.LState) int {
	_, err := io.WriteString(le.ci.Conn, "\u001b[2J")
	if err != nil {
		log.Println(err.Error())
	}
	return 0
}

func (le *LuaExtender) write(l *lua.LState) int {
	a := l.ToString(1)

	_, err := io.WriteString(le.ci.Conn, a)
	if err != nil {
		log.Println(err.Error())
	}

	return 0
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
