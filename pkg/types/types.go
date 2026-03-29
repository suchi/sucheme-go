// Package types defines the core value types for the Scheme interpreter.
package types

import (
	"fmt"
	"math"
	"strings"
)

// ValueType represents the type tag for a Scheme value.
type ValueType int

const (
	TypeInteger ValueType = iota
	TypeFloat
	TypeBoolean
	TypeCharacter
	TypeString
	TypeSymbol
	TypePair
	TypeVector
	TypeNil
	TypeProcedure
	TypeVoid
	TypePort
	TypeEof
	TypeValues
	TypeContinuation
	TypeSyntaxRules
	TypeUndefined
	TypeBytevector
	TypeRecord
)

// String returns the name of the value type.
func (t ValueType) String() string {
	switch t {
	case TypeInteger:
		return "integer"
	case TypeFloat:
		return "float"
	case TypeBoolean:
		return "boolean"
	case TypeCharacter:
		return "character"
	case TypeString:
		return "string"
	case TypeSymbol:
		return "symbol"
	case TypePair:
		return "pair"
	case TypeVector:
		return "vector"
	case TypeNil:
		return "nil"
	case TypeProcedure:
		return "procedure"
	case TypeVoid:
		return "void"
	case TypePort:
		return "port"
	case TypeEof:
		return "eof"
	case TypeValues:
		return "values"
	case TypeContinuation:
		return "continuation"
	case TypeSyntaxRules:
		return "syntax-rules"
	case TypeUndefined:
		return "undefined"
	case TypeBytevector:
		return "bytevector"
	case TypeRecord:
		return "record"
	default:
		return "unknown"
	}
}

// Env is an interface for environments, to break the circular dependency
// between the types and env packages.
type Env interface{}

// Value represents a Scheme value as a tagged union.
type Value struct {
	Type      ValueType
	IntVal    int64
	FloatVal  float64
	BoolVal   bool
	CharVal   rune
	StrVal    string
	SymbolVal string
	PairVal   *Pair
	VecVal    []*Value
	ProcVal   *Procedure
	ByteVec   []byte
	ValuesVal []*Value      // for TypeValues (multiple return values)
	SyntaxVal *SyntaxRules  // for TypeSyntaxRules
	PortVal   any           // for TypePort (points to evaluator.Port)
	RecordVal *Record       // for TypeRecord
}

// Pair is a cons cell.
type Pair struct {
	Car *Value
	Cdr *Value
}

// LambdaParams describes the parameter list of a lambda.
type LambdaParams struct {
	Fixed   []string
	Rest    string
	HasRest bool
	IsRest  bool // entire params is a single rest param (lambda x body)
}

// Lambda represents a user-defined procedure.
type Lambda struct {
	Params LambdaParams
	Body   []*Value
	Env    Env
	Name   string
}

// SyntaxRules represents a compiled syntax-rules macro transformer.
type SyntaxRules struct {
	Name     string       // macro name (for error messages)
	Ellipsis string       // ellipsis symbol (default "...")
	Literals []string     // literal identifiers
	Rules    []SyntaxRule // (pattern template) pairs
	DefEnv   Env          // environment where macro was defined (used for hygienic expansion in future)
}

// SyntaxRule is a single (pattern template) pair in syntax-rules.
type SyntaxRule struct {
	Pattern  *Value // the pattern S-expression
	Template *Value // the template S-expression
}

// Record represents a record instance with a type tag and named fields.
type Record struct {
	TypeName string
	Fields   map[string]*Value
}

// NewRecord creates a record value.
func NewRecord(typeName string, fields map[string]*Value) *Value {
	return &Value{Type: TypeRecord, RecordVal: &Record{TypeName: typeName, Fields: fields}}
}

// NewSyntaxRules creates a syntax-rules value.
func NewSyntaxRules(sr *SyntaxRules) *Value {
	return &Value{Type: TypeSyntaxRules, SyntaxVal: sr}
}

// BuiltinFunc is the signature for builtin functions.
type BuiltinFunc func(args []*Value) (*Value, error)

// Procedure represents either a builtin or user-defined procedure.
type Procedure struct {
	Builtin   BuiltinFunc
	Lambda    *Lambda
	Name      string
	IsBuiltin bool
	ParamRef  any // for parameter objects (points to evaluator.Parameter)
}

// ---------------------------------------------------------------------------
// Constructors
// ---------------------------------------------------------------------------

// NewInteger creates an integer value.
func NewInteger(n int64) *Value {
	return &Value{Type: TypeInteger, IntVal: n}
}

// NewFloat creates a float value.
func NewFloat(f float64) *Value {
	return &Value{Type: TypeFloat, FloatVal: f}
}

// NewBoolean creates a boolean value.
func NewBoolean(b bool) *Value {
	if b {
		return True
	}
	return False
}

// True and False are canonical boolean values.
var (
	True  = &Value{Type: TypeBoolean, BoolVal: true}
	False = &Value{Type: TypeBoolean, BoolVal: false}
)

