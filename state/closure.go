package state

import (
	. "luago/api"
	"luago/chunk"
)

type closure struct {
	proto  *chunk.Prototype
	goFunc GoFunction
	upvals []*upvalue
}

type upvalue struct {
	val *luaValue
}

func newLuaClosure(proto *chunk.Prototype) *closure {
	c := &closure{proto: proto}
	if nUpvals := len(proto.Upvalues); nUpvals > 0 {
		c.upvals = make([]*upvalue, nUpvals)
	}
	return c
}

func newGoClosure(f GoFunction, nUpvals int) *closure {
	c := &closure{goFunc: f}
	if nUpvals > 0 {
		c.upvals = make([]*upvalue, nUpvals)
	}
	return c
}
