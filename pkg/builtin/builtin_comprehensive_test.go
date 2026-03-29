package builtin

import (
	"math"
	"testing"

	"github.com/suchi/sucheme-go/pkg/types"
)

// =========================================================================
// String operations (42.8% coverage - many uncovered)
// =========================================================================

func TestNumberToString(t *testing.T) {
	e := setup()
	r := call(t, e, "number->string", types.NewInteger(42))
	if r.StrVal != "42" {
		t.Errorf("expected '42', got %q", r.StrVal)
	}
	r = call(t, e, "number->string", types.NewFloat(3.14))
	if r.StrVal != "3.14" {
		t.Errorf("expected '3.14', got %q", r.StrVal)
	}
}

func TestNumberToStringTypeError(t *testing.T) {
	e := setup()
	err := callErr(t, e, "number->string", types.NewString("x"))
	if err == nil {
		t.Error("expected type error")
	}
}

func TestStringToList(t *testing.T) {
	e := setup()
	r := call(t, e, "string->list", types.NewString("abc"))
	lst, _ := r.ToList()
	if len(lst) != 3 || lst[0].CharVal != 'a' || lst[1].CharVal != 'b' || lst[2].CharVal != 'c' {
		t.Errorf("expected (#\\a #\\b #\\c), got %s", r.String())
	}
}

func TestStringToListEmpty(t *testing.T) {
	e := setup()
	r := call(t, e, "string->list", types.NewString(""))
	if r.Type != types.TypeNil {
		t.Errorf("expected (), got %s", r.String())
	}
}

func TestListToString(t *testing.T) {
	e := setup()
	lst := types.FromList([]*types.Value{types.NewChar('h'), types.NewChar('i')})
	r := call(t, e, "list->string", lst)
	if r.StrVal != "hi" {
		t.Errorf("expected 'hi', got %q", r.StrVal)
	}
}

func TestListToStringEmpty(t *testing.T) {
	e := setup()
	r := call(t, e, "list->string", types.Nil())
	if r.StrVal != "" {
		t.Errorf("expected empty string, got %q", r.StrVal)
	}
}

func TestListToStringTypeError(t *testing.T) {
	e := setup()
	lst := types.FromList([]*types.Value{types.NewInteger(1)})
	err := callErr(t, e, "list->string", lst)
	if err == nil {
		t.Error("expected error for non-character in list")
	}
}

func TestStringLessThan(t *testing.T) {
	e := setup()
	r := call(t, e, "string<?", types.NewString("abc"), types.NewString("abd"))
	if !r.BoolVal {
		t.Error("expected #t for string<? abc abd")
	}
	r = call(t, e, "string<?", types.NewString("abd"), types.NewString("abc"))
	if r.BoolVal {
		t.Error("expected #f for string<? abd abc")
	}
}

func TestStringGreaterThan(t *testing.T) {
	e := setup()
	r := call(t, e, "string>?", types.NewString("abd"), types.NewString("abc"))
	if !r.BoolVal {
		t.Error("expected #t for string>? abd abc")
	}
}

func TestStringCopy(t *testing.T) {
	e := setup()
	r := call(t, e, "string-copy", types.NewString("hello"))
	if r.StrVal != "hello" {
		t.Errorf("expected 'hello', got %q", r.StrVal)
	}
}

func TestStringSetBoundary(t *testing.T) {
	e := setup()
	s := types.NewString("abc")
	call(t, e, "string-set!", s, types.NewInteger(0), types.NewChar('A'))
	if s.StrVal != "Abc" {
		t.Errorf("expected 'Abc', got %q", s.StrVal)
	}
	call(t, e, "string-set!", s, types.NewInteger(2), types.NewChar('C'))
	if s.StrVal != "AbC" {
		t.Errorf("expected 'AbC', got %q", s.StrVal)
	}
}

func TestStringSetOutOfRange(t *testing.T) {
	e := setup()
	err := callErr(t, e, "string-set!", types.NewString("abc"), types.NewInteger(5), types.NewChar('x'))
	if err == nil {
		t.Error("expected out of range error")
	}
}

func TestStringSetNegativeIndex(t *testing.T) {
	e := setup()
	err := callErr(t, e, "string-set!", types.NewString("abc"), types.NewInteger(-1), types.NewChar('x'))
	if err == nil {
		t.Error("expected out of range error for negative index")
	}
}

