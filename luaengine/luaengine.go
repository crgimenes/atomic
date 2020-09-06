package luaengine

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sync"

	"github.com/crgimenes/atomic/charset"
	"github.com/crgimenes/atomic/client"
	lua "github.com/yuin/gopher-lua"
)

type LuaExtender struct {
	mutex        sync.RWMutex
	luaState     *lua.LState
	triggerList  map[string]*lua.LFunction
	ci           *client.Instance
	captureInput bool
	echo         bool
	inputField   string
	inputTrigger chan struct{}
}

func New() *LuaExtender {
	le := &LuaExtender{}
	le.triggerList = make(map[string]*lua.LFunction)
	le.inputTrigger = make(chan struct{})
	le.luaState = lua.NewState()
	le.luaState.SetGlobal("pwd", le.luaState.NewFunction(le.pwd))
	le.luaState.SetGlobal("trigger", le.luaState.NewFunction(le.trigger))
	le.luaState.SetGlobal("quit", le.luaState.NewFunction(le.quit))
	le.luaState.SetGlobal("write", le.luaState.NewFunction(le.write))
	le.luaState.SetGlobal("cls", le.luaState.NewFunction(le.cls))
	le.luaState.SetGlobal("setANSI", le.luaState.NewFunction(le.setANSI))
	le.luaState.SetGlobal("resetScreen", le.luaState.NewFunction(le.resetScreen))
	le.luaState.SetGlobal("setEcho", le.luaState.NewFunction(le.setEcho))
	le.luaState.SetGlobal("getField", le.luaState.NewFunction(le.getField))
	le.luaState.SetGlobal("writeFromASCII", le.luaState.NewFunction(le.writeFromASCII))

	return le
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
	err = le.luaState.DoString(string(b))
	return err
}

func (le *LuaExtender) Input(s string) {
	if le.echo {
		if s[0] == '\x1b' {
			return
		}
		le.writeString(s)
	}
	if s == "\r" {
		if le.captureInput {
			le.captureInput = false
			le.inputTrigger <- struct{}{}
		}
	}
	if le.captureInput {
		le.inputField += s
	}

	fmt.Printf("input: %q\n", s)
}

func (le *LuaExtender) writeString(s string) {
	_, err := io.WriteString(le.ci.Conn, s)
	if err != nil {
		log.Println(err.Error())
	}
}

func (le *LuaExtender) getField(l *lua.LState) int {
	le.echo = true
	le.captureInput = true
	le.inputField = ""

	<-le.inputTrigger

	res := lua.LString(le.inputField)
	le.inputField = ""
	le.echo = false
	l.Push(res)
	return 1
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
	le.echo = b
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
			s += ";"
		}
		s += v
	}
	s += "m"
	le.writeString(s)
	return 0
}

func (le *LuaExtender) writeFromASCII(l *lua.LState) int {
	fileName := l.ToString(1)
	f, err := os.Open(fileName) // For read access.
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	r := bufio.NewReader(f)
	for {
		b, err := r.ReadByte()
		if err == io.EOF {
			break
		}
		le.writeString(string(charset.ASCII[b]))
	}
	return 0
}

func (le *LuaExtender) resetScreen(l *lua.LState) int {
	le.writeString("\u001bc")
	return 0
}

func (le *LuaExtender) cls(l *lua.LState) int {
	le.writeString("\u001b[2J")
	return 0
}

func (le *LuaExtender) write(l *lua.LState) int {
	s := l.ToString(1)
	le.writeString(s)
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
