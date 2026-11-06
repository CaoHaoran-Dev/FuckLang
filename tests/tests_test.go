package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fucklang/internal/evaluator"
	"fucklang/internal/lexer"
	"fucklang/internal/parser"
)

// helper: 执行一段 FuckLang 代码并返回它的输出结果
func testEval(input string) string {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return "SYNTAX_ERROR"
	}

	env := evaluator.NewEnvironment()
	obj := evaluator.Eval(program, env)
	if obj != nil {
		return obj.Inspect()
	}
	return "NULL"
}

// 1. 核心逻辑断言测试 (原 evaluator_test.go)
func TestEvaluatorLogic(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1 + 1", "2"},
		{"2 * (3 + 4)", "14"},
		{"\"foo\" + \"bar\"", "foobar"},
		{"damn 1 < 2 { ass \"yes\" }", "yes"},
		{"shit x = 5; x * 2", "10"},
		{"len([1, 2, 3])", "3"},
		{"{\"a\": 1}[\"a\"]", "1"},
		{"1 / 0", "ERROR: division by zero"},
		{"shit arr = [1]; push(arr, 2); len(arr)", "2"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		if evaluated != tt.expected {
			t.Errorf("for %s, expected %s, got %v", tt.input, tt.expected, evaluated)
		}
	}
}

// 2. 全量文件执行测试 (原 e2e_test.go)
func TestAllFuckFiles(t *testing.T) {
	files, err := filepath.Glob("*.fuck")
	if err != nil {
		t.Fatalf("Failed to find test files: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("No .fuck files found in /tests directory")
	}

	for _, file := range files {
		file := file // capture range variable
		t.Run(filepath.Base(file), func(t *testing.T) {
			source, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("Failed to read %s: %v", file, err)
			}

			l := lexer.New(string(source))
			p := parser.New(l)
			program := p.ParseProgram()

			// 检查语法错误
			if len(p.Errors()) > 0 {
				t.Fatalf("Syntax errors in %s:\n%s", file, strings.Join(p.Errors(), "\n"))
			}

			// 检查运行时错误
			env := evaluator.NewEnvironment()
			evaluated := evaluator.Eval(program, env)
			if evaluated != nil && evaluated.Type() == "ERROR" {
				t.Fatalf("Runtime error in %s: %s", file, evaluated.Inspect())
			}
		})
	}
}