func TestStringSetTypeErrors(t *testing.T) {
	e := setup()
	err := callErr(t, e, "string-set!", types.NewInteger(42), types.NewInteger(0), types.NewChar('x'))
	if err == nil {
		t.Error("expected type error for non-string")
	}
	err = callErr(t, e, "string-set!", types.NewString("a"), types.NewString("0"), types.NewChar('x'))
	if err == nil {
		t.Error("expected type error for non-integer index")
	}
	err = callErr(t, e, "string-set!", types.NewString("a"), types.NewInteger(0), types.NewInteger(65))
	if err == nil {
		t.Error("expected type error for non-character value")
	}
}

func TestMakeString(t *testing.T) {
	e := setup()
	r := call(t, e, "make-string", types.NewInteger(5))
	if len([]rune(r.StrVal)) != 5 {
		t.Errorf("expected string of length 5, got %d", len([]rune(r.StrVal)))
	}
}

func TestMakeStringWithFill(t *testing.T) {
	e := setup()
	r := call(t, e, "make-string", types.NewInteger(3), types.NewChar('x'))
	if r.StrVal != "xxx" {
		t.Errorf("expected 'xxx', got %q", r.StrVal)
	}
}

func TestMakeStringZero(t *testing.T) {
	e := setup()
	r := call(t, e, "make-string", types.NewInteger(0))
	if r.StrVal != "" {
		t.Errorf("expected empty string, got %q", r.StrVal)
	}
}

func TestMakeStringNegative(t *testing.T) {
	e := setup()
	err := callErr(t, e, "make-string", types.NewInteger(-1))
	if err == nil {
		t.Error("expected error for negative length")
	}
}

func TestMakeStringFillTypeError(t *testing.T) {
	e := setup()
	err := callErr(t, e, "make-string", types.NewInteger(3), types.NewInteger(65))
	if err == nil {
		t.Error("expected type error for non-character fill")
	}
}

func TestSubstringBoundary(t *testing.T) {
	e := setup()
	r := call(t, e, "substring", types.NewString("hello"), types.NewInteger(0), types.NewInteger(5))
	if r.StrVal != "hello" {
		t.Errorf("expected 'hello', got %q", r.StrVal)
	}
	r = call(t, e, "substring", types.NewString("hello"), types.NewInteger(2), types.NewInteger(2))
	if r.StrVal != "" {
		t.Errorf("expected empty string, got %q", r.StrVal)
	}
}

func TestSubstringOutOfRange(t *testing.T) {
	e := setup()
	err := callErr(t, e, "substring", types.NewString("hello"), types.NewInteger(0), types.NewInteger(10))
	if err == nil {
		t.Error("expected out of range error")
	}
}

func TestSubstringNegativeStart(t *testing.T) {
	e := setup()
	err := callErr(t, e, "substring", types.NewString("hello"), types.NewInteger(-1), types.NewInteger(3))
	if err == nil {
		t.Error("expected error for negative start")
	}
}

func TestSubstringEndBeforeStart(t *testing.T) {
	e := setup()
	err := callErr(t, e, "substring", types.NewString("hello"), types.NewInteger(3), types.NewInteger(1))
	if err == nil {
		t.Error("expected error for end < start")
	}
}

func TestStringToNumberWithRadix(t *testing.T) {
	e := setup()
	err := callErr(t, e, "string->number", types.NewString("10"), types.NewInteger(2))
	if err == nil {
		t.Error("expected error for radix (not yet supported)")
	}
}

func TestStringToNumberEmpty(t *testing.T) {
	e := setup()
	r := call(t, e, "string->number", types.NewString(""))
	if r.Type != types.TypeBoolean || r.BoolVal {
		t.Error("expected #f for empty string")
	}
}

func TestStringToNumberWhitespace(t *testing.T) {
	e := setup()
	r := call(t, e, "string->number", types.NewString("  42  "))
	if r.IntVal != 42 {
		t.Errorf("expected 42, got %s", r.String())
	}
}

func TestStringAppendTypeError(t *testing.T) {
	e := setup()
	err := callErr(t, e, "string-append", types.NewString("a"), types.NewInteger(1))
	if err == nil {
		t.Error("expected type error")
	}
}

