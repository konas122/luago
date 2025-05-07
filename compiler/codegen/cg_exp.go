package codegen

import (
	. "luago/compiler/ast"
	. "luago/compiler/lexer"
	. "luago/vm"
)

// kind of operands
const (
	ARG_CONST = 1 // const index
	ARG_REG   = 2 // register index
	ARG_UPVAL = 4 // upvalue index
	ARG_RK    = ARG_REG | ARG_CONST
	ARG_RU    = ARG_REG | ARG_UPVAL
	ARG_RUK   = ARG_REG | ARG_UPVAL | ARG_CONST
)

// todo: rename to evalExp()?
func cgExp(f *funcInfo, node Exp, a, n int) {
	switch exp := node.(type) {
	case *NilExp:
		f.emitLoadNil(exp.Line, a, n)
	case *FalseExp:
		f.emitLoadBool(exp.Line, a, 0, 0)
	case *TrueExp:
		f.emitLoadBool(exp.Line, a, 1, 0)
	case *IntegerExp:
		f.emitLoadK(exp.Line, a, exp.Val)
	case *FloatExp:
		f.emitLoadK(exp.Line, a, exp.Val)
	case *StringExp:
		f.emitLoadK(exp.Line, a, exp.Str)
	case *ParensExp:
		cgExp(f, exp.Exp, a, 1)
	case *VarargExp:
		cgVarargExp(f, exp, a, n)
	case *FuncDefExp:
		cgFuncDefExp(f, exp, a)
	case *TableConstructorExp:
		cgTableConstructorExp(f, exp, a)
	case *UnopExp:
		cgUnopExp(f, exp, a)
	case *BinopExp:
		cgBinopExp(f, exp, a)
	case *ConcatExp:
		cgConcatExp(f, exp, a)
	case *NameExp:
		cgNameExp(f, exp, a)
	case *TableAccessExp:
		cgTableAccessExp(f, exp, a)
	case *FuncCallExp:
		cgFuncCallExp(f, exp, a, n)
	}
}

func cgVarargExp(f *funcInfo, node *VarargExp, a, n int) {
	if !f.isVararg {
		panic("cannot use '...' outside a vararg function")
	}
	f.emitVararg(node.Line, a, n)
}

// f[a] := function(args) body end
func cgFuncDefExp(f *funcInfo, node *FuncDefExp, a int) {
	subFI := newFuncInfo(f, node)
	f.subFuncs = append(f.subFuncs, subFI)

	for _, param := range node.ParList {
		subFI.addLocVar(param, 0)
	}

	cgBlock(subFI, node.Block)
	subFI.exitScope(subFI.pc() + 2)
	subFI.emitReturn(node.LastLine, 0, 0)

	bx := len(f.subFuncs) - 1
	f.emitClosure(node.LastLine, a, bx)
}

func cgTableConstructorExp(f *funcInfo, node *TableConstructorExp, a int) {
	nArr := 0
	for _, keyExp := range node.KeyExps {
		if keyExp == nil {
			nArr++
		}
	}
	nExps := len(node.KeyExps)
	multRet := nExps > 0 && isVarargOrFuncCall(node.ValExps[nExps-1])

	f.emitNewTable(node.Line, a, nArr, nExps-nArr)

	arrIdx := 0
	for i, keyExp := range node.KeyExps {
		valExp := node.ValExps[i]

		if keyExp == nil {
			arrIdx++
			tmp := f.allocReg()
			if i == nExps-1 && multRet {
				cgExp(f, valExp, tmp, -1)
			} else {
				cgExp(f, valExp, tmp, 1)
			}

			if arrIdx%50 == 0 || arrIdx == nArr { // LFIELDS_PER_FLUSH
				n := arrIdx % 50
				if n == 0 {
					n = 50
				}
				f.freeRegs(n)
				line := lastLineOf(valExp)
				c := (arrIdx-1)/50 + 1 // todo: c > 0xFF
				if i == nExps-1 && multRet {
					f.emitSetList(line, a, 0, c)
				} else {
					f.emitSetList(line, a, n, c)
				}
			}

			continue
		}

		b := f.allocReg()
		cgExp(f, keyExp, b, 1)
		c := f.allocReg()
		cgExp(f, valExp, c, 1)
		f.freeRegs(2)

		line := lastLineOf(valExp)
		f.emitSetTable(line, a, b, c)
	}
}

// r[a] := op exp
func cgUnopExp(f *funcInfo, node *UnopExp, a int) {
	oldRegs := f.usedRegs
	b, _ := expToOpArg(f, node.Exp, ARG_REG)
	f.emitUnaryOp(node.Line, node.Op, a, b)
	f.usedRegs = oldRegs
}

