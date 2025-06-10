package object

import (
	"fmt"
)

var Builtins = []struct {
	Name    string
	Builtin *Builtin
}{
	{"len", &Builtin{Fn: lenBuiltin}},
	{"first", &Builtin{Fn: firstBuiltin}},
	{"last", &Builtin{Fn: lastBuiltin}},
	{"rest", &Builtin{Fn: restBuiltin}},
	{"push", &Builtin{Fn: pushBuiltin}},
	{"print", &Builtin{Fn: printBuiltin}},
}

func lenBuiltin(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments, got=%d, want=1", len(args))
	}

	switch arg := args[0].(type) {
	case *String:
		return &Number{Value: float64(len(arg.Value))}
	case *Array:
		return &Number{Value: float64(len(arg.Elements))}
	case *Hash:
		return &Number{Value: float64(len(arg.Pairs))}
	default:
		return newError("argument to `len` not supported, got %s", args[0].Type())
	}
}

func firstBuiltin(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	if args[0].Type() != ARRAY_OBJ {
		return newError("argument to `first` must be ARRAY, got %s", args[0].Type())
	}

	arr := args[0].(*Array)
	if len(arr.Elements) > 0 {
		return arr.Elements[0]
	}

	return nil
}

func lastBuiltin(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	if args[0].Type() != ARRAY_OBJ {
		return newError("argument to `last` must be ARRAY, got %s", args[0].Type())
	}

	arr := args[0].(*Array)
	length := len(arr.Elements)
	if length > 0 {
		return arr.Elements[length-1]
	}

	return nil
}

func restBuiltin(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	if args[0].Type() != ARRAY_OBJ {
		return newError("argument to `rest` must be ARRAY, got %s", args[0].Type())
	}

	arr := args[0].(*Array)
	length := len(arr.Elements)
	if length > 0 {
		newElements := make([]Object, length-1, length-1)
		copy(newElements, arr.Elements[1:length])
		return &Array{Elements: newElements}
	}

	return nil
}

func pushBuiltin(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments. got=%d, want=2", len(args))
	}
	if args[0].Type() != ARRAY_OBJ {
		return newError("argument to `push` must be ARRAY, got %s", args[0].Type())
	}

	arr := args[0].(*Array)
	length := len(arr.Elements)

	newElements := make([]Object, length+1, length+1)
	copy(newElements, arr.Elements)
	newElements[length] = args[1]

	return &Array{Elements: newElements}
}

func printBuiltin(args ...Object) Object {
	for _, arg := range args {
		fmt.Println(arg.Inspect())
	}
	return nil
}

func newError(format string, a ...interface{}) *Error {
	return &Error{Message: fmt.Sprintf(format, a...)}
}

func GetBuiltinByName(name string) *Builtin {
	for _, def := range Builtins {
		if def.Name == name {
			return def.Builtin
		}
	}
	return nil
}
