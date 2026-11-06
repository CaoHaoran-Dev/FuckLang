package tests

import (
	"testing"

	"fucklang/internal/evaluator"
	"fucklang/internal/lexer"
	"fucklang/internal/parser"
)

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

func TestEvaluatorLogic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// --- 算术运算 ---
		{"加法", "1 + 1", "2"},
		{"乘法与括号", "2 * (3 + 4)", "14"},
		{"除法", "10 / 2", "5"},
		{"负数", "-5", "-5"},
		{"除以零", "1 / 0", "ERROR: division by zero"},

		// --- 字符串 ---
		{"字符串拼接", "\"foo\" + \"bar\"", "foobar"},
		{"字符串比较", "\"a\" == \"a\"", "true"},

		// --- 条件 ---
		{"单行条件", "damn 1 < 2 { ass \"yes\" }", "yes"},

		// --- 变量与赋值 ---
		{"变量赋值", "shit x = 5; x * 2", "10"},

		// --- 内置函数 ---
		{"数组长度", "len([1, 2, 3])", "3"},
		{"数组推送", "shit arr = [1]; push(arr, 2); len(arr)", "2"},

		// --- 字典 ---
		{"字典访问", "shit h = {\"a\": 1}; h[\"a\"]", "1"},

		// --- 函数与闭包 ---
		{"单行函数", "shit f = bitch(x) { ass x * 2 }; f(5)", "10"},
		{"多行模拟函数", "bitch square(n) { ass n * n }; square(5)", "25"},
		{"闭包捕获", "bitch makeAdder(n) { ass bitch(x) { ass x + n } }; shit add5 = makeAdder(5); add5(10)", "15"},

		// --- 循环 ---
		{"累加循环", "shit i = 0; shit sum = 0; hell i < 5 { shit sum = sum + i; shit i = i + 1 }; sum", "10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			if evaluated != tt.expected {
				t.Errorf("输入: %s\n期望: %s\n实际: %v", tt.input, tt.expected, evaluated)
			}
		})
	}
}