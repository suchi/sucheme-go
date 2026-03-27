package builtin

import (
	"testing"

	"github.com/suchi/sucheme-go/pkg/env"
	"github.com/suchi/sucheme-go/pkg/types"
)

func setup() *env.Env {
	e := env.Global()
	RegisterBuiltins(e)
	return e
}

func call(t *testing.T, e *env.Env, name string, args ...*types.Value) *types.Value {
	t.Helper()
	v, ok := e.Get(name)
	if !ok {
		t.Fatalf("builtin %q not found", name)
	}
	result, err := v.ProcVal.Builtin(args)
	if err != nil {
		t.Fatalf("%s: unexpected error: %v", name, err)
	}
	return result
}

func callErr(t *testing.T, e *env.Env, name string, args ...*types.Value) error {
	t.Helper()
	v, ok := e.Get(name)
	if !ok {
		t.Fatalf("builtin %q not found", name)
	}
	_, err := v.ProcVal.Builtin(args)
	return err
}

// =========================================================================
// Arithmetic tests
// =========================================================================

func TestAdd(t *testing.T) {
	e := setup()
	tests := []struct {
		args   []*types.Value
		expect int64
	}{
		{nil, 0},
		{[]*types.Value{types.NewInteger(1)}, 1},
		{[]*types.Value{types.NewInteger(1), types.NewInteger(2)}, 3},
		{[]*types.Value{types.NewInteger(1), types.NewInteger(2), types.NewInteger(3)}, 6},
		{[]*types.Value{types.NewInteger(-5), types.NewInteger(5)}, 0},
		{[]*types.Value{types.NewInteger(0)}, 0},
	}
	for _, tt := range tests {
		result := call(t, e, "+", tt.args...)
		if result.IntVal != tt.expect {
			t.Errorf("(+ %v) = %d, want %d", tt.args, result.IntVal, tt.expect)
		}
	}
}

func TestAddFloat(t *testing.T) {
	e := setup()
	result := call(t, e, "+", types.NewInteger(1), types.NewFloat(2.5))
	if result.FloatVal != 3.5 {
		t.Errorf("expected 3.5, got %v", result.FloatVal)
	}
}

func TestAddTypeError(t *testing.T) {
	e := setup()
	err := callErr(t, e, "+", types.NewString("a"))
	if err == nil {
		t.Error("expected type error")
	}
}

func TestSub(t *testing.T) {
	e := setup()
	// Unary negation
	r := call(t, e, "-", types.NewInteger(5))
	if r.IntVal != -5 {
		t.Errorf("expected -5, got %d", r.IntVal)
	}
	// Binary
	r = call(t, e, "-", types.NewInteger(10), types.NewInteger(3))
	if r.IntVal != 7 {
		t.Errorf("expected 7, got %d", r.IntVal)
	}
	// Multiple
	r = call(t, e, "-", types.NewInteger(10), types.NewInteger(3), types.NewInteger(2))
	if r.IntVal != 5 {
		t.Errorf("expected 5, got %d", r.IntVal)
	}
}

func TestSubNoArgs(t *testing.T) {
	e := setup()
	err := callErr(t, e, "-")
	if err == nil {
		t.Error("expected error for no arguments")
	}
}

func TestMul(t *testing.T) {
	e := setup()
	r := call(t, e, "*")
	if r.IntVal != 1 {
		t.Errorf("expected 1 for (*), got %d", r.IntVal)
	}
	r = call(t, e, "*", types.NewInteger(3), types.NewInteger(4))
	if r.IntVal != 12 {
		t.Errorf("expected 12, got %d", r.IntVal)
	}
	r = call(t, e, "*", types.NewInteger(0), types.NewInteger(999))
	if r.IntVal != 0 {
		t.Errorf("expected 0, got %d", r.IntVal)
	}
}

func TestDiv(t *testing.T) {
	e := setup()
	r := call(t, e, "/", types.NewInteger(10), types.NewInteger(2))
	if r.IntVal != 5 {
		t.Errorf("expected 5, got %d", r.IntVal)
	}
	// Non-exact division returns float
	r = call(t, e, "/", types.NewInteger(10), types.NewInteger(3))
	if r.Type != types.TypeFloat {
		t.Errorf("expected float for 10/3, got %s", r.Type)
	}
}

func TestDivByZero(t *testing.T) {
	e := setup()
	err := callErr(t, e, "/", types.NewInteger(1), types.NewInteger(0))
	if err == nil {
		t.Error("expected division by zero error")
	}
}

