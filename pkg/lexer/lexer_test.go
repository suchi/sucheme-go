package lexer

import (
	"testing"
)

func mustToken(t *testing.T, l *Lexer) Token {
	t.Helper()
	tok, err := l.NextToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return tok
}

func expectToken(t *testing.T, tok Token, typ TokenType, val string) {
	t.Helper()
	if tok.Type != typ {
		t.Fatalf("expected type %s, got %s (value=%q)", typ, tok.Type, tok.Value)
	}
	if tok.Value != val {
		t.Fatalf("expected value %q, got %q", val, tok.Value)
	}
}

func TestIntegers(t *testing.T) {
	cases := []struct {
		input string
		value string
	}{
		{"42", "42"},
		{"0", "0"},
		{"-1", "-1"},
		{"+5", "+5"},
		{"12345678", "12345678"},
	}
	for _, c := range cases {
		l := New(c.input)
		tok := mustToken(t, l)
		expectToken(t, tok, TokenInteger, c.value)
	}
}

func TestFloats(t *testing.T) {
	cases := []struct {
		input string
		value string
	}{
		{"3.14", "3.14"},
		{".5", ".5"},
		{"1.", "1."},
		{"1e10", "1e10"},
		{"1.5e-3", "1.5e-3"},
		{"-2.5", "-2.5"},
		{"+.5", "+.5"},
		{"+inf.0", "+inf.0"},
		{"-inf.0", "-inf.0"},
		{"+nan.0", "+nan.0"},
	}
	for _, c := range cases {
		l := New(c.input)
		tok := mustToken(t, l)
		expectToken(t, tok, TokenFloat, c.value)
	}
}

func TestStrings(t *testing.T) {
	cases := []struct {
		input string
		value string
	}{
		{`"hello"`, "hello"},
		{`"hello world"`, "hello world"},
		{`""`, ""},
		{`"a\nb"`, "a\nb"},
		{`"a\tb"`, "a\tb"},
		{`"a\\b"`, "a\\b"},
		{`"a\"b"`, "a\"b"},
		{`"a\rb"`, "a\rb"},
		{`"\a"`, "\a"},
		{`"\b"`, "\b"},
		{`"\0"`, "\x00"},
		{`"\x41;"`, "A"},
	}
	for _, c := range cases {
		l := New(c.input)
		tok := mustToken(t, l)
		expectToken(t, tok, TokenString, c.value)
	}
}

func TestStringLineContinuation(t *testing.T) {
	input := "\"hello\\\n   world\""
	l := New(input)
	tok := mustToken(t, l)
	expectToken(t, tok, TokenString, "helloworld")
}

func TestUnterminatedString(t *testing.T) {
	l := New(`"hello`)
	_, err := l.NextToken()
	if err == nil {
		t.Fatal("expected error for unterminated string")
	}
}

func TestBooleans(t *testing.T) {
	cases := []struct {
		input string
		value string
	}{
		{"#t", "#t"},
		{"#f", "#f"},
		{"#true", "#t"},
		{"#false", "#f"},
	}
	for _, c := range cases {
		l := New(c.input)
		tok := mustToken(t, l)
		expectToken(t, tok, TokenBoolean, c.value)
	}
}

func TestCharacters(t *testing.T) {
	cases := []struct {
		input string
		value string
	}{
		{`#\a`, "a"},
		{`#\A`, "A"},
		{`#\space`, " "},
		{`#\newline`, "\n"},
		{`#\tab`, "\t"},
		{`#\return`, "\r"},
		{`#\null`, "\x00"},
		{`#\alarm`, "\a"},
		{`#\backspace`, "\b"},
		{`#\escape`, "\x1b"},
		{`#\delete`, "\x7f"},
		{`#\x41`, "A"},
	}
	for _, c := range cases {
		l := New(c.input)
		tok := mustToken(t, l)
		expectToken(t, tok, TokenCharacter, c.value)
	}
}

func TestSymbols(t *testing.T) {
	cases := []string{
		"foo", "bar-baz", "list->string", "string?", "set!", "+", "-",
		"<=", ">=", "...", "call/cc",
	}
	for _, c := range cases {
		l := New(c)
		tok := mustToken(t, l)
		expectToken(t, tok, TokenSymbol, c)
	}
}

func TestParensAndBrackets(t *testing.T) {
	l := New("( ) [ ]")
	expectToken(t, mustToken(t, l), TokenLeftParen, "(")
	expectToken(t, mustToken(t, l), TokenRightParen, ")")
	expectToken(t, mustToken(t, l), TokenLeftBracket, "[")
	expectToken(t, mustToken(t, l), TokenRightBracket, "]")
}

