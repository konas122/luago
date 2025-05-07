package codegen

import . "luago/compiler/ast"

func cgStat(f *funcInfo, node Stat) {
	switch stat := node.(type) {
	case *FuncCallStat:
		cgFuncCallStat(f, stat)
	case *BreakStat:
		cgBreakStat(f, stat)
	case *DoStat:
		cgDoStat(f, stat)
	case *WhileStat:
		cgWhileStat(f, stat)
	case *RepeatStat:
		cgRepeatStat(f, stat)
	case *IfStat:
		cgIfStat(f, stat)
	case *ForNumStat:
		cgForNumStat(f, stat)
	case *ForInStat:
		cgForInStat(f, stat)
	case *AssignStat:
		cgAssignStat(f, stat)
	case *LocalVarDeclStat:
		cgLocalVarDeclStat(f, stat)
	case *LocalFuncDefStat:
		cgLocalFuncDefStat(f, stat)
	case *LabelStat, *GotoStat:
		panic("label and goto statements are not supported!")
	}
}

func cgLocalFuncDefStat(f *funcInfo, node *LocalFuncDefStat) {
	r := f.addLocVar(node.Name, f.pc()+2)
	cgFuncDefExp(f, node.Exp, r)
}

func cgFuncCallStat(f *funcInfo, node *FuncCallStat) {
	r := f.allocReg()
	cgFuncCallExp(f, node, r, 0)
	f.freeReg()
}

func cgBreakStat(f *funcInfo, node *BreakStat) {
	pc := f.emitJmp(node.Line, 0, 0)
	f.addBreakJmp(pc)
}

func cgDoStat(f *funcInfo, node *DoStat) {
	f.enterScope(false)
	cgBlock(f, node.Block)
	f.closeOpenUpvals(node.Block.LastLine)
	f.exitScope(f.pc() + 1)
}

/*
        _________________
       |     false? jmp  |
       |                 |
while exp do block end <-'
      ^           \
      |___________/
           jmp
*/
func cgWhileStat(f *funcInfo, node *WhileStat) {
	pcBeforeExp := f.pc()

	oldRegs := f.usedRegs
	a, _ := expToOpArg(f, node.Exp, ARG_REG)
	f.usedRegs = oldRegs

	line := lastLineOf(node.Exp)
	f.emitTest(line, a, 0)
	pcJmpToEnd := f.emitJmp(line, 0, 0)

	f.enterScope(true)
	cgBlock(f, node.Block)
	f.closeOpenUpvals(node.Block.LastLine)
	f.emitJmp(node.Block.LastLine, 0, pcBeforeExp-f.pc()-1)
	f.exitScope(f.pc())

	f.fixSbx(pcJmpToEnd, f.pc()-pcJmpToEnd)
}

/*
        ______________
       |  false? jmp  |
       V              /
repeat block until exp
*/
func cgRepeatStat(f *funcInfo, node *RepeatStat) {
	f.enterScope(true)

	pcBeforeBlock := f.pc()
	cgBlock(f, node.Block)

	oldRegs := f.usedRegs
	a, _ := expToOpArg(f, node.Exp, ARG_REG)
	f.usedRegs = oldRegs

	line := lastLineOf(node.Exp)
	f.emitTest(line, a, 0)
	f.emitJmp(line, f.getJmpArgA(), pcBeforeBlock-f.pc()-1)
	f.closeOpenUpvals(line)

	f.exitScope(f.pc() + 1)
}

