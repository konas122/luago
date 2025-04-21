package state

type luaState struct {
	stack *luaStack
}

func New(stackSize ...int) *luaState {
	size := 20
	if len(stackSize) >= 1 {
		size = stackSize[0]
	}
	return &luaState{
		stack: newLuaStack(size),
	}
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
