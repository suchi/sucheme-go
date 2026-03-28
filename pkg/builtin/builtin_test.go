package builtin

import (
	"math"
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

func TestQuotient(t *testing.T) {
	e := setup()
	r := call(t, e, "quotient", types.NewInteger(10), types.NewInteger(3))
	if r.IntVal != 3 {
		t.Errorf("expected 3, got %d", r.IntVal)
	}
	r = call(t, e, "quotient", types.NewInteger(-10), types.NewInteger(3))
	if r.IntVal != -3 {
		t.Errorf("expected -3, got %d", r.IntVal)
	}
}

func TestQuotientByZero(t *testing.T) {
	e := setup()
	err := callErr(t, e, "quotient", types.NewInteger(1), types.NewInteger(0))
	if err == nil {
		t.Error("expected division by zero")
	}
}

func TestRemainder(t *testing.T) {
	e := setup()
	r := call(t, e, "remainder", types.NewInteger(10), types.NewInteger(3))
	if r.IntVal != 1 {
		t.Errorf("expected 1, got %d", r.IntVal)
	}
}

func TestAbs(t *testing.T) {
	e := setup()
	r := call(t, e, "abs", types.NewInteger(-42))
	if r.IntVal != 42 {
		t.Errorf("expected 42, got %d", r.IntVal)
	}
	r = call(t, e, "abs", types.NewFloat(-3.14))
	if r.FloatVal != 3.14 {
		t.Errorf("expected 3.14, got %f", r.FloatVal)
	}
}

func TestMinMax(t *testing.T) {
	e := setup()
	r := call(t, e, "min", types.NewInteger(3), types.NewInteger(1), types.NewInteger(5))
	if r.IntVal != 1 {
		t.Errorf("expected 1, got %d", r.IntVal)
	}
	r = call(t, e, "max", types.NewInteger(3), types.NewInteger(1), types.NewInteger(5))
	if r.IntVal != 5 {
		t.Errorf("expected 5, got %d", r.IntVal)
	}
}

func TestFloorCeilingRound(t *testing.T) {
	e := setup()
	r := call(t, e, "floor", types.NewFloat(3.7))
	if r.FloatVal != 3.0 {
		t.Errorf("expected 3.0, got %f", r.FloatVal)
	}
	r = call(t, e, "ceiling", types.NewFloat(3.2))
	if r.FloatVal != 4.0 {
		t.Errorf("expected 4.0, got %f", r.FloatVal)
	}
	// R7RS round: ties-to-even
	r = call(t, e, "round", types.NewFloat(2.5))
	if r.FloatVal != 2.0 {
		t.Errorf("expected 2.0 (even), got %f", r.FloatVal)
	}
	r = call(t, e, "round", types.NewFloat(3.5))
	if r.FloatVal != 4.0 {
		t.Errorf("expected 4.0 (even), got %f", r.FloatVal)
	}
}

func TestExpt(t *testing.T) {
	e := setup()
	r := call(t, e, "expt", types.NewInteger(2), types.NewInteger(10))
	if r.IntVal != 1024 {
		t.Errorf("expected 1024, got %d", r.IntVal)
	}
}

func TestSqrt(t *testing.T) {
	e := setup()
	r := call(t, e, "sqrt", types.NewFloat(4.0))
	if r.FloatVal != 2.0 {
		t.Errorf("expected 2.0, got %f", r.FloatVal)
	}
}

func TestEqv(t *testing.T) {
	e := setup()
	r := call(t, e, "eqv?", types.NewInteger(1), types.NewInteger(1))
	if !r.BoolVal {
		t.Error("expected #t")
	}
}

func TestBooleanEq(t *testing.T) {
	e := setup()
	r := call(t, e, "boolean=?", types.True, types.True)
	if !r.BoolVal {
		t.Error("expected #t")
	}
	r = call(t, e, "boolean=?", types.True, types.False)
	if r.BoolVal {
		t.Error("expected #f")
	}
}

func TestAssoc(t *testing.T) {
	e := setup()
	alist := types.FromList([]*types.Value{
		types.Cons(types.NewSymbol("a"), types.NewInteger(1)),
		types.Cons(types.NewSymbol("b"), types.NewInteger(2)),
	})
	r := call(t, e, "assoc", types.NewSymbol("b"), alist)
	if r.Type != types.TypePair || r.PairVal.Cdr.IntVal != 2 {
		t.Errorf("expected (b . 2), got %s", r.String())
	}
	r = call(t, e, "assoc", types.NewSymbol("z"), alist)
	if r.BoolVal != false || r.Type != types.TypeBoolean {
		t.Error("expected #f for missing key")
	}
}

func TestMember(t *testing.T) {
	e := setup()
	lst := types.FromList([]*types.Value{types.NewInteger(1), types.NewInteger(2), types.NewInteger(3)})
	r := call(t, e, "member", types.NewInteger(2), lst)
	if r.Type != types.TypePair || r.PairVal.Car.IntVal != 2 {
		t.Errorf("expected (2 3), got %s", r.String())
	}
	r = call(t, e, "member", types.NewInteger(9), lst)
	if r.Type != types.TypeBoolean {
		t.Error("expected #f for missing member")
	}
}

func TestTruncate(t *testing.T) {
	e := setup()
	r := call(t, e, "truncate", types.NewFloat(3.7))
	if r.FloatVal != 3.0 {
		t.Errorf("expected 3.0, got %f", r.FloatVal)
	}
	r = call(t, e, "truncate", types.NewFloat(-3.7))
	if r.FloatVal != -3.0 {
		t.Errorf("expected -3.0, got %f", r.FloatVal)
	}
}

func TestExactInexact(t *testing.T) {
	e := setup()
	r := call(t, e, "exact->inexact", types.NewInteger(42))
	if r.Type != types.TypeFloat || r.FloatVal != 42.0 {
		t.Errorf("expected 42.0 float, got %s", r.String())
	}
	r = call(t, e, "inexact->exact", types.NewFloat(42.0))
	if r.Type != types.TypeInteger || r.IntVal != 42 {
		t.Errorf("expected 42 integer, got %s", r.String())
	}
}

func TestInexactExactNonInteger(t *testing.T) {
	e := setup()
	err := callErr(t, e, "inexact->exact", types.NewFloat(3.5))
	if err == nil {
		t.Error("expected error for non-integer float")
	}
}

func TestExactInexactTypeError(t *testing.T) {
	e := setup()
	err := callErr(t, e, "exact->inexact", types.NewString("hello"))
	if err == nil {
		t.Error("expected error for non-number")
	}
}

func TestZeroPositiveNegative(t *testing.T) {
	e := setup()
	tests := []struct {
		fn   string
		arg  *types.Value
		want bool
	}{
		{"zero?", types.NewInteger(0), true},
		{"zero?", types.NewInteger(1), false},
		{"positive?", types.NewInteger(5), true},
		{"positive?", types.NewInteger(-5), false},
		{"negative?", types.NewInteger(-3), true},
		{"negative?", types.NewInteger(3), false},
		{"odd?", types.NewInteger(3), true},
		{"odd?", types.NewInteger(4), false},
		{"even?", types.NewInteger(4), true},
		{"even?", types.NewInteger(3), false},
	}
	for _, tt := range tests {
		r := call(t, e, tt.fn, tt.arg)
		if r.BoolVal != tt.want {
			t.Errorf("(%s %s) = %v, want %v", tt.fn, tt.arg.String(), r.BoolVal, tt.want)
		}
	}
}

func TestAssv(t *testing.T) {
	e := setup()
	alist := types.FromList([]*types.Value{
		types.Cons(types.NewInteger(1), types.NewString("one")),
		types.Cons(types.NewInteger(2), types.NewString("two")),
	})
	r := call(t, e, "assv", types.NewInteger(2), alist)
	if r.Type != types.TypePair {
		t.Errorf("expected pair, got %s", r.String())
	}
}

func TestMemv(t *testing.T) {
	e := setup()
	lst := types.FromList([]*types.Value{types.NewInteger(1), types.NewInteger(2), types.NewInteger(3)})
	r := call(t, e, "memv", types.NewInteger(2), lst)
	if r.Type != types.TypePair || r.PairVal.Car.IntVal != 2 {
		t.Errorf("expected (2 3), got %s", r.String())
	}
}

func TestSqrtNegative(t *testing.T) {
	e := setup()
	err := callErr(t, e, "sqrt", types.NewFloat(-1.0))
	if err == nil {
		t.Error("expected error for sqrt of negative")
	}
}

func TestRemainderByZero(t *testing.T) {
	e := setup()
	err := callErr(t, e, "remainder", types.NewInteger(5), types.NewInteger(0))
	if err == nil {
		t.Error("expected error for division by zero")
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

// =========================================================================
// PR #7 Review Remediation: Edge case tests
// =========================================================================

// --- abs: MinInt64 overflow ---
func TestAbsMinInt64Overflow(t *testing.T) {
	e := setup()
	err := callErr(t, e, "abs", types.NewInteger(math.MinInt64))
	if err == nil {
		t.Error("expected overflow error for abs(MinInt64)")
	}
}

func TestAbsZero(t *testing.T) {
	e := setup()
	r := call(t, e, "abs", types.NewInteger(0))
	if r.IntVal != 0 {
		t.Errorf("expected 0, got %d", r.IntVal)
	}
}

func TestAbsPositive(t *testing.T) {
	e := setup()
	r := call(t, e, "abs", types.NewInteger(42))
	if r.IntVal != 42 {
		t.Errorf("expected 42, got %d", r.IntVal)
	}
}

func TestAbsTypeError(t *testing.T) {
	e := setup()
	err := callErr(t, e, "abs", types.NewString("hello"))
	if err == nil {
		t.Error("expected type error for abs on string")
	}
}

// --- round: additional edge cases ---
func TestRoundInteger(t *testing.T) {
	e := setup()
	r := call(t, e, "round", types.NewInteger(5))
	if r.Type != types.TypeInteger || r.IntVal != 5 {
		t.Errorf("expected integer 5, got %s", r.String())
	}
}

func TestRoundNegativeHalf(t *testing.T) {
	e := setup()
	r := call(t, e, "round", types.NewFloat(-0.5))
	if r.FloatVal != 0.0 {
		t.Errorf("expected 0.0 (even), got %f", r.FloatVal)
	}
}

func TestRoundNegativeTwoPointFive(t *testing.T) {
	e := setup()
	r := call(t, e, "round", types.NewFloat(-2.5))
	if r.FloatVal != -2.0 {
		t.Errorf("expected -2.0 (even), got %f", r.FloatVal)
	}
}

func TestRoundTypeError(t *testing.T) {
	e := setup()
	err := callErr(t, e, "round", types.NewString("x"))
	if err == nil {
		t.Error("expected type error for round on string")
	}
}

// --- expt: edge cases ---
func TestExptNegativeExponent(t *testing.T) {
	e := setup()
	r := call(t, e, "expt", types.NewInteger(2), types.NewInteger(-1))
	if r.Type != types.TypeFloat || r.FloatVal != 0.5 {
		t.Errorf("expected 0.5 float, got %s", r.String())
	}
}

func TestExptZeroBase(t *testing.T) {
	e := setup()
	r := call(t, e, "expt", types.NewInteger(0), types.NewInteger(0))
	if r.IntVal != 1 {
		t.Errorf("expected 1 for (expt 0 0), got %d", r.IntVal)
	}
}

func TestExptFloatBase(t *testing.T) {
	e := setup()
	r := call(t, e, "expt", types.NewFloat(2.0), types.NewInteger(3))
	if r.Type != types.TypeFloat || r.FloatVal != 8.0 {
		t.Errorf("expected 8.0, got %s", r.String())
	}
}

func TestExptTypeError(t *testing.T) {
	e := setup()
	err := callErr(t, e, "expt", types.NewString("a"), types.NewInteger(1))
	if err == nil {
		t.Error("expected type error for expt on string")
	}
}

// --- inexact->exact: out of range ---
func TestInexactExactOutOfRange(t *testing.T) {
	e := setup()
	err := callErr(t, e, "inexact->exact", types.NewFloat(1e19))
	if err == nil {
		t.Error("expected out of range error for very large float")
	}
}

func TestInexactExactAlreadyExact(t *testing.T) {
	e := setup()
	r := call(t, e, "inexact->exact", types.NewInteger(42))
	if r.Type != types.TypeInteger || r.IntVal != 42 {
		t.Errorf("expected integer 42, got %s", r.String())
	}
}

func TestInexactExactNegativeZero(t *testing.T) {
	e := setup()
	negZero := math.Copysign(0, -1)
	r := call(t, e, "inexact->exact", types.NewFloat(negZero))
	if r.Type != types.TypeInteger || r.IntVal != 0 {
		t.Errorf("expected 0, got %s", r.String())
	}
}

// --- boolean=?: edge cases ---
func TestBooleanEqTypeError(t *testing.T) {
	e := setup()
	err := callErr(t, e, "boolean=?", types.True, types.NewInteger(1))
	if err == nil {
		t.Error("expected type error for boolean=? with non-boolean")
	}
}

func TestBooleanEqSingleArg(t *testing.T) {
	e := setup()
	err := callErr(t, e, "boolean=?", types.True)
	if err == nil {
		t.Error("expected error for boolean=? with single argument")
	}
}

func TestBooleanEqMultiple(t *testing.T) {
	e := setup()
	r := call(t, e, "boolean=?", types.False, types.False, types.False)
	if !r.BoolVal {
		t.Error("expected #t for (boolean=? #f #f #f)")
	}
	r = call(t, e, "boolean=?", types.True, types.True, types.False)
	if r.BoolVal {
		t.Error("expected #f for (boolean=? #t #t #f)")
	}
}

// --- eqv?: additional cases ---
func TestEqvDifferentTypes(t *testing.T) {
	e := setup()
	r := call(t, e, "eqv?", types.NewInteger(1), types.NewString("1"))
	if r.BoolVal {
		t.Error("expected #f for eqv? of different types")
	}
}

func TestEqvBooleans(t *testing.T) {
	e := setup()
	r := call(t, e, "eqv?", types.True, types.True)
	if !r.BoolVal {
		t.Error("expected #t for (eqv? #t #t)")
	}
	r = call(t, e, "eqv?", types.True, types.False)
	if r.BoolVal {
		t.Error("expected #f for (eqv? #t #f)")
	}
}

func TestEqvChars(t *testing.T) {
	e := setup()
	r := call(t, e, "eqv?", types.NewChar('a'), types.NewChar('a'))
	if !r.BoolVal {
		t.Error("expected #t for (eqv? #\\a #\\a)")
	}
	r = call(t, e, "eqv?", types.NewChar('a'), types.NewChar('b'))
	if r.BoolVal {
		t.Error("expected #f for (eqv? #\\a #\\b)")
	}
}

func TestEqvArgCount(t *testing.T) {
	e := setup()
	err := callErr(t, e, "eqv?", types.NewInteger(1))
	if err == nil {
		t.Error("expected error for eqv? with 1 argument")
	}
}

// --- assq: tests ---
func TestAssq(t *testing.T) {
	e := setup()
	sym := types.NewSymbol("b")
	alist := types.FromList([]*types.Value{
		types.Cons(types.NewSymbol("a"), types.NewInteger(1)),
		types.Cons(sym, types.NewInteger(2)),
		types.Cons(types.NewSymbol("c"), types.NewInteger(3)),
	})
	r := call(t, e, "assq", sym, alist)
	if r.Type != types.TypePair || r.PairVal.Cdr.IntVal != 2 {
		t.Errorf("expected (b . 2), got %s", r.String())
	}
}

func TestAssqNotFound(t *testing.T) {
	e := setup()
	alist := types.FromList([]*types.Value{
		types.Cons(types.NewSymbol("a"), types.NewInteger(1)),
	})
	r := call(t, e, "assq", types.NewSymbol("z"), alist)
	if r.Type != types.TypeBoolean || r.BoolVal {
		t.Error("expected #f for missing key in assq")
	}
}

func TestAssqNonList(t *testing.T) {
	e := setup()
	err := callErr(t, e, "assq", types.NewSymbol("a"), types.NewInteger(42))
	if err == nil {
		t.Error("expected error for assq on non-list")
	}
}

// --- assv: additional tests ---
func TestAssvNotFound(t *testing.T) {
	e := setup()
	alist := types.FromList([]*types.Value{
		types.Cons(types.NewInteger(1), types.NewString("one")),
	})
	r := call(t, e, "assv", types.NewInteger(99), alist)
	if r.Type != types.TypeBoolean || r.BoolVal {
		t.Error("expected #f for missing key in assv")
	}
}

// --- memq: tests ---
func TestMemq(t *testing.T) {
	e := setup()
	sym := types.NewSymbol("b")
	lst := types.FromList([]*types.Value{types.NewSymbol("a"), sym, types.NewSymbol("c")})
	r := call(t, e, "memq", sym, lst)
	if r.Type != types.TypePair || r.PairVal.Car.SymbolVal != "b" {
		t.Errorf("expected list starting with b, got %s", r.String())
	}
}

func TestMemqNotFound(t *testing.T) {
	e := setup()
	lst := types.FromList([]*types.Value{types.NewInteger(1), types.NewInteger(2)})
	r := call(t, e, "memq", types.NewInteger(9), lst)
	if r.Type != types.TypeBoolean || r.BoolVal {
		t.Error("expected #f for missing element in memq")
	}
}

func TestMemqEmptyList(t *testing.T) {
	e := setup()
	r := call(t, e, "memq", types.NewInteger(1), types.Nil())
	if r.Type != types.TypeBoolean || r.BoolVal {
		t.Error("expected #f for memq on empty list")
	}
}

// --- member/memv/memq: improper list error ---
func TestMemberImproperList(t *testing.T) {
	e := setup()
	// (1 . 2) is an improper list
	improper := types.Cons(types.NewInteger(1), types.NewInteger(2))
	err := callErr(t, e, "member", types.NewInteger(9), improper)
	if err == nil {
		t.Error("expected error for member on improper list")
	}
}

func TestMemvImproperList(t *testing.T) {
	e := setup()
	improper := types.Cons(types.NewInteger(1), types.NewInteger(2))
	err := callErr(t, e, "memv", types.NewInteger(9), improper)
	if err == nil {
		t.Error("expected error for memv on improper list")
	}
}

func TestMemqImproperList(t *testing.T) {
	e := setup()
	improper := types.Cons(types.NewInteger(1), types.NewInteger(2))
	err := callErr(t, e, "memq", types.NewInteger(9), improper)
	if err == nil {
		t.Error("expected error for memq on improper list")
	}
}

// --- member/memv on empty list ---
func TestMemberEmptyList(t *testing.T) {
	e := setup()
	r := call(t, e, "member", types.NewInteger(1), types.Nil())
	if r.Type != types.TypeBoolean || r.BoolVal {
		t.Error("expected #f for member on empty list")
	}
}

func TestMemvEmptyList(t *testing.T) {
	e := setup()
	r := call(t, e, "memv", types.NewInteger(1), types.Nil())
	if r.Type != types.TypeBoolean || r.BoolVal {
		t.Error("expected #f for memv on empty list")
	}
}

// --- member/memv/memq: non-list argument ---
func TestMemberNonList(t *testing.T) {
	e := setup()
	err := callErr(t, e, "member", types.NewInteger(1), types.NewInteger(42))
	if err == nil {
		t.Error("expected error for member on non-list")
	}
}

func TestMemvNonList(t *testing.T) {
	e := setup()
	err := callErr(t, e, "memv", types.NewInteger(1), types.NewString("hello"))
	if err == nil {
		t.Error("expected error for memv on non-list")
	}
}

func TestMemqNonList(t *testing.T) {
	e := setup()
	err := callErr(t, e, "memq", types.NewInteger(1), types.NewString("hello"))
	if err == nil {
		t.Error("expected error for memq on non-list")
	}
}

// --- quotient: additional edge cases ---
func TestQuotientNegativeDivisor(t *testing.T) {
	e := setup()
	r := call(t, e, "quotient", types.NewInteger(10), types.NewInteger(-3))
	if r.IntVal != -3 {
		t.Errorf("expected -3, got %d", r.IntVal)
	}
}

func TestQuotientTypeError(t *testing.T) {
	e := setup()
	err := callErr(t, e, "quotient", types.NewFloat(10.0), types.NewInteger(3))
	if err == nil {
		t.Error("expected type error for quotient with float")
	}
}

// --- remainder: additional edge cases ---
func TestRemainderNegative(t *testing.T) {
	e := setup()
	// Go remainder: sign follows dividend
	r := call(t, e, "remainder", types.NewInteger(-10), types.NewInteger(3))
	if r.IntVal != -1 {
		t.Errorf("expected -1 for (remainder -10 3), got %d", r.IntVal)
	}
}

// --- floor/ceiling/truncate: type errors ---
func TestFloorTypeError(t *testing.T) {
	e := setup()
	err := callErr(t, e, "floor", types.NewString("x"))
	if err == nil {
		t.Error("expected type error for floor on string")
	}
}

func TestCeilingTypeError(t *testing.T) {
	e := setup()
	err := callErr(t, e, "ceiling", types.NewString("x"))
	if err == nil {
		t.Error("expected type error for ceiling on string")
	}
}

func TestTruncateTypeError(t *testing.T) {
	e := setup()
	err := callErr(t, e, "truncate", types.NewString("x"))
	if err == nil {
		t.Error("expected type error for truncate on string")
	}
}

// --- floor/ceiling/truncate: integer passthrough ---
func TestFloorInteger(t *testing.T) {
	e := setup()
	r := call(t, e, "floor", types.NewInteger(5))
	if r.Type != types.TypeInteger || r.IntVal != 5 {
		t.Errorf("expected integer 5, got %s", r.String())
	}
}

func TestCeilingInteger(t *testing.T) {
	e := setup()
	r := call(t, e, "ceiling", types.NewInteger(5))
	if r.Type != types.TypeInteger || r.IntVal != 5 {
		t.Errorf("expected integer 5, got %s", r.String())
	}
}

func TestTruncateInteger(t *testing.T) {
	e := setup()
	r := call(t, e, "truncate", types.NewInteger(5))
	if r.Type != types.TypeInteger || r.IntVal != 5 {
		t.Errorf("expected integer 5, got %s", r.String())
	}
}

// --- assoc/assv/assq: comparison table verification ---
// assoc uses equal?, assv uses eqv?, assq uses eq?
func TestAssocComparisonSemantics(t *testing.T) {
	e := setup()
	// assoc should find by structural equality (equal?)
	alist := types.FromList([]*types.Value{
		types.Cons(types.NewString("key"), types.NewInteger(1)),
	})
	r := call(t, e, "assoc", types.NewString("key"), alist)
	if r.Type != types.TypePair {
		t.Error("assoc should find string key by equal? (structural equality)")
	}
}

func TestAssvComparisonSemantics(t *testing.T) {
	e := setup()
	// assv should find by eqv? (value equality for primitives)
	alist := types.FromList([]*types.Value{
		types.Cons(types.NewInteger(42), types.NewString("found")),
	})
	r := call(t, e, "assv", types.NewInteger(42), alist)
	if r.Type != types.TypePair {
		t.Error("assv should find integer key by eqv?")
	}
}

func TestAssqComparisonSemantics(t *testing.T) {
	e := setup()
	// assq uses eq? (pointer equality) - different string objects should NOT match
	alist := types.FromList([]*types.Value{
		types.Cons(types.NewString("key"), types.NewInteger(1)),
	})
	r := call(t, e, "assq", types.NewString("key"), alist)
	if r.Type == types.TypePair {
		t.Error("assq should NOT find different string objects (eq? is pointer equality)")
	}
}

// --- member/memv/memq: comparison table verification ---
func TestMemberComparisonSemantics(t *testing.T) {
	e := setup()
	// member uses equal? - should find structurally equal strings
	lst := types.FromList([]*types.Value{types.NewString("hello"), types.NewString("world")})
	r := call(t, e, "member", types.NewString("hello"), lst)
	if r.Type != types.TypePair {
		t.Error("member should find by equal? (structural equality)")
	}
}

func TestMemvComparisonSemantics(t *testing.T) {
	e := setup()
	// memv uses eqv? - should find by value equality for primitives
	lst := types.FromList([]*types.Value{types.NewChar('a'), types.NewChar('b')})
	r := call(t, e, "memv", types.NewChar('a'), lst)
	if r.Type != types.TypePair {
		t.Error("memv should find char by eqv?")
	}
}

func TestMemqComparisonSemantics(t *testing.T) {
	e := setup()
	// memq uses eq? - different string objects should NOT match
	lst := types.FromList([]*types.Value{types.NewString("hello")})
	r := call(t, e, "memq", types.NewString("hello"), lst)
	if r.Type == types.TypePair {
		t.Error("memq should NOT find different string objects (eq? is pointer equality)")
	}
}

// --- min/max: edge cases ---
func TestMinSingle(t *testing.T) {
	e := setup()
	r := call(t, e, "min", types.NewInteger(42))
	if r.IntVal != 42 {
		t.Errorf("expected 42, got %d", r.IntVal)
	}
}

func TestMaxSingle(t *testing.T) {
	e := setup()
	r := call(t, e, "max", types.NewInteger(42))
	if r.IntVal != 42 {
		t.Errorf("expected 42, got %d", r.IntVal)
	}
}

func TestMinNoArgs(t *testing.T) {
	e := setup()
	err := callErr(t, e, "min")
	if err == nil {
		t.Error("expected error for min with no arguments")
	}
}

func TestMaxNoArgs(t *testing.T) {
	e := setup()
	err := callErr(t, e, "max")
	if err == nil {
		t.Error("expected error for max with no arguments")
	}
}

func TestMinMixedTypes(t *testing.T) {
	e := setup()
	r := call(t, e, "min", types.NewFloat(1.5), types.NewInteger(2))
	if r.Type != types.TypeFloat || r.FloatVal != 1.5 {
		t.Errorf("expected 1.5, got %s", r.String())
	}
}

func TestMaxMixedTypes(t *testing.T) {
	e := setup()
	r := call(t, e, "max", types.NewFloat(1.5), types.NewInteger(2))
	if r.Type != types.TypeInteger || r.IntVal != 2 {
		t.Errorf("expected 2, got %s", r.String())
	}
}
