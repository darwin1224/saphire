package interpreter

import (
	"fmt"
	"io"
	"math"
	"os"

	"github.com/darwin1224/saphire/ast"
	"github.com/darwin1224/saphire/object"
)

var (
	NIL   = &object.Nil{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

// Interpreter owns the built-ins for one execution session.
type Interpreter struct {
	builtins map[string]*object.Builtin
}

// New creates an interpreter whose print built-in writes to out.
func New(out io.Writer) *Interpreter {
	if out == nil {
		out = io.Discard
	}
	sessionBuiltins := make(map[string]*object.Builtin, len(builtins)+1)
	for name, builtin := range builtins {
		sessionBuiltins[name] = builtin
	}
	sessionBuiltins["print"] = &object.Builtin{Fn: printBuiltin(out)}
	return &Interpreter{builtins: sessionBuiltins}
}

// Eval evaluates using standard output. Use New to supply a session output writer.
func Eval(node ast.Node, env *object.Environment) object.Object {
	return New(os.Stdout).Eval(node, env)
}

func (i *Interpreter) Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return i.evalProgram(node, env)
	case *ast.ExpressionStatement:
		return i.Eval(node.Expression, env)
	case *ast.UnaryExpression:
		right := i.Eval(node.Right, env)
		if isError(right) {
			return right
		}

		return evalUnaryExpression(node.Operator, right)
	case *ast.BinaryExpression:
		left := i.Eval(node.Left, env)
		if isError(left) {
			return left
		}

		right := i.Eval(node.Right, env)
		if isError(right) {
			return right
		}

		return evalBinaryExpression(node.Operator, left, right)
	case *ast.LetStatement:
		val := i.Eval(node.Value, env)
		if isError(val) {
			return val
		}

		env.Set(node.Name.Value, val)
		return NIL
	case *ast.IfExpression:
		return i.evalIfExpression(node, env)
	case *ast.BlockStatement:
		return i.evalBlockStatement(node, env)
	case *ast.ReturnStatement:
		val := i.Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}

		return &object.ReturnValue{Value: val}
	case *ast.Identifier:
		return i.evalIdentifier(node, env)
	case *ast.NumberLiteral:
		return &object.Number{Value: node.Value}
	case *ast.Boolean:
		return boolToBooleanObject(node.Value)
	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Env: env, Body: body}
	case *ast.CallExpression:
		function := i.Eval(node.Function, env)
		if isError(function) {
			return function
		}

		args := i.evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		return i.applyFunction(function, args)
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.ArrayLiteral:
		elems := i.evalExpressions(node.Elements, env)
		if len(elems) == 1 && isError(elems[0]) {
			return elems[0]
		}

		return &object.Array{Elements: elems}
	case *ast.IndexExpression:
		left := i.Eval(node.Left, env)
		if isError(left) {
			return left
		}

		index := i.Eval(node.Index, env)
		if isError(index) {
			return index
		}

		return evalIndexExpression(left, index)
	case *ast.HashLiteral:
		return i.evalHashLiteral(node, env)
	}

	return NIL
}

func (i *Interpreter) evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object = NIL

	for _, statement := range program.Statements {
		result = i.Eval(statement, env)

		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}

func (i *Interpreter) evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object = NIL

	for _, statement := range block.Statements {
		result = i.Eval(statement, env)

		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}

	return result
}

func evalUnaryExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusUnaryOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NIL:
		return TRUE
	default:
		return FALSE
	}
}

func evalMinusUnaryOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.NUMBER_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}

	value := right.(*object.Number).Value
	return &object.Number{Value: -value}
}