// =========================================================================
// Type predicates (63.9% coverage)
// =========================================================================

func TestProcedurePredicate(t *testing.T) {
	e := setup()
	fn := types.NewProcedure("test", func(args []*types.Value) (*types.Value, error) { return types.Nil(), nil })
	r := call(t, e, "procedure?", fn)
	if !r.BoolVal {
		t.Error("expected #t for procedure?")
	}
	r = call(t, e, "procedure?", types.NewInteger(1))
	if r.BoolVal {
		t.Error("expected #f for integer")
	}
}

func TestListPredicate(t *testing.T) {
	e := setup()
	r := call(t, e, "list?", types.Nil())
	if !r.BoolVal {
		t.Error("expected #t for ()")
	}
	lst := types.FromList([]*types.Value{types.NewInteger(1)})
	r = call(t, e, "list?", lst)
	if !r.BoolVal {
		t.Error("expected #t for proper list")
	}
	r = call(t, e, "list?", types.Cons(types.NewInteger(1), types.NewInteger(2)))
	if r.BoolVal {
		t.Error("expected #f for improper list")
	}
}

func TestVoidPredicate(t *testing.T) {
	e := setup()
	r := call(t, e, "void?", types.Void())
	if !r.BoolVal {
		t.Error("expected #t for void?")
	}
	r = call(t, e, "void?", types.NewInteger(0))
	if r.BoolVal {
		t.Error("expected #f for integer")
	}
}

// =========================================================================
// Comparison operations (65.1% coverage)
// =========================================================================

func TestGreaterThan(t *testing.T) {
	e := setup()
	r := call(t, e, ">", types.NewInteger(3), types.NewInteger(2))
	if !r.BoolVal {
		t.Error("expected #t")
	}
	r = call(t, e, ">", types.NewInteger(2), types.NewInteger(3))
	if r.BoolVal {
		t.Error("expected #f")
	}
}

func TestLessEqual(t *testing.T) {
	e := setup()
	r := call(t, e, "<=", types.NewInteger(2), types.NewInteger(3))
	if !r.BoolVal {
		t.Error("expected #t")
	}
	r = call(t, e, "<=", types.NewInteger(3), types.NewInteger(3))
	if !r.BoolVal {
		t.Error("expected #t for equal values")
	}
	r = call(t, e, "<=", types.NewInteger(4), types.NewInteger(3))
	if r.BoolVal {
		t.Error("expected #f")
	}
}

func TestGreaterEqual(t *testing.T) {
	e := setup()
	r := call(t, e, ">=", types.NewInteger(3), types.NewInteger(2))
	if !r.BoolVal {
		t.Error("expected #t")
	}
	r = call(t, e, ">=", types.NewInteger(3), types.NewInteger(3))
	if !r.BoolVal {
		t.Error("expected #t for equal values")
	}
	r = call(t, e, ">=", types.NewInteger(2), types.NewInteger(3))
	if r.BoolVal {
		t.Error("expected #f")
	}
}

func TestChainedComparisons(t *testing.T) {
	e := setup()
	r := call(t, e, "<", types.NewInteger(1), types.NewInteger(2), types.NewInteger(3), types.NewInteger(4))
	if !r.BoolVal {
		t.Error("expected #t for (< 1 2 3 4)")
	}
	r = call(t, e, "<", types.NewInteger(1), types.NewInteger(3), types.NewInteger(2))
	if r.BoolVal {
		t.Error("expected #f for (< 1 3 2)")
	}
}

func TestComparisonMixed(t *testing.T) {
	e := setup()
	r := call(t, e, "<", types.NewInteger(1), types.NewFloat(1.5))
	if !r.BoolVal {
		t.Error("expected #t for (< 1 1.5)")
	}
}

// =========================================================================
// Arithmetic edge cases
// =========================================================================

func TestSubFloat(t *testing.T) {
	e := setup()
	r := call(t, e, "-", types.NewFloat(3.5))
	if r.FloatVal != -3.5 {
		t.Errorf("expected -3.5, got %f", r.FloatVal)
	}
}

