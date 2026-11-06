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

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
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
		return nil
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
		return newError("identifier not found: " + node.Value)
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)
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
		return applyFunction(function, args)
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
		return evalIndexExpression(left, index)
	case *ast.HashLiteral:
		return evalHashLiteral(node, env)
	}
	return nil
}

func evalProgram(program *ast.Program, env *Environment) object.Object {
	var result object.Object
	for _, stmt := range program.Statements {
		result = Eval(stmt, env)
		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}
	return result
}

func evalBlock(block *ast.BlockStatement, env *Environment) object.Object {
	var result object.Object
	for _, stmt := range block.Statements {
		result = Eval(stmt, env)
		if result != nil {
			if result.Type() == object.RETURN_VALUE || result.Type() == object.ERROR {
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
			if result.Type() == object.RETURN_VALUE || result.Type() == object.ERROR {
				return result
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

func applyFunction(fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		outer, _ := fn.Env.(*Environment)
		extendedEnv := NewEnclosedEnvironment(outer)
		for i, param := range fn.Parameters {
			if i < len(args) {
				extendedEnv.Set(param.Name.Value, args[i])
			}
		}
		evaluated := Eval(fn.Body, extendedEnv)
		if rv, ok := evaluated.(*object.ReturnValue); ok {
			return rv.Value
		}
		return evaluated
	case *object.Builtin:
		return fn.Fn(args...)
	default:
		return newError("not a function: %s", fn.Type())
	}
}

func evalIndexExpression(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY && index.Type() == object.INTEGER:
		arr := left.(*object.Array)
		idx := index.(*object.Integer).Value
		if idx < 0 || int(idx) >= len(arr.Elements) {
			return newError("index out of bounds: %d", idx)
		}
		return arr.Elements[idx]
	case left.Type() == object.HASH:
		hash := left.(*object.Hash)
		key, ok := index.(object.Hashable)
		if !ok {
			return newError("unusable as hash key: %s", index.Type())
		}
		pair, ok := hash.Pairs[key.HashKey()]
		if !ok {
			return NULL
		}
		return pair.Value
	default:
		return newError("index operator not supported: %s", left.Type())
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

	if arr, ok := left.(*object.Array); ok {
		if index.Type() != object.INTEGER {
			return newError("array index must be INTEGER")
		}
		idx := index.(*object.Integer).Value
		if idx < 0 || int(idx) >= len(arr.Elements) {
			return newError("index out of bounds: %d", idx)
		}
		arr.Elements[idx] = value
		return value
	}
	return newError("index assignment not supported")
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return nativeBoolToBooleanObject(!isTruthy(right))
	case "-":
		if right.Type() == object.INTEGER {
			val := right.(*object.Integer).Value
			return &object.Integer{Value: -val}
		}
		return newError("unknown operator: -%s", right.Type())
	}
	return newError("unknown operator: %s%s", operator, right.Type())
}

func evalInfixExpression(operator string, left, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER && right.Type() == object.INTEGER:
		return evalIntegerInfix(operator, left, right)
	case left.Type() == object.STRING && right.Type() == object.STRING:
		return evalStringInfix(operator, left, right)
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalIntegerInfix(operator string, left, right object.Object) object.Object {
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
			return newError("division by zero")
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
	return newError("unknown operator: %s", operator)
}

func evalStringInfix(operator string, left, right object.Object) object.Object {
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
	return newError("unknown operator: %s", operator)
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