func TestModulo(t *testing.T) {
	e := setup()
	r := call(t, e, "modulo", types.NewInteger(10), types.NewInteger(3))
	if r.IntVal != 1 {
		t.Errorf("expected 1, got %d", r.IntVal)
	}
	// Scheme modulo: sign follows divisor
	r = call(t, e, "modulo", types.NewInteger(-7), types.NewInteger(3))
	if r.IntVal != 2 {
		t.Errorf("expected 2 for (modulo -7 3), got %d", r.IntVal)
	}
}

func TestModuloByZero(t *testing.T) {
	e := setup()
	err := callErr(t, e, "modulo", types.NewInteger(5), types.NewInteger(0))
	if err == nil {
		t.Error("expected division by zero error")
	}
}

// =========================================================================
// Comparison tests
// =========================================================================

func TestNumericEq(t *testing.T) {
	e := setup()
	tests := []struct {
		a, b   *types.Value
		expect bool
	}{
		{types.NewInteger(1), types.NewInteger(1), true},
		{types.NewInteger(1), types.NewInteger(2), false},
		{types.NewInteger(1), types.NewFloat(1.0), true},
		{types.NewFloat(1.5), types.NewFloat(1.5), true},
	}
	for _, tt := range tests {
		r := call(t, e, "=", tt.a, tt.b)
		if r.BoolVal != tt.expect {
			t.Errorf("(= %v %v) = %v, want %v", tt.a, tt.b, r.BoolVal, tt.expect)
		}
	}
}

func TestLessThan(t *testing.T) {
	e := setup()
	r := call(t, e, "<", types.NewInteger(1), types.NewInteger(2))
	if !r.BoolVal {
		t.Error("expected (< 1 2) = #t")
	}
	r = call(t, e, "<", types.NewInteger(2), types.NewInteger(1))
	if r.BoolVal {
		t.Error("expected (< 2 1) = #f")
	}
	// Chained
	r = call(t, e, "<", types.NewInteger(1), types.NewInteger(2), types.NewInteger(3))
	if !r.BoolVal {
		t.Error("expected (< 1 2 3) = #t")
	}
}

func TestNot(t *testing.T) {
	e := setup()
	r := call(t, e, "not", types.False)
	if !r.BoolVal {
		t.Error("expected (not #f) = #t")
	}
	r = call(t, e, "not", types.True)
	if r.BoolVal {
		t.Error("expected (not #t) = #f")
	}
	r = call(t, e, "not", types.NewInteger(0))
	if r.BoolVal {
		t.Error("expected (not 0) = #f (0 is truthy in Scheme)")
	}
}

func TestEqPredicate(t *testing.T) {
	e := setup()
	r := call(t, e, "eq?", types.NewInteger(1), types.NewInteger(1))
	if !r.BoolVal {
		t.Error("expected (eq? 1 1) = #t")
	}
	// eq? on different string objects is #f (pointer equality)
	r = call(t, e, "eq?", types.NewString("a"), types.NewString("a"))
	if r.BoolVal {
		t.Error("expected (eq? \"a\" \"a\") = #f for distinct string objects")
	}
	// eq? on same string object is #t
	s := types.NewString("hello")
	r = call(t, e, "eq?", s, s)
	if !r.BoolVal {
		t.Error("expected (eq? s s) = #t for same object")
	}
}

func TestEqualPredicate(t *testing.T) {
	e := setup()
	a := types.FromList([]*types.Value{types.NewInteger(1), types.NewInteger(2)})
	b := types.FromList([]*types.Value{types.NewInteger(1), types.NewInteger(2)})
	r := call(t, e, "equal?", a, b)
	if !r.BoolVal {
		t.Error("expected (equal? '(1 2) '(1 2)) = #t")
	}
}

// =========================================================================
// List tests
// =========================================================================

func TestCons(t *testing.T) {
	e := setup()
	r := call(t, e, "cons", types.NewInteger(1), types.Nil())
	if r.Type != types.TypePair || r.PairVal.Car.IntVal != 1 {
		t.Error("cons failed")
	}
}

func TestCar(t *testing.T) {
	e := setup()
	pair := types.Cons(types.NewInteger(42), types.Nil())
	r := call(t, e, "car", pair)
	if r.IntVal != 42 {
		t.Errorf("expected 42, got %d", r.IntVal)
	}
}

func TestCarError(t *testing.T) {
	e := setup()
	err := callErr(t, e, "car", types.NewInteger(1))
	if err == nil {
		t.Error("expected error for car on non-pair")
	}
}

