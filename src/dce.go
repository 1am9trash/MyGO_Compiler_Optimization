package main

func flattenBlocks(blocks [][]interface{}) []interface{} {
    var instrs []interface{}
    for _, block := range blocks {
        instrs = append(instrs, block...)
    }
    return instrs
}

func DropUnusedInstruction(function *Function) bool {
    changed := false
    used := make(map[string]bool)
    blocks := FormBlocks(function.Instrs)

    for _, block := range blocks {
        for _, instr := range block {
            args, ok := getInstructionValue(instr, "args")
            if !ok {
                continue
            }
            for _, arg := range args {
                used[arg] = true
            }
        }
    }

    for blockId, block := range blocks {
        var newBlock []interface{}
        for _, instr := range block {
            if nonLabelInstr, ok := instr.(Instruction); ok {
                if nonLabelInstr.Dest == "" || used[nonLabelInstr.Dest] {
                    newBlock = append(newBlock, nonLabelInstr)
                }
            } else {
                newBlock = append(newBlock, instr)
            }
        }
        if len(newBlock) < len(block) {
            changed = true
            blocks[blockId] = newBlock
        }
    }

    function.Instrs = flattenBlocks(blocks)

    return changed
}

func DropReassignment(function *Function) bool {
    changed := false
    drop := make(map[int]bool)
    blocks := FormBlocks(function.Instrs)

    for blockId, block := range blocks {
        lastDef := make(map[string]int)
        for instrId, instr := range block {
            if nonLabelInstr, ok := instr.(Instruction); ok {
                args, ok := getInstructionValue(nonLabelInstr, "args")
                if ok {
                    for _, arg := range args {
                        delete(lastDef, arg)
                    }
                }

                if nonLabelInstr.Dest != "" {
                    if dropId, ok := lastDef[nonLabelInstr.Dest]; ok {
                        changed = true
                        drop[blockId * len(blocks) + dropId] = true
                    }
                    lastDef[nonLabelInstr.Dest] = instrId
                }
            }
        }
    }

    var instrs []interface{}
    for blockId, block := range blocks {
        for instrId, instr := range block {
            if _, ok := drop[blockId * len(blocks) + instrId]; !ok {
                instrs = append(instrs, instr)
            } 
        }
    }
    function.Instrs = instrs

    return changed
}

func TrivalDCE(function *Function) {
    for change := true; change; change = DropReassignment(function) || DropUnusedInstruction(function) {}
}
