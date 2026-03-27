package types

import (
	"math"
	"testing"
)

// ---------------------------------------------------------------------------
// Constructor tests
// ---------------------------------------------------------------------------

func TestNewInteger(t *testing.T) {
	v := NewInteger(42)
	if v.Type != TypeInteger {
		t.Fatalf("expected TypeInteger, got %v", v.Type)
	}
	if v.IntVal != 42 {
		t.Fatalf("expected 42, got %d", v.IntVal)
	}
}

func TestNewFloat(t *testing.T) {
	v := NewFloat(3.14)
	if v.Type != TypeFloat {
		t.Fatalf("expected TypeFloat, got %v", v.Type)
	}
	if v.FloatVal != 3.14 {
		t.Fatalf("expected 3.14, got %f", v.FloatVal)
	}
}

func TestNewBoolean(t *testing.T) {
	tr := NewBoolean(true)
	fa := NewBoolean(false)
	if tr != True {
		t.Fatal("NewBoolean(true) should return canonical True")
	}
	if fa != False {
		t.Fatal("NewBoolean(false) should return canonical False")
	}
	if tr.Type != TypeBoolean || !tr.BoolVal {
		t.Fatal("True should be boolean true")
	}
	if fa.Type != TypeBoolean || fa.BoolVal {
		t.Fatal("False should be boolean false")
	}
}

func TestNewChar(t *testing.T) {
	v := NewChar('A')
	if v.Type != TypeCharacter || v.CharVal != 'A' {
		t.Fatalf("expected character A, got %v", v)
	}
}

func TestNewString(t *testing.T) {
	v := NewString("hello")
	if v.Type != TypeString || v.StrVal != "hello" {
		t.Fatalf("expected string hello, got %v", v)
	}
}

func TestNewSymbol(t *testing.T) {
	v := NewSymbol("foo")
	if v.Type != TypeSymbol || v.SymbolVal != "foo" {
		t.Fatalf("expected symbol foo, got %v", v)
	}
}

func TestCons(t *testing.T) {
	car := NewInteger(1)
	cdr := NewInteger(2)
	p := Cons(car, cdr)
	if p.Type != TypePair {
		t.Fatal("expected TypePair")
	}
	if p.PairVal.Car != car || p.PairVal.Cdr != cdr {
		t.Fatal("car/cdr mismatch")
	}
}

func TestNewPair(t *testing.T) {
	p := NewPair(NewInteger(1), Nil())
	if p.Type != TypePair {
		t.Fatal("expected TypePair")
	}
}

func TestNewVector(t *testing.T) {
	elems := []*Value{NewInteger(1), NewInteger(2), NewInteger(3)}
	v := NewVector(elems)
	if v.Type != TypeVector || len(v.VecVal) != 3 {
		t.Fatalf("expected vector of length 3, got %v", v)
	}
}

func TestNilVoidEofUndefined(t *testing.T) {
	if Nil().Type != TypeNil {
		t.Fatal("Nil should be TypeNil")
	}
	if Void().Type != TypeVoid {
		t.Fatal("Void should be TypeVoid")
	}
	if Eof().Type != TypeEof {
		t.Fatal("Eof should be TypeEof")
	}
	if Undefined().Type != TypeUndefined {
		t.Fatal("Undefined should be TypeUndefined")
	}
}

func TestNewProcedure(t *testing.T) {
	fn := func(args []*Value) (*Value, error) { return Nil(), nil }
	v := NewProcedure("test", fn)
	if v.Type != TypeProcedure {
		t.Fatal("expected TypeProcedure")
	}
	if !v.ProcVal.IsBuiltin || v.ProcVal.Name != "test" {
		t.Fatal("expected builtin procedure named test")
	}
}

func TestNewLambdaProcedure(t *testing.T) {
	lam := &Lambda{
		Params: LambdaParams{Fixed: []string{"x"}, HasRest: false},
		Body:   []*Value{NewSymbol("x")},
		Name:   "my-func",
	}
	v := NewLambdaProcedure("my-func", lam)
	if v.Type != TypeProcedure || v.ProcVal.IsBuiltin {
		t.Fatal("expected non-builtin procedure")
	}
	if v.ProcVal.Lambda != lam {
		t.Fatal("lambda mismatch")
	}
}