func TestCdr(t *testing.T) {
	e := setup()
	pair := types.Cons(types.NewInteger(1), types.Cons(types.NewInteger(2), types.Nil()))
	r := call(t, e, "cdr", pair)
	if r.Type != types.TypePair || r.PairVal.Car.IntVal != 2 {
		t.Error("cdr failed")
	}
}

func TestList(t *testing.T) {
	e := setup()
	r := call(t, e, "list", types.NewInteger(1), types.NewInteger(2), types.NewInteger(3))
	lst, err := r.ToList()
	if err != nil || len(lst) != 3 {
		t.Errorf("expected list of 3, got %v", r)
	}
}

func TestListEmpty(t *testing.T) {
	e := setup()
	r := call(t, e, "list")
	if r.Type != types.TypeNil {
		t.Error("expected empty list")
	}
}

func TestNull(t *testing.T) {
	e := setup()
	r := call(t, e, "null?", types.Nil())
	if !r.BoolVal {
		t.Error("expected (null? '()) = #t")
	}
	r = call(t, e, "null?", types.NewInteger(1))
	if r.BoolVal {
		t.Error("expected (null? 1) = #f")
	}
}

func TestLength(t *testing.T) {
	e := setup()
	lst := types.FromList([]*types.Value{types.NewInteger(1), types.NewInteger(2), types.NewInteger(3)})
	r := call(t, e, "length", lst)
	if r.IntVal != 3 {
		t.Errorf("expected 3, got %d", r.IntVal)
	}
}

func TestLengthEmpty(t *testing.T) {
	e := setup()
	r := call(t, e, "length", types.Nil())
	if r.IntVal != 0 {
		t.Errorf("expected 0, got %d", r.IntVal)
	}
}

func TestAppend(t *testing.T) {
	e := setup()
	a := types.FromList([]*types.Value{types.NewInteger(1), types.NewInteger(2)})
	b := types.FromList([]*types.Value{types.NewInteger(3), types.NewInteger(4)})
	r := call(t, e, "append", a, b)
	lst, _ := r.ToList()
	if len(lst) != 4 {
		t.Errorf("expected 4 elements, got %d", len(lst))
	}
}

func TestAppendEmpty(t *testing.T) {
	e := setup()
	r := call(t, e, "append")
	if r.Type != types.TypeNil {
		t.Error("expected empty list for (append)")
	}
}

func TestReverse(t *testing.T) {
	e := setup()
	lst := types.FromList([]*types.Value{types.NewInteger(1), types.NewInteger(2), types.NewInteger(3)})
	r := call(t, e, "reverse", lst)
	result, _ := r.ToList()
	if len(result) != 3 || result[0].IntVal != 3 || result[2].IntVal != 1 {
		t.Error("reverse failed")
	}
}

func TestListRef(t *testing.T) {
	e := setup()
	lst := types.FromList([]*types.Value{types.NewInteger(10), types.NewInteger(20), types.NewInteger(30)})
	r := call(t, e, "list-ref", lst, types.NewInteger(1))
	if r.IntVal != 20 {
		t.Errorf("expected 20, got %d", r.IntVal)
	}
}

func TestListRefOutOfRange(t *testing.T) {
	e := setup()
	lst := types.FromList([]*types.Value{types.NewInteger(1)})
	err := callErr(t, e, "list-ref", lst, types.NewInteger(5))
	if err == nil {
		t.Error("expected out of range error")
	}
}

func TestSetCar(t *testing.T) {
	e := setup()
	pair := types.Cons(types.NewInteger(1), types.Nil())
	call(t, e, "set-car!", pair, types.NewInteger(42))
	if pair.PairVal.Car.IntVal != 42 {
		t.Error("set-car! failed")
	}
}

func TestSetCdr(t *testing.T) {
	e := setup()
	pair := types.Cons(types.NewInteger(1), types.Nil())
	call(t, e, "set-cdr!", pair, types.NewInteger(99))
	if pair.PairVal.Cdr.IntVal != 99 {
		t.Error("set-cdr! failed")
	}
}

func TestListCopy(t *testing.T) {
	e := setup()
	// Proper list copy
	lst := types.FromList([]*types.Value{types.NewInteger(1), types.NewInteger(2), types.NewInteger(3)})
	r := call(t, e, "list-copy", lst)
	result, _ := r.ToList()
	if len(result) != 3 || result[0].IntVal != 1 {
		t.Error("list-copy: proper list copy failed")
	}
	// Mutation of copy should not affect original
	call(t, e, "set-car!", r, types.NewInteger(99))
	if lst.PairVal.Car.IntVal != 1 {
		t.Error("list-copy: copy should be independent of original")
	}
}