func TestQuoteTokens(t *testing.T) {
	l := New("' ` , ,@")
	expectToken(t, mustToken(t, l), TokenQuote, "'")
	expectToken(t, mustToken(t, l), TokenQuasiquote, "`")
	expectToken(t, mustToken(t, l), TokenUnquote, ",")
	expectToken(t, mustToken(t, l), TokenUnquoteSplicing, ",@")
}

func TestDot(t *testing.T) {
	l := New("(a . b)")
	expectToken(t, mustToken(t, l), TokenLeftParen, "(")
	expectToken(t, mustToken(t, l), TokenSymbol, "a")
	expectToken(t, mustToken(t, l), TokenDot, ".")
	expectToken(t, mustToken(t, l), TokenSymbol, "b")
	expectToken(t, mustToken(t, l), TokenRightParen, ")")
}

func TestHashLeftParen(t *testing.T) {
	l := New("#(1 2 3)")
	expectToken(t, mustToken(t, l), TokenHashLeftParen, "#(")
	expectToken(t, mustToken(t, l), TokenInteger, "1")
	expectToken(t, mustToken(t, l), TokenInteger, "2")
	expectToken(t, mustToken(t, l), TokenInteger, "3")
	expectToken(t, mustToken(t, l), TokenRightParen, ")")
}

func TestHashU8LeftParen(t *testing.T) {
	l := New("#u8(1 2 3)")
	expectToken(t, mustToken(t, l), TokenHashU8LeftParen, "#u8(")
	expectToken(t, mustToken(t, l), TokenInteger, "1")
	expectToken(t, mustToken(t, l), TokenInteger, "2")
	expectToken(t, mustToken(t, l), TokenInteger, "3")
	expectToken(t, mustToken(t, l), TokenRightParen, ")")
}

func TestRadixNumbers(t *testing.T) {
	cases := []struct {
		input string
		value string
	}{
		{"#b101", "#b101"},
		{"#o17", "#o17"},
		{"#d42", "#d42"},
		{"#xff", "#xff"},
		{"#xFF", "#xFF"},
		{"#b-101", "#b-101"},
		{"#x+ff", "#x+ff"},
	}
	for _, c := range cases {
		l := New(c.input)
		tok := mustToken(t, l)
		expectToken(t, tok, TokenInteger, c.value)
	}
}

func TestLineComment(t *testing.T) {
	l := New("; this is a comment\n42")
	tok := mustToken(t, l)
	expectToken(t, tok, TokenInteger, "42")
}

func TestBlockComment(t *testing.T) {
	l := New("#| this is a block comment |# 42")
	tok := mustToken(t, l)
	expectToken(t, tok, TokenInteger, "42")
}

func TestNestedBlockComment(t *testing.T) {
	l := New("#| outer #| inner |# still comment |# 42")
	tok := mustToken(t, l)
	expectToken(t, tok, TokenInteger, "42")
}

func TestDatumCommentToken(t *testing.T) {
	l := New("#; foo 42")
	tok := mustToken(t, l)
	// The lexer returns #; as a symbol token for the parser to handle
	expectToken(t, tok, TokenSymbol, "#;")
	tok = mustToken(t, l)
	expectToken(t, tok, TokenSymbol, "foo")
	tok = mustToken(t, l)
	expectToken(t, tok, TokenInteger, "42")
}

func TestLineColumnTracking(t *testing.T) {
	l := New("foo\nbar")
	tok := mustToken(t, l)
	if tok.Line != 1 || tok.Column != 1 {
		t.Fatalf("expected 1:1, got %d:%d", tok.Line, tok.Column)
	}
	tok = mustToken(t, l)
	if tok.Line != 2 || tok.Column != 1 {
		t.Fatalf("expected 2:1, got %d:%d", tok.Line, tok.Column)
	}
}

func TestEof(t *testing.T) {
	l := New("")
	tok := mustToken(t, l)
	if tok.Type != TokenEof {
		t.Fatalf("expected Eof, got %s", tok.Type)
	}
	// Calling again should still return Eof
	tok = mustToken(t, l)
	if tok.Type != TokenEof {
		t.Fatalf("expected Eof on second call, got %s", tok.Type)
	}
}

