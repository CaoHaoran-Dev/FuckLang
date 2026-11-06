package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"fucklang/internal/evaluator"
	"fucklang/internal/lexer"
	"fucklang/internal/parser"
)

func main() {
	if len(os.Args) < 2 {
		startRepl()
		return
	}

	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	run(string(data))
}

func run(source string) {
	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Println("Syntax Error:")
		for _, msg := range p.Errors() {
			fmt.Println("  ", msg)
		}
		os.Exit(1)
	}

	env := evaluator.NewEnvironment()
	evaluated := evaluator.Eval(program, env)
	if evaluated != nil && evaluated.Type() == "ERROR" {
		fmt.Printf("Runtime Error: %s\n", evaluated.Inspect())
		os.Exit(1)
	}
}

func startRepl() {
	fmt.Println("Welcome to FuckLang REPL")
	fmt.Println("Press Ctrl+C or Ctrl+D to exit")
	fmt.Println("--------------------------------")
	scanner := bufio.NewScanner(os.Stdin)
	env := evaluator.NewEnvironment()

	for {
		fmt.Print(">> ")
		if !scanner.Scan() {
			fmt.Println("\nBye!")
			break
		}
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		l := lexer.New(line)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			for _, msg := range p.Errors() {
				fmt.Println("Syntax Error:", msg)
			}
			continue
		}

		evaluated := evaluator.Eval(program, env)
		if evaluated != nil && evaluated.Type() == "ERROR" {
			fmt.Printf("Runtime Error: %s\n", evaluated.Inspect())
		} else if evaluated != nil && evaluated.Type() != "NULL" {
			fmt.Println(evaluated.Inspect())
		}
	}
}