func TestListCopyImproper(t *testing.T) {
	e := setup()
	// Improper list: (1 . 2)
	pair := types.Cons(types.NewInteger(1), types.NewInteger(2))
	r := call(t, e, "list-copy", pair)
	if r.PairVal.Car.IntVal != 1 || r.PairVal.Cdr.IntVal != 2 {
		t.Error("list-copy: improper list copy failed")
	}
}

func TestListCopyNonList(t *testing.T) {
	e := setup()
	err := callErr(t, e, "list-copy", types.NewInteger(42))
	if err == nil {
		t.Error("expected error for list-copy on non-list")
	}
}

// =========================================================================
// String tests
// =========================================================================

func TestStringLength(t *testing.T) {
	e := setup()
	r := call(t, e, "string-length", types.NewString("hello"))
	if r.IntVal != 5 {
		t.Errorf("expected 5, got %d", r.IntVal)
	}
}

func TestStringLengthEmpty(t *testing.T) {
	e := setup()
	r := call(t, e, "string-length", types.NewString(""))
	if r.IntVal != 0 {
		t.Errorf("expected 0, got %d", r.IntVal)
	}
}

func TestStringRef(t *testing.T) {
	e := setup()
	r := call(t, e, "string-ref", types.NewString("abc"), types.NewInteger(1))
	if r.CharVal != 'b' {
		t.Errorf("expected 'b', got %c", r.CharVal)
	}
}

func TestStringRefOutOfRange(t *testing.T) {
	e := setup()
	err := callErr(t, e, "string-ref", types.NewString("a"), types.NewInteger(5))
	if err == nil {
		t.Error("expected out of range error")
	}
}

func TestStringAppend(t *testing.T) {
	e := setup()
	r := call(t, e, "string-append", types.NewString("hello"), types.NewString(" "), types.NewString("world"))
	if r.StrVal != "hello world" {
		t.Errorf("expected 'hello world', got %q", r.StrVal)
	}
}

func TestStringAppendEmpty(t *testing.T) {
	e := setup()
	r := call(t, e, "string-append")
	if r.StrVal != "" {
		t.Error("expected empty string")
	}
}

func TestSubstring(t *testing.T) {
	e := setup()
	r := call(t, e, "substring", types.NewString("hello"), types.NewInteger(1), types.NewInteger(4))
	if r.StrVal != "ell" {
		t.Errorf("expected 'ell', got %q", r.StrVal)
	}
}

func TestStringToNumber(t *testing.T) {
	e := setup()
	r := call(t, e, "string->number", types.NewString("42"))
	if r.IntVal != 42 {
		t.Errorf("expected 42, got %d", r.IntVal)
	}
	r = call(t, e, "string->number", types.NewString("3.14"))
	if r.Type != types.TypeFloat {
		t.Error("expected float")
	}
	r = call(t, e, "string->number", types.NewString("abc"))
	if r.Type != types.TypeBoolean || r.BoolVal {
		t.Error("expected #f for invalid number")
	}
}

func TestStringUpcase(t *testing.T) {
	e := setup()
	r := call(t, e, "string-upcase", types.NewString("hello"))
	if r.StrVal != "HELLO" {
		t.Errorf("expected 'HELLO', got %q", r.StrVal)
	}
}

func TestStringDowncase(t *testing.T) {
	e := setup()
	r := call(t, e, "string-downcase", types.NewString("HELLO"))
	if r.StrVal != "hello" {
		t.Errorf("expected 'hello', got %q", r.StrVal)
	}
}

func TestStringEq(t *testing.T) {
	e := setup()
	r := call(t, e, "string=?", types.NewString("abc"), types.NewString("abc"))
	if !r.BoolVal {
		t.Error("expected #t")
	}
	r = call(t, e, "string=?", types.NewString("abc"), types.NewString("def"))
	if r.BoolVal {
		t.Error("expected #f")
	}
}

func TestSymbolToString(t *testing.T) {
	e := setup()
	r := call(t, e, "symbol->string", types.NewSymbol("hello"))
	if r.StrVal != "hello" {
		t.Errorf("expected 'hello', got %q", r.StrVal)
	}
}

func TestStringToSymbol(t *testing.T) {
	e := setup()
	r := call(t, e, "string->symbol", types.NewString("hello"))
	if r.SymbolVal != "hello" {
		t.Errorf("expected symbol hello, got %q", r.SymbolVal)
	}
}

