package luaengine

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"github.com/crgimenes/atomic/client"
	"github.com/crgimenes/atomic/config"
	"github.com/crgimenes/atomic/term"
	lua "github.com/yuin/gopher-lua"
	parse "github.com/yuin/gopher-lua/parse"
)

// LuaExtender holds an instance of the moon interpreter and the state variables of the extensions we made.
type LuaExtender struct {
	Term        term.Term
	mutex       sync.RWMutex
	luaState    *lua.LState
	ci          *client.Instance
	triggerList map[string]*lua.LFunction
	Proto       *lua.FunctionProto
}

// New creates a new instance of LuaExtender.
func New(cfg config.Config) *LuaExtender {
	le := &LuaExtender{}
	le.triggerList = make(map[string]*lua.LFunction)
	le.luaState = lua.NewState()
	le.luaState.SetGlobal("pwd", le.luaState.NewFunction(le.pwd))
	le.luaState.SetGlobal("timer", le.luaState.NewFunction(le.timer))
	le.luaState.SetGlobal("trigger", le.luaState.NewFunction(le.trigger))
	le.luaState.SetGlobal("rmTrigger", le.luaState.NewFunction(le.removeTrigger))
	le.luaState.SetGlobal("clearTriggers", le.luaState.NewFunction(le.clearTriggers))
	le.luaState.SetGlobal("quit", le.luaState.NewFunction(le.quit))
	le.luaState.SetGlobal("cls", le.luaState.NewFunction(le.cls))
	le.luaState.SetGlobal("setEcho", le.luaState.NewFunction(le.setEcho))
	le.luaState.SetGlobal("getField", le.luaState.NewFunction(le.getField))
	le.luaState.SetGlobal("getPassword", le.luaState.NewFunction(le.getPassword))
	le.luaState.SetGlobal("write", le.luaState.NewFunction(le.write))
	le.luaState.SetGlobal("writeFromASCII", le.luaState.NewFunction(le.writeFromASCII))
	le.luaState.SetGlobal("inlineImagesProtocol", le.luaState.NewFunction(le.inlineImagesProtocol))

	return le
}

func (le *LuaExtender) Close() error {
	return le.Close()
}

func (le *LuaExtender) Input(s string) {
	le.Term.Input(s)
}

// GetState returns the state of the moon interpreter.
func (le *LuaExtender) GetState() *lua.LState {
	return le.luaState
}

func (le *LuaExtender) cls(l *lua.LState) int {
	err := le.Term.Clear()
	if err != nil {
		fmt.Println(err)
	}
	return 0
}

func (le *LuaExtender) writeFromASCII(l *lua.LState) int {
	s := l.ToString(1)
	le.Term.WriteFromASCII(s)

	return 0
}

func (le *LuaExtender) inlineImagesProtocol(l *lua.LState) int {
	s := l.ToString(1)

	f, err := os.Open(s)
	if err != nil {
		fmt.Println(err)
	}
	reader := bufio.NewReader(f)
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		fmt.Println(err)
	}

	encoded := base64.StdEncoding.EncodeToString(content)

	le.Term.WriteString("\033]1337;File=inline=1;preserveAspectRatio=1:")
	le.Term.WriteString(encoded)
	le.Term.WriteString("\a")

	return 0
}

func (le *LuaExtender) write(l *lua.LState) int {
	s := l.ToString(1)
	le.Term.WriteString(s)

	return 0
}

// CompileLua reads the passed lua file from disk and compiles it.
func (le *LuaExtender) Compile(filePath string) (*lua.FunctionProto, error) {
	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(file)
	chunk, err := parse.Parse(reader, filePath)
	if err != nil {
		return nil, err
	}
	proto, err := lua.Compile(chunk, filePath)
	if err != nil {
		return nil, err
	}
	return proto, nil
}

// DoCompiledFile takes a FunctionProto, as returned by CompileLua, and runs it in the LState. It is equivalent
// to calling DoFile on the LState with the original source file.
func (le *LuaExtender) DoCompiledFile(L *lua.LState, proto *lua.FunctionProto) error {
	lfunc := L.NewFunctionFromProto(proto)
	L.Push(lfunc)
	return L.PCall(0, lua.MultRet, nil)
}

// InitState starts the lua interpreter with a script.
func (le *LuaExtender) InitState(ci *client.Instance) error {
	le.ci = ci
	le.Term = term.Term{
		C: le.ci.Conn,
	}
	le.Term.Init()
	return le.DoCompiledFile(le.luaState, le.Proto)
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

// RunTrigger executes a pre-configured trigger.
func (le *LuaExtender) RunTrigger(name string) (bool, error) {
	le.mutex.Lock()
	f, ok := le.triggerList[name]
	le.mutex.Unlock()
	if !ok {
		return false, nil
	}

	err := le.luaState.CallByParam(lua.P{
		Fn:      f,    // Lua function
		NRet:    0,    // number of returned values
		Protect: true, // return err or panic
	})
	return true, err
}

func (le *LuaExtender) setEcho(l *lua.LState) int {
	b := l.ToBool(1)
	le.Term.SetEcho(b)
	return 0
}

func (le *LuaExtender) removeTrigger(l *lua.LState) int {
	n := l.ToString(1) // name
	le.mutex.Lock()
	delete(le.triggerList, n)
	le.mutex.Unlock()
	return 0
}

func (le *LuaExtender) clearTriggers(l *lua.LState) int {
	le.mutex.Lock()
	le.triggerList = make(map[string]*lua.LFunction)
	le.mutex.Unlock()
	return 0
}

func (le *LuaExtender) timer(l *lua.LState) int {
	n := l.ToString(1)   // name
	t := l.ToInt(2)      // timer
	f := l.ToFunction(3) // function

	if n == "" {
		n = "timer"
	}

	le.mutex.Lock()
	le.triggerList[n] = f
	le.mutex.Unlock()

	go func() {
		for {
			<-time.After(time.Duration(t) * time.Millisecond)
			le.mutex.Lock()
			_, ok := le.triggerList[n]
			le.mutex.Unlock()
			if !le.ci.IsConnected || !ok {
				return
			}
			ok, err := le.RunTrigger(n)
			if err != nil {
				log.Println(n, "timer trigger error", err)
				return
			}
			if !ok {
				return
			}
		}
	}()

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
	le.ci.IsConnected = false
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