// NewChar creates a character value.
func NewChar(c rune) *Value {
	return &Value{Type: TypeCharacter, CharVal: c}
}

// NewString creates a string value.
func NewString(s string) *Value {
	return &Value{Type: TypeString, StrVal: s}
}

// NewSymbol creates a symbol value.
func NewSymbol(s string) *Value {
	return &Value{Type: TypeSymbol, SymbolVal: s}
}

// Cons creates a pair (cons cell).
func Cons(car, cdr *Value) *Value {
	return &Value{
		Type:    TypePair,
		PairVal: &Pair{Car: car, Cdr: cdr},
	}
}

// NewPair is an alias for Cons.
func NewPair(car, cdr *Value) *Value {
	return Cons(car, cdr)
}

// NewVector creates a vector value.
func NewVector(elems []*Value) *Value {
	return &Value{Type: TypeVector, VecVal: elems}
}

// NilVal is the canonical nil (empty list) value.
var NilVal = &Value{Type: TypeNil}

// Nil returns the empty list value.
func Nil() *Value {
	return NilVal
}

// VoidVal is the canonical void value.
var VoidVal = &Value{Type: TypeVoid}

// Void returns the void value.
func Void() *Value {
	return VoidVal
}

// EofVal is the canonical EOF value.
var EofVal = &Value{Type: TypeEof}

// Eof returns the EOF value.
func Eof() *Value {
	return EofVal
}

// UndefinedVal is the canonical undefined value.
var UndefinedVal = &Value{Type: TypeUndefined}

// Undefined returns the undefined value.
func Undefined() *Value {
	return UndefinedVal
}

// NewProcedure creates a builtin procedure value.
func NewProcedure(name string, fn BuiltinFunc) *Value {
	return &Value{
		Type: TypeProcedure,
		ProcVal: &Procedure{
			Builtin:   fn,
			Name:      name,
			IsBuiltin: true,
		},
	}
}

// NewLambdaProcedure creates a user-defined procedure value.
func NewLambdaProcedure(name string, lam *Lambda) *Value {
	return &Value{
		Type: TypeProcedure,
		ProcVal: &Procedure{
			Lambda:    lam,
			Name:      name,
			IsBuiltin: false,
		},
	}
}

// NewBytevector creates a bytevector value.
func NewBytevector(b []byte) *Value {
	return &Value{Type: TypeBytevector, ByteVec: b}
}

// NewValues creates a multiple-values value.
func NewValues(vals []*Value) *Value {
	return &Value{Type: TypeValues, ValuesVal: vals}
}

// ---------------------------------------------------------------------------
// List helpers
// ---------------------------------------------------------------------------

// FromList builds a proper list from a slice of values.
func FromList(vals []*Value) *Value {
	result := Nil()
	for i := len(vals) - 1; i >= 0; i-- {
		result = Cons(vals[i], result)
	}
	return result
}

// ToList converts a proper list value to a slice. Returns an error if
// the value is not a proper list.
func (v *Value) ToList() ([]*Value, error) {
	var result []*Value
	cur := v
	for cur.Type == TypePair {
		result = append(result, cur.PairVal.Car)
		cur = cur.PairVal.Cdr
	}
	if cur.Type != TypeNil {
		return nil, fmt.Errorf("not a proper list")
	}
	return result, nil
}

// IsList returns true if v is a proper list (nil-terminated chain of pairs).
func (v *Value) IsList() bool {
	_, err := v.ToList()
	return err == nil
}

// ListLength returns the length of a proper list, or -1 if not a proper list.
func (v *Value) ListLength() int {
	n := 0
	cur := v
	for cur.Type == TypePair {
		n++
		cur = cur.PairVal.Cdr
	}
	if cur.Type != TypeNil {
		return -1
	}
	return n
}

// ---------------------------------------------------------------------------
// Truth
// ---------------------------------------------------------------------------

// IsTrue returns true for all values except #f.
// In Scheme, only #f is false; everything else (including 0, "", '()) is true.
func (v *Value) IsTrue() bool {
	if v.Type == TypeBoolean && !v.BoolVal {
		return false
	}
	return true
}

// ---------------------------------------------------------------------------
// Equality
// ---------------------------------------------------------------------------

// Eq implements eq? (pointer/identity equality).
func (v *Value) Eq(other *Value) bool {
	if v == other {
		return true
	}
	// Same-valued booleans, nil, void, eof are eq?
	if v.Type != other.Type {
		return false
	}
	switch v.Type {
	case TypeBoolean:
		return v.BoolVal == other.BoolVal
	case TypeNil, TypeVoid, TypeEof, TypeUndefined:
		return true
	case TypeSymbol:
		return v.SymbolVal == other.SymbolVal
	case TypeInteger:
		return v.IntVal == other.IntVal
	case TypeCharacter:
		return v.CharVal == other.CharVal
	default:
		return false
	}
}

// Eqv implements eqv? (like eq? but also compares numbers and characters by value).
func (v *Value) Eqv(other *Value) bool {
	if v.Eq(other) {
		return true
	}
	if v.Type != other.Type {
		return false
	}
	switch v.Type {
	case TypeFloat:
		return v.FloatVal == other.FloatVal
	default:
		return false
	}
}

