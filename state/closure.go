package state

import (
	. "luago/api"
	"luago/chunk"
)

type closure struct {
	proto  *chunk.Prototype
	goFunc GoFunction
}

func newLuaClosure(proto *chunk.Prototype) *closure {
	return &closure{proto: proto}
}

func newGoClosure(f GoFunction) *closure {
	return &closure{goFunc: f}
}
