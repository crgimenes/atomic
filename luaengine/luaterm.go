package luaengine

import (
	"bufio"
	"encoding/base64"
	"io/ioutil"
	"log"
	"os"

	lua "github.com/yuin/gopher-lua"
)

func (le *LuaExtender) write(l *lua.LState) int {
	s := l.ToString(1)
	le.Term.WriteString(s)

	return 0
}

func (le *LuaExtender) writeFromASCII(l *lua.LState) int {
	s := l.ToString(1)
	le.Term.WriteFromASCII(s)

	return 0
}

func (le *LuaExtender) cls(l *lua.LState) int {
	err := le.Term.Clear()
	if err != nil {
		log.Println(err)
	}
	return 0
}

func (le *LuaExtender) setMaxInputLength(l *lua.LState) int {
	le.Term.MaxInputLength = int(l.ToNumber(1))
	return 0
}

func (le *LuaExtender) setEcho(l *lua.LState) int {
	b := l.ToBool(1)
	le.Term.SetEcho(b)
	return 0
}

func (le *LuaExtender) setInputLimit(l *lua.LState) int {
	i := l.ToInt(1)
	le.Term.SetInputLimit(i)
	return 0
}

func (le *LuaExtender) setOutputDelay(l *lua.LState) int {
	i := l.ToInt(1)
	le.Term.SetOutputDelay(i)
	return 0
}

func (le *LuaExtender) setOutputMode(l *lua.LState) int {
	s := l.ToString(1)
	le.Term.SetOutputMode(s)
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

func (le *LuaExtender) getField(l *lua.LState) int {
	res := lua.LString(le.Term.GetField())
	l.Push(res)
	return 1
}

func (le *LuaExtender) getOutputMode(l *lua.LState) int {
	res := lua.LString(le.Term.GetOutputDisplay())
	l.Push(res)
	return 1
}

func (le *LuaExtender) getPassword(l *lua.LState) int {
	res := lua.LString(le.Term.GetField())
	l.Push(res)
	return 1
}

func (le *LuaExtender) drawBox(l *lua.LState) int {
	x := l.ToInt(1)
	y := l.ToInt(2)
	w := l.ToInt(3)
	h := l.ToInt(4)
	le.Term.DrawBox(x, y, w, h)
	return 0
}

func (le *LuaExtender) getSize(l *lua.LState) int {
	w, h := le.Term.GetSize()
	res := lua.LNumber(w)
	l.Push(res)
	res = lua.LNumber(h)
	l.Push(res)
	return 2
}

func (le *LuaExtender) print(l *lua.LState) int {
	col := l.ToInt(1)
	lin := l.ToInt(2)
	s := l.ToString(3)
	le.Term.Print(lin, col, s)
	return 0
}

func (le *LuaExtender) termLoader(L *lua.LState) int {
	var termAPI = map[string]lua.LGFunction{
		"cls":                  le.cls,
		"drawBox":              le.drawBox,
		"inlineImagesProtocol": le.inlineImagesProtocol,
		"getField":             le.getField,
		"getPassword":          le.getPassword,
		"getOutputMode":        le.getOutputMode,
		"getSize":              le.getSize,
		"print":                le.print,
		"setMaxInputLength":    le.setMaxInputLength,
		"setEcho":              le.setEcho,
		"setInputLimit":        le.setInputLimit,
		"setOutputDelay":       le.setOutputDelay,
		"setOutputMode":        le.setOutputMode,
		"write":                le.write,
		"writeFromASCII":       le.writeFromASCII,
	}

	t := le.luaState.NewTable()
	le.luaState.SetFuncs(t, termAPI)
	le.luaState.Push(t)
	return 1
}
