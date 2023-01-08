package luaengine

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

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
	Sessions     *map[string]*database.User
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

// Winsize stores the Height and Width of a terminal.
type Winsize struct {
	Height uint16
	Width  uint16
	x      uint16 // unused
	y      uint16 // unused
}

// New creates a new instance of LuaExtender.
func New(cfg config.Config,
	Sessions *map[string]*database.User,
	user *database.User,
	term *term.Term,
	serverConn *ssh.ServerConn,
	conn ssh.Channel,
) *LuaExtender {

	le := &LuaExtender{
		Sessions:    Sessions,
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
	le.luaState.SetGlobal("exec", le.luaState.NewFunction(le.exec))
	le.luaState.SetGlobal("execWithTriggers", le.luaState.NewFunction(le.execWithTriggers))
	le.luaState.SetGlobal("execNonInteractive", le.luaState.NewFunction(le.execNonInteractive))
	le.luaState.SetGlobal("fileExists", le.luaState.NewFunction(le.fileExists))
	le.luaState.SetGlobal("getEnv", le.luaState.NewFunction(le.getEnv))
	le.luaState.SetGlobal("logf", le.luaState.NewFunction(le.logf))
	le.luaState.SetGlobal("pwd", le.luaState.NewFunction(le.pwd))
	le.luaState.SetGlobal("quit", le.luaState.NewFunction(le.quit))
	le.luaState.SetGlobal("rmTrigger", le.luaState.NewFunction(le.removeTrigger))
	le.luaState.SetGlobal("timer", le.luaState.NewFunction(le.timer))
	le.luaState.SetGlobal("trigger", le.luaState.NewFunction(le.trigger))
	le.luaState.SetGlobal("getUser", le.luaState.NewFunction(le.getUser))
	le.luaState.SetGlobal("hasGroup", le.luaState.NewFunction(le.hasGroup))
	le.luaState.SetGlobal("readFile", le.luaState.NewFunction(le.readFile))

	le.luaState.PreloadModule("term", le.termLoader)
	return le
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
	sessionID := fmt.Sprintf("%x", le.ServerConn.SessionID())
	delete(*le.Sessions, sessionID)
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
	name := l.ToString(1)
	args := make([]string, l.GetTop()-1)
	for i := 2; i <= l.GetTop(); i++ {
		args[i-2] = l.ToString(i)
	}

	le.ExternalExec = true
	npty, ntty, err := pty.Open()
	if err != nil {
		log.Printf("could not start pty (%s)", err)
	}

	cmd := exec.Command(name, args...)
	cmd.Stdout = ntty
	cmd.Stdin = ntty
	cmd.Stderr = ntty
	setCtrlTerm(cmd)

	if err := cmd.Start(); err != nil {
		cmdlog := name + " " + strings.Join(args, " ")
		log.Printf("failed to start %v (%s)", cmdlog, err)
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

			le.Term.WriteString(string(b[:n]))
			//_, err = le.Conn.Write(b[:n])
			//if err != nil {
			//	log.Printf("1 <- n: %v (%v)", n, err)
			//	return
			//}
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
			w = le.Term.Width
			h = le.Term.Height
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

func (le *LuaExtender) execWithTriggers(l *lua.LState) int {
	name := l.ToString(1)
	args := make([]string, l.GetTop()-1)
	for i := 2; i <= l.GetTop(); i++ {
		args[i-2] = l.ToString(i)
	}

	npty, ntty, err := pty.Open()
	if err != nil {
		log.Printf("could not start pty (%s)", err)
	}

	cmd := exec.Command(name, args...)
	cmd.Stdout = ntty
	cmd.Stdin = ntty
	cmd.Stderr = ntty
	setCtrlTerm(cmd)

	if err := cmd.Start(); err != nil {
		cmdlog := name + " " + strings.Join(args, " ")
		log.Printf("failed to start %v (%s)", cmdlog, err)
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

			le.Term.WriteString(string(b[:n]))
			//_, err = le.Conn.Write(b[:n])
			//if err != nil {
			//	log.Printf("1 <- n: %v (%v)", n, err)
			//	return
			//}
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

			k := string(b[:n])
			ok, err := le.RunTrigger(k)
			if err != nil {
				log.Println("error RunTrigger", err.Error())
				break
			}
			if !ok {
				le.Input(k)
			}

			if !le.IsConnected {
				return
			}
		}
	}()

	cmd.Wait()
	le.ExternalExec = false

	return 0
}

func (le *LuaExtender) execNonInteractive(l *lua.LState) int {
	var outb, errb bytes.Buffer
	name := l.ToString(1)
	args := make([]string, l.GetTop()-1)
	for i := 2; i <= l.GetTop(); i++ {
		args[i-2] = l.ToString(i)
	}

	cmd := exec.Command(name, args...)
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	//cmd.Stdin =

	if err := cmd.Start(); err != nil {
		cmdlog := name + " " + strings.Join(args, " ")
		log.Printf("failed to start %v (%s)", cmdlog, err)
		return 0
	}

	cmd.Wait()

	le.Term.WriteString(outb.String())
	if errb.String() != "" {
		cmdlog := name + " " + strings.Join(args, " ")
		log.Printf("exec %v: %v", cmdlog, errb.String())
	}

	le.ExternalExec = false

	l.Push(lua.LString(outb.String()))
	return 1
}

func (le *LuaExtender) readFile(l *lua.LState) int {
	file := l.ToString(1)
	content, err := os.ReadFile(file)
	if err != nil {
		log.Printf("error reading file %v, %v", file, err)
		return 0
	}
	l.Push(lua.LString(string(content)))
	return 1
}