// Equal implements equal? (deep structural equality).
func (v *Value) Equal(other *Value) bool {
	if v.Eqv(other) {
		return true
	}
	if v.Type != other.Type {
		return false
	}
	switch v.Type {
	case TypeString:
		return v.StrVal == other.StrVal
	case TypePair:
		return v.PairVal.Car.Equal(other.PairVal.Car) && v.PairVal.Cdr.Equal(other.PairVal.Cdr)
	case TypeVector:
		if len(v.VecVal) != len(other.VecVal) {
			return false
		}
		for i, el := range v.VecVal {
			if !el.Equal(other.VecVal[i]) {
				return false
			}
		}
		return true
	case TypeBytevector:
		if len(v.ByteVec) != len(other.ByteVec) {
			return false
		}
		for i, b := range v.ByteVec {
			if b != other.ByteVec[i] {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// ---------------------------------------------------------------------------
// Display / String
// ---------------------------------------------------------------------------

// Display returns the Scheme display representation (no quotes on strings).
func (v *Value) Display() string {
	switch v.Type {
	case TypeString:
		return v.StrVal
	case TypeCharacter:
		return string(v.CharVal)
	default:
		return v.String()
	}
}

// String returns the Scheme write representation (debug-friendly).
func (v *Value) String() string {
	switch v.Type {
	case TypeInteger:
		return fmt.Sprintf("%d", v.IntVal)
	case TypeFloat:
		return formatFloat(v.FloatVal)
	case TypeBoolean:
		if v.BoolVal {
			return "#t"
		}
		return "#f"
	case TypeCharacter:
		return formatChar(v.CharVal)
	case TypeString:
		return fmt.Sprintf("%q", v.StrVal)
	case TypeSymbol:
		return v.SymbolVal
	case TypePair:
		return formatList(v)
	case TypeVector:
		return formatVector(v.VecVal)
	case TypeNil:
		return "()"
	case TypeProcedure:
		if v.ProcVal != nil && v.ProcVal.Name != "" {
			return fmt.Sprintf("#<procedure %s>", v.ProcVal.Name)
		}
		return "#<procedure>"
	case TypeVoid:
		return "#<void>"
	case TypePort:
		return "#<port>"
	case TypeEof:
		return "#<eof>"
	case TypeValues:
		return "#<values>"
	case TypeContinuation:
		return "#<continuation>"
	case TypeSyntaxRules:
		return "#<syntax-rules>"
	case TypeUndefined:
		return "#<undefined>"
	case TypeBytevector:
		return formatBytevector(v.ByteVec)
	case TypeRecord:
		if v.RecordVal != nil {
			return fmt.Sprintf("#<%s>", v.RecordVal.TypeName)
		}
		return "#<record>"
	default:
		return "#<unknown>"
	}
}

func formatFloat(f float64) string {
	if math.IsInf(f, 1) {
		return "+inf.0"
	}
	if math.IsInf(f, -1) {
		return "-inf.0"
	}
	if math.IsNaN(f) {
		return "+nan.0"
	}
	s := fmt.Sprintf("%g", f)
	// Ensure there's a decimal point so it reads as a float
	if !strings.Contains(s, ".") && !strings.Contains(s, "e") && !strings.Contains(s, "E") {
		s += ".0"
	}
	return s
}

func formatChar(c rune) string {
	switch c {
	case ' ':
		return `#\space`
	case '\n':
		return `#\newline`
	case '\t':
		return `#\tab`
	case '\r':
		return `#\return`
	case '\x00':
		return `#\null`
	case '\a':
		return `#\alarm`
	case '\b':
		return `#\backspace`
	case '\x1b':
		return `#\escape`
	case '\x7f':
		return `#\delete`
	default:
		return fmt.Sprintf(`#\%c`, c)
	}
}

func formatList(v *Value) string {
	var buf strings.Builder
	buf.WriteByte('(')
	cur := v
	first := true
	for cur.Type == TypePair {
		if !first {
			buf.WriteByte(' ')
		}
		first = false
		buf.WriteString(cur.PairVal.Car.String())
		cur = cur.PairVal.Cdr
	}
	if cur.Type != TypeNil {
		buf.WriteString(" . ")
		buf.WriteString(cur.String())
	}
	buf.WriteByte(')')
	return buf.String()
}

func formatVector(elems []*Value) string {
	var buf strings.Builder
	buf.WriteString("#(")
	for i, e := range elems {
		if i > 0 {
			buf.WriteByte(' ')
		}
		buf.WriteString(e.String())
	}
	buf.WriteByte(')')
	return buf.String()
}

func formatBytevector(b []byte) string {
	var buf strings.Builder
	buf.WriteString("#u8(")
	for i, x := range b {
		if i > 0 {
			buf.WriteByte(' ')
		}
		fmt.Fprintf(&buf, "%d", x)
	}
	buf.WriteByte(')')
	return buf.String()
}
