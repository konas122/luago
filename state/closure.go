package state

import "luago/chunk"

type closure struct {
	proto *chunk.Prototype
}

func newLuaClosure(proto *chunk.Prototype) *closure {
	return &closure{proto: proto}
}
