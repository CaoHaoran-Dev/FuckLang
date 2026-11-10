package evaluator

import (
	"fmt"
	"io"
	"sort"

	"fucklang/internal/ast"
	"fucklang/internal/object"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

type Environment struct {
	store map[string]object.Object
	outer *Environment
}

func NewEnvironment() *Environment {
	env := &Environment{store: make(map[string]object.Object)}
	for name, fn := range builtins {
		env.Set(name, fn)
	}
	// 命名空间 hash
	env.Set("str", namespaceHash(strBuiltins))
	env.Set("arr", namespaceHash(arrBuiltins))
	env.Set("file", namespaceHash(fileBuiltins))
	return env
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

func (e *Environment) Get(name string) (object.Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		return e.outer.Get(name)
	}
	return obj, ok
}

func (e *Environment) Set(name string, val object.Object) object.Object {
	e.store[name] = val
	return val
}

// Dump 打印环境变量（供 REPL :env 使用）
func (e *Environment) Dump(w io.Writer) {
	names := make([]string, 0, len(e.store))
	for n := range e.store {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		v := e.store[n]
		if v == nil {
			fmt.Fprintf(w, "%s = <nil>\n", n)
			continue
		}
		fmt.Fprintf(w, "%s = %s\n", n, v.Inspect())
	}
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func newErrorAt(line int, format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...), Line: line}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR
	}
	return false
}

func namespaceHash(m map[string]*object.Builtin) *object.Hash {
	pairs := make(map[object.HashKey]object.HashPair, len(m))
	for name, fn := range m {
		key := &object.String{Value: name}
		pairs[key.HashKey()] = object.HashPair{Key: key, Value: fn}
	}
	return &object.Hash{Pairs: pairs}
}

func Eval(node ast.Node, env *Environment) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.BlockStatement:
		return evalBlock(node, env)
	case *ast.PrintStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		if val != nil {
			fmt.Println(val.Inspect())
		}
		return NULL
	case *ast.AssignStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)
		return val
	case *ast.CompoundAssignStatement:
		return evalCompoundAssign(node, env)
	case *ast.IndexAssignStatement:
		return evalIndexAssign(node, env)
	case *ast.ReturnStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	case *ast.BreakStatement:
		return &object.BreakValue{}
	case *ast.ContinueStatement:
		return &object.ContinueValue{}
	case *ast.AssertStatement:
		return evalAssert(node, env)
	case *ast.LoopStatement:
		return evalLoop(node, env)
	case *ast.ForStatement:
		return evalFor(node, env)
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.Identifier:
		if val, ok := env.Get(node.Value); ok {
			return val
		}
		return newErrorAt(node.Token.Line, "identifier not found: %s", node.Value)
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Token.Line, node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Token.Line, node.Operator, left, right)
	case *ast.FunctionLiteral:
		return &object.Function{Parameters: node.Parameters, Body: node.Body, Env: env}
	case *ast.CallExpression:
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}
		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}
		return applyFunction(node.Token.Line, function, args)
	case *ast.ArrayLiteral:
		elements := evalExpressions(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &object.Array{Elements: elements}
	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}
		return evalIndexExpression(node.Token.Line, left, index)
	case *ast.MemberExpression:
		return evalMemberExpression(node, env)
	case *ast.HashLiteral:
		return evalHashLiteral(node, env)
	}
	return nil
}

func evalAssert(node *ast.AssertStatement, env *Environment) object.Object {
	cond := Eval(node.Condition, env)
	if isError(cond) {
		return cond
	}
	if !isTruthy(cond) {
		msg := "assertion failed"
		if node.Message != nil {
			m := Eval(node.Message, env)
			if isError(m) {
				return m
			}
			msg = m.Inspect()
		}
		return newErrorAt(node.Token.Line, "wtf: %s", msg)
	}
	return NULL
}

func evalCompoundAssign(node *ast.CompoundAssignStatement, env *Environment) object.Object {
	cur, ok := env.Get(node.Name.Value)
	if !ok {
		return newErrorAt(node.Token.Line, "identifier not found: %s", node.Name.Value)
	}
	val := Eval(node.Value, env)
	if isError(val) {
		return val
	}
	res := evalInfixExpression(node.Token.Line, node.Operator, cur, val)
	if isError(res) {
		return res
	}
	env.Set(node.Name.Value, res)
	return res
}

