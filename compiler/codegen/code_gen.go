package codegen

import (
	. "luago/chunk"
	. "luago/compiler/ast"
)

func GenProto(chunk *Block) *Prototype {
	fd := &FuncDefExp{
		LastLine: chunk.LastLine,
		IsVararg: true,
		Block:    chunk,
	}

	f := newFuncInfo(nil, fd)
	f.addLocVar("_ENV", 0)
	cgFuncDefExp(f, fd, 0)
	return toProto(f.subFuncs[0])
}
