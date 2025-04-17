package state

import "luago/chunk"

type luaState struct {
	stack *luaStack
	proto *chunk.Prototype
	pc    int
}

func New(stackSize int, proto *chunk.Prototype) *luaState {
	return &luaState{
		stack: newLuaStack(stackSize),
		proto: proto,
		pc:    0,
	}
}
