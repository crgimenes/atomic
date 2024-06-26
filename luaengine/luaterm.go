package luaengine

import (
	"bufio"
	"encoding/base64"
	"io"
	"log"
	"os"
	"strings"

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
	content, err := io.ReadAll(reader)
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
	row := l.ToInt(1)
	col := l.ToInt(2)
	width := l.ToInt(3)
	height := l.ToInt(4)
	le.Term.DrawBox(row, col, width, height)
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
	row := l.ToInt(1)
	col := l.ToInt(2)
	s := l.ToString(3)
	le.Term.Print(row, col, s)
	return 0
}

func (le *LuaExtender) moveCursor(l *lua.LState) int {
	row := l.ToInt(1)
	col := l.ToInt(2)
	le.Term.MoveCursor(row, col)
	return 0
}

/*
ResetScreen()
EnterScreen()
ExitScreen()
SetOutputMode(mode string)
SetInputLimit(limit int)
SetOutputDelay(delay int)
DrawBox(row, col, width, height int)
GetSize() (int, int)
SetColor(color string)
SetBackgroundColor(color string)
SetColorRGB(r, g, b int)
SetBackgroundColorRGB(r, g, b int)
SetBold()
SetUnderline()
SetBlink()
SetReverse()
SetInvisible()
Reset()
SetCursorVisible(visible bool)
*/

func (le *LuaExtender) resetScreen(l *lua.LState) int {
	le.Term.ResetScreen()
	return 0
}

func (le *LuaExtender) enterScreen(l *lua.LState) int {
	le.Term.EnterScreen()
	return 0
}

func (le *LuaExtender) exitScreen(l *lua.LState) int {
	le.Term.ExitScreen()
	return 0
}

func (le *LuaExtender) printMultipleLines(l *lua.LState) int {
	row := l.ToInt(1)
	col := l.ToInt(2)
	s := l.ToString(3)
	s = strings.Replace(s, "\r", "", -1)
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		le.Term.MoveCursor(row+i, col)
		le.Term.WriteString(line)
	}
	return 0
}

func (le *LuaExtender) termLoader(L *lua.LState) int {
	var termAPI = map[string]lua.LGFunction{
		"cls":                  le.cls,
		"drawBox":              le.drawBox,
		"enterScreen":          le.enterScreen,
		"exitScreen":           le.exitScreen,
		"getField":             le.getField,
		"getOutputMode":        le.getOutputMode,
		"getPassword":          le.getPassword,
		"getSize":              le.getSize,
		"inlineImagesProtocol": le.inlineImagesProtocol,
		"moveCursor":           le.moveCursor,
		"print":                le.print,
		"resetScreen":          le.resetScreen,
		"setEcho":              le.setEcho,
		"setInputLimit":        le.setInputLimit,
		"setMaxInputLength":    le.setMaxInputLength,
		"setOutputDelay":       le.setOutputDelay,
		"setOutputMode":        le.setOutputMode,
		"write":                le.write,
		"writeFromASCII":       le.writeFromASCII,
		"printMultipleLines":   le.printMultipleLines,
	}

	t := le.luaState.NewTable()
	le.luaState.SetFuncs(t, termAPI)
	le.luaState.Push(t)
	return 1
}
