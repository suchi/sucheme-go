package lexer

import (
	"strings"
	"testing"
)

// =========================================================================
// Empty / whitespace-only inputs
// =========================================================================

func TestEmptyInput(t *testing.T) {
	l := New("")
	tok := mustToken(t, l)
	if tok.Type != TokenEof {
		t.Fatalf("expected Eof for empty input, got %s", tok.Type)
	}
}

func TestWhitespaceOnly(t *testing.T) {
	inputs := []string{"   ", "\t\t", "\n\n\n", "  \t \n  "}
	for _, input := range inputs {
		l := New(input)
		tok := mustToken(t, l)
		if tok.Type != TokenEof {
			t.Fatalf("expected Eof for whitespace-only %q, got %s", input, tok.Type)
		}
	}
}

func TestCommentOnly(t *testing.T) {
	l := New("; just a comment")
	tok := mustToken(t, l)
	if tok.Type != TokenEof {
		t.Fatalf("expected Eof for comment-only input, got %s", tok.Type)
	}
}

// =========================================================================
// Unicode and special characters
// =========================================================================

func TestUnicodeSymbols(t *testing.T) {
	l := New("λ")
	tok := mustToken(t, l)
	expectToken(t, tok, TokenSymbol, "λ")
}

func TestUnicodeString(t *testing.T) {
	l := New(`"日本語テスト"`)
	tok := mustToken(t, l)
	expectToken(t, tok, TokenString, "日本語テスト")
}

func TestUnicodeCharacterLiteral(t *testing.T) {
	l := New(`#\λ`)
	tok := mustToken(t, l)
	expectToken(t, tok, TokenCharacter, "λ")
}

// =========================================================================
// String edge cases
// =========================================================================

func TestStringWithOnlyEscapes(t *testing.T) {
	l := New(`"\n\t\r"`)
	tok := mustToken(t, l)
	expectToken(t, tok, TokenString, "\n\t\r")
}

func TestStringHexEscapeLowercase(t *testing.T) {
	l := New(`"\x61;"`)
	tok := mustToken(t, l)
	expectToken(t, tok, TokenString, "a")
}

func TestStringHexEscapeUppercase(t *testing.T) {
	l := New(`"\x4F;"`)
	tok := mustToken(t, l)
	expectToken(t, tok, TokenString, "O")
}

func TestStringWithNullChar(t *testing.T) {
	l := New(`"\0"`)
	tok := mustToken(t, l)
	expectToken(t, tok, TokenString, "\x00")
}

func TestStringLineContinuationMultipleSpaces(t *testing.T) {
	input := "\"abc\\\n      def\""
	l := New(input)
	tok := mustToken(t, l)
	expectToken(t, tok, TokenString, "abcdef")
}

func TestStringLineContinuationNoSpaces(t *testing.T) {
	input := "\"abc\\\ndef\""
	l := New(input)
	tok := mustToken(t, l)
	expectToken(t, tok, TokenString, "abcdef")
}

// =========================================================================
// Number edge cases
// =========================================================================

func TestZero(t *testing.T) {
	l := New("0")
	tok := mustToken(t, l)
	expectToken(t, tok, TokenInteger, "0")
}

func TestNegativeZero(t *testing.T) {
	l := New("-0")
	tok := mustToken(t, l)
	expectToken(t, tok, TokenInteger, "-0")
}

func TestPositivePrefixNumber(t *testing.T) {
	l := New("+42")
	tok := mustToken(t, l)
	expectToken(t, tok, TokenInteger, "+42")
}

func TestFloatWithExponent(t *testing.T) {
	cases := []struct {
		input, value string
	}{
		{"1e5", "1e5"},
		{"1E5", "1E5"},
		{"1.5e+2", "1.5e+2"},
		{"1.5e-2", "1.5e-2"},
		{"-1e10", "-1e10"},
		{"+1e10", "+1e10"},
	}
	for _, c := range cases {
		l := New(c.input)
		tok := mustToken(t, l)
		expectToken(t, tok, TokenFloat, c.value)
	}
}

