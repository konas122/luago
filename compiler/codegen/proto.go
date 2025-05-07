package codegen

import . "luago/chunk"

func toProto(f *funcInfo) *Prototype {
	proto := &Prototype{
		LineDefined:     uint32(f.line),
		LastLineDefined: uint32(f.lastLine),
		NumParams:       byte(f.numParams),
		MaxStackSize:    byte(f.maxRegs),
		Code:            f.insts,
		Constants:       getConstants(f),
		Upvalues:        getUpvalues(f),
		Protos:          toProtos(f.subFuncs),
		LineInfo:        f.lineNums,
		LocVars:         getLocVars(f),
		UpvalueNames:    getUpvalueNames(f),
	}

	if f.line == 0 {
		proto.LastLineDefined = 0
	}
	if proto.MaxStackSize < 2 {
		proto.MaxStackSize = 2 // todo
	}
	if f.isVararg {
		proto.IsVararg = 1 // todo
	}

	return proto
}

func toProtos(fis []*funcInfo) []*Prototype {
	protos := make([]*Prototype, len(fis))
	for i, f := range fis {
		protos[i] = toProto(f)
	}
	return protos
}

func getConstants(f *funcInfo) []interface{} {
	consts := make([]interface{}, len(f.constants))
	for k, idx := range f.constants {
		consts[idx] = k
	}
	return consts
}

func getLocVars(f *funcInfo) []LocVar {
	locVars := make([]LocVar, len(f.locVars))
	for i, locVar := range f.locVars {
		locVars[i] = LocVar{
			VarName: locVar.name,
			StartPC: uint32(locVar.startPC),
			EndPC:   uint32(locVar.endPC),
		}
	}
	return locVars
}

func getUpvalues(f *funcInfo) []Upvalue {
	upvals := make([]Upvalue, len(f.upvalues))
	for _, uv := range f.upvalues {
		if uv.locVarSlot >= 0 { // instack
			upvals[uv.index] = Upvalue{Instack: 1, Idx: byte(uv.locVarSlot)}
		} else {
			upvals[uv.index] = Upvalue{Instack: 0, Idx: byte(uv.upvalIndex)}
		}
	}
	return upvals
}

func getUpvalueNames(f *funcInfo) []string {
	names := make([]string, len(f.upvalues))
	for name, uv := range f.upvalues {
		names[uv.index] = name
	}
	return names
}
