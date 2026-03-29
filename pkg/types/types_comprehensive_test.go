package types

import (
	"math"
	"testing"
)

// =========================================================================
// ValueType.String() comprehensive coverage
// =========================================================================

func TestValueTypeStringAll(t *testing.T) {
	cases := []struct {
		vt   ValueType
		want string
	}{
		{TypeInteger, "integer"},
		{TypeFloat, "float"},
		{TypeBoolean, "boolean"},
		{TypeCharacter, "character"},
		{TypeString, "string"},
		{TypeSymbol, "symbol"},
		{TypePair, "pair"},
		{TypeNil, "nil"},
		{TypeVector, "vector"},
		{TypeProcedure, "procedure"},
		{TypeVoid, "void"},
		{TypeEof, "eof"},
		{TypeUndefined, "undefined"},
		{TypePort, "port"},
		{TypeValues, "values"},
		{TypeContinuation, "continuation"},
		{TypeBytevector, "bytevector"},
		{TypeSyntaxRules, "syntax-rules"},
		{TypeRecord, "record"},
	}
	for _, c := range cases {
		got := c.vt.String()
		if got != c.want {
			t.Errorf("ValueType(%d).String() = %q, want %q", c.vt, got, c.want)
		}
	}
}

func TestValueTypeStringUnknown(t *testing.T) {
	if ValueType(9999).String() != "unknown" {
		t.Error("expected 'unknown' for invalid ValueType")
	}
}

// =========================================================================
// Value.String() coverage for all types
// =========================================================================

func TestStringRecord(t *testing.T) {
	v := NewRecord("point", map[string]*Value{"x": NewInteger(1), "y": NewInteger(2)})
	if v.String() != "#<point>" {
		t.Errorf("expected #<point>, got %s", v.String())
	}
}

func TestStringSyntaxRulesComprehensive(t *testing.T) {
	v := NewSyntaxRules(nil)
	if v.String() != "#<syntax-rules>" {
		t.Errorf("expected #<syntax-rules>, got %s", v.String())
	}
}

// =========================================================================
// NewRecord / NewSyntaxRules constructors
// =========================================================================

func TestNewRecordConstructor(t *testing.T) {
	fields := map[string]*Value{"x": NewInteger(1), "y": NewInteger(2)}
	v := NewRecord("point", fields)
	if v.Type != TypeRecord {
		t.Fatal("expected TypeRecord")
	}
	if v.RecordVal.TypeName != "point" {
		t.Errorf("expected type name 'point', got %s", v.RecordVal.TypeName)
	}
	if len(v.RecordVal.Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(v.RecordVal.Fields))
	}
}

func TestNewSyntaxRulesConstructor(t *testing.T) {
	v := NewSyntaxRules(nil)
	if v.Type != TypeSyntaxRules {
		t.Fatal("expected TypeSyntaxRules")
	}
}

// =========================================================================
// Equality edge cases
// =========================================================================

func TestEqSameObject(t *testing.T) {
	values := []*Value{
		NewInteger(42), NewFloat(3.14), True, False,
		NewChar('a'), NewString("test"), NewSymbol("sym"),
		Nil(), Void(), Eof(), Undefined(),
	}
	for _, v := range values {
		if !v.Eq(v) {
			t.Errorf("value should be eq? to itself: %s", v.String())
		}
	}
}

func TestEqDifferentTypes(t *testing.T) {
	pairs := [][2]*Value{
		{NewInteger(1), NewString("1")},
		{NewInteger(0), False},
		{NewString(""), Nil()},
		{NewChar('0'), NewInteger(48)},
	}
	for _, p := range pairs {
		if p[0].Eq(p[1]) {
			t.Errorf("different types should not be eq?: %s and %s", p[0], p[1])
		}
	}
}

func TestEqvFloatSpecial(t *testing.T) {
	nan1 := NewFloat(math.NaN())
	nan2 := NewFloat(math.NaN())
	if nan1.Eqv(nan2) {
		t.Error("NaN should not be eqv? to NaN")
	}

	inf1 := NewFloat(math.Inf(1))
	inf2 := NewFloat(math.Inf(1))
	if !inf1.Eqv(inf2) {
		t.Error("+inf.0 should be eqv? to +inf.0")
	}

	ninf1 := NewFloat(math.Inf(-1))
	ninf2 := NewFloat(math.Inf(-1))
	if !ninf1.Eqv(ninf2) {
		t.Error("-inf.0 should be eqv? to -inf.0")
	}
}

func TestEqualDeepNested(t *testing.T) {
	inner := FromList([]*Value{NewInteger(3)})
	mid := FromList([]*Value{NewInteger(2), inner})
	outer1 := FromList([]*Value{NewInteger(1), mid})

	inner2 := FromList([]*Value{NewInteger(3)})
	mid2 := FromList([]*Value{NewInteger(2), inner2})
	outer2 := FromList([]*Value{NewInteger(1), mid2})

	if !outer1.Equal(outer2) {
		t.Error("deeply nested equal structures should be equal?")
	}
}

func TestEqualVectorElementDifference(t *testing.T) {
	v1 := NewVector([]*Value{NewInteger(1), NewInteger(2), NewInteger(3)})
	v2 := NewVector([]*Value{NewInteger(1), NewInteger(99), NewInteger(3)})
	if v1.Equal(v2) {
		t.Error("vectors with different elements should not be equal?")
	}
}