func TestRadixBinary(t *testing.T) {
	cases := []struct {
		input string
		value string
	}{
		{"#b0", "#b0"},
		{"#b1", "#b1"},
		{"#b1010", "#b1010"},
		{"#b+1010", "#b+1010"},
		{"#b-1010", "#b-1010"},
	}
	for _, c := range cases {
		l := New(c.input)
		tok := mustToken(t, l)
		expectToken(t, tok, TokenInteger, c.value)
	}
}

func TestRadixOctal(t *testing.T) {
	l := New("#o777")
	tok := mustToken(t, l)
	expectToken(t, tok, TokenInteger, "#o777")
}

func TestRadixDecimal(t *testing.T) {
	l := New("#d999")
	tok := mustToken(t, l)
	expectToken(t, tok, TokenInteger, "#d999")
}

func TestRadixHex(t *testing.T) {
	cases := []string{"#xDEAD", "#xbeef", "#x0", "#xFf"}
	for _, c := range cases {
		l := New(c)
		tok := mustToken(t, l)
		expectToken(t, tok, TokenInteger, c)
	}
}

func TestRadixInvalidDigits(t *testing.T) {
	l := New("#b2")
	_, err := l.NextToken()
	if err == nil {
		t.Fatal("expected error for #b2 (not a valid binary digit)")
	}
}

// =========================================================================
// Symbol edge cases
// =========================================================================

func TestSymbolWithSpecialChars(t *testing.T) {
	cases := []string{
		"->", "<->", "string?", "set!", "null?",
		"a1", "x-y-z",
	}
	for _, c := range cases {
		l := New(c)
		tok := mustToken(t, l)
		expectToken(t, tok, TokenSymbol, c)
	}
}

func TestPlusMinusAsSymbol(t *testing.T) {
	for _, s := range []string{"+", "-"} {
		l := New(s + " ")
		tok := mustToken(t, l)
		expectToken(t, tok, TokenSymbol, s)
	}
}

// =========================================================================
// Multi-token sequences
// =========================================================================

func TestNestedParens(t *testing.T) {
	tokens, err := AllTokens("((()))")
	if err != nil {
		t.Fatal(err)
	}
	expected := []TokenType{TokenLeftParen, TokenLeftParen, TokenLeftParen, TokenRightParen, TokenRightParen, TokenRightParen}
	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
	}
	for i, tok := range tokens {
		if tok.Type != expected[i] {
			t.Errorf("token %d: expected %s, got %s", i, expected[i], tok.Type)
		}
	}
}

func TestMixedExpression(t *testing.T) {
	tokens, err := AllTokens(`(define x "hello") ; comment`)
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 5 {
		t.Fatalf("expected 5 tokens, got %d", len(tokens))
	}
}

// =========================================================================
// Comment edge cases
// =========================================================================

func TestBlockCommentAtEndOfInput(t *testing.T) {
	l := New("#| comment |#")
	tok := mustToken(t, l)
	if tok.Type != TokenEof {
		t.Fatalf("expected Eof after block comment, got %s", tok.Type)
	}
}

func TestDeeplyNestedBlockComment(t *testing.T) {
	l := New("#| #| #| deep |# mid |# outer |# 42")
	tok := mustToken(t, l)
	expectToken(t, tok, TokenInteger, "42")
}

func TestDatumCommentBeforeList(t *testing.T) {
	l := New("#;(1 2 3)42")
	tok := mustToken(t, l)
	expectToken(t, tok, TokenSymbol, "#;")
}

// =========================================================================
// Token.String() coverage
// =========================================================================

func TestTokenTypeStringCoverage(t *testing.T) {
	types := []TokenType{
		TokenInteger, TokenFloat, TokenString, TokenSymbol,
		TokenBoolean, TokenCharacter, TokenLeftParen, TokenRightParen,
		TokenLeftBracket, TokenRightBracket, TokenQuote, TokenQuasiquote,
		TokenUnquote, TokenUnquoteSplicing, TokenDot, TokenHashLeftParen,
		TokenHashU8LeftParen, TokenEof,
	}
	for _, tt := range types {
		s := tt.String()
		if s == "" {
			t.Errorf("String() for token type %d should not be empty", tt)
		}
	}
}