// r[a] := exp1 op exp2
func cgBinopExp(f *funcInfo, node *BinopExp, a int) {
	switch node.Op {
	case TOKEN_OP_AND, TOKEN_OP_OR:
		oldRegs := f.usedRegs

		b, _ := expToOpArg(f, node.Exp1, ARG_REG)
		f.usedRegs = oldRegs
		if node.Op == TOKEN_OP_AND {
			f.emitTestSet(node.Line, a, b, 0)
		} else {
			f.emitTestSet(node.Line, a, b, 1)
		}
		pcOfJmp := f.emitJmp(node.Line, 0, 0)

		b, _ = expToOpArg(f, node.Exp2, ARG_REG)
		f.usedRegs = oldRegs
		f.emitMove(node.Line, a, b)
		f.fixSbx(pcOfJmp, f.pc()-pcOfJmp)

	default:
		oldRegs := f.usedRegs
		b, _ := expToOpArg(f, node.Exp1, ARG_RK)
		c, _ := expToOpArg(f, node.Exp2, ARG_RK)
		f.emitBinaryOp(node.Line, node.Op, a, b, c)
		f.usedRegs = oldRegs
	}
}

// r[a] := exp1 .. exp2
func cgConcatExp(f *funcInfo, node *ConcatExp, a int) {
	for _, subExp := range node.Exps {
		a := f.allocReg()
		cgExp(f, subExp, a, 1)
	}

	c := f.usedRegs - 1
	b := c - len(node.Exps) + 1
	f.freeRegs(c - b + 1)
	f.emitABC(node.Line, OP_CONCAT, a, b, c)
}

// r[a] := name
func cgNameExp(f *funcInfo, node *NameExp, a int) {
	if r := f.slotOfLocVar(node.Name); r >= 0 {
		f.emitMove(node.Line, a, r)
	} else if idx := f.indexOfUpval(node.Name); idx >= 0 {
		f.emitGetUpval(node.Line, a, idx)
	} else { // x => _ENV['x']
		taExp := &TableAccessExp{
			LastLine:  node.Line,
			PrefixExp: &NameExp{Line: node.Line, Name: "_ENV"},
			KeyExp:    &StringExp{Line: node.Line, Str: node.Name},
		}
		cgTableAccessExp(f, taExp, a)
	}
}

// r[a] := prefix[key]
func cgTableAccessExp(f *funcInfo, node *TableAccessExp, a int) {
	oldRegs := f.usedRegs
	b, kindB := expToOpArg(f, node.PrefixExp, ARG_RU)
	c, _ := expToOpArg(f, node.KeyExp, ARG_RK)
	f.usedRegs = oldRegs

	if kindB == ARG_UPVAL {
		f.emitGetTabUp(node.LastLine, a, b, c)
	} else {
		f.emitGetTable(node.LastLine, a, b, c)
	}
}

// r[a] := f(args)
func cgFuncCallExp(f *funcInfo, node *FuncCallExp, a, n int) {
	nArgs := prepFuncCall(f, node, a)
	f.emitCall(node.Line, a, nArgs, n)
}

// return f(args)
func cgTailCallExp(f *funcInfo, node *FuncCallExp, a int) {
	nArgs := prepFuncCall(f, node, a)
	f.emitTailCall(node.Line, a, nArgs)
}

func prepFuncCall(f *funcInfo, node *FuncCallExp, a int) int {
	nArgs := len(node.Args)
	lastArgIsVarargOrFuncCall := false

	cgExp(f, node.PrefixExp, a, 1)
	if node.NameExp != nil {
		f.allocReg()
		c, k := expToOpArg(f, node.NameExp, ARG_RK)
		f.emitSelf(node.Line, a, a, c)
		if k == ARG_REG {
			f.freeRegs(1)
		}
	}
	for i, arg := range node.Args {
		tmp := f.allocReg()
		if i == nArgs-1 && isVarargOrFuncCall(arg) {
			lastArgIsVarargOrFuncCall = true
			cgExp(f, arg, tmp, -1)
		} else {
			cgExp(f, arg, tmp, 1)
		}
	}
	f.freeRegs(nArgs)

	if node.NameExp != nil {
		f.freeReg()
		nArgs++
	}
	if lastArgIsVarargOrFuncCall {
		nArgs = -1
	}

	return nArgs
}

func expToOpArg(f *funcInfo, node Exp, argKinds int) (arg, argKind int) {
	if argKinds&ARG_CONST > 0 {
		idx := -1
		switch x := node.(type) {
		case *NilExp:
			idx = f.indexOfConstant(nil)
		case *FalseExp:
			idx = f.indexOfConstant(false)
		case *TrueExp:
			idx = f.indexOfConstant(true)
		case *IntegerExp:
			idx = f.indexOfConstant(x.Val)
		case *FloatExp:
			idx = f.indexOfConstant(x.Val)
		case *StringExp:
			idx = f.indexOfConstant(x.Str)
		}
		if idx >= 0 && idx <= 0xFF {
			return 0x100 + idx, ARG_CONST
		}
	}

	if nameExp, ok := node.(*NameExp); ok {
		if argKinds&ARG_REG > 0 {
			if r := f.slotOfLocVar(nameExp.Name); r >= 0 {
				return r, ARG_REG
			}
		}
		if argKinds&ARG_UPVAL > 0 {
			if idx := f.indexOfUpval(nameExp.Name); idx >= 0 {
				return idx, ARG_UPVAL
			}
		}
	}

	a := f.allocReg()
	cgExp(f, node, a, 1)
	return a, ARG_REG
}