func TestEqualImproperList(t *testing.T) {
	p1 := Cons(NewInteger(1), NewInteger(2))
	p2 := Cons(NewInteger(1), NewInteger(2))
	if !p1.Equal(p2) {
		t.Error("equal dotted pairs should be equal?")
	}
	p3 := Cons(NewInteger(1), NewInteger(3))
	if p1.Equal(p3) {
		t.Error("different dotted pairs should not be equal?")
	}
}

// =========================================================================
// List operations edge cases
// =========================================================================

func TestFromListSingle(t *testing.T) {
	list := FromList([]*Value{NewInteger(42)})
	if list.Type != TypePair || list.PairVal.Car.IntVal != 42 || list.PairVal.Cdr.Type != TypeNil {
		t.Fatal("single-element list error")
	}
}

func TestToListNilType(t *testing.T) {
	_, err := NewInteger(42).ToList()
	if err == nil {
		t.Fatal("expected error for ToList on integer")
	}
}

func TestListLengthLong(t *testing.T) {
	items := make([]*Value, 1000)
	for i := range items {
		items[i] = NewInteger(int64(i))
	}
	list := FromList(items)
	if list.ListLength() != 1000 {
		t.Errorf("expected length 1000, got %d", list.ListLength())
	}
}

func TestIsListOnAtom(t *testing.T) {
	if NewInteger(42).IsList() {
		t.Error("integer should not be a list")
	}
	if NewString("hello").IsList() {
		t.Error("string should not be a list")
	}
}

// =========================================================================
// Display edge cases
// =========================================================================

func TestDisplayStringComprehensive(t *testing.T) {
	s := NewString(`hello "world"`)
	disp := s.Display()
	if disp != `hello "world"` {
		t.Errorf("Display should not add quotes: got %q", disp)
	}
}

func TestDisplayCharNewline(t *testing.T) {
	c := NewChar('\n')
	if c.Display() != "\n" {
		t.Errorf("Display of newline char should be the newline itself")
	}
}

func TestDisplayNilComprehensive(t *testing.T) {
	if Nil().Display() != "()" {
		t.Error("Display of nil should be ()")
	}
}

func TestDisplayPairComprehensive(t *testing.T) {
	p := Cons(NewInteger(1), NewInteger(2))
	if p.Display() != "(1 . 2)" {
		t.Errorf("expected (1 . 2), got %s", p.Display())
	}
}

func TestDisplayVectorComprehensive(t *testing.T) {
	v := NewVector([]*Value{NewInteger(1), NewInteger(2)})
	d := v.Display()
	if d != "#(1 2)" {
		t.Errorf("expected #(1 2), got %s", d)
	}
}

// =========================================================================
// String formatting edge cases
// =========================================================================

func TestFormatFloatInteger(t *testing.T) {
	if NewFloat(1.0).String() != "1.0" {
		t.Errorf("expected 1.0, got %s", NewFloat(1.0).String())
	}
}

func TestFormatFloatVerySmall(t *testing.T) {
	v := NewFloat(1e-10)
	s := v.String()
	if s == "" {
		t.Error("string should not be empty for very small float")
	}
}

func TestStringEscapedInString(t *testing.T) {
	v := NewString(`he said "hi"`)
	s := v.String()
	if s != `"he said \"hi\""` {
		t.Errorf("expected escaped quotes, got %s", s)
	}
}

func TestStringWithBackslash(t *testing.T) {
	v := NewString(`a\b`)
	s := v.String()
	if s != `"a\\b"` {
		t.Errorf("expected escaped backslash, got %s", s)
	}
}

// =========================================================================
// IsTrue edge cases
// =========================================================================

func TestIsTrueAllTypes(t *testing.T) {
	trueValues := []*Value{
		NewInteger(0), NewInteger(1), NewFloat(0.0),
		NewString(""), NewChar(0), Nil(), Void(),
		NewVector(nil), NewBytevector(nil), True,
	}
	for _, v := range trueValues {
		if !v.IsTrue() {
			t.Errorf("expected %s to be truthy", v.String())
		}
	}
	if False.IsTrue() {
		t.Error("#f should be falsy")
	}
}

// =========================================================================
// Bytevector equality edge cases
// =========================================================================

func TestBytevectorEqualEmpty(t *testing.T) {
	b1 := NewBytevector(nil)
	b2 := NewBytevector([]byte{})
	if !b1.Equal(b2) {
		t.Error("empty bytevectors should be equal?")
	}
}

func TestBytevectorEqualLarge(t *testing.T) {
	a := make([]byte, 1000)
	b := make([]byte, 1000)
	for i := range a {
		a[i] = byte(i % 256)
		b[i] = byte(i % 256)
	}
	if !NewBytevector(a).Equal(NewBytevector(b)) {
		t.Error("identical large bytevectors should be equal?")
	}
	b[999] = 0
	if NewBytevector(a).Equal(NewBytevector(b)) {
		t.Error("different large bytevectors should not be equal?")
	}
}

// =========================================================================
// Values type
// =========================================================================

func TestNewValuesEmpty(t *testing.T) {
	v := NewValues(nil)
	if v.Type != TypeValues || len(v.ValuesVal) != 0 {
		t.Error("expected empty values")
	}
}

func TestNewValuesSingle(t *testing.T) {
	v := NewValues([]*Value{NewInteger(42)})
	if len(v.ValuesVal) != 1 || v.ValuesVal[0].IntVal != 42 {
		t.Error("expected values with single element 42")
	}
}
