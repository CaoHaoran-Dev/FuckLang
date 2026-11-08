package evaluator

import (
	"fmt"
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

// newError 构造无行号的运行时错误（用于拿不到 AST 节点行号的地方）
func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

// newErrorAt 构造带行号的运行时错误
func newErrorAt(line int, format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...), Line: line}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR
	}
	return false
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
	case *ast.LoopStatement:
		return evalLoop(node, env)
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
	case *ast.HashLiteral:
		return evalHashLiteral(node, env)
	}
	return nil
}

func evalProgram(program *ast.Program, env *Environment) object.Object {
	var result object.Object = NULL
	for _, stmt := range program.Statements {
		result = Eval(stmt, env)
		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		case *object.BreakValue:
			return newError("`break` outside of loop")
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
			if t == object.RETURN_VALUE || t == object.ERROR || t == object.BREAK {
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
			// 内置函数返回的 Error 若没有行号，补上调用点行号
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
	return evaluated
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