// =========================================================================
// Character tests
// =========================================================================

func TestCharToInteger(t *testing.T) {
	e := setup()
	r := call(t, e, "char->integer", types.NewChar('A'))
	if r.IntVal != 65 {
		t.Errorf("expected 65, got %d", r.IntVal)
	}
}

func TestIntegerToChar(t *testing.T) {
	e := setup()
	r := call(t, e, "integer->char", types.NewInteger(65))
	if r.CharVal != 'A' {
		t.Errorf("expected 'A', got %c", r.CharVal)
	}
}

func TestCharPredicates(t *testing.T) {
	e := setup()
	tests := []struct {
		fn   string
		ch   rune
		want bool
	}{
		{"char-alphabetic?", 'a', true},
		{"char-alphabetic?", '1', false},
		{"char-numeric?", '5', true},
		{"char-numeric?", 'a', false},
		{"char-whitespace?", ' ', true},
		{"char-whitespace?", 'a', false},
		{"char-upper-case?", 'A', true},
		{"char-upper-case?", 'a', false},
		{"char-lower-case?", 'a', true},
		{"char-lower-case?", 'A', false},
	}
	for _, tt := range tests {
		r := call(t, e, tt.fn, types.NewChar(tt.ch))
		if r.BoolVal != tt.want {
			t.Errorf("(%s #\\%c) = %v, want %v", tt.fn, tt.ch, r.BoolVal, tt.want)
		}
	}
}

func TestCharCase(t *testing.T) {
	e := setup()
	r := call(t, e, "char-upcase", types.NewChar('a'))
	if r.CharVal != 'A' {
		t.Errorf("expected 'A', got %c", r.CharVal)
	}
	r = call(t, e, "char-downcase", types.NewChar('A'))
	if r.CharVal != 'a' {
		t.Errorf("expected 'a', got %c", r.CharVal)
	}
}

func TestCharComparisons(t *testing.T) {
	e := setup()
	r := call(t, e, "char=?", types.NewChar('a'), types.NewChar('a'))
	if !r.BoolVal {
		t.Error("expected #t")
	}
	r = call(t, e, "char<?", types.NewChar('a'), types.NewChar('b'))
	if !r.BoolVal {
		t.Error("expected #t")
	}
}

// =========================================================================
// Vector tests
// =========================================================================

func TestMakeVector(t *testing.T) {
	e := setup()
	r := call(t, e, "make-vector", types.NewInteger(3), types.NewInteger(0))
	if r.Type != types.TypeVector || len(r.VecVal) != 3 {
		t.Error("make-vector failed")
	}
}

func TestVector(t *testing.T) {
	e := setup()
	r := call(t, e, "vector", types.NewInteger(1), types.NewInteger(2), types.NewInteger(3))
	if len(r.VecVal) != 3 {
		t.Errorf("expected 3 elements, got %d", len(r.VecVal))
	}
}

func TestVectorRef(t *testing.T) {
	e := setup()
	v := types.NewVector([]*types.Value{types.NewInteger(10), types.NewInteger(20)})
	r := call(t, e, "vector-ref", v, types.NewInteger(1))
	if r.IntVal != 20 {
		t.Errorf("expected 20, got %d", r.IntVal)
	}
}

func TestVectorRefOutOfRange(t *testing.T) {
	e := setup()
	v := types.NewVector([]*types.Value{types.NewInteger(1)})
	err := callErr(t, e, "vector-ref", v, types.NewInteger(5))
	if err == nil {
		t.Error("expected out of range error")
	}
}

func TestVectorSet(t *testing.T) {
	e := setup()
	v := types.NewVector([]*types.Value{types.NewInteger(1), types.NewInteger(2)})
	call(t, e, "vector-set!", v, types.NewInteger(0), types.NewInteger(99))
	if v.VecVal[0].IntVal != 99 {
		t.Error("vector-set! failed")
	}
}

func TestVectorToList(t *testing.T) {
	e := setup()
	v := types.NewVector([]*types.Value{types.NewInteger(1), types.NewInteger(2)})
	r := call(t, e, "vector->list", v)
	lst, _ := r.ToList()
	if len(lst) != 2 {
		t.Error("vector->list failed")
	}
}

func TestListToVector(t *testing.T) {
	e := setup()
	lst := types.FromList([]*types.Value{types.NewInteger(1), types.NewInteger(2)})
	r := call(t, e, "list->vector", lst)
	if r.Type != types.TypeVector || len(r.VecVal) != 2 {
		t.Error("list->vector failed")
	}
}

