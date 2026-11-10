package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"fucklang/internal/ast"
	"fucklang/internal/evaluator"
	"fucklang/internal/lexer"
	"fucklang/internal/object"
	"fucklang/internal/parser"
)

const historyFile = ".fucklang_history"

// ---------- 平台相关 ----------

// isWindows 当前是否运行在 Windows
func isWindows() bool { return runtime.GOOS == "windows" }

// clearScreen 跨平台清屏
func clearScreen() {
	if isWindows() {
		// Windows 优先用 cmd 的 cls，兼容老 cmd 和 Windows Terminal
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		if err := cmd.Run(); err == nil {
			return
		}
		// 回退到 ANSI（Windows 10+ 支持）
		fmt.Print("\033[2J\033[H")
		return
	}
	// Unix: ANSI 序列
	fmt.Print("\033[2J\033[H")
}

// historyPath 跨平台历史文件路径
func historyPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// 回退到当前目录
		return historyFile
	}
	return filepath.Join(home, historyFile)
}

// eofHint 返回平台相关的 EOF 提示
func eofHint() string {
	if isWindows() {
		return "Ctrl+Z then Enter"
	}
	return "Ctrl+D"
}

// normalizeLine 处理 Windows 的 \r\n 和残留 \r
func normalizeLine(s string) string {
	s = strings.TrimRight(s, "\r")
	return s
}

// ---------- 主入口 ----------

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

	// 统一换行符：把 \r\n 转成 \n，避免 lexer 把 \r 当非法字符
	source := strings.ReplaceAll(string(data), "\r\n", "\n")
	run(source)
}

// ---------- 错误格式化 ----------

func formatRuntimeError(errObj *object.Error, source string) string {
	if errObj.Line <= 0 {
		return errObj.Message
	}
	line := errObj.Line
	lines := strings.Split(source, "\n")
	if line > len(lines) {
		return fmt.Sprintf("line %d: %s", line, errObj.Message)
	}
	srcLine := lines[line-1]
	return fmt.Sprintf("line %d: %s\n  %s\n  %s^",
		line, errObj.Message, srcLine, strings.Repeat(" ", leadingSpaces(srcLine)))
}

func leadingSpaces(s string) int {
	n := 0
	for _, c := range s {
		if c == ' ' || c == '\t' {
			n++
		} else {
			break
		}
	}
	return n
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
		fmt.Printf("Runtime Error: %s\n", formatRuntimeError(evaluated.(*object.Error), source))
		os.Exit(1)
	}
}

// ---------- REPL 状态 ----------

type replState struct {
	env     *evaluator.Environment
	history []string
	buffer  strings.Builder
}

func (s *replState) pushHistory(line string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}
	if len(s.history) > 0 && s.history[len(s.history)-1] == line {
		return
	}
	s.history = append(s.history, line)
}

func (s *replState) loadHistory() {
	data, err := os.ReadFile(historyPath())
	if err != nil {
		return
	}
	content := strings.ReplaceAll(string(data), "\r\n", "\n")
	for _, line := range strings.Split(content, "\n") {
		if line != "" {
			s.history = append(s.history, line)
		}
	}
}

func (s *replState) saveHistory() {
	f, err := os.OpenFile(historyPath(), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return
	}
	defer f.Close()
	start := 0
	if len(s.history) > 1000 {
		start = len(s.history) - 1000
	}
	for _, line := range s.history[start:] {
		// 统一写 \n；Windows 记事本也能正确识别（现代 Windows 都支持）
		fmt.Fprintln(f, line)
	}
}

// ---------- 元命令 ----------

func (s *replState) handleMeta(line string) bool {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, ":") {
		return false
	}
	parts := strings.SplitN(trimmed[1:], " ", 2)
	cmd := parts[0]
	arg := ""
	if len(parts) == 2 {
		arg = strings.TrimSpace(parts[1])
	}

	switch cmd {
	case "help", "h", "?":
		printHelp()
	case "env":
		s.env.Dump(os.Stdout)
	case "ast":
		if arg == "" {
			fmt.Println("Usage: :ast <expression>")
			return true
		}
		printAST(arg)
	case "type":
		if arg == "" {
			fmt.Println("Usage: :type <expression>")
			return true
		}
		printType(arg, s.env)
	case "time":
		if arg == "" {
			fmt.Println("Usage: :time <expression>")
			return true
		}
		printTime(arg, s.env)
	case "history":
		for i, h := range s.history {
			fmt.Printf("%4d  %s\n", i+1, h)
		}
	case "clear":
		s.buffer.Reset()
		fmt.Println("(cleared)")
	case "cls", "clr":
		clearScreen()
	case "quit", "q", "exit":
		s.saveHistory()
		fmt.Println("Bye!")
		os.Exit(0)
	default:
		fmt.Printf("Unknown meta command: :%s (type :help for help)\n", cmd)
	}
	return true
}

