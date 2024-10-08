MyGo Compiler Optimization
---

## Intro

Compiler optimization strategies implemented in Go, working directly on Bril's intermediate representation (IR) for back-end optimizations without involving front-end syntax or semantic analysis. Runtime optimizations are also beyond the scope of this project.

## Prerequisites
- **Bril**: Install [Bril](https://github.com/sampsyo/bril) to run this project.
- **Bril IR Syntax**: Implemented an interpreter for Bril in the Go environment, based on the IR syntax. For more details, please refer to [documentation](https://capra.cs.cornell.edu/bril/lang/syntax.html).

## Running the Project

Use the following command to run the project.

```sh
$ bril2json < examples/<filename>.bril | go run src/*.go --output=[text|json]
```

Replace `<filename>` with the name of your Bril file and specify the desired output format (`text` or `json`).

## Optimization Strategies

Currently, the following optimization strategies are implemented:
- **DCE (Dead Code Elimination)**: Removes code that does not affect the program's output.
- **LVN (Local Value Numbering)**: Identifies and eliminates redundant calculations within basic blocks.

## A Simple Example

Consider the following Bril code:

```=
# examples/idchain.bril
@main {
x: int = const 4;
jmp .label;
.label:
copy1: int = id x;
copy2: int = id copy1;
copy3: int = id copy2;
print copy3;
}
```

To optimize this code, run:

```sh
$ bril2json < examples/idchain.bril | go run src/*.go --output=text
```

The output might be:

```=
@main {
x: int = const 4;
jmp .label;
.label:
print x;
}
```

In this example, redundant assignments are removed, simplifying the code while preserving its functionality.