func TestMulFloat(t *testing.T) {
	e := setup()
	r := call(t, e, "*", types.NewFloat(2.5), types.NewInteger(2))
	if r.FloatVal != 5.0 {
		t.Errorf("expected 5.0, got %f", r.FloatVal)
	}
}

func TestDivFloat(t *testing.T) {
	e := setup()
	r := call(t, e, "/", types.NewFloat(10.0), types.NewFloat(3.0))
	if r.Type != types.TypeFloat {
		t.Error("expected float result")
	}
}

func TestDivNoArgs(t *testing.T) {
	e := setup()
	err := callErr(t, e, "/")
	if err == nil {
		t.Error("expected error for / with no args")
	}
}

func TestDivFloatByZero(t *testing.T) {
	e := setup()
	err := callErr(t, e, "/", types.NewFloat(1.0), types.NewFloat(0.0))
	if err == nil {
		t.Error("expected error for float division by zero")
	}
}

// =========================================================================
// Vector edge cases (69.5% coverage)
// =========================================================================

func TestVectorLength(t *testing.T) {
	e := setup()
	v := types.NewVector([]*types.Value{types.NewInteger(1), types.NewInteger(2), types.NewInteger(3)})
	r := call(t, e, "vector-length", v)
	if r.IntVal != 3 {
		t.Errorf("expected 3, got %d", r.IntVal)
	}
}

func TestVectorLengthEmpty(t *testing.T) {
	e := setup()
	v := types.NewVector(nil)
	r := call(t, e, "vector-length", v)
	if r.IntVal != 0 {
		t.Errorf("expected 0, got %d", r.IntVal)
	}
}

func TestVectorSetOutOfRange(t *testing.T) {
	e := setup()
	v := types.NewVector([]*types.Value{types.NewInteger(1)})
	err := callErr(t, e, "vector-set!", v, types.NewInteger(5), types.NewInteger(99))
	if err == nil {
		t.Error("expected out of range error")
	}
}

func TestVectorFill(t *testing.T) {
	e := setup()
	v := types.NewVector([]*types.Value{types.NewInteger(1), types.NewInteger(2), types.NewInteger(3)})
	call(t, e, "vector-fill!", v, types.NewInteger(0))
	for i, elem := range v.VecVal {
		if elem.IntVal != 0 {
			t.Errorf("element %d: expected 0, got %d", i, elem.IntVal)
		}
	}
}

func TestMakeVectorNoFill(t *testing.T) {
	e := setup()
	r := call(t, e, "make-vector", types.NewInteger(3))
	if r.Type != types.TypeVector || len(r.VecVal) != 3 {
		t.Error("make-vector without fill failed")
	}
}

// =========================================================================
// Bytevector edge cases (66.7% coverage)
// =========================================================================

func TestBytevectorLength(t *testing.T) {
	e := setup()
	bv := types.NewBytevector([]byte{1, 2, 3})
	r := call(t, e, "bytevector-length", bv)
	if r.IntVal != 3 {
		t.Errorf("expected 3, got %d", r.IntVal)
	}
}

func TestBytevectorLengthEmpty(t *testing.T) {
	e := setup()
	bv := types.NewBytevector(nil)
	r := call(t, e, "bytevector-length", bv)
	if r.IntVal != 0 {
		t.Errorf("expected 0, got %d", r.IntVal)
	}
}

func TestBytevectorRefOutOfRange(t *testing.T) {
	e := setup()
	bv := types.NewBytevector([]byte{1})
	err := callErr(t, e, "bytevector-u8-ref", bv, types.NewInteger(5))
	if err == nil {
		t.Error("expected out of range error")
	}
}

func TestBytevectorSetOutOfRange(t *testing.T) {
	e := setup()
	bv := types.NewBytevector([]byte{0})
	err := callErr(t, e, "bytevector-u8-set!", bv, types.NewInteger(5), types.NewInteger(1))
	if err == nil {
		t.Error("expected out of range error")
	}
}

func TestBytevectorCopy(t *testing.T) {
	e := setup()
	bv := types.NewBytevector([]byte{1, 2, 3})
	r := call(t, e, "bytevector-copy", bv)
	if len(r.ByteVec) != 3 || r.ByteVec[0] != 1 || r.ByteVec[2] != 3 {
		t.Error("bytevector-copy failed")
	}
	r.ByteVec[0] = 99
	if bv.ByteVec[0] != 1 {
		t.Error("copy should be independent")
	}
}

