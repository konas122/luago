package state

// [-0, +1, –]
// http://www.lua.org/manual/5.3/manual.html#lua_pushnil
func (lstate *luaState) PushNil() {
	lstate.stack.push(nil)
}

// [-0, +1, –]
// http://www.lua.org/manual/5.3/manual.html#lua_pushboolean
func (lstate *luaState) PushBoolean(b bool) {
	lstate.stack.push(b)
}

// [-0, +1, –]
// http://www.lua.org/manual/5.3/manual.html#lua_pushinteger
func (lstate *luaState) PushInteger(n int64) {
	lstate.stack.push(n)
}

// [-0, +1, –]
// http://www.lua.org/manual/5.3/manual.html#lua_pushnumber
func (lstate *luaState) PushNumber(n float64) {
	lstate.stack.push(n)
}

// [-0, +1, m]
// http://www.lua.org/manual/5.3/manual.html#lua_pushstring
func (lstate *luaState) PushString(s string) {
	lstate.stack.push(s)
}
