package parser

import (
	"math"
	"testing"

	"github.com/suchi/sucheme-go/pkg/types"
)

func mustParse(t *testing.T, input string) *types.Value {
	t.Helper()
	v, err := ParseString(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if v == nil {
		t.Fatal("expected value, got nil")
	}
	return v
}

func mustParseAll(t *testing.T, input string) []*types.Value {
	t.Helper()
	vals, err := ParseAllString(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	return vals
}

func expectError(t *testing.T, input string) {
	t.Helper()
	_, err := ParseString(input)
	if err == nil {
		t.Fatalf("expected error for input %q", input)
	}
}

// ---------------------------------------------------------------------------
// Atoms
// ---------------------------------------------------------------------------

func TestParseInteger(t *testing.T) {
	v := mustParse(t, "42")
	if v.Type != types.TypeInteger || v.IntVal != 42 {
		t.Fatalf("expected 42, got %v", v)
	}
}

func TestParseNegativeInteger(t *testing.T) {
	v := mustParse(t, "-7")
	if v.Type != types.TypeInteger || v.IntVal != -7 {
		t.Fatalf("expected -7, got %v", v)
	}
}

func TestParseRadixIntegers(t *testing.T) {
	cases := []struct {
		input string
		want  int64
	}{
		{"#b101", 5},
		{"#o17", 15},
		{"#d42", 42},
		{"#xff", 255},
		{"#b-101", -5},
	}
	for _, c := range cases {
		v := mustParse(t, c.input)
		if v.Type != types.TypeInteger || v.IntVal != c.want {
			t.Fatalf("input %q: expected %d, got %v", c.input, c.want, v)
		}
	}
}

func TestParseFloat(t *testing.T) {
	v := mustParse(t, "3.14")
	if v.Type != types.TypeFloat || v.FloatVal != 3.14 {
		t.Fatalf("expected 3.14, got %v", v)
	}
}

func TestParseNegativeFloat(t *testing.T) {
	v := mustParse(t, "-2.5")
	if v.Type != types.TypeFloat || v.FloatVal != -2.5 {
		t.Fatalf("expected -2.5, got %v", v)
	}
}

func TestParseFloatNoLeadingDigit(t *testing.T) {
	v := mustParse(t, ".5")
	if v.Type != types.TypeFloat || v.FloatVal != 0.5 {
		t.Fatalf("expected 0.5, got %v", v)
	}
}

func TestParseSpecialFloats(t *testing.T) {
	v := mustParse(t, "+inf.0")
	if v.Type != types.TypeFloat || !math.IsInf(v.FloatVal, 1) {
		t.Fatalf("expected +inf.0, got %v", v)
	}
	v = mustParse(t, "-inf.0")
	if v.Type != types.TypeFloat || !math.IsInf(v.FloatVal, -1) {
		t.Fatalf("expected -inf.0, got %v", v)
	}
	v = mustParse(t, "+nan.0")
	if v.Type != types.TypeFloat || !math.IsNaN(v.FloatVal) {
		t.Fatalf("expected +nan.0, got %v", v)
	}
}

func TestParseBoolean(t *testing.T) {
	v := mustParse(t, "#t")
	if v.Type != types.TypeBoolean || !v.BoolVal {
		t.Fatal("expected #t")
	}
	v = mustParse(t, "#f")
	if v.Type != types.TypeBoolean || v.BoolVal {
		t.Fatal("expected #f")
	}
	v = mustParse(t, "#true")
	if v.Type != types.TypeBoolean || !v.BoolVal {
		t.Fatal("expected #true -> #t")
	}
	v = mustParse(t, "#false")
	if v.Type != types.TypeBoolean || v.BoolVal {
		t.Fatal("expected #false -> #f")
	}
}

func TestParseString(t *testing.T) {
	v := mustParse(t, `"hello world"`)
	if v.Type != types.TypeString || v.StrVal != "hello world" {
		t.Fatalf("expected hello world, got %v", v)
	}
}

func TestParseStringEscapes(t *testing.T) {
	v := mustParse(t, `"a\nb"`)
	if v.StrVal != "a\nb" {
		t.Fatalf("expected a\\nb, got %q", v.StrVal)
	}
}

func TestParseSymbol(t *testing.T) {
	v := mustParse(t, "foo")
	if v.Type != types.TypeSymbol || v.SymbolVal != "foo" {
		t.Fatalf("expected symbol foo, got %v", v)
	}
}

func TestParseCharacter(t *testing.T) {
	v := mustParse(t, `#\A`)
	if v.Type != types.TypeCharacter || v.CharVal != 'A' {
		t.Fatalf("expected character A, got %v", v)
	}
	v = mustParse(t, `#\space`)
	if v.CharVal != ' ' {
		t.Fatalf("expected space, got %v", v)
	}
}

// ---------------------------------------------------------------------------
// Lists
// ---------------------------------------------------------------------------

func TestParseEmptyList(t *testing.T) {
	v := mustParse(t, "()")
	if v.Type != types.TypeNil {
		t.Fatalf("expected nil, got %v", v)
	}
}

func TestParseSimpleList(t *testing.T) {
	v := mustParse(t, "(1 2 3)")
	list, err := v.ToList()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(list))
	}
	if list[0].IntVal != 1 || list[1].IntVal != 2 || list[2].IntVal != 3 {
		t.Fatal("element values mismatch")
	}
}

