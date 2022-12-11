package luaengine

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/creack/pty"
	"github.com/crgimenes/atomic/client"
	"github.com/crgimenes/atomic/config"
	"github.com/crgimenes/atomic/term"
	lua "github.com/yuin/gopher-lua"
	parse "github.com/yuin/gopher-lua/parse"
)

// LuaExtender holds an instance of the moon interpreter and the state variables of the extensions we made.
type LuaExtender struct {
	Term         term.Term
	mutex        sync.RWMutex
	luaState     *lua.LState
	ci           *client.Instance
	triggerList  map[string]*lua.LFunction
	Proto        *lua.FunctionProto
	ExternalExec bool
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
	le.luaState.SetGlobal("fileExists", le.luaState.NewFunction(le.fileExists))
	le.luaState.SetGlobal("exec", le.luaState.NewFunction(le.exec))
	le.luaState.SetGlobal("getEnv", le.luaState.NewFunction(le.getEnv))
	le.luaState.SetGlobal("setOutputMode", le.luaState.NewFunction(le.setOutputMode))
	le.luaState.SetGlobal("limitInputLength", le.luaState.NewFunction(le.limitInputLength))

	return le
}

func (le *LuaExtender) setOutputMode(l *lua.LState) int {
	s := l.ToString(1)
	le.Term.SetOutputMode(s)
	return 0
}

func (le *LuaExtender) limitInputLength(l *lua.LState) int {
	i := l.ToInt(1)
	le.Term.SetInputLimit(i)
	return 0
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

// getEnv returns the value of the environment variable named by the key.
func (le *LuaExtender) getEnv(l *lua.LState) int {
	key := l.ToString(1)
	value := le.ci.Environment[key]
	l.Push(lua.LString(value))
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

func (le *LuaExtender) fileExists(l *lua.LState) int {
	filename := l.ToString(1)
	res := lua.LBool(fileExists(filename))
	l.Push(res)
	return 1
}

func (le *LuaExtender) exec(l *lua.LState) int {
	execFile := l.ToString(1)
	le.ExternalExec = true
	le.Term.Cls()
	npty, ntty, err := pty.Open()
	if err != nil {
		log.Printf("Could not start pty (%s)", err)
	}

	cmd := exec.Command(execFile)
	cmd.Stdout = ntty
	cmd.Stdin = ntty
	cmd.Stderr = ntty
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setctty: true,
		Setsid:  true, // TODO: Evaluate me?
	}

	if err := cmd.Start(); err != nil {
		log.Printf("Failed to start %v (%s)", execFile, err)
		return 0
	}

	go func() {
		/*
			n, err := io.Copy(le.ci.Conn, npty)
			if err != nil {
				log.Printf("1 n: %v (%v)", n, err)
			}
		*/
		for {
			b := make([]byte, 8)
			n, err := npty.Read(b)
			if err != nil {
				log.Printf("1 -> n: %v (%v)", n, err)
				return
			}
			_, err = le.ci.Conn.Write(b[:n])
			if err != nil {
				log.Printf("1 <- n: %v (%v)", n, err)
				return
			}
		}
	}()

	go func() {
		b := make([]byte, 1024)
		/*
			n, err := io.Copy(npty, le.ci.Conn)
			if err != nil {
				log.Printf("2 n: %v (%v)", n, err)
			}
		*/
		for {
			n, err := le.ci.Conn.Read(b)
			if err != nil {
				log.Printf("2 -> n: %v (%v)", n, err)
				return
			}
			_, err = npty.Write(b[:n])
			if err != nil {
				log.Printf("2 <- n: %v (%v) %q", n, err, b[:n])
				//le.ci.Conn.Write(b[:n])
				//le.Input(string(b[:n]))
				ok, err := le.RunTrigger(string(b[:n]))
				if err != nil {
					log.Println("error RunTrigger", err.Error())
					return
				}
				if !ok {
					le.Input(string(b[:n]))
				}

				return
			}
		}
	}()

	go func() {
		var w, h uint32
		var sizeAux string

		for {
			// TODO: improve using a channels
			if !le.ExternalExec {
				return
			}
			time.Sleep(100 * time.Millisecond)
			w = le.ci.W
			h = le.ci.H
			if sizeAux == fmt.Sprintf("%d;%d", w, h) {
				continue
			}

			SetWinsize(npty.Fd(), w, h)
			sizeAux = fmt.Sprintf("%d;%d", w, h)
		}
	}()

	cmd.Wait()
	le.ExternalExec = false

	return 0
}

// Winsize stores the Height and Width of a terminal.
type Winsize struct {
	Height uint16
	Width  uint16
	x      uint16 // unused
	y      uint16 // unused
}

func SetWinsize(fd uintptr, w, h uint32) {
	ws := &Winsize{Width: uint16(w), Height: uint16(h)}
	syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCSWINSZ), uintptr(unsafe.Pointer(ws)))
}
