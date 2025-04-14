package state

import (
	"fmt"
	. "luago/api"
)

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_typename
func (lstate *luaState) TypeName(tp LuaType) string {
	switch tp {
	case LUA_TNONE:
		return "no value"
	case LUA_TNIL:
		return "nil"
	case LUA_TBOOLEAN:
		return "boolean"
	case LUA_TNUMBER:
		return "number"
	case LUA_TSTRING:
		return "string"
	case LUA_TTABLE:
		return "table"
	case LUA_TFUNCTION:
		return "function"
	case LUA_TTHREAD:
		return "thread"
	default:
		return "userdata"
	}
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_type
func (lstate *luaState) Type(idx int) LuaType {
	if lstate.stack.isValid(idx) {
		val := lstate.stack.get(idx)
		return typeOf(val)
	}
	return LUA_TNONE
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_isnone
func (lstate *luaState) IsNone(idx int) bool {
	return lstate.Type(idx) == LUA_TNONE
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_isnil
func (lstate *luaState) IsNil(idx int) bool {
	return lstate.Type(idx) == LUA_TNIL
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_isnoneornil
func (lstate *luaState) IsNoneOrNil(idx int) bool {
	return lstate.Type(idx) <= LUA_TNIL
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_isboolean
func (lstate *luaState) IsBoolean(idx int) bool {
	return lstate.Type(idx) == LUA_TBOOLEAN
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_istable
func (lstate *luaState) IsTable(idx int) bool {
	return lstate.Type(idx) == LUA_TTABLE
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_isfunction
func (lstate *luaState) IsFunction(idx int) bool {
	return lstate.Type(idx) == LUA_TFUNCTION
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_isthread
func (lstate *luaState) IsThread(idx int) bool {
	return lstate.Type(idx) == LUA_TTHREAD
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_isstring
func (lstate *luaState) IsString(idx int) bool {
	t := lstate.Type(idx)
	return t == LUA_TSTRING || t == LUA_TNUMBER
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_isnumber
func (lstate *luaState) IsNumber(idx int) bool {
	_, ok := lstate.ToNumberX(idx)
	return ok
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_isinteger
func (lstate *luaState) IsInteger(idx int) bool {
	val := lstate.stack.get(idx)
	_, ok := val.(int64)
	return ok
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_toboolean
func (lstate *luaState) ToBoolean(idx int) bool {
	val := lstate.stack.get(idx)
	return convertToBoolean(val)
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_tointeger
func (lstate *luaState) ToInteger(idx int) int64 {
	i, _ := lstate.ToIntegerX(idx)
	return i
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_tointegerx
func (lstate *luaState) ToIntegerX(idx int) (int64, bool) {
	val := lstate.stack.get(idx)
	i, ok := val.(int64)
	return i, ok
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_tonumber
func (lstate *luaState) ToNumber(idx int) float64 {
	n, _ := lstate.ToNumberX(idx)
	return n
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_tonumberx
func (lstate *luaState) ToNumberX(idx int) (float64, bool) {
	val := lstate.stack.get(idx)
	switch x := val.(type) {
	case float64:
		return x, true
	case int64:
		return float64(x), true
	default:
		return 0, false
	}
}

// [-0, +0, m]
// http://www.lua.org/manual/5.3/manual.html#lua_tostring
func (lstate *luaState) ToString(idx int) string {
	s, _ := lstate.ToStringX(idx)
	return s
}

func (lstate *luaState) ToStringX(idx int) (string, bool) {
	val := lstate.stack.get(idx)

	switch x := val.(type) {
	case string:
		return x, true
	case int64, float64:
		s := fmt.Sprintf("%v", x) // todo
		lstate.stack.set(idx, s)
		return s, true
	default:
		return "", false
	}
}