/*
         _________________       _________________       _____________
        / false? jmp      |     / false? jmp      |     / false? jmp  |
       /                  V    /                  V    /              V
if exp1 then block1 elseif exp2 then block2 elseif true then block3 end <-.
                   \                       \                       \      |
                    \_______________________\_______________________\_____|
                    jmp                     jmp                     jmp
*/
func cgIfStat(f *funcInfo, node *IfStat) {
	pcJmpToEnds := make([]int, len(node.Exps))
	pcJmpToNextExp := -1

	for i, exp := range node.Exps {
		if pcJmpToNextExp >= 0 {
			f.fixSbx(pcJmpToNextExp, f.pc()-pcJmpToNextExp)
		}

		oldRegs := f.usedRegs
		a, _ := expToOpArg(f, exp, ARG_REG)
		f.usedRegs = oldRegs

		line := lastLineOf(exp)
		f.emitTest(line, a, 0)
		pcJmpToNextExp = f.emitJmp(line, 0, 0)

		block := node.Blocks[i]
		f.enterScope(false)
		cgBlock(f, block)
		f.closeOpenUpvals(block.LastLine)
		f.exitScope(f.pc() + 1)
		if i < len(node.Exps)-1 {
			pcJmpToEnds[i] = f.emitJmp(block.LastLine, 0, 0)
		} else {
			pcJmpToEnds[i] = pcJmpToNextExp
		}
	}

	for _, pc := range pcJmpToEnds {
		f.fixSbx(pc, f.pc()-pc)
	}
}

func cgForNumStat(f *funcInfo, node *ForNumStat) {
	forIndexVar := "(for index)"
	forLimitVar := "(for limit)"
	forStepVar := "(for step)"

	f.enterScope(true)

	cgLocalVarDeclStat(f, &LocalVarDeclStat{
		NameList: []string{forIndexVar, forLimitVar, forStepVar},
		ExpList:  []Exp{node.InitExp, node.LimitExp, node.StepExp},
	})
	f.addLocVar(node.VarName, f.pc()+2)

	a := f.usedRegs - 4
	pcForPrep := f.emitForPrep(node.LineOfDo, a, 0)
	cgBlock(f, node.Block)
	f.closeOpenUpvals(node.Block.LastLine)
	pcForLoop := f.emitForLoop(node.LineOfFor, a, 0)

	f.fixSbx(pcForPrep, pcForLoop-pcForPrep-1)
	f.fixSbx(pcForLoop, pcForPrep-pcForLoop)

	f.exitScope(f.pc())
	f.fixEndPC(forIndexVar, 1)
	f.fixEndPC(forLimitVar, 1)
	f.fixEndPC(forStepVar, 1)
}

func cgForInStat(f *funcInfo, node *ForInStat) {
	forGeneratorVar := "(for generator)"
	forStateVar := "(for state)"
	forControlVar := "(for control)"

	f.enterScope(true)

	cgLocalVarDeclStat(f, &LocalVarDeclStat{
		//LastLine: 0,
		NameList: []string{forGeneratorVar, forStateVar, forControlVar},
		ExpList:  node.ExpList,
	})
	for _, name := range node.NameList {
		f.addLocVar(name, f.pc()+2)
	}

	pcJmpToTFC := f.emitJmp(node.LineOfDo, 0, 0)
	cgBlock(f, node.Block)
	f.closeOpenUpvals(node.Block.LastLine)
	f.fixSbx(pcJmpToTFC, f.pc()-pcJmpToTFC)

	line := lineOf(node.ExpList[0])
	rGenerator := f.slotOfLocVar(forGeneratorVar)
	f.emitTForCall(line, rGenerator, len(node.NameList))
	f.emitTForLoop(line, rGenerator+2, pcJmpToTFC-f.pc()-1)

	f.exitScope(f.pc() - 1)
	f.fixEndPC(forGeneratorVar, 2)
	f.fixEndPC(forStateVar, 2)
	f.fixEndPC(forControlVar, 2)
}

