package parser

import (
	"strings"
	"testing"

	"github.com/suchi/sucheme-go/pkg/types"
)

// =========================================================================
// Deep nesting
// =========================================================================

func TestParseDeepNesting(t *testing.T) {
	depth := 100
	input := strings.Repeat("(", depth) + "42" + strings.Repeat(")", depth)
	v := mustParse(t, input)
	for i := 0; i < depth; i++ {
		if v.Type != types.TypePair {
			t.Fatalf("expected pair at depth %d, got %s", i, v.Type)
		}
		v = v.PairVal.Car
	}
	if v.Type != types.TypeInteger || v.IntVal != 42 {
		t.Fatalf("expected 42, got %v", v)
	}
}

// =========================================================================
// Empty list variations
// =========================================================================

func TestParseEmptyListBrackets(t *testing.T) {
	v := mustParse(t, "[]")
	if v.Type != types.TypeNil {
		t.Fatalf("expected nil for [], got %s", v.Type)
	}
}

func TestParseEmptyListWithSpaces(t *testing.T) {
	v := mustParse(t, "(   )")
	if v.Type != types.TypeNil {
		t.Fatalf("expected nil for (   ), got %s", v.Type)
	}
}

func TestParseEmptyListWithComments(t *testing.T) {
	v := mustParse(t, "( ; comment\n)")
	if v.Type != types.TypeNil {
		t.Fatalf("expected nil, got %s", v.Type)
	}
}

// =========================================================================
// Dotted pair edge cases
// =========================================================================

func TestDottedPairWithLists(t *testing.T) {
	v := mustParse(t, "((1 2) . (3 4))")
	if v.Type != types.TypePair {
		t.Fatal("expected pair")
	}
	car := v.PairVal.Car
	carList, err := car.ToList()
	if err != nil {
		t.Fatalf("car should be list: %v", err)
	}
	if len(carList) != 2 {
		t.Fatalf("expected car of length 2, got %d", len(carList))
	}
}

func TestDottedPairNested(t *testing.T) {
	v := mustParse(t, "(1 . (2 . (3 . ())))")
	list, err := v.ToList()
	if err != nil {
		t.Fatalf("should be a proper list: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(list))
	}
}

// =========================================================================
// Quote abbreviation edge cases
// =========================================================================

func TestQuoteNested(t *testing.T) {
	v := mustParse(t, "''x")
	list, _ := v.ToList()
	if list[0].SymbolVal != "quote" {
		t.Fatal("expected outer quote")
	}
	inner, _ := list[1].ToList()
	if inner[0].SymbolVal != "quote" {
		t.Fatal("expected inner quote")
	}
}

func TestQuoteEmptyList(t *testing.T) {
	v := mustParse(t, "'()")
	list, _ := v.ToList()
	if len(list) != 2 || list[0].SymbolVal != "quote" || list[1].Type != types.TypeNil {
		t.Fatalf("expected (quote ()), got %s", v.String())
	}
}

func TestQuoteVector(t *testing.T) {
	v := mustParse(t, "'#(1 2 3)")
	list, _ := v.ToList()
	if list[0].SymbolVal != "quote" || list[1].Type != types.TypeVector {
		t.Fatalf("expected (quote #(1 2 3)), got %s", v.String())
	}
}

// =========================================================================
// Vector edge cases
// =========================================================================

func TestVectorSingleElement(t *testing.T) {
	v := mustParse(t, "#(42)")
	if v.Type != types.TypeVector || len(v.VecVal) != 1 || v.VecVal[0].IntVal != 42 {
		t.Fatalf("expected #(42), got %s", v.String())
	}
}

func TestVectorMixedTypes(t *testing.T) {
	v := mustParse(t, `#(1 "two" #t)`)
	if len(v.VecVal) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(v.VecVal))
	}
}

func TestVectorNestedVector(t *testing.T) {
	v := mustParse(t, "#(#(1) #(2))")
	if len(v.VecVal) != 2 || v.VecVal[0].Type != types.TypeVector {
		t.Fatal("expected nested vectors")
	}
}

// =========================================================================
// Bytevector edge cases
// =========================================================================

func TestBytevectorEmpty(t *testing.T) {
	v := mustParse(t, "#u8()")
	if v.Type != types.TypeBytevector || len(v.ByteVec) != 0 {
		t.Fatal("expected empty bytevector")
	}
}