func TestTokenStringFormat(t *testing.T) {
	tok := Token{Type: TokenSymbol, Value: "foo", Line: 3, Column: 5}
	s := tok.String()
	if !strings.Contains(s, "foo") {
		t.Errorf("token String() should contain the value, got %q", s)
	}
}

// =========================================================================
// AllTokens error case
// =========================================================================

func TestAllTokensError(t *testing.T) {
	_, err := AllTokens(`"unterminated`)
	if err == nil {
		t.Fatal("expected error from AllTokens with unterminated string")
	}
}

// =========================================================================
// Dot-prefix edge cases
// =========================================================================

func TestDotFollowedByDigit(t *testing.T) {
	l := New(".123")
	tok := mustToken(t, l)
	expectToken(t, tok, TokenFloat, ".123")
}

func TestDotAsDelimiter(t *testing.T) {
	l := New("(a . b)")
	expectToken(t, mustToken(t, l), TokenLeftParen, "(")
	expectToken(t, mustToken(t, l), TokenSymbol, "a")
	expectToken(t, mustToken(t, l), TokenDot, ".")
	expectToken(t, mustToken(t, l), TokenSymbol, "b")
	expectToken(t, mustToken(t, l), TokenRightParen, ")")
}

// =========================================================================
// Character literal edge cases
// =========================================================================

func TestCharLiteralDigit(t *testing.T) {
	l := New(`#\0`)
	tok := mustToken(t, l)
	expectToken(t, tok, TokenCharacter, "0")
}

func TestCharLiteralSpecialChars(t *testing.T) {
	cases := []struct {
		input string
		value string
	}{
		{`#\ `, " "},
		{`#\space`, " "},
		{`#\newline`, "\n"},
		{`#\tab`, "\t"},
		{`#\return`, "\r"},
	}
	for _, c := range cases {
		l := New(c.input)
		tok := mustToken(t, l)
		expectToken(t, tok, TokenCharacter, c.value)
	}
}

// =========================================================================
// Signed number vs symbol disambiguation
// =========================================================================

func TestSignedInf(t *testing.T) {
	l := New("+inf.0")
	tok := mustToken(t, l)
	expectToken(t, tok, TokenFloat, "+inf.0")

	l = New("-inf.0")
	tok = mustToken(t, l)
	expectToken(t, tok, TokenFloat, "-inf.0")
}

func TestSignedNan(t *testing.T) {
	l := New("+nan.0")
	tok := mustToken(t, l)
	expectToken(t, tok, TokenFloat, "+nan.0")
}

func TestSignedDotNumber(t *testing.T) {
	l := New("+.5")
	tok := mustToken(t, l)
	expectToken(t, tok, TokenFloat, "+.5")

	l = New("-.5")
	tok = mustToken(t, l)
	expectToken(t, tok, TokenFloat, "-.5")
}

// =========================================================================
// Multiple tokens on same line
// =========================================================================

func TestMultipleTokensNoSpaces(t *testing.T) {
	tokens, err := AllTokens("(+1 2)")
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 4 {
		t.Fatalf("expected 4 tokens, got %d", len(tokens))
	}
}

func TestConsecutiveStrings(t *testing.T) {
	tokens, err := AllTokens(`"a""b""c"`)
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 3 {
		t.Fatalf("expected 3 tokens, got %d", len(tokens))
	}
}

// =========================================================================
// Large input
// =========================================================================

func TestLargeInput(t *testing.T) {
	size := 5000
	if testing.Short() {
		size = 100
	}
	input := strings.Repeat("(", size) + strings.Repeat(")", size)
	tokens, err := AllTokens(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != size*2 {
		t.Fatalf("expected %d tokens, got %d", size*2, len(tokens))
	}
}

func TestLongString(t *testing.T) {
	size := 10000
	if testing.Short() {
		size = 500
	}
	long := strings.Repeat("a", size)
	input := `"` + long + `"`
	l := New(input)
	tok := mustToken(t, l)
	if tok.Type != TokenString || len(tok.Value) != size {
		t.Fatalf("expected string of length %d, got length %d", size, len(tok.Value))
	}
}

func TestLongSymbol(t *testing.T) {
	sym := strings.Repeat("x", 1000)
	l := New(sym)
	tok := mustToken(t, l)
	expectToken(t, tok, TokenSymbol, sym)
}