func evalFor(node *ast.ForStatement, env *Environment) object.Object {
	if node.Init != nil {
		res := Eval(node.Init, env)
		if isError(res) {
			return res
		}
	}
	for {
		cond := Eval(node.Cond, env)
		if isError(cond) {
			return cond
		}
		if !isTruthy(cond) {
			break
		}
		result := Eval(node.Body, env)
		if result != nil {
			t := result.Type()
			if t == object.RETURN_VALUE || t == object.ERROR {
				return result
			}
			if t == object.BREAK {
				return NULL
			}
		}
		if node.Post != nil {
			res := Eval(node.Post, env)
			if isError(res) {
				return res
			}
			if rv, ok := res.(*object.ReturnValue); ok {
				return rv
			}
		}
	}
	return NULL
}

func evalProgram(program *ast.Program, env *Environment) object.Object {
	var result object.Object = NULL
	for _, stmt := range program.Statements {
		result = Eval(stmt, env)

		if rv, ok := result.(*object.ReturnValue); ok {
			return rv.Value
		}
		if result != nil && result.Type() == object.ERROR {
			return result
		}
		if result != nil && result.Type() == object.BREAK {
			return newError("`break` outside of loop")
		}
		if result != nil && result.Type() == object.CONTINUE {
			return newError("`fk` outside of loop")
		}
	}
	return result
}

func evalBlock(block *ast.BlockStatement, env *Environment) object.Object {
	var result object.Object = NULL
	for _, stmt := range block.Statements {
		result = Eval(stmt, env)
		if result != nil {
			t := result.Type()
			if t == object.RETURN_VALUE || t == object.ERROR ||
				t == object.BREAK || t == object.CONTINUE {
				return result
			}
		}
	}
	return result
}

func evalLoop(ls *ast.LoopStatement, env *Environment) object.Object {
	for {
		cond := Eval(ls.Condition, env)
		if isError(cond) {
			return cond
		}
		if !isTruthy(cond) {
			break
		}
		result := Eval(ls.Body, env)
		if result != nil {
			t := result.Type()
			if t == object.RETURN_VALUE || t == object.ERROR {
				return result
			}
			if t == object.BREAK {
				return NULL
			}
		}
	}
	return NULL
}

func evalIfExpression(ie *ast.IfExpression, env *Environment) object.Object {
	condition := Eval(ie.Condition, env)
	if isError(condition) {
		return condition
	}
	if isTruthy(condition) {
		return Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
	}
	return NULL
}

func evalExpressions(exprs []ast.Expression, env *Environment) []object.Object {
	var result []object.Object
	for _, e := range exprs {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}
	return result
}

func applyFunction(line int, fn object.Object, args []object.Object) object.Object {
	function, ok := fn.(*object.Function)
	if !ok {
		if builtin, ok := fn.(*object.Builtin); ok {
			result := builtin.Fn(args...)
			if errObj, ok := result.(*object.Error); ok && errObj.Line == 0 {
				errObj.Line = line
			}
			return result
		}
		return newErrorAt(line, "not a function: %s", fn.Type())
	}

	if len(args) != len(function.Parameters) {
		return newErrorAt(line, "wrong number of arguments: got=%d, want=%d", len(args), len(function.Parameters))
	}

	outer, ok := function.Env.(*Environment)
	if !ok {
		return newErrorAt(line, "internal error: function environment is invalid")
	}
	extendedEnv := NewEnclosedEnvironment(outer)

	for i, param := range function.Parameters {
		extendedEnv.Set(param.Name.Value, args[i])
	}

	evaluated := Eval(function.Body, extendedEnv)
	if rv, ok := evaluated.(*object.ReturnValue); ok {
		return rv.Value
	}

	if evaluated != nil && evaluated.Type() == object.BREAK {
		return newErrorAt(line, "`break` cannot cross function boundary")
	}
	if evaluated != nil && evaluated.Type() == object.CONTINUE {
		return newErrorAt(line, "`fk` cannot cross function boundary")
	}

	return evaluated
}

func evalMemberExpression(node *ast.MemberExpression, env *Environment) object.Object {
	obj := Eval(node.Object, env)
	if isError(obj) {
		return obj
	}
	if h, ok := obj.(*object.Hash); ok {
		key := &object.String{Value: node.Member.Value}
		pair, ok := h.Pairs[key.HashKey()]
		if !ok {
			return newErrorAt(node.Token.Line, "member not found: %s", node.Member.Value)
		}
		return pair.Value
	}
	return newErrorAt(node.Token.Line, "member access not supported on %s", obj.Type())
}

