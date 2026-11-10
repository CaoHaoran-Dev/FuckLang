package evaluator

import (
	"os"

	"fucklang/internal/object"
)

var fileBuiltins = map[string]*object.Builtin{
	"read": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			path, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `file.read` must be STRING")
			}
			data, err := os.ReadFile(path.Value)
			if err != nil {
				return newError("file.read: %s", err.Error())
			}
			return &object.String{Value: string(data)}
		},
	},
	"write": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			path, ok1 := args[0].(*object.String)
			content, ok2 := args[1].(*object.String)
			if !ok1 || !ok2 {
				return newError("arguments to `file.write` must be STRING")
			}
			if err := os.WriteFile(path.Value, []byte(content.Value), 0644); err != nil {
				return newError("file.write: %s", err.Error())
			}
			return NULL
		},
	},
	"exists": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			path, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `file.exists` must be STRING")
			}
			_, err := os.Stat(path.Value)
			return nativeBoolToBooleanObject(err == nil)
		},
	},
}