func TestNewBytevector(t *testing.T) {
	v := NewBytevector([]byte{1, 2, 3})
	if v.Type != TypeBytevector || len(v.ByteVec) != 3 {
		t.Fatalf("expected bytevector of length 3")
	}
}

func TestNewValues(t *testing.T) {
	v := NewValues([]*Value{NewInteger(1), NewInteger(2)})
	if v.Type != TypeValues || len(v.ValuesVal) != 2 {
		t.Fatal("expected values with 2 elements")
	}
}

// ---------------------------------------------------------------------------
// List tests
// ---------------------------------------------------------------------------

func TestFromListAndToList(t *testing.T) {
	input := []*Value{NewInteger(1), NewInteger(2), NewInteger(3)}
	list := FromList(input)
	result, err := list.ToList()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(result))
	}
	for i, v := range result {
		if v.IntVal != input[i].IntVal {
			t.Fatalf("element %d mismatch", i)
		}
	}
}

func TestToListImproper(t *testing.T) {
	improper := Cons(NewInteger(1), NewInteger(2))
	_, err := improper.ToList()
	if err == nil {
		t.Fatal("expected error for improper list")
	}
}

func TestEmptyList(t *testing.T) {
	list := FromList(nil)
	if list.Type != TypeNil {
		t.Fatal("empty FromList should be nil")
	}
	result, err := list.ToList()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatal("expected empty slice")
	}
}

func TestIsList(t *testing.T) {
	if !Nil().IsList() {
		t.Fatal("nil should be a list")
	}
	if !FromList([]*Value{NewInteger(1)}).IsList() {
		t.Fatal("(1) should be a list")
	}
	if Cons(NewInteger(1), NewInteger(2)).IsList() {
		t.Fatal("(1 . 2) should not be a list")
	}
}

func TestListLength(t *testing.T) {
	if Nil().ListLength() != 0 {
		t.Fatal("nil length should be 0")
	}
	list := FromList([]*Value{NewInteger(1), NewInteger(2)})
	if list.ListLength() != 2 {
		t.Fatal("expected length 2")
	}
	if Cons(NewInteger(1), NewInteger(2)).ListLength() != -1 {
		t.Fatal("improper list length should be -1")
	}
}

// ---------------------------------------------------------------------------
// IsTrue tests
// ---------------------------------------------------------------------------

