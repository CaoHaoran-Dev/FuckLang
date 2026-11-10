package evaluator

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"fucklang/internal/object"
)

var ioBuiltins = map[string]*object.Builtin{
	"suck": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newError("wrong number of arguments. got=%d, want=0", len(args))
			}
			reader := bufio.NewReader(os.Stdin)
			line, err := reader.ReadString('\n')
			if err != nil && line == "" {
				return &object.String{Value: ""}
			}
			return &object.String{Value: strings.TrimRight(line, "\r\n")}
		},
	},
	"fucking": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments. got=%d, want>=1", len(args))
			}
			format, ok := args[0].(*object.String)
			if !ok {
				return newError("first argument to `fucking` must be STRING")
			}
			converted := make([]interface{}, 0, len(args)-1)
			for _, a := range args[1:] {
				converted = append(converted, objectToGo(a))
			}
			fmt.Println(fmt.Sprintf(format.Value, converted...))
			return NULL
		},
	},
	"fuckoff": {
		Fn: func(args ...object.Object) object.Object {
			code := 0
			if len(args) >= 1 {
				if i, ok := args[0].(*object.Integer); ok {
					code = int(i.Value)
				} else {
					return newError("argument to `fuckoff` must be INTEGER")
				}
			}
			if len(args) == 2 {
				if s, ok := args[1].(*object.String); ok {
					fmt.Fprintln(os.Stderr, s.Value)
				}
			}
			os.Exit(code)
			return NULL
		},
	},
}

func objectToGo(o object.Object) interface{} {
	switch v := o.(type) {
	case *object.Integer:
		return v.Value
	case *object.String:
		return v.Value
	case *object.Boolean:
		return v.Value
	case *object.Null:
		return nil
	default:
		return v.Inspect()
	}
}