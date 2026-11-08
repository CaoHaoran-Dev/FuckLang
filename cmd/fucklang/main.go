package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"fucklang/internal/evaluator"
	"fucklang/internal/lexer"
	"fucklang/internal/object"
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

func formatRuntimeError(errObj *object.Error) string {
	if errObj.Line > 0 {
		return fmt.Sprintf("line %d: %s", errObj.Line, errObj.Message)
	}
	return errObj.Message
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
		fmt.Printf("Runtime Error: %s\n", formatRuntimeError(evaluated.(*object.Error)))
		os.Exit(1)
	}
}

// 统计一行里未配平的括号数量，用于 REPL 多行输入。
// 跳过字符串和 // 行注释。
func bracketBalance(s string) int {
	balance := 0
	inString := false
	inComment := false
	for i := 0; i < len(s); i++ {
		c := s[i]

		if inComment {
			if c == '\n' {
				inComment = false
			}
			continue
		}

		if inString {
			if c == '\\' && i+1 < len(s) {
				i++
				continue
			}
			if c == '"' {
				inString = false
			}
			continue
		}

		switch c {
		case '/':
			if i+1 < len(s) && s[i+1] == '/' {
				inComment = true
				i++
				continue
			}
		case '"':
			inString = true
		case '{', '(', '[':
			balance++
		case '}', ')', ']':
			balance--
		}
	}
	return balance
}

func startRepl() {
	fmt.Println("Welcome to FuckLang REPL")
	fmt.Println("Press Ctrl+C or Ctrl+D to exit")
	fmt.Println("--------------------------------")
	scanner := bufio.NewScanner(os.Stdin)
	env := evaluator.NewEnvironment()

	var buffer strings.Builder

	for {
		if buffer.Len() == 0 {
			fmt.Print(">> ")
		} else {
			fmt.Print(".. ")
		}
		if !scanner.Scan() {
			fmt.Println("\nBye!")
			break
		}
		line := scanner.Text()
		buffer.WriteString(line)
		buffer.WriteString("\n")

		src := buffer.String()
		if strings.TrimSpace(src) == "" {
			buffer.Reset()
			continue
		}
		// 括号未配平，继续读下一行
		if bracketBalance(src) > 0 {
			continue
		}

		buffer.Reset()

		l := lexer.New(src)
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
			fmt.Printf("Runtime Error: %s\n", formatRuntimeError(evaluated.(*object.Error)))
		} else if evaluated != nil && evaluated.Type() != "NULL" {
			fmt.Println(evaluated.Inspect())
		}
	}
}