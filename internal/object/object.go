package object

import (
	"strconv"
	"strings"

	"fucklang/internal/ast"
)

type Type string

const (
	INTEGER      Type = "INTEGER"
	STRING       Type = "STRING"
	BOOLEAN      Type = "BOOLEAN"
	NULL         Type = "NULL"
	RETURN_VALUE Type = "RETURN_VALUE"
	FUNCTION     Type = "FUNCTION"
	ARRAY        Type = "ARRAY"
)

type Object interface {
	Type() Type
	Inspect() string
}

type Integer struct {
	Value int64
}

func (i *Integer) Type() Type      { return INTEGER }
func (i *Integer) Inspect() string { return strconv.FormatInt(i.Value, 10) }

type String struct {
	Value string
}

func (s *String) Type() Type      { return STRING }
func (s *String) Inspect() string { return s.Value }

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() Type      { return BOOLEAN }
func (b *Boolean) Inspect() string { return strconv.FormatBool(b.Value) }

type Null struct{}

func (n *Null) Type() Type      { return NULL }
func (n *Null) Inspect() string { return "null" }

type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() Type      { return RETURN_VALUE }
func (rv *ReturnValue) Inspect() string { return rv.Value.Inspect() }

type Function struct {
	Parameters []*ast.Parameter
	Body       *ast.BlockStatement
	Env        interface{} // *evaluator.Environment，避免循环依赖用 interface
}

func (f *Function) Type() Type      { return FUNCTION }
func (f *Function) Inspect() string { return "bitch(...) {...}" }

type Array struct {
	Elements []Object
}

func (a *Array) Type() Type { return ARRAY }
func (a *Array) Inspect() string {
	var parts []string
	for _, e := range a.Elements {
		parts = append(parts, e.Inspect())
	}
	return "[" + strings.Join(parts, ", ") + "]"
}