func TestBytevectorBoundary(t *testing.T) {
	v := mustParse(t, "#u8(0 255)")
	if v.ByteVec[0] != 0 || v.ByteVec[1] != 255 {
		t.Fatal("bytevector boundary values mismatch")
	}
}

func TestBytevector256Error(t *testing.T) {
	expectError(t, "#u8(256)")
}

func TestBytevectorFloatError(t *testing.T) {
	expectError(t, "#u8(1.5)")
}

// =========================================================================
// Datum comment edge cases
// =========================================================================

func TestDatumCommentNested(t *testing.T) {
	v := mustParse(t, "#;(define x 42) 99")
	if v.IntVal != 99 {
		t.Fatalf("expected 99, got %v", v)
	}
}

func TestDatumCommentVector(t *testing.T) {
	v := mustParse(t, "#;#(1 2 3) 42")
	if v.IntVal != 42 {
		t.Fatalf("expected 42, got %v", v)
	}
}

func TestDatumCommentMultiple(t *testing.T) {
	vals := mustParseAll(t, "#;a #;b 42")
	if len(vals) != 1 || vals[0].IntVal != 42 {
		t.Fatalf("expected [42], got %v", vals)
	}
}

// =========================================================================
// Error cases
// =========================================================================

func TestMultipleDots(t *testing.T) {
	expectError(t, "(1 . 2 . 3)")
}

func TestDotAtEnd(t *testing.T) {
	expectError(t, "(1 .)")
}

// =========================================================================
// ParseAll edge cases
// =========================================================================

func TestParseAllSingleExpression(t *testing.T) {
	vals := mustParseAll(t, "42")
	if len(vals) != 1 || vals[0].IntVal != 42 {
		t.Fatal("expected single value 42")
	}
}

func TestParseAllWithDatumComments(t *testing.T) {
	vals := mustParseAll(t, "#;a 1 #;b 2 #;c 3")
	if len(vals) != 3 {
		t.Fatalf("expected 3 values, got %d", len(vals))
	}
}

func TestParseAllComplex(t *testing.T) {
	vals := mustParseAll(t, `(define (f x) (+ x 1)) (f 42) "hello" #t`)
	if len(vals) != 4 {
		t.Fatalf("expected 4 expressions, got %d", len(vals))
	}
}

// =========================================================================
// Round-trip additional cases
// =========================================================================

func TestRoundTripComplex(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{`#\A`, `#\A`},
		{`#\space`, `#\space`},
		{`#\newline`, `#\newline`},
		{"((1 2) (3 4))", "((1 2) (3 4))"},
		{"(a (b c) d)", "(a (b c) d)"},
		{"#u8()", "#u8()"},
		{"#()", "#()"},
	}
	for _, c := range cases {
		v := mustParse(t, c.input)
		got := v.String()
		if got != c.want {
			t.Errorf("round-trip %q: got %q, want %q", c.input, got, c.want)
		}
	}
}

// =========================================================================
// Special float parsing
// =========================================================================

func TestParseNan(t *testing.T) {
	v := mustParse(t, "+nan.0")
	if v.Type != types.TypeFloat {
		t.Fatal("expected float for +nan.0")
	}
}

func TestParseTrailingDot(t *testing.T) {
	v := mustParse(t, "1.")
	if v.Type != types.TypeFloat || v.FloatVal != 1.0 {
		t.Fatalf("expected 1.0, got %v", v)
	}
}

func TestParseLeadingDot(t *testing.T) {
	v := mustParse(t, ".5")
	if v.Type != types.TypeFloat || v.FloatVal != 0.5 {
		t.Fatalf("expected 0.5, got %v", v)
	}
}

// =========================================================================
// Multiple expressions in sequence
// =========================================================================

func TestParseSequentialReads(t *testing.T) {
	p := New("1 2 3")
	for i := int64(1); i <= 3; i++ {
		v, err := p.Parse()
		if err != nil {
			t.Fatal(err)
		}
		if v.IntVal != i {
			t.Errorf("expected %d, got %d", i, v.IntVal)
		}
	}
	v, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if v != nil {
		t.Fatalf("expected nil at end, got %v", v)
	}
}