func TestParseNestedList(t *testing.T) {
	v := mustParse(t, "(1 (2 3) 4)")
	list, err := v.ToList()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(list))
	}
	inner, err := list[1].ToList()
	if err != nil {
		t.Fatal(err)
	}
	if len(inner) != 2 || inner[0].IntVal != 2 || inner[1].IntVal != 3 {
		t.Fatal("inner list mismatch")
	}
}

func TestParseDottedPair(t *testing.T) {
	v := mustParse(t, "(1 . 2)")
	if v.Type != types.TypePair {
		t.Fatal("expected pair")
	}
	if v.PairVal.Car.IntVal != 1 || v.PairVal.Cdr.IntVal != 2 {
		t.Fatal("dotted pair values mismatch")
	}
}

func TestParseImproperList(t *testing.T) {
	v := mustParse(t, "(1 2 . 3)")
	if v.Type != types.TypePair {
		t.Fatal("expected pair")
	}
	if v.PairVal.Car.IntVal != 1 {
		t.Fatal("first element mismatch")
	}
	rest := v.PairVal.Cdr
	if rest.PairVal.Car.IntVal != 2 || rest.PairVal.Cdr.IntVal != 3 {
		t.Fatal("rest elements mismatch")
	}
}

func TestParseBrackets(t *testing.T) {
	v := mustParse(t, "[1 2 3]")
	list, err := v.ToList()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(list))
	}
}

// ---------------------------------------------------------------------------
// Vectors
// ---------------------------------------------------------------------------

func TestParseVector(t *testing.T) {
	v := mustParse(t, "#(1 2 3)")
	if v.Type != types.TypeVector {
		t.Fatal("expected vector")
	}
	if len(v.VecVal) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(v.VecVal))
	}
	if v.VecVal[0].IntVal != 1 || v.VecVal[1].IntVal != 2 || v.VecVal[2].IntVal != 3 {
		t.Fatal("vector element mismatch")
	}
}

func TestParseEmptyVector(t *testing.T) {
	v := mustParse(t, "#()")
	if v.Type != types.TypeVector || len(v.VecVal) != 0 {
		t.Fatal("expected empty vector")
	}
}

// ---------------------------------------------------------------------------
// Bytevectors
// ---------------------------------------------------------------------------

func TestParseBytevector(t *testing.T) {
	v := mustParse(t, "#u8(1 2 255)")
	if v.Type != types.TypeBytevector {
		t.Fatal("expected bytevector")
	}
	if len(v.ByteVec) != 3 {
		t.Fatalf("expected 3 bytes, got %d", len(v.ByteVec))
	}
	if v.ByteVec[0] != 1 || v.ByteVec[1] != 2 || v.ByteVec[2] != 255 {
		t.Fatal("bytevector values mismatch")
	}
}

func TestParseBytevectorOutOfRange(t *testing.T) {
	expectError(t, "#u8(256)")
}

func TestParseBytevectorNonInteger(t *testing.T) {
	expectError(t, "#u8(foo)")
}

// ---------------------------------------------------------------------------
// Quote shorthands
// ---------------------------------------------------------------------------

func TestParseQuote(t *testing.T) {
	v := mustParse(t, "'foo")
	list, err := v.ToList()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 || list[0].SymbolVal != "quote" || list[1].SymbolVal != "foo" {
		t.Fatal("expected (quote foo)")
	}
}

func TestParseQuasiquote(t *testing.T) {
	v := mustParse(t, "`foo")
	list, err := v.ToList()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 || list[0].SymbolVal != "quasiquote" {
		t.Fatal("expected (quasiquote foo)")
	}
}

func TestParseUnquote(t *testing.T) {
	v := mustParse(t, ",foo")
	list, err := v.ToList()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 || list[0].SymbolVal != "unquote" {
		t.Fatal("expected (unquote foo)")
	}
}

func TestParseUnquoteSplicing(t *testing.T) {
	v := mustParse(t, ",@foo")
	list, err := v.ToList()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 || list[0].SymbolVal != "unquote-splicing" {
		t.Fatal("expected (unquote-splicing foo)")
	}
}

func TestParseQuoteList(t *testing.T) {
	v := mustParse(t, "'(1 2 3)")
	list, err := v.ToList()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 || list[0].SymbolVal != "quote" {
		t.Fatal("expected (quote (1 2 3))")
	}
	inner, err := list[1].ToList()
	if err != nil {
		t.Fatal(err)
	}
	if len(inner) != 3 {
		t.Fatal("expected inner list of 3 elements")
	}
}

// ---------------------------------------------------------------------------
// Datum comment
// ---------------------------------------------------------------------------

