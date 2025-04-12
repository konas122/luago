package vm

const MAXARG_Bx = 1<<18 - 1       // 262143
const MAXARG_sBx = MAXARG_Bx >> 1 // 131071

/*
31        22       13      5    0
+--------+^-------+^------+^-----
|b=9bits |c=9bits |a=8bits| op=6|
+--------+^-------+^------+^-----
|    bx=18bits    |a=8bits| op=6|
+--------+^-------+^------+^-----
|   sbx=18bits    |a=8bits| op=6|
+--------+^-------+^------+^-----
|    ax=26bits            | op=6|
+--------+^-------+^------+^-----
31      23      15        7    0
*/

type Instruction uint32

func (ins Instruction) Opcode() int {
	return int(ins & 0x3F)
}

func (ins Instruction) ABC() (a, b, c int) {
	a = int(ins >> 6 & 0xFF)
	c = int(ins >> 14 & 0x1FF)
	b = int(ins >> 23 & 0x1FF)
	return
}

func (ins Instruction) ABx() (a, bx int) {
	a = int(ins >> 6 & 0xFF)
	bx = int(ins >> 14)
	return
}

func (ins Instruction) AsBx() (a, sbx int) {
	a, bx := ins.ABx()
	return a, bx - MAXARG_sBx
}

func (ins Instruction) Ax() int {
	return int(ins >> 6)
}

func (ins Instruction) OpName() string {
	return opcodes[ins.Opcode()].name
}

func (ins Instruction) OpMode() byte {
	return opcodes[ins.Opcode()].opMode
}

func (ins Instruction) BMode() byte {
	return opcodes[ins.Opcode()].argBMode
}

func (ins Instruction) CMode() byte {
	return opcodes[ins.Opcode()].argCMode
}
