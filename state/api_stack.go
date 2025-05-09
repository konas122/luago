package state

import . "luago/api"

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_gettop
func (lstate *luaState) GetTop() int {
	return lstate.stack.top
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_absindex
func (lstate *luaState) AbsIndex(idx int) int {
	return lstate.stack.absIndex(idx)
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_checkstack
func (lstate *luaState) CheckStack(n int) bool {
	lstate.stack.check(n)
	return true // never fails
}

// [-n, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_pop
func (lstate *luaState) Pop(n int) {
	for i := 0; i < n; i++ {
		lstate.stack.pop()
	}
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_copy
func (lstate *luaState) Copy(fromIdx, toIdx int) {
	val := lstate.stack.get(fromIdx)
	lstate.stack.set(toIdx, val)
}

// [-0, +1, –]
// http://www.lua.org/manual/5.3/manual.html#lua_pushvalue
func (lstate *luaState) PushValue(idx int) {
	val := lstate.stack.get(idx)
	lstate.stack.push(val)
}

// [-1, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_replace
func (lstate *luaState) Replace(idx int) {
	val := lstate.stack.pop()
	lstate.stack.set(idx, val)
}

// [-1, +1, –]
// http://www.lua.org/manual/5.3/manual.html#lua_insert
func (lstate *luaState) Insert(idx int) {
	lstate.Rotate(idx, 1)
}

// [-1, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_remove
func (lstate *luaState) Remove(idx int) {
	lstate.Rotate(idx, -1)
	lstate.Pop(1)
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_rotate
func (lstate *luaState) Rotate(idx, n int) {
	t := lstate.stack.top - 1           /* end of stack segment being rotated */
	p := lstate.stack.absIndex(idx) - 1 /* start of segment */
	var m int                           /* end of prefix */
	if n >= 0 {
		m = t - n
	} else {
		m = p - n - 1
	}
	lstate.stack.reverse(p, m)   /* reverse the prefix with length 'n' */
	lstate.stack.reverse(m+1, t) /* reverse the suffix */
	lstate.stack.reverse(p, t)   /* reverse the entire segment */
}

// [-?, +?, –]
// http://www.lua.org/manual/5.3/manual.html#lua_settop
func (lstate *luaState) SetTop(idx int) {
	newTop := lstate.stack.absIndex(idx)
	if newTop < 0 {
		panic("stack underflow!")
	}

	n := lstate.stack.top - newTop
	if n > 0 {
		for i := 0; i < n; i++ {
			lstate.stack.pop()
		}
	} else if n < 0 {
		for i := 0; i > n; i-- {
			lstate.stack.push(nil)
		}
	}
}

// [-?, +?, –]
// http://www.lua.org/manual/5.3/manual.html#lua_xmove
func (self *luaState) XMove(to LuaState, n int) {
	vals := self.stack.popN(n)
	to.(*luaState).stack.pushN(vals, n)
}
