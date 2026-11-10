package evaluator

import "fucklang/internal/object"

var arrBuiltins = map[string]*object.Builtin{
	"len": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			arr, ok := args[0].(*object.Array)
			if !ok {
				return newError("argument to `arr.len` must be ARRAY, got %s", args[0].Type())
			}
			return &object.Integer{Value: int64(len(arr.Elements))}
		},
	},
	"push": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			arr, ok := args[0].(*object.Array)
			if !ok {
				return newError("argument to `arr.push` must be ARRAY")
			}
			arr.Elements = append(arr.Elements, args[1])
			return arr
		},
	},
	"pop": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			arr, ok := args[0].(*object.Array)
			if !ok {
				return newError("argument to `arr.pop` must be ARRAY")
			}
			if len(arr.Elements) == 0 {
				return NULL
			}
			last := arr.Elements[len(arr.Elements)-1]
			arr.Elements = arr.Elements[:len(arr.Elements)-1]
			return last
		},
	},
	"slice": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newError("wrong number of arguments. got=%d, want=3", len(args))
			}
			arr, ok := args[0].(*object.Array)
			if !ok {
				return newError("first argument to `arr.slice` must be ARRAY")
			}
			start, ok1 := args[1].(*object.Integer)
			end, ok2 := args[2].(*object.Integer)
			if !ok1 || !ok2 {
				return newError("indices to `arr.slice` must be INTEGER")
			}
			s, e := start.Value, end.Value
			if s < 0 || e > int64(len(arr.Elements)) || s > e {
				return newError("arr.slice: index out of bounds [%d:%d]", s, e)
			}
			out := make([]object.Object, e-s)
			copy(out, arr.Elements[s:e])
			return &object.Array{Elements: out}
		},
	},
	"contains": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			arr, ok := args[0].(*object.Array)
			if !ok {
				return newError("first argument to `arr.contains` must be ARRAY")
			}
			for _, e := range arr.Elements {
				if objectsEqual(e, args[1]) {
					return TRUE
				}
			}
			return FALSE
		},
	},
	"keys": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			h, ok := args[0].(*object.Hash)
			if !ok {
				return newError("argument to `arr.keys` must be HASH")
			}
			out := make([]object.Object, 0, len(h.Pairs))
			for _, p := range h.Pairs {
				out = append(out, p.Key)
			}
			return &object.Array{Elements: out}
		},
	},
	"values": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			h, ok := args[0].(*object.Hash)
			if !ok {
				return newError("argument to `arr.values` must be HASH")
			}
			out := make([]object.Object, 0, len(h.Pairs))
			for _, p := range h.Pairs {
				out = append(out, p.Value)
			}
			return &object.Array{Elements: out}
		},
	},
}

func objectsEqual(a, b object.Object) bool {
	if a.Type() != b.Type() {
		return false
	}
	switch av := a.(type) {
	case *object.Integer:
		return av.Value == b.(*object.Integer).Value
	case *object.String:
		return av.Value == b.(*object.String).Value
	case *object.Boolean:
		return av.Value == b.(*object.Boolean).Value
	default:
		return a == b
	}
}