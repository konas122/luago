package state

import (
	. "luago/api"
	"luago/number"
	"math"
)

type operator struct {
	metamethod  string
	integerFunc func(int64, int64) int64
	floatFunc   func(float64, float64) float64
}

var (
	iadd  = func(a, b int64) int64 { return a + b }
	fadd  = func(a, b float64) float64 { return a + b }
	isub  = func(a, b int64) int64 { return a - b }
	fsub  = func(a, b float64) float64 { return a - b }
	imul  = func(a, b int64) int64 { return a * b }
	fmul  = func(a, b float64) float64 { return a * b }
	imod  = number.IMod
	fmod  = number.FMod
	pow   = math.Pow
	div   = func(a, b float64) float64 { return a / b }
	iidiv = number.IFloorDiv
	fidiv = number.FFloorDiv
	band  = func(a, b int64) int64 { return a & b }
	bor   = func(a, b int64) int64 { return a | b }
	bxor  = func(a, b int64) int64 { return a ^ b }
	shl   = number.ShiftLeft
	shr   = number.ShiftRight
	iunm  = func(a, _ int64) int64 { return -a }
	funm  = func(a, _ float64) float64 { return -a }
	bnot  = func(a, _ int64) int64 { return ^a }
)

var operators = []operator{
	{"__add", iadd, fadd},
	{"__sub", isub, fsub},
	{"__mul", imul, fmul},
	{"__mod", imod, fmod},
	{"__pow", nil, pow},
	{"__div", nil, div},
	{"__idiv", iidiv, fidiv},
	{"__band", band, nil},
	{"__bor", bor, nil},
	{"__bxor", bxor, nil},
	{"__shl", shl, nil},
	{"__shr", shr, nil},
	{"__unm", iunm, funm},
	{"__bnot", bnot, nil},
}

// [-(2|1), +1, e]
// http://www.lua.org/manual/5.3/manual.html#lua_arith
func (self *luaState) Arith(op ArithOp) {
	var a, b luaValue
	b = self.stack.pop()
	if op != LUA_OPUNM && op != LUA_OPBNOT {
		a = self.stack.pop()
	} else {
		a = b
	}

	operator := operators[op]
	if result := _arith(a, b, operator); result != nil {
		self.stack.push(result)
		return
	}

	mm := operator.metamethod
	if result, ok := callMetamethod(a, b, mm, self); ok {
		self.stack.push(result)
		return
	}

	panic("arithmetic error!")
}

func _arith(a, b luaValue, op operator) luaValue {
	if op.floatFunc == nil { // bitwise
		if x, ok := convertToInteger(a); ok {
			if y, ok := convertToInteger(b); ok {
				return op.integerFunc(x, y)
			}
		}
	} else {
		if op.integerFunc != nil {
			if x, ok := a.(int64); ok {
				if y, ok := b.(int64); ok {
					return op.integerFunc(x, y)
				}
			}
		}
		if x, ok := convertToFloat(a); ok {
			if y, ok := convertToFloat(b); ok {
				return op.floatFunc(x, y)
			}
		}
	}
	return nil
}
