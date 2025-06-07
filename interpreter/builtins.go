package interpreter

import (
	"fmt"

	"github.com/darwin1224/saphire/object"
)

var builtins = map[string]*object.Builtin{
	"len":   &object.Builtin{Fn: lenBuiltin},
	"first": &object.Builtin{Fn: firstBuiltin},
	"last":  &object.Builtin{Fn: lastBuiltin},
	"rest":  &object.Builtin{Fn: restBuiltin},
	"push":  &object.Builtin{Fn: pushBuiltin},
	"print": &object.Builtin{Fn: printBuiltin},
}

func lenBuiltin(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments, got=%d, want=1", len(args))
	}

	switch arg := args[0].(type) {
	case *object.String:
		return &object.Number{Value: float64(len(arg.Value))}
	case *object.Array:
		return &object.Number{Value: float64(len(arg.Elements))}
	case *object.Hash:
		return &object.Number{Value: float64(len(arg.Pairs))}
	default:
		return newError("argument to `len` not supported, got %s", args[0].Type())
	}
}

func firstBuiltin(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	if args[0].Type() != object.ARRAY_OBJ {
		return newError("argument to `first` must be ARRAY, got %s", args[0].Type())
	}

	arr := args[0].(*object.Array)
	if len(arr.Elements) > 0 {
		return arr.Elements[0]
	}

	return NIL
}

func lastBuiltin(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	if args[0].Type() != object.ARRAY_OBJ {
		return newError("argument to `last` must be ARRAY, got %s", args[0].Type())
	}

	arr := args[0].(*object.Array)
	length := len(arr.Elements)
	if length > 0 {
		return arr.Elements[length-1]
	}

	return NIL
}

func restBuiltin(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	if args[0].Type() != object.ARRAY_OBJ {
		return newError("argument to `rest` must be ARRAY, got %s", args[0].Type())
	}

	arr := args[0].(*object.Array)
	length := len(arr.Elements)
	if length > 0 {
		newElements := make([]object.Object, length-1, length-1)
		copy(newElements, arr.Elements[1:length])
		return &object.Array{Elements: newElements}
	}

	return NIL
}

func pushBuiltin(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("wrong number of arguments. got=%d, want=2", len(args))
	}
	if args[0].Type() != object.ARRAY_OBJ {
		return newError("argument to `push` must be ARRAY, got %s", args[0].Type())
	}

	arr := args[0].(*object.Array)
	length := len(arr.Elements)

	newElements := make([]object.Object, length+1, length+1)
	copy(newElements, arr.Elements)
	newElements[length] = args[1]

	return &object.Array{Elements: newElements}
}

func printBuiltin(args ...object.Object) object.Object {
	for _, arg := range args {
		fmt.Println(arg.Inspect())
	}
	return NIL
}