func TestPeekToken(t *testing.T) {
	l := New("42 foo")
	tok, err := l.PeekToken()
	if err != nil {
		t.Fatal(err)
	}
	expectToken(t, tok, TokenInteger, "42")
	// Peek again should return same token
	tok, err = l.PeekToken()
	if err != nil {
		t.Fatal(err)
	}
	expectToken(t, tok, TokenInteger, "42")
	// NextToken should consume it
	tok = mustToken(t, l)
	expectToken(t, tok, TokenInteger, "42")
	tok = mustToken(t, l)
	expectToken(t, tok, TokenSymbol, "foo")
}

func TestAllTokens(t *testing.T) {
	tokens, err := AllTokens("(+ 1 2)")
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 5 {
		t.Fatalf("expected 5 tokens, got %d", len(tokens))
	}
	expectToken(t, tokens[0], TokenLeftParen, "(")
	expectToken(t, tokens[1], TokenSymbol, "+")
	expectToken(t, tokens[2], TokenInteger, "1")
	expectToken(t, tokens[3], TokenInteger, "2")
	expectToken(t, tokens[4], TokenRightParen, ")")
}

func TestWhitespace(t *testing.T) {
	l := New("  \t\n  42")
	tok := mustToken(t, l)
	expectToken(t, tok, TokenInteger, "42")
}

func TestComplexExpression(t *testing.T) {
	tokens, err := AllTokens("(define (fact n) (if (<= n 1) 1 (* n (fact (- n 1)))))")
	if err != nil {
		t.Fatal(err)
	}
	// Just verify it tokenizes without errors and has the right count
	if len(tokens) != 28 {
		t.Fatalf("expected 28 tokens, got %d", len(tokens))
	}
}

func TestTokenString(t *testing.T) {
	tok := Token{Type: TokenInteger, Value: "42", Line: 1, Column: 1}
	s := tok.String()
	if s == "" {
		t.Fatal("token String() should not be empty")
	}
}

func TestUnexpectedCharacter(t *testing.T) {
	l := New("\x01")
	_, err := l.NextToken()
	if err == nil {
		t.Fatal("expected error for unexpected character")
	}
}

func TestUnexpectedEndAfterHash(t *testing.T) {
	l := New("#")
	_, err := l.NextToken()
	if err == nil {
		t.Fatal("expected error for # at end of input")
	}
}

func TestInvalidBooleanLiteral(t *testing.T) {
	l := New("#trux")
	_, err := l.NextToken()
	if err == nil {
		t.Fatal("expected error for invalid boolean")
	}
}

func TestInvalidFalseLiteral(t *testing.T) {
	l := New("#falsx")
	_, err := l.NextToken()
	if err == nil {
		t.Fatal("expected error for invalid boolean")
	}
}

func TestRadixNoDigits(t *testing.T) {
	l := New("#b ")
	_, err := l.NextToken()
	if err == nil {
		t.Fatal("expected error for radix with no digits")
	}
}

func TestUnknownCharacterName(t *testing.T) {
	l := New(`#\foobar`)
	_, err := l.NextToken()
	if err == nil {
		t.Fatal("expected error for unknown character name")
	}
}

func TestCharacterEndOfInput(t *testing.T) {
	l := New(`#\`)
	_, err := l.NextToken()
	if err == nil {
		t.Fatal("expected error for character at end of input")
	}
}

func TestBytevectorMissingEight(t *testing.T) {
	l := New("#u(")
	_, err := l.NextToken()
	if err == nil {
		t.Fatal("expected error for #u without 8")
	}
}

func TestBytevectorMissingParen(t *testing.T) {
	l := New("#u8x")
	_, err := l.NextToken()
	if err == nil {
		t.Fatal("expected error for #u8 without (")
	}
}

func TestDoubleDotError(t *testing.T) {
	l := New("..")
	_, err := l.NextToken()
	if err == nil {
		t.Fatal("expected error for ..")
	}
}

func TestEllipsis(t *testing.T) {
	l := New("...")
	tok := mustToken(t, l)
	expectToken(t, tok, TokenSymbol, "...")
}

func TestUnknownEscapeInString(t *testing.T) {
	l := New(`"\q"`)
	_, err := l.NextToken()
	if err == nil {
		t.Fatal("expected error for unknown escape")
	}
}

func TestStringEscapeEndOfInput(t *testing.T) {
	l := New(`"\`)
	_, err := l.NextToken()
	if err == nil {
		t.Fatal("expected error for escape at end of input")
	}
}

func TestUnexpectedHashChar(t *testing.T) {
	l := New("#z")
	_, err := l.NextToken()
	if err == nil {
		t.Fatal("expected error for #z")
	}
}