func evalIndexExpression(line int, left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY && index.Type() == object.INTEGER:
		arr := left.(*object.Array)
		idx := index.(*object.Integer).Value
		if idx < 0 || int(idx) >= len(arr.Elements) {
			return newErrorAt(line, "index out of bounds: %d", idx)
		}
		return arr.Elements[idx]
	case left.Type() == object.HASH:
		hash := left.(*object.Hash)
		key, ok := index.(object.Hashable)
		if !ok {
			return newErrorAt(line, "unusable as hash key: %s", index.Type())
		}
		pair, ok := hash.Pairs[key.HashKey()]
		if !ok {
			return NULL
		}
		return pair.Value
	default:
		return newErrorAt(line, "index operator not supported: %s", left.Type())
	}
}

func evalIndexAssign(node *ast.IndexAssignStatement, env *Environment) object.Object {
	left := Eval(node.Left, env)
	if isError(left) {
		return left
	}
	index := Eval(node.Index, env)
	if isError(index) {
		return index
	}
	value := Eval(node.Value, env)
	if isError(value) {
		return value
	}

	line := node.Token.Line
	switch l := left.(type) {
	case *object.Array:
		if index.Type() != object.INTEGER {
			return newErrorAt(line, "array index must be INTEGER")
		}
		idx := index.(*object.Integer).Value
		if idx < 0 || int(idx) >= len(l.Elements) {
			return newErrorAt(line, "index out of bounds: %d", idx)
		}
		l.Elements[idx] = value
		return value
	case *object.Hash:
		key, ok := index.(object.Hashable)
		if !ok {
			return newErrorAt(line, "unusable as hash key: %s", index.Type())
		}
		l.Pairs[key.HashKey()] = object.HashPair{Key: index, Value: value}
		return value
	default:
		return newErrorAt(line, "index assignment not supported on %s", left.Type())
	}
}

func evalPrefixExpression(line int, operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return nativeBoolToBooleanObject(!isTruthy(right))
	case "-":
		if right.Type() == object.INTEGER {
			val := right.(*object.Integer).Value
			return &object.Integer{Value: -val}
		}
		return newErrorAt(line, "unknown operator: -%s", right.Type())
	}
	return newErrorAt(line, "unknown operator: %s%s", operator, right.Type())
}

func evalInfixExpression(line int, operator string, left, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER && right.Type() == object.INTEGER:
		return evalIntegerInfix(line, operator, left, right)
	case left.Type() == object.STRING && right.Type() == object.STRING:
		return evalStringInfix(line, operator, left, right)
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	case left.Type() != right.Type():
		return newErrorAt(line, "type mismatch: %s %s %s", left.Type(), operator, right.Type())
	default:
		return newErrorAt(line, "unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalIntegerInfix(line int, operator string, left, right object.Object) object.Object {
	lv := left.(*object.Integer).Value
	rv := right.(*object.Integer).Value
	switch operator {
	case "+":
		return &object.Integer{Value: lv + rv}
	case "-":
		return &object.Integer{Value: lv - rv}
	case "*":
		return &object.Integer{Value: lv * rv}
	case "/":
		if rv == 0 {
			return newErrorAt(line, "division by zero")
		}
		return &object.Integer{Value: lv / rv}
	case "<":
		return nativeBoolToBooleanObject(lv < rv)
	case ">":
		return nativeBoolToBooleanObject(lv > rv)
	case "<=":
		return nativeBoolToBooleanObject(lv <= rv)
	case ">=":
		return nativeBoolToBooleanObject(lv >= rv)
	case "==":
		return nativeBoolToBooleanObject(lv == rv)
	case "!=":
		return nativeBoolToBooleanObject(lv != rv)
	}
	return newErrorAt(line, "unknown operator: %s", operator)
}

func evalStringInfix(line int, operator string, left, right object.Object) object.Object {
	lv := left.(*object.String).Value
	rv := right.(*object.String).Value
	switch operator {
	case "+":
		return &object.String{Value: lv + rv}
	case "==":
		return nativeBoolToBooleanObject(lv == rv)
	case "!=":
		return nativeBoolToBooleanObject(lv != rv)
	}
	return newErrorAt(line, "unknown operator: %s", operator)
}

func evalHashLiteral(node *ast.HashLiteral, env *Environment) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)
	for keyNode, valueNode := range node.Pairs {
		key := Eval(keyNode, env)
		if isError(key) {
			return key
		}
		hashKey, ok := key.(object.Hashable)
		if !ok {
			return newError("unusable as hash key: %s", key.Type())
		}
		value := Eval(valueNode, env)
		if isError(value) {
			return value
		}
		pairs[hashKey.HashKey()] = object.HashPair{Key: key, Value: value}
	}
	return &object.Hash{Pairs: pairs}
}

func isTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}