func TestVectorCopy(t *testing.T) {
	e := setup()
	v := types.NewVector([]*types.Value{types.NewInteger(1), types.NewInteger(2)})
	r := call(t, e, "vector-copy", v)
	if len(r.VecVal) != 2 {
		t.Error("vector-copy failed")
	}
	// Mutation of copy should not affect original
	r.VecVal[0] = types.NewInteger(99)
	if v.VecVal[0].IntVal != 1 {
		t.Error("vector-copy: copy should be independent")
	}
}

// =========================================================================
// Bytevector tests
// =========================================================================

func TestMakeBytevector(t *testing.T) {
	e := setup()
	r := call(t, e, "make-bytevector", types.NewInteger(5), types.NewInteger(42))
	if r.Type != types.TypeBytevector || len(r.ByteVec) != 5 || r.ByteVec[0] != 42 {
		t.Error("make-bytevector failed")
	}
}

func TestBytevectorRef(t *testing.T) {
	e := setup()
	bv := types.NewBytevector([]byte{10, 20, 30})
	r := call(t, e, "bytevector-u8-ref", bv, types.NewInteger(1))
	if r.IntVal != 20 {
		t.Errorf("expected 20, got %d", r.IntVal)
	}
}

func TestBytevectorSet(t *testing.T) {
	e := setup()
	bv := types.NewBytevector([]byte{0, 0, 0})
	call(t, e, "bytevector-u8-set!", bv, types.NewInteger(1), types.NewInteger(255))
	if bv.ByteVec[1] != 255 {
		t.Error("bytevector-u8-set! failed")
	}
}

func TestBytevectorAppend(t *testing.T) {
	e := setup()
	a := types.NewBytevector([]byte{1, 2})
	b := types.NewBytevector([]byte{3, 4})
	r := call(t, e, "bytevector-append", a, b)
	if len(r.ByteVec) != 4 || r.ByteVec[2] != 3 {
		t.Error("bytevector-append failed")
	}
}

// =========================================================================
// Type predicates tests
// =========================================================================

func TestTypePredicates(t *testing.T) {
	e := setup()
	tests := []struct {
		fn    string
		arg   *types.Value
		want  bool
	}{
		{"number?", types.NewInteger(1), true},
		{"number?", types.NewFloat(1.0), true},
		{"number?", types.NewString("a"), false},
		{"integer?", types.NewInteger(1), true},
		{"integer?", types.NewFloat(1.0), false},
		{"boolean?", types.True, true},
		{"boolean?", types.NewInteger(1), false},
		{"string?", types.NewString("a"), true},
		{"string?", types.NewInteger(1), false},
		{"symbol?", types.NewSymbol("a"), true},
		{"symbol?", types.NewString("a"), false},
		{"pair?", types.Cons(types.NewInteger(1), types.Nil()), true},
		{"pair?", types.Nil(), false},
		{"null?", types.Nil(), true},
		{"null?", types.NewInteger(0), false},
		{"char?", types.NewChar('a'), true},
		{"char?", types.NewString("a"), false},
		{"vector?", types.NewVector(nil), true},
		{"vector?", types.Nil(), false},
		{"bytevector?", types.NewBytevector(nil), true},
		{"bytevector?", types.NewInteger(0), false},
	}
	for _, tt := range tests {
		r := call(t, e, tt.fn, tt.arg)
		if r.BoolVal != tt.want {
			t.Errorf("(%s %v) = %v, want %v", tt.fn, tt.arg, r.BoolVal, tt.want)
		}
	}
}

// =========================================================================
// I/O tests
// =========================================================================

func TestDisplay(t *testing.T) {
	e := setup()
	r := call(t, e, "display", types.NewString("hello"))
	if r.Type != types.TypeVoid {
		t.Error("display should return void")
	}
}

func TestError(t *testing.T) {
	e := setup()
	err := callErr(t, e, "error", types.NewString("test error"))
	if err == nil {
		t.Error("expected error")
	}
}

// =========================================================================
// Values test
// =========================================================================

func TestValues(t *testing.T) {
	e := setup()
	r := call(t, e, "values", types.NewInteger(1))
	if r.IntVal != 1 {
		t.Error("single value should return as-is")
	}
	r = call(t, e, "values", types.NewInteger(1), types.NewInteger(2))
	if r.Type != types.TypeValues || len(r.ValuesVal) != 2 {
		t.Error("multiple values failed")
	}
}