func TestIsTrue(t *testing.T) {
	cases := []struct {
		val  *Value
		want bool
	}{
		{True, true},
		{False, false},
		{NewInteger(0), true},
		{NewString(""), true},
		{Nil(), true},
		{Void(), true},
		{NewBoolean(true), true},
		{NewBoolean(false), false},
	}
	for _, c := range cases {
		if c.val.IsTrue() != c.want {
			t.Errorf("IsTrue(%v) = %v, want %v", c.val, c.val.IsTrue(), c.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Equality tests
// ---------------------------------------------------------------------------

func TestEq(t *testing.T) {
	// Same pointer
	v := NewInteger(42)
	if !v.Eq(v) {
		t.Fatal("value should be eq? to itself")
	}

	// Symbols with same name
	if !NewSymbol("foo").Eq(NewSymbol("foo")) {
		t.Fatal("symbols with same name should be eq?")
	}

	// Integers with same value
	if !NewInteger(5).Eq(NewInteger(5)) {
		t.Fatal("integers with same value should be eq?")
	}

	// Characters
	if !NewChar('a').Eq(NewChar('a')) {
		t.Fatal("same chars should be eq?")
	}

	// Booleans
	if !True.Eq(NewBoolean(true)) {
		t.Fatal("booleans should be eq?")
	}

	// Nil
	if !Nil().Eq(Nil()) {
		t.Fatal("nil should be eq? to nil")
	}

	// Different types
	if NewInteger(1).Eq(NewFloat(1.0)) {
		t.Fatal("different types should not be eq?")
	}

	// Different strings are not eq?
	if NewString("a").Eq(NewString("a")) {
		t.Fatal("different string objects should not be eq?")
	}
}

func TestEqv(t *testing.T) {
	// Floats
	if !NewFloat(3.14).Eqv(NewFloat(3.14)) {
		t.Fatal("same floats should be eqv?")
	}
	if NewFloat(3.14).Eqv(NewFloat(2.71)) {
		t.Fatal("different floats should not be eqv?")
	}
	// eqv? includes eq?
	if !NewInteger(5).Eqv(NewInteger(5)) {
		t.Fatal("eqv? should include eq? cases")
	}
}

func TestEqual(t *testing.T) {
	// Strings
	if !NewString("hello").Equal(NewString("hello")) {
		t.Fatal("equal strings should be equal?")
	}
	if NewString("hello").Equal(NewString("world")) {
		t.Fatal("different strings should not be equal?")
	}

	// Pairs
	list1 := FromList([]*Value{NewInteger(1), NewInteger(2)})
	list2 := FromList([]*Value{NewInteger(1), NewInteger(2)})
	if !list1.Equal(list2) {
		t.Fatal("equal lists should be equal?")
	}

	list3 := FromList([]*Value{NewInteger(1), NewInteger(3)})
	if list1.Equal(list3) {
		t.Fatal("different lists should not be equal?")
	}

	// Vectors
	v1 := NewVector([]*Value{NewInteger(1), NewInteger(2)})
	v2 := NewVector([]*Value{NewInteger(1), NewInteger(2)})
	if !v1.Equal(v2) {
		t.Fatal("equal vectors should be equal?")
	}

	v3 := NewVector([]*Value{NewInteger(1)})
	if v1.Equal(v3) {
		t.Fatal("different-length vectors should not be equal?")
	}

	// Bytevectors
	b1 := NewBytevector([]byte{1, 2})
	b2 := NewBytevector([]byte{1, 2})
	if !b1.Equal(b2) {
		t.Fatal("equal bytevectors should be equal?")
	}
	b3 := NewBytevector([]byte{1, 3})
	if b1.Equal(b3) {
		t.Fatal("different bytevectors should not be equal?")
	}
	b4 := NewBytevector([]byte{1})
	if b1.Equal(b4) {
		t.Fatal("different-length bytevectors should not be equal?")
	}

	// Different types
	if NewInteger(1).Equal(NewString("1")) {
		t.Fatal("different types should not be equal?")
	}
}

// ---------------------------------------------------------------------------
// Display / String tests
// ---------------------------------------------------------------------------

func TestStringRepresentations(t *testing.T) {
	cases := []struct {
		val  *Value
		want string
	}{
		{NewInteger(42), "42"},
		{NewInteger(-1), "-1"},
		{NewFloat(3.14), "3.14"},
		{NewFloat(1.0), "1.0"},
		{NewFloat(0.0), "0.0"},
		{True, "#t"},
		{False, "#f"},
		{NewChar('A'), `#\A`},
		{NewChar(' '), `#\space`},
		{NewChar('\n'), `#\newline`},
		{NewChar('\t'), `#\tab`},
		{NewChar('\r'), `#\return`},
		{NewChar('\x00'), `#\null`},
		{NewChar('\a'), `#\alarm`},
		{NewChar('\b'), `#\backspace`},
		{NewChar('\x1b'), `#\escape`},
		{NewChar('\x7f'), `#\delete`},
		{NewString("hello"), `"hello"`},
		{NewSymbol("foo"), "foo"},
		{Nil(), "()"},
		{Void(), "#<void>"},
		{Eof(), "#<eof>"},
		{Undefined(), "#<undefined>"},
	}
	for _, c := range cases {
		got := c.val.String()
		if got != c.want {
			t.Errorf("String() of %v: got %q, want %q", c.val.Type, got, c.want)
		}
	}
}

func TestStringList(t *testing.T) {
	list := FromList([]*Value{NewInteger(1), NewInteger(2), NewInteger(3)})
	if list.String() != "(1 2 3)" {
		t.Fatalf("expected (1 2 3), got %s", list.String())
	}
}

func TestStringDottedPair(t *testing.T) {
	p := Cons(NewInteger(1), NewInteger(2))
	if p.String() != "(1 . 2)" {
		t.Fatalf("expected (1 . 2), got %s", p.String())
	}
}

func TestStringImproperList(t *testing.T) {
	// (1 2 . 3)
	p := Cons(NewInteger(1), Cons(NewInteger(2), NewInteger(3)))
	if p.String() != "(1 2 . 3)" {
		t.Fatalf("expected (1 2 . 3), got %s", p.String())
	}
}

func TestStringVector(t *testing.T) {
	v := NewVector([]*Value{NewInteger(1), NewSymbol("a")})
	if v.String() != "#(1 a)" {
		t.Fatalf("expected #(1 a), got %s", v.String())
	}
}

func TestStringBytevector(t *testing.T) {
	v := NewBytevector([]byte{1, 2, 3})
	if v.String() != "#u8(1 2 3)" {
		t.Fatalf("expected #u8(1 2 3), got %s", v.String())
	}
}

func TestStringProcedure(t *testing.T) {
	fn := func(args []*Value) (*Value, error) { return Nil(), nil }
	v := NewProcedure("add", fn)
	if v.String() != "#<procedure add>" {
		t.Fatalf("expected #<procedure add>, got %s", v.String())
	}

	// Anonymous procedure
	anon := &Value{Type: TypeProcedure, ProcVal: &Procedure{Name: ""}}
	if anon.String() != "#<procedure>" {
		t.Fatalf("expected #<procedure>, got %s", anon.String())
	}
}

func TestStringSpecialFloats(t *testing.T) {
	if NewFloat(math.Inf(1)).String() != "+inf.0" {
		t.Fatal("expected +inf.0")
	}
	if NewFloat(math.Inf(-1)).String() != "-inf.0" {
		t.Fatal("expected -inf.0")
	}
	if NewFloat(math.NaN()).String() != "+nan.0" {
		t.Fatal("expected +nan.0")
	}
}

func TestDisplay(t *testing.T) {
	// Display of string omits quotes
	s := NewString("hello")
	if s.Display() != "hello" {
		t.Fatalf("expected hello, got %s", s.Display())
	}
	// Display of char returns the character itself
	c := NewChar('A')
	if c.Display() != "A" {
		t.Fatalf("expected A, got %s", c.Display())
	}
	// Display of integer is same as String
	n := NewInteger(42)
	if n.Display() != "42" {
		t.Fatalf("expected 42, got %s", n.Display())
	}
}

func TestValueTypeString(t *testing.T) {
	if TypeInteger.String() != "integer" {
		t.Fatalf("expected 'integer', got %s", TypeInteger.String())
	}
	if TypeProcedure.String() != "procedure" {
		t.Fatalf("expected 'procedure', got %s", TypeProcedure.String())
	}
	if ValueType(999).String() != "unknown" {
		t.Fatalf("expected 'unknown' for invalid type")
	}
}

func TestStringOtherTypes(t *testing.T) {
	port := &Value{Type: TypePort}
	if port.String() != "#<port>" {
		t.Fatalf("expected #<port>, got %s", port.String())
	}
	vals := NewValues([]*Value{NewInteger(1)})
	if vals.String() != "#<values>" {
		t.Fatalf("expected #<values>, got %s", vals.String())
	}
	cont := &Value{Type: TypeContinuation}
	if cont.String() != "#<continuation>" {
		t.Fatalf("expected #<continuation>, got %s", cont.String())
	}
	sr := &Value{Type: TypeSyntaxRules}
	if sr.String() != "#<syntax-rules>" {
		t.Fatalf("expected #<syntax-rules>, got %s", sr.String())
	}
}

func TestEmptyVector(t *testing.T) {
	v := NewVector(nil)
	if v.String() != "#()" {
		t.Fatalf("expected #(), got %s", v.String())
	}
}

func TestEmptyBytevector(t *testing.T) {
	v := NewBytevector(nil)
	if v.String() != "#u8()" {
		t.Fatalf("expected #u8(), got %s", v.String())
	}
}

func TestNestedList(t *testing.T) {
	inner := FromList([]*Value{NewInteger(3), NewInteger(4)})
	outer := FromList([]*Value{NewInteger(1), NewInteger(2), inner})
	if outer.String() != "(1 2 (3 4))" {
		t.Fatalf("expected (1 2 (3 4)), got %s", outer.String())
	}
}

func TestEqualNestedStructures(t *testing.T) {
	inner1 := FromList([]*Value{NewInteger(1)})
	inner2 := FromList([]*Value{NewInteger(1)})
	v1 := NewVector([]*Value{inner1})
	v2 := NewVector([]*Value{inner2})
	if !v1.Equal(v2) {
		t.Fatal("nested structures should be equal?")
	}
}
