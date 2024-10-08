package main

import (
    "encoding/json"
    "flag"
    "fmt"
)

func OptimizeProgram(program *Program, format string) {
    for funcId := range program.Functions {
        Lvn(&(program.Functions[funcId]))
        TrivalDCE(&(program.Functions[funcId]))
    }

    if format == "json" {
        jsonData, err := json.MarshalIndent(program, "", "  ")
        if err != nil {
            fmt.Println("Failed to marshal Program:", err)
            return
        }
        fmt.Printf("%s", jsonData)
    } else if format == "text" {
        fmt.Printf("%v\n", ParseProgram2Text(program))
    }
}

func main() {
    format := flag.String("output", "", "output format (json or text)")
    flag.Parse()

    program := ReadProgramJsonFromStdin()
    OptimizeProgram(&program, *format)
}
