package parser

import (
	. "luago/compiler/ast"
	. "luago/compiler/lexer"
	"luago/number"
	"math"
)

func optimizeLogicalOr(exp *BinopExp) Exp {
	if isTrue(exp.Exp1) {
		return exp.Exp1 // true or x => true
	}
	if isFalse(exp.Exp1) && !isVarargOrFuncCall(exp.Exp2) {
		return exp.Exp2 // false or x => x
	}
	return exp
}

func optimizeLogicalAnd(exp *BinopExp) Exp {
	if isFalse(exp.Exp1) {
		return exp.Exp1 // false and x => false
	}
	if isTrue(exp.Exp1) && !isVarargOrFuncCall(exp.Exp2) {
		return exp.Exp2 // true and x => x
	}
	return exp
}

func optimizeBitwiseBinaryOp(exp *BinopExp) Exp {
	if i, ok := castToInt(exp.Exp1); ok {
		if j, ok := castToInt(exp.Exp2); ok {
			switch exp.Op {
			case TOKEN_OP_BAND:
				return &IntegerExp{Line: exp.Line, Val: i & j}
			case TOKEN_OP_BOR:
				return &IntegerExp{Line: exp.Line, Val: i | j}
			case TOKEN_OP_BXOR:
				return &IntegerExp{Line: exp.Line, Val: i ^ j}
			case TOKEN_OP_SHL:
				return &IntegerExp{Line: exp.Line, Val: number.ShiftLeft(i, j)}
			case TOKEN_OP_SHR:
				return &IntegerExp{Line: exp.Line, Val: number.ShiftRight(i, j)}
			}
		}
	}
	return exp
}

func optimizeArithBinaryOp(exp *BinopExp) Exp {
	if x, ok := exp.Exp1.(*IntegerExp); ok {
		if y, ok := exp.Exp2.(*IntegerExp); ok {
			switch exp.Op {
			case TOKEN_OP_ADD:
				return &IntegerExp{Line: exp.Line, Val: x.Val + y.Val}
			case TOKEN_OP_SUB:
				return &IntegerExp{Line: exp.Line, Val: x.Val - y.Val}
			case TOKEN_OP_MUL:
				return &IntegerExp{Line: exp.Line, Val: x.Val * y.Val}
			case TOKEN_OP_IDIV:
				if y.Val != 0 {
					return &IntegerExp{Line: exp.Line, Val: number.IFloorDiv(x.Val, y.Val)}
				}
			case TOKEN_OP_MOD:
				if y.Val != 0 {
					return &IntegerExp{Line: exp.Line, Val: number.IMod(x.Val, y.Val)}
				}
			}
		}
	}
	if f, ok := castToFloat(exp.Exp1); ok {
		if g, ok := castToFloat(exp.Exp2); ok {
			switch exp.Op {
			case TOKEN_OP_ADD:
				return &FloatExp{Line: exp.Line, Val: f + g}
			case TOKEN_OP_SUB:
				return &FloatExp{Line: exp.Line, Val: f - g}
			case TOKEN_OP_MUL:
				return &FloatExp{Line: exp.Line, Val: f * g}
			case TOKEN_OP_DIV:
				if g != 0 {
					return &FloatExp{Line: exp.Line, Val: f / g}
				}
			case TOKEN_OP_IDIV:
				if g != 0 {
					return &FloatExp{Line: exp.Line, Val: number.FFloorDiv(f, g)}
				}
			case TOKEN_OP_MOD:
				if g != 0 {
					return &FloatExp{Line: exp.Line, Val: number.FMod(f, g)}
				}
			case TOKEN_OP_POW:
				return &FloatExp{Line: exp.Line, Val: math.Pow(f, g)}
			}
		}
	}
	return exp
}

func optimizePow(exp Exp) Exp {
	if binop, ok := exp.(*BinopExp); ok {
		if binop.Op == TOKEN_OP_POW {
			binop.Exp2 = optimizePow(binop.Exp2)
		}
		return optimizeArithBinaryOp(binop)
	}
	return exp
}

func optimizeUnaryOp(exp *UnopExp) Exp {
	switch exp.Op {
	case TOKEN_OP_UNM:
		return optimizeUnm(exp)
	case TOKEN_OP_NOT:
		return optimizeNot(exp)
	case TOKEN_OP_BNOT:
		return optimizeBnot(exp)
	default:
		return exp
	}
}

func optimizeUnm(exp *UnopExp) Exp {
	switch x := exp.Exp.(type) { // number?
	case *IntegerExp:
		x.Val = -x.Val
		return x
	case *FloatExp:
		if x.Val != 0 {
			x.Val = -x.Val
			return x
		}
	}
	return exp
}

func optimizeNot(exp *UnopExp) Exp {
	switch exp.Exp.(type) {
	case *NilExp, *FalseExp: // false
		return &TrueExp{Line: exp.Line}
	case *TrueExp, *IntegerExp, *FloatExp, *StringExp: // true
		return &FalseExp{Line: exp.Line}
	default:
		return exp
	}
}

func optimizeBnot(exp *UnopExp) Exp {
	switch x := exp.Exp.(type) { // number?
	case *IntegerExp:
		x.Val = ^x.Val
		return x
	case *FloatExp:
		if i, ok := number.FloatToInteger(x.Val); ok {
			return &IntegerExp{Line: x.Line, Val: ^i}
		}
	}
	return exp
}

func isFalse(exp Exp) bool {
	switch exp.(type) {
	case *FalseExp, *NilExp:
		return true
	default:
		return false
	}
}

func isTrue(exp Exp) bool {
	switch exp.(type) {
	case *TrueExp, *IntegerExp, *FloatExp, *StringExp:
		return true
	default:
		return false
	}
}

// todo
func isVarargOrFuncCall(exp Exp) bool {
	switch exp.(type) {
	case *VarargExp, *FuncCallExp:
		return true
	}
	return false
}

func castToInt(exp Exp) (int64, bool) {
	switch x := exp.(type) {
	case *IntegerExp:
		return x.Val, true
	case *FloatExp:
		return number.FloatToInteger(x.Val)
	default:
		return 0, false
	}
}

func castToFloat(exp Exp) (float64, bool) {
	switch x := exp.(type) {
	case *IntegerExp:
		return float64(x.Val), true
	case *FloatExp:
		return x.Val, true
	default:
		return 0, false
	}
}
