package luaengine

import lua "github.com/yuin/gopher-lua"

func (le *LuaExtender) termLoader(L *lua.LState) int {
	var termAPI = map[string]lua.LGFunction{
		"write": le.write,
	}

	t := le.luaState.NewTable()
	le.luaState.SetFuncs(t, termAPI)
	le.luaState.Push(t)
	return 1
}
