package luaengine

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"crg.eti.br/go/atomic/config"
	"crg.eti.br/go/atomic/database"
	"crg.eti.br/go/atomic/term"
	"github.com/creack/pty"
	lua "github.com/yuin/gopher-lua"
	parse "github.com/yuin/gopher-lua/parse"
	"golang.org/x/crypto/ssh"
)

// LuaExtender holds an instance of the moon interpreter and the state variables of the extensions we made.
type LuaExtender struct {
	mutex        sync.RWMutex
	luaState     *lua.LState
	triggerList  map[string]*lua.LFunction
	Proto        *lua.FunctionProto
	ExternalExec bool
	Users        *map[string]*database.User
	User         *database.User
	Term         *term.Term
	ServerConn   *ssh.ServerConn
	Conn         ssh.Channel
	IsConnected  bool
	Environment  map[string]string
}

type KeyValue struct {
	Key   string
	Value string
}

// New creates a new instance of LuaExtender.
func New(cfg config.Config,
	users *map[string]*database.User,
	user *database.User,
	term *term.Term,
	serverConn *ssh.ServerConn,
	conn ssh.Channel,
) *LuaExtender {

	le := &LuaExtender{
		Users:       users,
		User:        user,
		Term:        term,
		ServerConn:  serverConn,
		Conn:        conn,
		Environment: make(map[string]string),
		IsConnected: true,
	}
	le.triggerList = make(map[string]*lua.LFunction)
	le.luaState = lua.NewState()
	le.luaState.SetGlobal("clearTriggers", le.luaState.NewFunction(le.ClearTriggers))
	le.luaState.SetGlobal("cls", le.luaState.NewFunction(le.cls))
	le.luaState.SetGlobal("exec", le.luaState.NewFunction(le.exec))
	le.luaState.SetGlobal("fileExists", le.luaState.NewFunction(le.fileExists))
	le.luaState.SetGlobal("getEnv", le.luaState.NewFunction(le.getEnv))
	le.luaState.SetGlobal("getField", le.luaState.NewFunction(le.getField))
	le.luaState.SetGlobal("getOutputMode", le.luaState.NewFunction(le.getOutputMode))
	le.luaState.SetGlobal("getPassword", le.luaState.NewFunction(le.getPassword))
	le.luaState.SetGlobal("inlineImagesProtocol", le.luaState.NewFunction(le.inlineImagesProtocol))
	le.luaState.SetGlobal("limitInputLength", le.luaState.NewFunction(le.limitInputLength))
	le.luaState.SetGlobal("logf", le.luaState.NewFunction(le.logf))
	le.luaState.SetGlobal("pwd", le.luaState.NewFunction(le.pwd))
	le.luaState.SetGlobal("quit", le.luaState.NewFunction(le.quit))
	le.luaState.SetGlobal("rmTrigger", le.luaState.NewFunction(le.removeTrigger))
	le.luaState.SetGlobal("setEcho", le.luaState.NewFunction(le.setEcho))
	le.luaState.SetGlobal("setInputLimit", le.luaState.NewFunction(le.setInputLimit))
	le.luaState.SetGlobal("setOutputDelay", le.luaState.NewFunction(le.setOutputDelay))
	le.luaState.SetGlobal("setOutputMode", le.luaState.NewFunction(le.setOutputMode))
	le.luaState.SetGlobal("timer", le.luaState.NewFunction(le.timer))
	le.luaState.SetGlobal("trigger", le.luaState.NewFunction(le.trigger))
	//le.luaState.SetGlobal("write", le.luaState.NewFunction(le.write))
	le.luaState.SetGlobal("writeFromASCII", le.luaState.NewFunction(le.writeFromASCII))
	le.luaState.SetGlobal("getUser", le.luaState.NewFunction(le.getUser))
	le.luaState.SetGlobal("hasGroup", le.luaState.NewFunction(le.hasGroup))
	le.luaState.SetGlobal("setMaxInputLength", le.luaState.NewFunction(le.setMaxInputLength))

	le.luaState.PreloadModule("term", le.termLoader)
	return le
}

func (le *LuaExtender) setMaxInputLength(l *lua.LState) int {
	le.Term.MaxInputLength = int(l.ToNumber(1))
	return 0
}

func (le *LuaExtender) hasGroup(l *lua.LState) int {
	group := l.ToString(1)
	groups := strings.Split(le.User.Groups, ",")
	for _, g := range groups {
		if g == group {
			l.Push(lua.LBool(true))
			return 1
		}
	}
	l.Push(lua.LBool(false))
	return 1
}