func printHelp() {
	fmt.Printf(`Meta commands:
  :help / :h / :?     Show this help
  :env                List environment variables
  :ast <expr>         Print AST of an expression (indented tree)
  :type <expr>        Show the type of an expression
  :time <expr>        Measure evaluation time
  :history            Show input history
  :clear              Clear the current multi-line input buffer
  :cls / :clr         Clear the screen
  :quit / :q / :exit  Save history and quit

Keywords: fuck shit damn else hell break bitch ass wtf fk orgy
          bullshit nobullshit str arr file

Builtins: len push type suck fucking fuckoff

Namespaces:
  str.{len,upper,lower,trim,split,contains,replace,join}
  arr.{len,push,pop,slice,contains,keys,values}
  file.{read,write,exists}

Platform: %s
Exit: %s
Tip: multi-line prompt shows '..N' where N is bracket depth.
`, runtime.GOOS, eofHint())
}

// ---------- AST 树形打印 ----------

func printAST(src string) {
	l := lexer.New(src)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		fmt.Println("Syntax Error:")
		for _, m := range p.Errors() {
			fmt.Println("  ", m)
		}
		return
	}
	if len(program.Statements) == 0 {
		fmt.Println("(empty)")
		return
	}
	for _, stmt := range program.Statements {
		printNode(stmt, 0)
	}
}

func indent(n int) string { return strings.Repeat("  ", n) }

func printNode(node ast.Node, depth int) {
	pad := indent(depth)
	switch n := node.(type) {
	case *ast.Program:
		fmt.Printf("%sProgram\n", pad)
		for _, s := range n.Statements {
			printNode(s, depth+1)
		}
	case *ast.ExpressionStatement:
		fmt.Printf("%sExpressionStatement\n", pad)
		printNode(n.Expression, depth+1)
	case *ast.PrintStatement:
		fmt.Printf("%sPrintStatement\n", pad)
		printNode(n.Value, depth+1)
	case *ast.AssignStatement:
		fmt.Printf("%sAssignStatement %s\n", pad, n.Name.Value)
		printNode(n.Value, depth+1)
	case *ast.IndexAssignStatement:
		fmt.Printf("%sIndexAssignStatement\n", pad)
		fmt.Printf("%s  target:\n", pad)
		printNode(n.Left, depth+2)
		fmt.Printf("%s  index:\n", pad)
		printNode(n.Index, depth+2)
		fmt.Printf("%s  value:\n", pad)
		printNode(n.Value, depth+2)
	case *ast.CompoundAssignStatement:
		fmt.Printf("%sCompoundAssignStatement %s %s=\n", pad, n.Name.Value, n.Operator)
		printNode(n.Value, depth+1)
	case *ast.ReturnStatement:
		fmt.Printf("%sReturnStatement\n", pad)
		printNode(n.Value, depth+1)
	case *ast.BreakStatement:
		fmt.Printf("%sBreakStatement\n", pad)
	case *ast.ContinueStatement:
		fmt.Printf("%sContinueStatement\n", pad)
	case *ast.AssertStatement:
		fmt.Printf("%sAssertStatement\n", pad)
		fmt.Printf("%s  cond:\n", pad)
		printNode(n.Condition, depth+2)
		if n.Message != nil {
			fmt.Printf("%s  msg:\n", pad)
			printNode(n.Message, depth+2)
		}
	case *ast.LoopStatement:
		fmt.Printf("%sLoopStatement (hell)\n", pad)
		fmt.Printf("%s  cond:\n", pad)
		printNode(n.Condition, depth+2)
		printNode(n.Body, depth+1)
	case *ast.ForStatement:
		fmt.Printf("%sForStatement (orgy)\n", pad)
		fmt.Printf("%s  init:\n", pad)
		printNode(n.Init, depth+2)
		fmt.Printf("%s  cond:\n", pad)
		printNode(n.Cond, depth+2)
		fmt.Printf("%s  post:\n", pad)
		printNode(n.Post, depth+2)
		printNode(n.Body, depth+1)
	case *ast.BlockStatement:
		fmt.Printf("%sBlockStatement\n", pad)
		for _, s := range n.Statements {
			printNode(s, depth+1)
		}
	case *ast.Identifier:
		fmt.Printf("%sIdentifier %q\n", pad, n.Value)
	case *ast.IntegerLiteral:
		fmt.Printf("%sInteger %d\n", pad, n.Value)
	case *ast.StringLiteral:
		fmt.Printf("%sString %q\n", pad, n.Value)
	case *ast.Boolean:
		fmt.Printf("%sBoolean %v\n", pad, n.Value)
	case *ast.PrefixExpression:
		fmt.Printf("%sPrefixExpression %s\n", pad, n.Operator)
		printNode(n.Right, depth+1)
	case *ast.InfixExpression:
		fmt.Printf("%sInfixExpression %s\n", pad, n.Operator)
		printNode(n.Left, depth+1)
		printNode(n.Right, depth+1)
	case *ast.IfExpression:
		fmt.Printf("%sIfExpression (damn)\n", pad)
		fmt.Printf("%s  cond:\n", pad)
		printNode(n.Condition, depth+2)
		printNode(n.Consequence, depth+1)
		if n.Alternative != nil {
			fmt.Printf("%s  else:\n", pad)
			printNode(n.Alternative, depth+2)
		}
	case *ast.FunctionLiteral:
		names := make([]string, len(n.Parameters))
		for i, p := range n.Parameters {
			names[i] = p.Name.Value
		}
		fmt.Printf("%sFunctionLiteral (%s)\n", pad, strings.Join(names, ", "))
		printNode(n.Body, depth+1)
	case *ast.CallExpression:
		fmt.Printf("%sCallExpression\n", pad)
		fmt.Printf("%s  fn:\n", pad)
		printNode(n.Function, depth+2)
		if len(n.Arguments) > 0 {
			fmt.Printf("%s  args:\n", pad)
			for _, a := range n.Arguments {
				printNode(a, depth+2)
			}
		}
	case *ast.ArrayLiteral:
		fmt.Printf("%sArrayLiteral\n", pad)
		for _, e := range n.Elements {
			printNode(e, depth+1)
		}
	case *ast.HashLiteral:
		fmt.Printf("%sHashLiteral\n", pad)
		for k, v := range n.Pairs {
			printNode(k, depth+1)
			printNode(v, depth+1)
		}
	case *ast.IndexExpression:
		fmt.Printf("%sIndexExpression\n", pad)
		printNode(n.Left, depth+1)
		printNode(n.Index, depth+1)
	case *ast.MemberExpression:
		fmt.Printf("%sMemberExpression .%s\n", pad, n.Member.Value)
		printNode(n.Object, depth+1)
	default:
		fmt.Printf("%s%T\n", pad, node)
	}
}

