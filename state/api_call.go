package state

import (
	. "luago/api"
	"luago/chunk"
	"luago/compiler"
	"luago/vm"
)

// [-0, +1, –]
// http://www.lua.org/manual/5.3/manual.html#lua_load
func (self *luaState) Load(binchunk []byte, chunkName, mode string) int {
	var proto *chunk.Prototype
	if chunk.IsBinaryChunk(binchunk) {
		proto = chunk.Undump(binchunk)
	} else {
		proto = compiler.Compile(string(binchunk), chunkName)
	}

	c := newLuaClosure(proto)
	self.stack.push(c)
	if len(proto.Upvalues) > 0 {
		env := self.registry.get(LUA_RIDX_GLOBALS)
		c.upvals[0] = &upvalue{&env}
	}
	return LUA_OK
}

// [-(nargs+1), +nresults, e]
// http://www.lua.org/manual/5.3/manual.html#lua_call
func (self *luaState) Call(nArgs, nResults int) {
	val := self.stack.get(-(nArgs + 1))
	c, ok := val.(*closure)

	if !ok {
		if mf := getMetafield(val, "__call", self); mf != nil {
			if c, ok = mf.(*closure); ok {
				self.stack.push(val)
				self.Insert(-(nArgs + 2))
				nArgs += 1
			}
		}
	}

	if ok {
		if c.proto != nil {
			self.callLuaClosure(nArgs, nResults, c)
		} else {
			self.callGoClosure(nArgs, nResults, c)
		}
	} else {
		panic("not a function")
	}
}

func (self *luaState) callGoClosure(nArgs, nResults int, c *closure) {
	newStack := newLuaStack(nArgs+LUA_MINSTACK, self)
	newStack.closure = c

	args := self.stack.popN(nArgs)
	newStack.pushN(args, nArgs)
	self.stack.pop()

	self.pushLuaStack(newStack)
	r := c.goFunc(self)
	self.popLuaStack()

	if nResults != 0 {
		results := newStack.popN(r)
		self.stack.check(len(results))
		self.stack.pushN(results, nResults)
	}
}

func (self *luaState) callLuaClosure(nArgs, nResults int, c *closure) {
	nRegs := int(c.proto.MaxStackSize)
	nParams := int(c.proto.NumParams)
	isVararg := c.proto.IsVararg == 1

	newStack := newLuaStack(nRegs+LUA_MINSTACK, self)
	newStack.closure = c

	funcAndArgs := self.stack.popN(nArgs + 1)
	newStack.pushN(funcAndArgs[1:], nParams)
	newStack.top = nRegs
	if nArgs > nParams && isVararg {
		newStack.varargs = funcAndArgs[nParams+1:]
	}
	// run closure
	self.pushLuaStack(newStack)
	self.runLuaClosure()
	self.popLuaStack()

	// return results
	if nResults != 0 {
		results := newStack.popN(newStack.top - nRegs)
		self.stack.check(len(results))
		self.stack.pushN(results, nResults)
	}
}

func (self *luaState) runLuaClosure() {
	for {
		inst := vm.Instruction(self.Fetch())
		inst.Execute(self)
		if inst.Opcode() == vm.OP_RETURN {
			break
		}
	}
}

// Calls a function in protected mode.
// http://www.lua.org/manual/5.3/manual.html#lua_pcallG
func (self *luaState) PCall(nArgs, nResults, msgh int) (status int) {
	caller := self.stack
	status = LUA_ERRRUN

	// catch error
	defer func() {
		if err := recover(); err != nil {
			if msgh != 0 {
				panic(err)
			}
			for self.stack != caller {
				self.popLuaStack()
			}
			self.stack.push(err)
		}
	}()

	self.Call(nArgs, nResults)
	status = LUA_OK
	return
}
