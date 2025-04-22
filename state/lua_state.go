package state

import . "luago/api"

type luaState struct {
	registry *luaTable
	stack    *luaStack
}

func New(stackSize ...int) *luaState {
	size := LUA_MINSTACK
	if len(stackSize) >= 1 {
		size = max(stackSize[0], size)
	}
	registry := newLuaTable(0, 0)
	registry.put(LUA_RIDX_GLOBALS, newLuaTable(0, 0))

	ls := &luaState{registry: registry}
	ls.pushLuaStack(newLuaStack(size, ls))
	return ls
}

func (self *luaState) pushLuaStack(stack *luaStack) {
	stack.prev = self.stack
	self.stack = stack
}

func (self *luaState) popLuaStack() {
	stack := self.stack
	self.stack = stack.prev
	stack.prev = nil
}
