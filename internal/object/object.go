package object

import (
	"fmt"
	"hash/fnv"
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
	ERROR        Type = "ERROR"
	FUNCTION     Type = "FUNCTION"
	BUILTIN      Type = "BUILTIN"
	ARRAY        Type = "ARRAY"
	HASH         Type = "HASH"
	BREAK        Type = "BREAK"
)

type Object interface {
	Type() Type
	Inspect() string
}

type Integer struct{ Value int64 }

func (i *Integer) Type() Type      { return INTEGER }
func (i *Integer) Inspect() string { return strconv.FormatInt(i.Value, 10) }

type String struct{ Value string }

func (s *String) Type() Type      { return STRING }
func (s *String) Inspect() string { return s.Value }

type Boolean struct{ Value bool }

func (b *Boolean) Type() Type      { return BOOLEAN }
func (b *Boolean) Inspect() string { return strconv.FormatBool(b.Value) }

type Null struct{}

func (n *Null) Type() Type      { return NULL }
func (n *Null) Inspect() string { return "null" }

type ReturnValue struct{ Value Object }

func (rv *ReturnValue) Type() Type      { return RETURN_VALUE }
func (rv *ReturnValue) Inspect() string { return rv.Value.Inspect() }

// BreakValue 在循环体内传播 break 信号
type BreakValue struct{}

func (bv *BreakValue) Type() Type      { return BREAK }
func (bv *BreakValue) Inspect() string { return "break" }

type Error struct {
	Message string
	Line    int
}

func (e *Error) Type() Type      { return ERROR }
func (e *Error) Inspect() string { return "ERROR: " + e.Message }

type Env interface{}

type Function struct {
	Parameters []*ast.Parameter
	Body       *ast.BlockStatement
	Env        Env
}

func (f *Function) Type() Type      { return FUNCTION }
func (f *Function) Inspect() string { return "bitch(...) {...}" }

type BuiltinFunction func(args ...Object) Object

type Builtin struct{ Fn BuiltinFunction }

func (b *Builtin) Type() Type      { return BUILTIN }
func (b *Builtin) Inspect() string { return "builtin function" }

type Array struct{ Elements []Object }

func (a *Array) Type() Type { return ARRAY }
func (a *Array) Inspect() string {
	var parts []string
	for _, e := range a.Elements {
		parts = append(parts, e.Inspect())
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

type HashKey struct {
	Type  Type
	Value uint64
}

func (b *Boolean) HashKey() HashKey {
	var value uint64
	if b.Value {
		value = 1
	} else {
		value = 0
	}
	return HashKey{Type: b.Type(), Value: value}
}

func (i *Integer) HashKey() HashKey {
	return HashKey{Type: i.Type(), Value: uint64(i.Value)}
}

func (s *String) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))
	return HashKey{Type: s.Type(), Value: h.Sum64()}
}

type HashPair struct {
	Key   Object
	Value Object
}

type Hash struct{ Pairs map[HashKey]HashPair }

func (h *Hash) Type() Type { return HASH }
func (h *Hash) Inspect() string {
	var pairs []string
	for _, pair := range h.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s", pair.Key.Inspect(), pair.Value.Inspect()))
	}
	return "{" + strings.Join(pairs, ", ") + "}"
}

type Hashable interface {
	HashKey() HashKey
}