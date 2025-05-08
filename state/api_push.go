package state

import (
	"fmt"
	. "luago/api"
)

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

// [-0, +1, e]
// http://www.lua.org/manual/5.3/manual.html#lua_pushfstring
func (self *luaState) PushFString(fmtStr string, a ...interface{}) {
	str := fmt.Sprintf(fmtStr, a...)
	self.stack.push(str)
}

// [-0, +1, –]
// http://www.lua.org/manual/5.3/manual.html#lua_pushcfunction
func (self *luaState) PushGoFunction(f GoFunction) {
	self.stack.push(newGoClosure(f, 0))
}

// [-n, +1, m]
// http://www.lua.org/manual/5.3/manual.html#lua_pushcclosure
func (self *luaState) PushGoClosure(f GoFunction, n int) {
	closure := newGoClosure(f, n)
	for i := n; i > 0; i-- {
		val := self.stack.pop()
		closure.upvals[i-1] = &upvalue{&val}
	}
	self.stack.push(closure)
}

// [-0, +1, –]
// http://www.lua.org/manual/5.3/manual.html#lua_pushglobaltable
func (self *luaState) PushGlobalTable() {
	global := self.registry.get(LUA_RIDX_GLOBALS)
	self.stack.push(global)
}
