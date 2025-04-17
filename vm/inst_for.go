package vm

import . "luago/api"

// R(A)-=R(A+2); pc+=sBx
func forPrep(i Instruction, vm LuaVM) {
	a, sBx := i.AsBx()
	a += 1

	if vm.Type(a) == LUA_TSTRING {
		vm.PushNumber(vm.ToNumber(a))
		vm.Replace(a)
	}
	if vm.Type(a+1) == LUA_TSTRING {
		vm.PushNumber(vm.ToNumber(a + 1))
		vm.Replace(a + 1)
	}
	if vm.Type(a+2) == LUA_TSTRING {
		vm.PushNumber(vm.ToNumber(a + 2))
		vm.Replace(a + 2)
	}

	// R(A) -= R(A+2)
	vm.PushValue(a)
	vm.PushValue(a + 2)
	vm.Arith(LUA_OPSUB)
	vm.Replace(a)
	// pc += sBx
	vm.AddPC(sBx)
}

// R(A)+=R(A+2);
// if R(A) <?= R(A+1) then {
//   pc+=sBx; R(A+3)=R(A)
// }
func forLoop(i Instruction, vm LuaVM) {
	a, sBx := i.AsBx()
	a += 1

	// R(A) += R(A+2)
	vm.PushValue(a)
	vm.PushValue(a + 2)
	vm.Arith(LUA_OPADD)
	vm.Replace(a)
	// R(a) <?= R(A+1)
	isPositive := vm.ToNumber(a+2) >= 0
	if isPositive && vm.Compare(a, a+1, LUA_OPLE) ||
		!isPositive && vm.Compare(a+1, a, LUA_OPLE) {
		// pc += sBx
		vm.AddPC(sBx)
		// R(A+3) = R(A)
		vm.Copy(a, a+3)
	}
}