func evalBinaryExpression(operator string, left, right object.Object) object.Object {
	switch {
	case left.Type() == object.NUMBER_OBJ && right.Type() == object.NUMBER_OBJ:
		return evalNumberBinaryExpression(operator, left, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringBinaryExpression(operator, left, right)
	case operator == "==":
		return boolToBooleanObject(left == right)
	case operator == "!=":
		return boolToBooleanObject(left != right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalNumberBinaryExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Number).Value
	rightVal := right.(*object.Number).Value

	switch operator {
	case "+":
		return &object.Number{Value: leftVal + rightVal}
	case "-":
		return &object.Number{Value: leftVal - rightVal}
	case "*":
		return &object.Number{Value: leftVal * rightVal}
	case "**":
		return &object.Number{Value: math.Pow(leftVal, rightVal)}
	case "/":
		return &object.Number{Value: leftVal / rightVal}
	case "%":
		if rightVal == 0 {
			return newError("modulo by zero")
		}
		return &object.Number{Value: math.Mod(leftVal, rightVal)}
	case "<":
		return boolToBooleanObject(leftVal < rightVal)
	case ">":
		return boolToBooleanObject(leftVal > rightVal)
	case "<=":
		return boolToBooleanObject(leftVal <= rightVal)
	case ">=":
		return boolToBooleanObject(leftVal >= rightVal)
	case "==":
		return boolToBooleanObject(leftVal == rightVal)
	case "!=":
		return boolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalStringBinaryExpression(operator string, left, right object.Object) object.Object {
	if operator != "+" {
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}

	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value
	return &object.String{Value: leftVal + rightVal}
}

func (i *Interpreter) evalIfExpression(ie *ast.IfExpression, env *object.Environment) object.Object {
	condition := i.Eval(ie.Condition, env)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return i.Eval(ie.Consequence, env)
	}
	if ie.Alternative != nil {
		return i.Eval(ie.Alternative, env)
	}

	return NIL
}

func (i *Interpreter) evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}

	if builtin, ok := i.builtins[node.Value]; ok {
		return builtin
	}

	return newError("identifier not found: %s", node.Value)
}

func (i *Interpreter) evalExpressions(exps []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object

	for _, e := range exps {
		evaluated := i.Eval(e, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

func evalIndexExpression(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.NUMBER_OBJ:
		return evalArrayIndexExpression(left, index)
	case left.Type() == object.HASH_OBJ:
		return evalHashIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func evalArrayIndexExpression(array, index object.Object) object.Object {
	arrayObject := array.(*object.Array)
	idx := index.(*object.Number).Value
	max := float64(len(arrayObject.Elements) - 1)

	if math.IsNaN(idx) || math.IsInf(idx, 0) || math.Trunc(idx) != idx {
		return newError("array index must be a finite integer")
	}
	if idx < 0 || idx > max {
		return NIL
	}
	return arrayObject.Elements[int64(idx)]
}

func (i *Interpreter) evalHashLiteral(node *ast.HashLiteral, env *object.Environment) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)

	for _, pair := range node.Pairs {
		key := i.Eval(pair.Key, env)
		if isError(key) {
			return key
		}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return newError("unusable as hash key: %s", key.Type())
		}
		if number, ok := key.(*object.Number); ok && math.IsNaN(number.Value) {
			return newError("unusable as hash key: NaN")
		}

		value := i.Eval(pair.Value, env)
		if isError(value) {
			return value
		}

		hashed := hashKey.HashKey()
		pairs[hashed] = object.HashPair{Key: key, Value: value}
	}

	return &object.Hash{Pairs: pairs}
}

func evalHashIndexExpression(hash, index object.Object) object.Object {
	hashObject := hash.(*object.Hash)

	key, ok := index.(object.Hashable)
	if !ok {
		return newError("unusable as hash key: %s", index.Type())
	}
	if number, ok := index.(*object.Number); ok && math.IsNaN(number.Value) {
		return newError("unusable as hash key: NaN")
	}

	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return NIL
	}

	return pair.Value
}

func boolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func isTruthy(obj object.Object) bool {
	switch obj {
	case NIL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

func (i *Interpreter) applyFunction(fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		if len(args) != len(fn.Parameters) {
			return newError("wrong number of arguments, got=%d, want=%d", len(args), len(fn.Parameters))
		}
		extendedEnv := extendFunctionEnv(fn, args)
		evaluated := i.Eval(fn.Body, extendedEnv)
		return unwrapReturnValue(evaluated)
	case *object.Builtin:
		return fn.Fn(args...)
	default:
		return newError("not a function: %s", fn.Type())
	}
}

func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	for paramIdx, param := range fn.Parameters {
		env.Set(param.Value, args[paramIdx])
	}

	return env
}

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}
	return obj
}