func TestMakeBytevectorNoFill(t *testing.T) {
	e := setup()
	r := call(t, e, "make-bytevector", types.NewInteger(3))
	if r.Type != types.TypeBytevector || len(r.ByteVec) != 3 {
		t.Error("make-bytevector without fill failed")
	}
}

func TestBytevectorAppendEmpty(t *testing.T) {
	e := setup()
	r := call(t, e, "bytevector-append")
	if r.Type != types.TypeBytevector || len(r.ByteVec) != 0 {
		t.Error("expected empty bytevector from (bytevector-append)")
	}
}

// =========================================================================
// I/O operations (40% coverage)
// =========================================================================

func TestNewline(t *testing.T) {
	e := setup()
	r := call(t, e, "newline")
	if r.Type != types.TypeVoid {
		t.Error("newline should return void")
	}
}

func TestWrite(t *testing.T) {
	e := setup()
	r := call(t, e, "write", types.NewString("hello"))
	if r.Type != types.TypeVoid {
		t.Error("write should return void")
	}
}

func TestErrorWithIrritants(t *testing.T) {
	e := setup()
	err := callErr(t, e, "error", types.NewString("msg"), types.NewInteger(1), types.NewInteger(2))
	if err == nil {
		t.Error("expected error")
	}
}

// =========================================================================
// Arithmetic: special float values
// =========================================================================

func TestAddInfNan(t *testing.T) {
	e := setup()
	r := call(t, e, "+", types.NewFloat(math.Inf(1)), types.NewFloat(1.0))
	if !math.IsInf(r.FloatVal, 1) {
		t.Error("expected +inf.0")
	}
}

func TestZeroPredicateFloat(t *testing.T) {
	e := setup()
	r := call(t, e, "zero?", types.NewFloat(0.0))
	if !r.BoolVal {
		t.Error("expected #t for (zero? 0.0)")
	}
}

func TestPositivePredicateFloat(t *testing.T) {
	e := setup()
	r := call(t, e, "positive?", types.NewFloat(0.5))
	if !r.BoolVal {
		t.Error("expected #t for (positive? 0.5)")
	}
}

func TestNegativePredicateFloat(t *testing.T) {
	e := setup()
	r := call(t, e, "negative?", types.NewFloat(-0.5))
	if !r.BoolVal {
		t.Error("expected #t for (negative? -0.5)")
	}
}

// =========================================================================
// List operations additional coverage
// =========================================================================

func TestListTailBasic(t *testing.T) {
	e := setup()
	lst := types.FromList([]*types.Value{types.NewInteger(1), types.NewInteger(2), types.NewInteger(3)})
	r := call(t, e, "list-tail", lst, types.NewInteger(1))
	items, _ := r.ToList()
	if len(items) != 2 || items[0].IntVal != 2 {
		t.Errorf("expected (2 3), got %s", r.String())
	}
}

func TestListTailZero(t *testing.T) {
	e := setup()
	lst := types.FromList([]*types.Value{types.NewInteger(1)})
	r := call(t, e, "list-tail", lst, types.NewInteger(0))
	items, _ := r.ToList()
	if len(items) != 1 {
		t.Error("list-tail 0 should return the original list")
	}
}

func TestAppendSingle(t *testing.T) {
	e := setup()
	lst := types.FromList([]*types.Value{types.NewInteger(1), types.NewInteger(2)})
	r := call(t, e, "append", lst)
	items, _ := r.ToList()
	if len(items) != 2 {
		t.Error("append with single list should return it")
	}
}

func TestReverseEmpty(t *testing.T) {
	e := setup()
	r := call(t, e, "reverse", types.Nil())
	if r.Type != types.TypeNil {
		t.Error("reverse of () should be ()")
	}
}

func TestReverseSingle(t *testing.T) {
	e := setup()
	lst := types.FromList([]*types.Value{types.NewInteger(1)})
	r := call(t, e, "reverse", lst)
	items, _ := r.ToList()
	if len(items) != 1 || items[0].IntVal != 1 {
		t.Error("reverse of (1) should be (1)")
	}
}
