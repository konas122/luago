package codegen

import . "luago/compiler/ast"

func cgBlock(f *funcInfo, node *Block) {
	for _, stat := range node.Stats {
		cgStat(f, stat)
	}
	if node.RetExps != nil {
		cgRetStat(f, node.RetExps, node.LastLine)
	}
}

func cgRetStat(f *funcInfo, exps []Exp, lastLine int) {
	nExps := len(exps)
	if nExps == 0 {
		f.emitReturn(lastLine, 0, 0)
		return
	}

	if nExps == 1 {
		if nameExp, ok := exps[0].(*NameExp); ok {
			if r := f.slotOfLocVar(nameExp.Name); r >= 0 {
				f.emitReturn(lastLine, r, 1)
				return
			}
		}
		if fcExp, ok := exps[0].(*FuncCallExp); ok {
			r := f.allocReg()
			cgTailCallExp(f, fcExp, r)
			f.freeReg()
			f.emitReturn(lastLine, r, -1)
			return
		}
	}

	multRet := isVarargOrFuncCall(exps[nExps-1])
	for i, exp := range exps {
		r := f.allocReg()
		if i == nExps-1 && multRet {
			cgExp(f, exp, r, -1)
		} else {
			cgExp(f, exp, r, 1)
		}
	}
	f.freeRegs(nExps)

	a := f.usedRegs // correct?
	if multRet {
		f.emitReturn(lastLine, a, -1)
	} else {
		f.emitReturn(lastLine, a, nExps)
	}
}