func cgLocalVarDeclStat(f *funcInfo, node *LocalVarDeclStat) {
	exps := removeTailNils(node.ExpList)
	nExps := len(exps)
	nNames := len(node.NameList)

	oldRegs := f.usedRegs
	if nExps == nNames {
		for _, exp := range exps {
			a := f.allocReg()
			cgExp(f, exp, a, 1)
		}
	} else if nExps > nNames {
		for i, exp := range exps {
			a := f.allocReg()
			if i == nExps-1 && isVarargOrFuncCall(exp) {
				cgExp(f, exp, a, 0)
			} else {
				cgExp(f, exp, a, 1)
			}
		}
	} else { // nNames > nExps
		multRet := false
		for i, exp := range exps {
			a := f.allocReg()
			if i == nExps-1 && isVarargOrFuncCall(exp) {
				multRet = true
				n := nNames - nExps + 1
				cgExp(f, exp, a, n)
				f.allocRegs(n - 1)
			} else {
				cgExp(f, exp, a, 1)
			}
		}
		if !multRet {
			n := nNames - nExps
			a := f.allocRegs(n)
			f.emitLoadNil(node.LastLine, a, n)
		}
	}

	f.usedRegs = oldRegs
	startPC := f.pc() + 1
	for _, name := range node.NameList {
		f.addLocVar(name, startPC)
	}
}

func cgAssignStat(f *funcInfo, node *AssignStat) {
	exps := removeTailNils(node.ExpList)
	nExps := len(exps)
	nVars := len(node.VarList)

	tRegs := make([]int, nVars)
	kRegs := make([]int, nVars)
	vRegs := make([]int, nVars)
	oldRegs := f.usedRegs

	for i, exp := range node.VarList {
		if taExp, ok := exp.(*TableAccessExp); ok {
			tRegs[i] = f.allocReg()
			cgExp(f, taExp.PrefixExp, tRegs[i], 1)
			kRegs[i] = f.allocReg()
			cgExp(f, taExp.KeyExp, kRegs[i], 1)
		} else {
			name := exp.(*NameExp).Name
			if f.slotOfLocVar(name) < 0 && f.indexOfUpval(name) < 0 {
				// global var
				kRegs[i] = -1
				if f.indexOfConstant(name) > 0xFF {
					kRegs[i] = f.allocReg()
				}
			}
		}
	}
	for i := 0; i < nVars; i++ {
		vRegs[i] = f.usedRegs + i
	}

	if nExps >= nVars {
		for i, exp := range exps {
			a := f.allocReg()
			if i >= nVars && i == nExps-1 && isVarargOrFuncCall(exp) {
				cgExp(f, exp, a, 0)
			} else {
				cgExp(f, exp, a, 1)
			}
		}
	} else { // nVars > nExps
		multRet := false
		for i, exp := range exps {
			a := f.allocReg()
			if i == nExps-1 && isVarargOrFuncCall(exp) {
				multRet = true
				n := nVars - nExps + 1
				cgExp(f, exp, a, n)
				f.allocRegs(n - 1)
			} else {
				cgExp(f, exp, a, 1)
			}
		}
		if !multRet {
			n := nVars - nExps
			a := f.allocRegs(n)
			f.emitLoadNil(node.LastLine, a, n)
		}
	}

	lastLine := node.LastLine
	for i, exp := range node.VarList {
		if nameExp, ok := exp.(*NameExp); ok {
			varName := nameExp.Name
			if a := f.slotOfLocVar(varName); a >= 0 {
				f.emitMove(lastLine, a, vRegs[i])
			} else if b := f.indexOfUpval(varName); b >= 0 {
				f.emitSetUpval(lastLine, vRegs[i], b)
			} else if a := f.slotOfLocVar("_ENV"); a >= 0 {
				if kRegs[i] < 0 {
					b := 0x100 + f.indexOfConstant(varName)
					f.emitSetTable(lastLine, a, b, vRegs[i])
				} else {
					f.emitSetTable(lastLine, a, kRegs[i], vRegs[i])
				}
			} else { // global var
				a := f.indexOfUpval("_ENV")
				if kRegs[i] < 0 {
					b := 0x100 + f.indexOfConstant(varName)
					f.emitSetTabUp(lastLine, a, b, vRegs[i])
				} else {
					f.emitSetTabUp(lastLine, a, kRegs[i], vRegs[i])
				}
			}
		} else {
			f.emitSetTable(lastLine, tRegs[i], kRegs[i], vRegs[i])
		}
	}

	// todo
	f.usedRegs = oldRegs
}
