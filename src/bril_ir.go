package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
    "strings"
)

type Label struct {
    Label string  `json:"label"`
}

type Instruction struct {
    Op     string      `json:"op"`
    Dest   string      `json:"dest,omitempty"`
    Args   []string    `json:"args,omitempty"`
    Funcs  []string    `json:"funcs,omitempty"`
    Labels []string    `json:"labels,omitempty"`
    Type   string      `json:"type,omitempty"`
    Value  interface{} `json:"value,omitempty"`
}

type Function struct {
    Instrs []interface{} `json:"instrs"`
    Name   string        `json:"name"`
    Args   []Argument    `json:"args,omitempty"`
    Type   string        `json:"type,omitempty"`
}

type Argument struct {
    Name string `json:"name"`
    Type string `json:"type"`
}

type Program struct {
    Functions []Function `json:"functions"`
}

func (instr Instruction) String() string {
    var sb strings.Builder
    if instr.Dest != "" {
        sb.WriteString(fmt.Sprintf("%v: %v = ", instr.Dest, instr.Type))
    }
    sb.WriteString(instr.Op)

    if len(instr.Funcs) > 0 {
        sb.WriteString(" @" + strings.Join(instr.Funcs, "@ "))
    }
    if len(instr.Args) > 0 {
        sb.WriteString(" " + strings.Join(instr.Args, " ") )
    }
    if len(instr.Labels) > 0 {
        sb.WriteString(" ." + strings.Join(instr.Labels, " ."))
    }

    switch v := instr.Value.(type) {
    case bool, float64:
        sb.WriteString(fmt.Sprintf(" %v", v))
    }

    sb.WriteString(";")
    return sb.String()
}

func ReadProgramJsonFromStdin() Program {
    input, err := ioutil.ReadAll(os.Stdin)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to read from stdin: %v\n", err)
        os.Exit(1)
    }
    var program Program
    err = json.Unmarshal(input, &program)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to unmarshal JSON: %v\n", err)
        os.Exit(1)
    }

    for funcId, function := range program.Functions {
        var convertedInstrs []interface{}
        for _, instr := range function.Instrs {
            if instrMap, ok := instr.(map[string]interface{}); ok {
                instrBytes, _ := json.Marshal(instrMap)
                if _, hasLabel := instrMap["label"]; hasLabel {
                    var label Label
                    _ = json.Unmarshal(instrBytes, &label)
                    convertedInstrs = append(convertedInstrs, label)
                } else {
                    var instr Instruction
                    _ = json.Unmarshal(instrBytes, &instr)
                    convertedInstrs = append(convertedInstrs, instr)
                }
            }
        }
        program.Functions[funcId].Instrs = convertedInstrs
    }

    return program
}

func getInstructionValue(instr interface{}, key string) ([]string, bool) {
    nonLabelInstr, ok := instr.(Instruction)
    if !ok {
        return nil, false
    }

    if key == "args" {
        return nonLabelInstr.Args, true
    } else if key == "funcs" {
        return nonLabelInstr.Funcs, true
    } else if key == "labels" {
        return nonLabelInstr.Labels, true
    }
    return nil, false
}

func FormBlocks(instrs []interface{}) [][]interface{} {
    terminators := map[string]bool {
        "br": true,
        "jmp": true,
        "ret": true,
    }

    var blocks [][]interface{}
    curBlock := []interface{}{}
    for _, instr := range instrs {
        switch v := instr.(type) {
        case Instruction:
            curBlock = append(curBlock, instr)
            if terminators[v.Op] {
                blocks = append(blocks, curBlock)
                curBlock = []interface{}{}
            }
        case Label:
            if len(curBlock) > 0 {
                blocks = append(blocks, curBlock)
            }
            curBlock = []interface{}{instr}
        }
    }
    if len(curBlock) > 0 {
        blocks = append(blocks, curBlock)
    }
    return blocks
}

func ParseInstruction2Text(instr interface{}) string {
    var sb strings.Builder
    switch v := instr.(type) {
    case Instruction:
        sb.WriteString(fmt.Sprintf("  %v\n", v))
    case Label:
        sb.WriteString(fmt.Sprintf(".%v:\n", v.Label))
    default:
        sb.WriteString(fmt.Sprintf("  %v\n", v))
    }
    return sb.String()
}

func ParseBlock2Text(block []interface{}) string {
    var sb strings.Builder
    for _, instr := range block {
        sb.WriteString(ParseInstruction2Text(instr))
    }
    return sb.String()
}

func ParseProgram2Text(program *Program) string {
    var sb strings.Builder
    for _, function := range program.Functions {
        sb.WriteString(fmt.Sprintf("@%v", function.Name))
        if len(function.Args) > 0 {
            sb.WriteString("(")
            for argId, arg := range function.Args {
                sb.WriteString(fmt.Sprintf("%v: %v", arg.Name, arg.Type))
                if argId != len(function.Args) - 1 {
                    sb.WriteString(", ")
                }
            }
            sb.WriteString(")")
            if function.Type != "" {
                sb.WriteString(fmt.Sprintf(": %v", function.Type))
            }
        }
        sb.WriteString(" {\n")
        for _, instr := range function.Instrs {
            sb.WriteString(ParseInstruction2Text(instr))
        }
        sb.WriteString("}\n")
    }
    return sb.String()
}
