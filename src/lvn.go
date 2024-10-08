package main

import (
    "fmt"
    "strings"
)

type Expression struct {
    Op    string
    Args  []string
    Const interface{}
}

type LvnTableElement struct {
    Id    int
    Value string
}

func (exp Expression) Key() string {
    var sb strings.Builder
    sb.WriteString(fmt.Sprintf("%v", exp.Op))

    if len(exp.Args) > 0 {
        sb.WriteString(" " + strings.Join(exp.Args, " ") )
    }

    switch v := exp.Const.(type) {
    case bool, float64:
        sb.WriteString(fmt.Sprintf(" %v", v))
    }

    return sb.String()
}

func ReadFirst(instrs []interface{}, table map[string]LvnTableElement, num2var map[int]string, var2num map[string]int) {
    assigned := make(map[string]bool)
    for _, instr := range instrs {
        if nonLabelInstr, ok := instr.(Instruction); ok {
            if args, ok := getInstructionValue(nonLabelInstr, "args"); ok {
                for _, arg := range args {
                    if !assigned[arg] {
                        id := len(var2num) + 1
                        expression := Expression{
                            Op:   "haved_assigned",
                            Args: []string{arg},
                        }
                        table[expression.Key()] = LvnTableElement{
                            Id:    id,
                            Value: arg,
                        }
                        num2var[id] = arg
                        var2num[arg] = id
                    }
                }
                if nonLabelInstr.Dest != "" {
                    assigned[nonLabelInstr.Dest] = true
                }
            }
        }
    }
}

func LastWrite(instrs []interface{}) map[int]bool {
    lastWrite := make(map[int]bool)
    used := make(map[string]bool)
    for i := len(instrs) - 1; i >= 0; i-- {
        if nonLabelInstr, ok := instrs[i].(Instruction); ok {
            if nonLabelInstr.Dest != "" && !used[nonLabelInstr.Dest] {
                lastWrite[i] = true
                used[nonLabelInstr.Dest] = true
            }
        }
    }
    return lastWrite
}

func Id2Text(id int, num2var map[int]string) string {
    if id >= 0 {
        return fmt.Sprintf("%v", num2var[id])
    } else {
        return fmt.Sprintf("lvn#%v", -id)
    }
}

func intArray2StringArray(arr []int, num2var map[int]string) []string {
    strArr := make([]string, len(arr))
    for i, v := range arr {
        strArr[i] = Id2Text(v, num2var)
    }
    return strArr
}

func LvnBlock(block []interface{}) []interface{} {
    table := make(map[string]LvnTableElement)
    num2var := make(map[int]string)
    var2num := make(map[string]int)

    renamedMap := make(map[string]int)
    renamedCnt := 0

    ReadFirst(block, table, num2var, var2num)
    lastWrite := LastWrite(block)

    var newBlock []interface{}
    for instrId, instr := range block {
        switch v := instr.(type) {
        case Instruction:
            if args, ok := getInstructionValue(v, "args"); ok {
                renamedArgs := make([]int, len(args))
                for argId, arg := range args {
                    if renamedArg, ok := renamedMap[arg]; ok {
                        arg = Id2Text(renamedArg, num2var)
                    }
                    renamedArgs[argId] = var2num[arg]
                }

                newArgs := make([]string, len(renamedArgs))
                for argId, arg := range renamedArgs {
                    newArgs[argId] = Id2Text(arg, num2var)
                }

                newInstr := Instruction{
                    Op:     v.Op,
                    Dest:   v.Dest,
                    Args:   newArgs,
                    Funcs:  v.Funcs,
                    Labels: v.Labels,
                    Type:   v.Type,
                    Value:  v.Value,
                }

                if v.Dest != "" {
                    expression := Expression{
                        Op:    v.Op,
                        Args:  intArray2StringArray(renamedArgs, num2var),
                        Const: v.Value,
                    }

                    if lastWrite[instrId] {
                        newInstr.Dest = v.Dest
                        delete(renamedMap, v.Dest)
                    } else {
                        id := -(renamedCnt + 1)
                        renamedCnt += 1
                        renamedMap[v.Dest] = id
                        newInstr.Dest = Id2Text(id, num2var)
                    }

                    var destId int
                    if element, ok := table[expression.Key()]; ok {
                        newInstr.Op = "id"
                        newInstr.Args = []string{Id2Text(element.Id, num2var)}
                        newInstr.Value = nil
                        destId = element.Id
                    } else {
                        if v.Op == "id" {
                            destId = var2num[args[0]]
                        } else {
                            destId = len(var2num) + 1
                            table[expression.Key()] = LvnTableElement{
                                Id:    destId,
                                Value: v.Dest,
                            }
                        }
                    }

                    var2num[newInstr.Dest] = destId
                    if _, ok := num2var[destId]; !ok {
                        num2var[destId] = newInstr.Dest
                    }
                }
                newBlock = append(newBlock, newInstr)
            }
        case Label:
            newBlock = append(newBlock, v)
        }
    }

    return newBlock
}

func Lvn(function *Function) {
    blocks := FormBlocks(function.Instrs)
    var instrs []interface{}
    for _, block := range blocks {
        newBlock := LvnBlock(block)
        instrs = append(instrs, newBlock...)
    }
    function.Instrs = instrs
}
