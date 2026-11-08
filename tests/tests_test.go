package tests

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
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
		{"字符串转义", "\"a\\nb\"", "a\nb"},

		// --- 注释 ---
		{"行注释", "// comment\n1 + 1", "2"},
		{"行尾注释", "1 + 1 // comment", "2"},

		// --- 条件 ---
		{"单行条件", "damn 1 < 2 { ass \"yes\" }", "yes"},
		{"else 分支", "damn 1 > 2 { ass \"a\" } else { ass \"b\" }", "b"},
		{"else if", "damn 1 > 2 { ass \"a\" } else damn 2 < 3 { ass \"c\" } else { ass \"b\" }", "c"},

		// --- 变量与赋值 ---
		{"变量赋值", "shit x = 5; x * 2", "10"},

		// --- 内置函数 ---
		{"数组长度", "len([1, 2, 3])", "3"},
		{"数组推送", "shit arr = [1]; push(arr, 2); len(arr)", "2"},
		{"哈希长度", "len({\"a\": 1, \"b\": 2})", "2"},

		// --- 字典 ---
		{"字典访问", "shit h = {\"a\": 1}; h[\"a\"]", "1"},
		{"字典索引赋值", "shit h = {\"a\": 1}; h[\"a\"] = 9; h[\"a\"]", "9"},
		{"字典新增键", "shit h = {}; h[\"k\"] = 7; h[\"k\"]", "7"},

		// --- 数组索引赋值 ---
		{"数组索引赋值", "shit a = [1, 2, 3]; a[1] = 20; a[1]", "20"},

		// --- 函数与闭包 ---
		{"单行函数", "shit f = bitch(x) { ass x * 2 }; f(5)", "10"},
		{"多行模拟函数", "bitch square(n) { ass n * n }; square(5)", "25"},
		{"闭包捕获", "bitch makeAdder(n) { ass bitch(x) { ass x + n } }; shit add5 = makeAdder(5); add5(10)", "15"},
		{"参数数量错误", "bitch f(a, b) { ass a + b }; f(1)", "ERROR: wrong number of arguments: got=1, want=2"},

		// --- 循环 ---
		{"累加循环", "shit i = 0; shit sum = 0; hell i < 5 { shit sum = sum + i; shit i = i + 1 }; sum", "10"},

		// --- break ---
		{"break 基本", "shit i = 0; shit sum = 0; hell true { damn i >= 3 { break }; shit sum = sum + i; shit i = i + 1 }; sum", "3"},
		{"break 不执行后续语句", "shit i = 0; shit x = 0; hell i < 10 { shit i = i + 1; break; shit x = 999 }; x", "0"},
		{"break 外层", "break", "ERROR: `break` outside of loop"},
		{"break 不穿透函数", "bitch f() { break }; f()", "ERROR: `break` cannot cross function boundary"},
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

// 黑盒测试：跑 tests/*.fuck，比对 stdout 与同名 .expected 文件
func TestFuckFiles(t *testing.T) {
	files, err := filepath.Glob("*.fuck")
	if err != nil {
		t.Fatalf("glob failed: %v", err)
	}
	if len(files) == 0 {
		t.Fatalf("no .fuck test files found")
	}
	for _, file := range files {
		file := file
		t.Run(filepath.Base(file), func(t *testing.T) {
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("read file failed: %v", err)
			}

			// 捕获 stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			l := lexer.New(string(data))
			p := parser.New(l)
			program := p.ParseProgram()

			var evalErr string
			if len(p.Errors()) > 0 {
				evalErr = "syntax errors: " + strings.Join(p.Errors(), "; ")
			} else {
				env := evaluator.NewEnvironment()
				obj := evaluator.Eval(program, env)
				if obj != nil && obj.Type() == "ERROR" {
					evalErr = "runtime error: " + obj.Inspect()
				}
			}

			w.Close()
			os.Stdout = oldStdout
			var buf bytes.Buffer
			io.Copy(&buf, r)
			r.Close()

			if evalErr != "" {
				t.Fatalf("%s", evalErr)
			}

			expectedPath := strings.TrimSuffix(file, ".fuck") + ".expected"
			expectedData, err := os.ReadFile(expectedPath)
			if err != nil {
				t.Fatalf("missing expected file %s: %v", expectedPath, err)
			}
			expected := string(expectedData)
			actual := buf.String()

			if actual != expected {
				t.Errorf("output mismatch for %s\n--- expected ---\n%s\n--- actual ---\n%s", file, expected, actual)
			}
		})
	}
}