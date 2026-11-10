package evaluator

import (
	"strings"

	"fucklang/internal/object"
)

var strBuiltins = map[string]*object.Builtin{
	"len": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			s, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `str.len` must be STRING, got %s", args[0].Type())
			}
			return &object.Integer{Value: int64(len(s.Value))}
		},
	},
	"upper": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			s, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `str.upper` must be STRING")
			}
			return &object.String{Value: strings.ToUpper(s.Value)}
		},
	},
	"lower": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			s, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `str.lower` must be STRING")
			}
			return &object.String{Value: strings.ToLower(s.Value)}
		},
	},
	"trim": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			s, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `str.trim` must be STRING")
			}
			return &object.String{Value: strings.TrimSpace(s.Value)}
		},
	},
	"split": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			s, ok1 := args[0].(*object.String)
			sep, ok2 := args[1].(*object.String)
			if !ok1 || !ok2 {
				return newError("arguments to `str.split` must be STRING")
			}
			parts := strings.Split(s.Value, sep.Value)
			elems := make([]object.Object, len(parts))
			for i, p := range parts {
				elems[i] = &object.String{Value: p}
			}
			return &object.Array{Elements: elems}
		},
	},
	"contains": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			s, ok1 := args[0].(*object.String)
			sub, ok2 := args[1].(*object.String)
			if !ok1 || !ok2 {
				return newError("arguments to `str.contains` must be STRING")
			}
			return nativeBoolToBooleanObject(strings.Contains(s.Value, sub.Value))
		},
	},
	"replace": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newError("wrong number of arguments. got=%d, want=3", len(args))
			}
			s, ok1 := args[0].(*object.String)
			old, ok2 := args[1].(*object.String)
			nw, ok3 := args[2].(*object.String)
			if !ok1 || !ok2 || !ok3 {
				return newError("arguments to `str.replace` must be STRING")
			}
			return &object.String{Value: strings.ReplaceAll(s.Value, old.Value, nw.Value)}
		},
	},
	"join": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			arr, ok1 := args[0].(*object.Array)
			sep, ok2 := args[1].(*object.String)
			if !ok1 || !ok2 {
				return newError("arguments to `str.join` must be ARRAY, STRING")
			}
			parts := make([]string, len(arr.Elements))
			for i, e := range arr.Elements {
				s, ok := e.(*object.String)
				if !ok {
					return newError("str.join: element %d is not STRING", i)
				}
				parts[i] = s.Value
			}
			return &object.String{Value: strings.Join(parts, sep.Value)}
		},
	},
}