func (le *LuaExtender) getUser(l *lua.LState) int {
	u := le.User
	tbl := l.NewTable()
	l.SetField(tbl, "id", lua.LNumber(u.ID))
	l.SetField(tbl, "nickname", lua.LString(u.Nickname))
	l.SetField(tbl, "email", lua.LString(u.Email))
	l.SetField(tbl, "groups", lua.LString(u.Groups))
	l.SetField(tbl, "created_at", lua.LString(u.CreatedAt))
	l.SetField(tbl, "updated_at", lua.LString(u.UpdatedAt))
	l.Push(tbl)
	return 1
}

func (le *LuaExtender) logf(l *lua.LState) int {
	format := l.ToString(1)
	args := make([]interface{}, l.GetTop()-1)
	for i := 2; i <= l.GetTop(); i++ {
		args[i-2] = l.ToString(i)
	}
	log.Printf(format, args...)
	return 0
}

func (le *LuaExtender) setOutputDelay(l *lua.LState) int {
	i := l.ToInt(1)
	le.Term.SetOutputDelay(i)
	return 0
}

func (le *LuaExtender) getOutputMode(l *lua.LState) int {
	res := lua.LString(le.Term.GetOutputDisplay())
	l.Push(res)
	return 1
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
		log.Println(err)
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
		log.Println(err)
	}
	reader := bufio.NewReader(f)
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Println(err)
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
func (le *LuaExtender) InitState() error {
	return le.DoCompiledFile(le.luaState, le.Proto)
}

func (le *LuaExtender) setInputLimit(l *lua.LState) int {
	i := l.ToInt(1)
	le.Term.SetInputLimit(i)
	return 0
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
	value := le.Environment[key]
	l.Push(lua.LString(value))
	return 1
}

// RunTrigger executes a pre-configured trigger.
func (le *LuaExtender) RunTrigger(name string) (bool, error) {
	f, ok := le.triggerList[name]
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

func (le *LuaExtender) ClearTriggers(l *lua.LState) int {
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
			if !le.IsConnected || !ok {
				return
			}
			le.mutex.Lock()
			ok, err := le.RunTrigger(n)
			le.mutex.Unlock()
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
	le.Conn.Close()
	le.IsConnected = false
	le.ServerConn.Conn.Close()
	delete(*le.Users, le.User.Nickname)
	return 0
}

func pwd() string {
	d, err := os.Getwd()
	if err != nil {
		log.Println(err)
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
		for {
			b := make([]byte, 8)
			n, err := npty.Read(b)
			if err != nil {
				log.Printf("1 -> n: %v (%v)", n, err)
				return
			}
			_, err = le.Conn.Write(b[:n])
			if err != nil {
				log.Printf("1 <- n: %v (%v)", n, err)
				return
			}
			if !le.IsConnected {
				return
			}
		}
	}()

	go func() {
		b := make([]byte, 1024)
		for {
			n, err := le.Conn.Read(b)
			if err != nil {
				log.Printf("2 -> n: %v (%v)", n, err)
				return
			}
			_, err = npty.Write(b[:n])
			if err != nil {
				log.Printf("2 <- n: %v (%v) %q", n, err, b[:n])
				le.mutex.Lock()
				ok, err := le.RunTrigger(string(b[:n]))
				le.mutex.Unlock()
				if err != nil {
					log.Println("error RunTrigger", err.Error())
					return
				}
				if !ok {
					le.Input(string(b[:n]))
				}

				return
			}
			if !le.IsConnected {
				return
			}
		}
	}()

	go func() {
		var w, h int
		var sizeAux string

		for {
			// TODO: improve using a channels
			le.mutex.Lock()
			if !le.ExternalExec {
				le.mutex.Unlock()
				return
			}
			le.mutex.Unlock()
			time.Sleep(100 * time.Millisecond)
			w = le.Term.W
			h = le.Term.H
			if sizeAux == fmt.Sprintf("%d;%d", w, h) {
				continue
			}

			SetWinsize(npty.Fd(), w, h)
			sizeAux = fmt.Sprintf("%d;%d", w, h)

			if !le.IsConnected {
				return
			}
		}
	}()

	cmd.Wait()
	le.mutex.Lock()
	le.ExternalExec = false
	le.mutex.Unlock()

	return 0
}

// Winsize stores the Height and Width of a terminal.
type Winsize struct {
	Height uint16
	Width  uint16
	x      uint16 // unused
	y      uint16 // unused
}

func SetWinsize(fd uintptr, w, h int) {
	ws := &Winsize{Width: uint16(w), Height: uint16(h)}
	syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCSWINSZ), uintptr(unsafe.Pointer(ws)))
}
