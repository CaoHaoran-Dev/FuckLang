package evaluator

import (
	"fucklang/internal/ast"
	"fucklang/internal/object"
	"fmt"
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
	return &Environment{store: make(map[string]object.Object)}
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
		if val != nil {
			fmt.Println(val.Inspect())
		}
		return nil
	case *ast.AssignStatement:
		val := Eval(node.Value, env)
		env.Set(node.Name.Value, val)
		return val
	case *ast.IndexAssignStatement:
		return evalIndexAssign(node, env)
	case *ast.ReturnStatement:
		val := Eval(node.Value, env)
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
		return NULL
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		right := Eval(node.Right, env)
		return evalInfixExpression(node.Operator, left, right)
	case *ast.FunctionLiteral:
		return &object.Function{
			Parameters: node.Parameters,
			Body:       node.Body,
			Env:        env,
		}
	case *ast.CallExpression:
		function := Eval(node.Function, env)
		args := evalExpressions(node.Arguments, env)
		return applyFunction(function, args)
	case *ast.ArrayLiteral:
		elements := evalExpressions(node.Elements, env)
		return &object.Array{Elements: elements}
	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		index := Eval(node.Index, env)
		return evalIndexExpression(left, index)
	}
	return nil
}

func evalProgram(program *ast.Program, env *Environment) object.Object {
	var result object.Object
	for _, stmt := range program.Statements {
		result = Eval(stmt, env)
		if rv, ok := result.(*object.ReturnValue); ok {
			return rv.Value
		}
	}
	return result
}

func evalBlock(block *ast.BlockStatement, env *Environment) object.Object {
	var result object.Object
	for _, stmt := range block.Statements {
		result = Eval(stmt, env)
		if rv, ok := result.(*object.ReturnValue); ok {
			return rv
		}
	}
	return result
}

func evalLoop(ls *ast.LoopStatement, env *Environment) object.Object {
	for {
		cond := Eval(ls.Condition, env)
		if !isTruthy(cond) {
			break
		}
		result := Eval(ls.Body, env)
		if rv, ok := result.(*object.ReturnValue); ok {
			return rv
		}
	}
	return NULL
}

func evalIfExpression(ie *ast.IfExpression, env *Environment) object.Object {
	condition := Eval(ie.Condition, env)
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
		if rv, ok := evaluated.(*object.ReturnValue); ok {
			return []object.Object{rv.Value}
		}
		result = append(result, evaluated)
	}
	return result
}

func applyFunction(fn object.Object, args []object.Object) object.Object {
	function, ok := fn.(*object.Function)
	if !ok {
		return NULL
	}
	outer, _ := function.Env.(*Environment)
	extendedEnv := NewEnclosedEnvironment(outer)
	for i, param := range function.Parameters {
		if i < len(args) {
			extendedEnv.Set(param.Name.Value, args[i])
		}
	}
	evaluated := Eval(function.Body, extendedEnv)
	if rv, ok := evaluated.(*object.ReturnValue); ok {
		return rv.Value
	}
	return evaluated
}

func evalIndexExpression(left, index object.Object) object.Object {
	if left.Type() == object.ARRAY && index.Type() == object.INTEGER {
		arr := left.(*object.Array)
		idx := index.(*object.Integer).Value
		if idx < 0 || int(idx) >= len(arr.Elements) {
			return NULL
		}
		return arr.Elements[idx]
	}
	return NULL
}

func evalIndexAssign(node *ast.IndexAssignStatement, env *Environment) object.Object {
	left := Eval(node.Left, env)
	index := Eval(node.Index, env)
	value := Eval(node.Value, env)
	arr, ok := left.(*object.Array)
	if !ok || index.Type() != object.INTEGER {
		return NULL
	}
	idx := index.(*object.Integer).Value
	if idx < 0 || int(idx) >= len(arr.Elements) {
		return NULL
	}
	arr.Elements[idx] = value
	return value
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
	}
	return NULL
}

func evalInfixExpression(operator string, left, right object.Object) object.Object {
	if left.Type() == object.INTEGER && right.Type() == object.INTEGER {
		return evalIntegerInfix(operator, left, right)
	}
	if left.Type() == object.STRING && right.Type() == object.STRING {
		return evalStringInfix(operator, left, right)
	}
	switch operator {
	case "==":
		return nativeBoolToBooleanObject(left == right)
	case "!=":
		return nativeBoolToBooleanObject(left != right)
	}
	return NULL
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
			return NULL
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
	return NULL
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
	return NULL
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