// ---------- :type / :time ----------

func printType(src string, env *evaluator.Environment) {
	l := lexer.New(src)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		fmt.Println("Syntax Error:")
		for _, m := range p.Errors() {
			fmt.Println("  ", m)
		}
		return
	}
	obj := evaluator.Eval(program, env)
	if obj == nil {
		fmt.Println("null")
		return
	}
	if obj.Type() == "ERROR" {
		fmt.Printf("ERROR: %s\n", obj.Inspect())
		return
	}
	fmt.Println(obj.Type())
}

func printTime(src string, env *evaluator.Environment) {
	l := lexer.New(src)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		fmt.Println("Syntax Error:")
		for _, m := range p.Errors() {
			fmt.Println("  ", m)
		}
		return
	}
	start := time.Now()
	obj := evaluator.Eval(program, env)
	elapsed := time.Since(start)
	if obj != nil && obj.Type() == "ERROR" {
		fmt.Printf("ERROR: %s (%.3fms)\n", obj.Inspect(), float64(elapsed.Microseconds())/1000.0)
		return
	}
	fmt.Printf("%.3fms\n", float64(elapsed.Microseconds())/1000.0)
}

// ---------- 主循环 ----------

func startRepl() {
	fmt.Println("Welcome to FuckLang REPL")
	fmt.Printf("Platform: %s | Exit: %s | Help: :help\n", runtime.GOOS, eofHint())
	fmt.Println(strings.Repeat("-", 50))

	s := &replState{env: evaluator.NewEnvironment()}
	s.loadHistory()

	scanner := bufio.NewScanner(os.Stdin)
	// 放宽单行上限，避免长表达式被截断（默认 64KB）
	scanner.Buffer(make([]byte, 0, 4096), 1024*1024)

	for {
		var prompt string
		if s.buffer.Len() == 0 {
			prompt = ">> "
		} else {
			balance := bracketBalance(s.buffer.String())
			prompt = fmt.Sprintf("..%d ", balance)
		}
		fmt.Print(prompt)

		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Printf("\nInput error: %v\n", err)
			}
			fmt.Println("\nBye!")
			s.saveHistory()
			break
		}
		line := normalizeLine(scanner.Text())

		// 元命令仅在缓冲为空时生效
		if s.buffer.Len() == 0 && s.handleMeta(line) {
			continue
		}

		s.pushHistory(line)
		s.buffer.WriteString(line)
		s.buffer.WriteString("\n")

		src := s.buffer.String()
		trimmed := strings.TrimSpace(src)
		if trimmed == "" {
			s.buffer.Reset()
			continue
		}
		// 纯注释行直接丢弃
		if strings.HasPrefix(trimmed, "//") && bracketBalance(src) == 0 {
			s.buffer.Reset()
			continue
		}
		if bracketBalance(src) > 0 {
			continue
		}

		s.buffer.Reset()

		l := lexer.New(src)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			for _, msg := range p.Errors() {
				fmt.Println("Syntax Error:", msg)
			}
			continue
		}

		evaluated := evaluator.Eval(program, s.env)
		if evaluated != nil && evaluated.Type() == "ERROR" {
			fmt.Printf("Runtime Error: %s\n", formatRuntimeError(evaluated.(*object.Error), src))
		} else if evaluated != nil && evaluated.Type() != "NULL" {
			fmt.Println(evaluated.Inspect())
		}
	}
}

// bracketBalance 返回括号净深度，跳过字符串和 // 注释
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