func TestDatumComment(t *testing.T) {
	v := mustParse(t, "#; foo 42")
	if v.Type != types.TypeInteger || v.IntVal != 42 {
		t.Fatalf("expected 42 after datum comment, got %v", v)
	}
}

func TestDatumCommentList(t *testing.T) {
	v := mustParse(t, "#; (1 2 3) 42")
	if v.Type != types.TypeInteger || v.IntVal != 42 {
		t.Fatalf("expected 42 after datum comment, got %v", v)
	}
}

// ---------------------------------------------------------------------------
// Multiple expressions
// ---------------------------------------------------------------------------

func TestParseAll(t *testing.T) {
	vals := mustParseAll(t, "1 2 3")
	if len(vals) != 3 {
		t.Fatalf("expected 3 values, got %d", len(vals))
	}
	if vals[0].IntVal != 1 || vals[1].IntVal != 2 || vals[2].IntVal != 3 {
		t.Fatal("values mismatch")
	}
}

func TestParseAllMixed(t *testing.T) {
	vals := mustParseAll(t, `(define x 42) (+ x 1)`)
	if len(vals) != 2 {
		t.Fatalf("expected 2 expressions, got %d", len(vals))
	}
}

func TestParseAllEmpty(t *testing.T) {
	vals := mustParseAll(t, "")
	if len(vals) != 0 {
		t.Fatal("expected no values for empty input")
	}
}

func TestParseAllWithComments(t *testing.T) {
	vals := mustParseAll(t, "; comment\n1 ; another\n2")
	if len(vals) != 2 {
		t.Fatalf("expected 2 values, got %d", len(vals))
	}
}

// ---------------------------------------------------------------------------
// Error cases
// ---------------------------------------------------------------------------

func TestParseUnmatchedRightParen(t *testing.T) {
	expectError(t, ")")
}

func TestParseUnmatchedRightBracket(t *testing.T) {
	expectError(t, "]")
}

func TestParseUnterminatedList(t *testing.T) {
	expectError(t, "(1 2")
}

func TestParseDotAtStartOfList(t *testing.T) {
	expectError(t, "(. 1)")
}

func TestParseUnexpectedDot(t *testing.T) {
	expectError(t, ".")
}

func TestParseUnterminatedVector(t *testing.T) {
	expectError(t, "#(1 2")
}

func TestParseUnterminatedBytevector(t *testing.T) {
	expectError(t, "#u8(1 2")
}

// ---------------------------------------------------------------------------
// Round-trip tests (parse then print)
// ---------------------------------------------------------------------------

func TestRoundTrip(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"42", "42"},
		{"-7", "-7"},
		{"3.14", "3.14"},
		{`"hello"`, `"hello"`},
		{"foo", "foo"},
		{"#t", "#t"},
		{"#f", "#f"},
		{"()", "()"},
		{"(1 2 3)", "(1 2 3)"},
		{"(1 . 2)", "(1 . 2)"},
		{"(1 2 . 3)", "(1 2 . 3)"},
		{"#(1 2 3)", "#(1 2 3)"},
		{"#u8(1 2 3)", "#u8(1 2 3)"},
		{"'foo", "(quote foo)"},
		{"`foo", "(quasiquote foo)"},
		{",foo", "(unquote foo)"},
		{",@foo", "(unquote-splicing foo)"},
	}
	for _, c := range cases {
		v := mustParse(t, c.input)
		got := v.String()
		if got != c.want {
			t.Errorf("round-trip %q: got %q, want %q", c.input, got, c.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Complex expressions
// ---------------------------------------------------------------------------

func TestParseDefine(t *testing.T) {
	v := mustParse(t, "(define (fact n) (if (<= n 1) 1 (* n (fact (- n 1)))))")
	list, err := v.ToList()
	if err != nil {
		t.Fatal(err)
	}
	if list[0].SymbolVal != "define" {
		t.Fatal("expected define")
	}
}

func TestParseLet(t *testing.T) {
	v := mustParse(t, "(let ((x 1) (y 2)) (+ x y))")
	list, err := v.ToList()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 3 || list[0].SymbolVal != "let" {
		t.Fatal("expected let with 3 elements")
	}
}

func TestParseNestedQuasiquote(t *testing.T) {
	v := mustParse(t, "`(a ,b ,@c)")
	list, err := v.ToList()
	if err != nil {
		t.Fatal(err)
	}
	if list[0].SymbolVal != "quasiquote" {
		t.Fatal("expected quasiquote")
	}
}

func TestParseEOFReturnsNil(t *testing.T) {
	p := New("")
	v, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if v != nil {
		t.Fatal("expected nil for empty input")
	}
}

func TestParseFromLexer(t *testing.T) {
	l := NewFromLexer(nil)
	_ = l // just verify it compiles
}

func TestParseBytevectorNegative(t *testing.T) {
	expectError(t, "#u8(-1)")
}

func TestDottedPairExtraAfterCdr(t *testing.T) {
	expectError(t, "(1 . 2 3)")
}
