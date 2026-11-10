package evaluator

import "fucklang/internal/object"

var builtins = map[string]*object.Builtin{}

func init() {
	for k, v := range coreBuiltins {
		builtins[k] = v
	}
	for k, v := range ioBuiltins {
		builtins[k] = v
	}
}