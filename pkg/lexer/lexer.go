// Package lexer implements a tokenizer for R7RS Scheme.
package lexer

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// TokenType represents the type of a lexer token.
type TokenType int

const (
	TokenLeftParen       TokenType = iota // (
	TokenRightParen                       // )
	TokenLeftBracket                      // [
	TokenRightBracket                     // ]
	TokenQuote                            // '
	TokenQuasiquote                       // `
	TokenUnquote                          // ,
	TokenUnquoteSplicing                  // ,@
	TokenDot                              // .
	TokenHashLeftParen                    // #(
	TokenHashU8LeftParen                  // #u8(
	TokenInteger                          // integer literal
	TokenFloat                            // float literal
	TokenString                           // string literal
	TokenSymbol                           // symbol / identifier
	TokenBoolean                          // #t, #f, #true, #false
	TokenCharacter                        // #\x
	TokenEof                              // end of input
)

// String returns the name of the token type.
func (t TokenType) String() string {
	switch t {
	case TokenLeftParen:
		return "LeftParen"
	case TokenRightParen:
		return "RightParen"
	case TokenLeftBracket:
		return "LeftBracket"
	case TokenRightBracket:
		return "RightBracket"
	case TokenQuote:
		return "Quote"
	case TokenQuasiquote:
		return "Quasiquote"
	case TokenUnquote:
		return "Unquote"
	case TokenUnquoteSplicing:
		return "UnquoteSplicing"
	case TokenDot:
		return "Dot"
	case TokenHashLeftParen:
		return "HashLeftParen"
	case TokenHashU8LeftParen:
		return "HashU8LeftParen"
	case TokenInteger:
		return "Integer"
	case TokenFloat:
		return "Float"
	case TokenString:
		return "String"
	case TokenSymbol:
		return "Symbol"
	case TokenBoolean:
		return "Boolean"
	case TokenCharacter:
		return "Character"
	case TokenEof:
		return "Eof"
	default:
		return "Unknown"
	}
}

// Token represents a single token from the lexer.
type Token struct {
	Type   TokenType
	Value  string
	Line   int
	Column int
}

// String returns a debug representation of the token.
func (tok Token) String() string {
	return fmt.Sprintf("Token(%s, %q, %d:%d)", tok.Type, tok.Value, tok.Line, tok.Column)
}

// Lexer tokenizes Scheme source code.
type Lexer struct {
	input  string
	pos    int
	line   int
	col    int
	peeked *Token
}

// New creates a new Lexer for the given input.
func New(input string) *Lexer {
	return &Lexer{
		input: input,
		pos:   0,
		line:  1,
		col:   1,
	}
}

// NextToken returns the next token from the input.
func (l *Lexer) NextToken() (Token, error) {
	if l.peeked != nil {
		tok := *l.peeked
		l.peeked = nil
		return tok, nil
	}
	return l.readToken()
}

// PeekToken returns the next token without consuming it.
func (l *Lexer) PeekToken() (Token, error) {
	if l.peeked != nil {
		return *l.peeked, nil
	}
	tok, err := l.readToken()
	if err != nil {
		return tok, err
	}
	l.peeked = &tok
	return tok, nil
}

func (l *Lexer) readToken() (Token, error) {
	l.skipWhitespaceAndComments()

	if l.pos >= len(l.input) {
		return Token{Type: TokenEof, Line: l.line, Column: l.col}, nil
	}

	line, col := l.line, l.col
	ch := l.peekChar()

	switch {
	case ch == '(':
		l.advance()
		return Token{Type: TokenLeftParen, Value: "(", Line: line, Column: col}, nil
	case ch == ')':
		l.advance()
		return Token{Type: TokenRightParen, Value: ")", Line: line, Column: col}, nil
	case ch == '[':
		l.advance()
		return Token{Type: TokenLeftBracket, Value: "[", Line: line, Column: col}, nil
	case ch == ']':
		l.advance()
		return Token{Type: TokenRightBracket, Value: "]", Line: line, Column: col}, nil
	case ch == '\'':
		l.advance()
		return Token{Type: TokenQuote, Value: "'", Line: line, Column: col}, nil
	case ch == '`':
		l.advance()
		return Token{Type: TokenQuasiquote, Value: "`", Line: line, Column: col}, nil
	case ch == ',':
		l.advance()
		if l.pos < len(l.input) && l.peekChar() == '@' {
			l.advance()
			return Token{Type: TokenUnquoteSplicing, Value: ",@", Line: line, Column: col}, nil
		}
		return Token{Type: TokenUnquote, Value: ",", Line: line, Column: col}, nil
	case ch == '#':
		return l.readHash(line, col)
	case ch == '"':
		return l.readString(line, col)
	case isDigit(ch):
		return l.readNumber(line, col)
	case ch == '+' || ch == '-':
		return l.readSignedOrSymbol(line, col)
	case ch == '.':
		return l.readDotOrNumber(line, col)
	case isSymbolInitial(ch):
		return l.readSymbol(line, col)
	default:
		l.advance()
		return Token{}, fmt.Errorf("%d:%d: unexpected character: %c", line, col, ch)
	}
}

// ---------------------------------------------------------------------------
// Character-level helpers
// ---------------------------------------------------------------------------

func (l *Lexer) peekChar() rune {
	r, _ := utf8.DecodeRuneInString(l.input[l.pos:])
	return r
}

func (l *Lexer) advance() rune {
	r, size := utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += size
	if r == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return r
}

func (l *Lexer) peekCharAt(offset int) (rune, bool) {
	p := l.pos + offset
	if p >= len(l.input) {
		return 0, false
	}
	r, _ := utf8.DecodeRuneInString(l.input[p:])
	return r, true
}

// ---------------------------------------------------------------------------
// Whitespace and comments
// ---------------------------------------------------------------------------

func (l *Lexer) skipWhitespaceAndComments() {
	for l.pos < len(l.input) {
		ch := l.peekChar()
		if unicode.IsSpace(ch) {
			l.advance()
			continue
		}
		if ch == ';' {
			l.skipLineComment()
			continue
		}
		if ch == '#' && l.pos+1 < len(l.input) {
			next, _ := l.peekCharAt(1)
			if next == '|' {
				l.skipBlockComment()
				continue
			}
			if next == ';' {
				// datum comment - handled by the parser, not here
				// We just return so the parser sees #;
				break
			}
		}
		break
	}
}

func (l *Lexer) skipLineComment() {
	for l.pos < len(l.input) {
		ch := l.advance()
		if ch == '\n' {
			break
		}
	}
}

func (l *Lexer) skipBlockComment() {
	// skip #|
	l.advance()
	l.advance()
	depth := 1
	for l.pos < len(l.input) && depth > 0 {
		ch := l.advance()
		if ch == '#' && l.pos < len(l.input) && l.peekChar() == '|' {
			l.advance()
			depth++
		} else if ch == '|' && l.pos < len(l.input) && l.peekChar() == '#' {
			l.advance()
			depth--
		}
	}
}

// ---------------------------------------------------------------------------
// Hash-prefixed tokens
// ---------------------------------------------------------------------------

func (l *Lexer) readHash(line, col int) (Token, error) {
	l.advance() // skip #

	if l.pos >= len(l.input) {
		return Token{}, fmt.Errorf("%d:%d: unexpected end of input after #", line, col)
	}

	ch := l.peekChar()

	switch {
	case ch == 't':
		return l.readBooleanTrue(line, col)
	case ch == 'f':
		return l.readBooleanFalse(line, col)
	case ch == '\\':
		return l.readCharacter(line, col)
	case ch == '(':
		l.advance()
		return Token{Type: TokenHashLeftParen, Value: "#(", Line: line, Column: col}, nil
	case ch == 'u' || ch == 'U':
		return l.readBytevectorPrefix(line, col)
	case ch == 'b' || ch == 'B':
		return l.readRadixNumber(2, line, col)
	case ch == 'o' || ch == 'O':
		return l.readRadixNumber(8, line, col)
	case ch == 'd' || ch == 'D':
		return l.readRadixNumber(10, line, col)
	case ch == 'x' || ch == 'X':
		return l.readRadixNumber(16, line, col)
	case ch == ';':
		// datum comment token - return it so parser can handle
		l.advance()
		return Token{Type: TokenSymbol, Value: "#;", Line: line, Column: col}, nil
	default:
		return Token{}, fmt.Errorf("%d:%d: unexpected character after #: %c", line, col, ch)
	}
}

func (l *Lexer) readBooleanTrue(line, col int) (Token, error) {
	l.advance() // skip 't'
	// Check for #true
	if l.pos < len(l.input) && l.peekChar() == 'r' {
		rest := "rue"
		for _, expected := range rest {
			if l.pos >= len(l.input) || l.peekChar() != expected {
				return Token{}, fmt.Errorf("%d:%d: invalid boolean literal", line, col)
			}
			l.advance()
		}
		if l.pos < len(l.input) && isSymbolSubsequent(l.peekChar()) {
			return Token{}, fmt.Errorf("%d:%d: invalid boolean literal", line, col)
		}
	}
	return Token{Type: TokenBoolean, Value: "#t", Line: line, Column: col}, nil
}

func (l *Lexer) readBooleanFalse(line, col int) (Token, error) {
	l.advance() // skip 'f'
	// Check for #false
	if l.pos < len(l.input) && l.peekChar() == 'a' {
		rest := "alse"
		for _, expected := range rest {
			if l.pos >= len(l.input) || l.peekChar() != expected {
				return Token{}, fmt.Errorf("%d:%d: invalid boolean literal", line, col)
			}
			l.advance()
		}
		if l.pos < len(l.input) && isSymbolSubsequent(l.peekChar()) {
			return Token{}, fmt.Errorf("%d:%d: invalid boolean literal", line, col)
		}
	}
	return Token{Type: TokenBoolean, Value: "#f", Line: line, Column: col}, nil
}

func (l *Lexer) readCharacter(line, col int) (Token, error) {
	l.advance() // skip backslash

	if l.pos >= len(l.input) {
		return Token{}, fmt.Errorf("%d:%d: unexpected end of input in character literal", line, col)
	}

	ch := l.advance()

	// Check for named characters
	if isLetter(ch) && l.pos < len(l.input) && isLetter(l.peekChar()) {
		var buf strings.Builder
		buf.WriteRune(ch)
		for l.pos < len(l.input) && isLetter(l.peekChar()) {
			buf.WriteRune(l.advance())
		}
		name := buf.String()
		switch strings.ToLower(name) {
		case "space":
			return Token{Type: TokenCharacter, Value: " ", Line: line, Column: col}, nil
		case "newline":
			return Token{Type: TokenCharacter, Value: "\n", Line: line, Column: col}, nil
		case "tab":
			return Token{Type: TokenCharacter, Value: "\t", Line: line, Column: col}, nil
		case "return":
			return Token{Type: TokenCharacter, Value: "\r", Line: line, Column: col}, nil
		case "null":
			return Token{Type: TokenCharacter, Value: "\x00", Line: line, Column: col}, nil
		case "alarm":
			return Token{Type: TokenCharacter, Value: "\a", Line: line, Column: col}, nil
		case "backspace":
			return Token{Type: TokenCharacter, Value: "\b", Line: line, Column: col}, nil
		case "escape":
			return Token{Type: TokenCharacter, Value: "\x1b", Line: line, Column: col}, nil
		case "delete":
			return Token{Type: TokenCharacter, Value: "\x7f", Line: line, Column: col}, nil
		default:
			return Token{}, fmt.Errorf("%d:%d: unknown character name: %s", line, col, name)
		}
	}

	// Check for hex scalar value: #\xNN
	if ch == 'x' && l.pos < len(l.input) && isHexDigit(l.peekChar()) {
		var buf strings.Builder
		for l.pos < len(l.input) && isHexDigit(l.peekChar()) {
			buf.WriteRune(l.advance())
		}
		var code rune
		fmt.Sscanf(buf.String(), "%x", &code)
		return Token{Type: TokenCharacter, Value: string(code), Line: line, Column: col}, nil
	}

	return Token{Type: TokenCharacter, Value: string(ch), Line: line, Column: col}, nil
}

func (l *Lexer) readBytevectorPrefix(line, col int) (Token, error) {
	l.advance() // skip 'u'
	if l.pos >= len(l.input) || l.peekChar() != '8' {
		return Token{}, fmt.Errorf("%d:%d: expected '8' after #u", line, col)
	}
	l.advance() // skip '8'
	if l.pos >= len(l.input) || l.peekChar() != '(' {
		return Token{}, fmt.Errorf("%d:%d: expected '(' after #u8", line, col)
	}
	l.advance() // skip '('
	return Token{Type: TokenHashU8LeftParen, Value: "#u8(", Line: line, Column: col}, nil
}

func (l *Lexer) readRadixNumber(radix int, line, col int) (Token, error) {
	l.advance() // skip radix prefix letter

	var buf strings.Builder
	switch radix {
	case 2:
		buf.WriteString("#b")
	case 8:
		buf.WriteString("#o")
	case 10:
		buf.WriteString("#d")
	case 16:
		buf.WriteString("#x")
	}

	// Optional sign
	if l.pos < len(l.input) && (l.peekChar() == '+' || l.peekChar() == '-') {
		buf.WriteRune(l.advance())
	}

	if l.pos >= len(l.input) || !isRadixDigit(l.peekChar(), radix) {
		return Token{}, fmt.Errorf("%d:%d: expected digit after radix prefix", line, col)
	}

	for l.pos < len(l.input) && isRadixDigit(l.peekChar(), radix) {
		buf.WriteRune(l.advance())
	}

	return Token{Type: TokenInteger, Value: buf.String(), Line: line, Column: col}, nil
}

// ---------------------------------------------------------------------------
// Strings
// ---------------------------------------------------------------------------

func (l *Lexer) readString(line, col int) (Token, error) {
	l.advance() // skip opening "

	var buf strings.Builder
	for l.pos < len(l.input) {
		ch := l.advance()
		if ch == '"' {
			return Token{Type: TokenString, Value: buf.String(), Line: line, Column: col}, nil
		}
		if ch == '\\' {
			if l.pos >= len(l.input) {
				return Token{}, fmt.Errorf("%d:%d: unexpected end of input in string", line, col)
			}
			esc := l.advance()
			switch esc {
			case 'n':
				buf.WriteByte('\n')
			case 't':
				buf.WriteByte('\t')
			case 'r':
				buf.WriteByte('\r')
			case '"':
				buf.WriteByte('"')
			case '\\':
				buf.WriteByte('\\')
			case 'a':
				buf.WriteByte('\a')
			case 'b':
				buf.WriteByte('\b')
			case '0':
				buf.WriteByte(0)
			case 'x':
				// Hex escape: \xNN;
				var hexBuf strings.Builder
				for l.pos < len(l.input) && l.peekChar() != ';' {
					hexBuf.WriteRune(l.advance())
				}
				if l.pos >= len(l.input) {
					return Token{}, fmt.Errorf("%d:%d: expected ';' after hex escape", line, col)
				}
				l.advance() // skip ;
				var code rune
				fmt.Sscanf(hexBuf.String(), "%x", &code)
				buf.WriteRune(code)
			case '\n':
				// Line continuation: skip whitespace after newline
				for l.pos < len(l.input) && (l.peekChar() == ' ' || l.peekChar() == '\t') {
					l.advance()
				}
			default:
				return Token{}, fmt.Errorf("%d:%d: unknown escape sequence: \\%c", line, col, esc)
			}
		} else {
			buf.WriteRune(ch)
		}
	}
	return Token{}, fmt.Errorf("%d:%d: unterminated string", line, col)
}

// ---------------------------------------------------------------------------
// Numbers
// ---------------------------------------------------------------------------

func (l *Lexer) readNumber(line, col int) (Token, error) {
	var buf strings.Builder
	isFloat := false

	for l.pos < len(l.input) && isDigit(l.peekChar()) {
		buf.WriteRune(l.advance())
	}

	// Check for decimal point
	if l.pos < len(l.input) && l.peekChar() == '.' {
		// Peek ahead: if next after '.' is a digit, it's a float
		if next, ok := l.peekCharAt(l.pos - l.pos + 1); ok && isDigit(next) {
			// Actually look at what's after the dot in the input
		}
		isFloat = true
		buf.WriteRune(l.advance()) // the dot
		for l.pos < len(l.input) && isDigit(l.peekChar()) {
			buf.WriteRune(l.advance())
		}
	}

	// Check for exponent
	if l.pos < len(l.input) && (l.peekChar() == 'e' || l.peekChar() == 'E') {
		isFloat = true
		buf.WriteRune(l.advance())
		if l.pos < len(l.input) && (l.peekChar() == '+' || l.peekChar() == '-') {
			buf.WriteRune(l.advance())
		}
		for l.pos < len(l.input) && isDigit(l.peekChar()) {
			buf.WriteRune(l.advance())
		}
	}

	// Check for rational (/) - treat as symbol if present
	if l.pos < len(l.input) && l.peekChar() == '/' {
		buf.WriteRune(l.advance())
		for l.pos < len(l.input) && isDigit(l.peekChar()) {
			buf.WriteRune(l.advance())
		}
		// Treat rationals as symbols for now
		return Token{Type: TokenSymbol, Value: buf.String(), Line: line, Column: col}, nil
	}

	if isFloat {
		return Token{Type: TokenFloat, Value: buf.String(), Line: line, Column: col}, nil
	}
	return Token{Type: TokenInteger, Value: buf.String(), Line: line, Column: col}, nil
}

func (l *Lexer) readSignedOrSymbol(line, col int) (Token, error) {
	sign := l.advance() // + or -
	if l.pos < len(l.input) && isDigit(l.peekChar()) {
		// Signed number
		var buf strings.Builder
		buf.WriteRune(sign)
		for l.pos < len(l.input) && isDigit(l.peekChar()) {
			buf.WriteRune(l.advance())
		}
		isFloat := false
		if l.pos < len(l.input) && l.peekChar() == '.' {
			isFloat = true
			buf.WriteRune(l.advance())
			for l.pos < len(l.input) && isDigit(l.peekChar()) {
				buf.WriteRune(l.advance())
			}
		}
		if l.pos < len(l.input) && (l.peekChar() == 'e' || l.peekChar() == 'E') {
			isFloat = true
			buf.WriteRune(l.advance())
			if l.pos < len(l.input) && (l.peekChar() == '+' || l.peekChar() == '-') {
				buf.WriteRune(l.advance())
			}
			for l.pos < len(l.input) && isDigit(l.peekChar()) {
				buf.WriteRune(l.advance())
			}
		}
		if isFloat {
			return Token{Type: TokenFloat, Value: buf.String(), Line: line, Column: col}, nil
		}
		return Token{Type: TokenInteger, Value: buf.String(), Line: line, Column: col}, nil
	}

	if l.pos < len(l.input) && l.peekChar() == '.' {
		// Could be +.5 style float
		if next, ok := l.peekCharAt(1); ok && isDigit(next) {
			var buf strings.Builder
			buf.WriteRune(sign)
			buf.WriteRune(l.advance()) // dot
			for l.pos < len(l.input) && isDigit(l.peekChar()) {
				buf.WriteRune(l.advance())
			}
			if l.pos < len(l.input) && (l.peekChar() == 'e' || l.peekChar() == 'E') {
				buf.WriteRune(l.advance())
				if l.pos < len(l.input) && (l.peekChar() == '+' || l.peekChar() == '-') {
					buf.WriteRune(l.advance())
				}
				for l.pos < len(l.input) && isDigit(l.peekChar()) {
					buf.WriteRune(l.advance())
				}
			}
			return Token{Type: TokenFloat, Value: buf.String(), Line: line, Column: col}, nil
		}
	}

	// Check for special number identifiers: +inf.0, -inf.0, +nan.0
	if l.pos+4 <= len(l.input) {
		rest := l.input[l.pos:]
		if strings.HasPrefix(rest, "inf.0") {
			for i := 0; i < 5; i++ {
				l.advance()
			}
			return Token{Type: TokenFloat, Value: string(sign) + "inf.0", Line: line, Column: col}, nil
		}
		if sign == '+' && strings.HasPrefix(rest, "nan.0") {
			for i := 0; i < 5; i++ {
				l.advance()
			}
			return Token{Type: TokenFloat, Value: "+nan.0", Line: line, Column: col}, nil
		}
	}

	// It's a symbol starting with + or -
	var buf strings.Builder
	buf.WriteRune(sign)
	for l.pos < len(l.input) && isSymbolSubsequent(l.peekChar()) {
		buf.WriteRune(l.advance())
	}
	return Token{Type: TokenSymbol, Value: buf.String(), Line: line, Column: col}, nil
}

func (l *Lexer) readDotOrNumber(line, col int) (Token, error) {
	l.advance() // skip .

	// Check for number starting with .
	if l.pos < len(l.input) && isDigit(l.peekChar()) {
		var buf strings.Builder
		buf.WriteByte('.')
		for l.pos < len(l.input) && isDigit(l.peekChar()) {
			buf.WriteRune(l.advance())
		}
		if l.pos < len(l.input) && (l.peekChar() == 'e' || l.peekChar() == 'E') {
			buf.WriteRune(l.advance())
			if l.pos < len(l.input) && (l.peekChar() == '+' || l.peekChar() == '-') {
				buf.WriteRune(l.advance())
			}
			for l.pos < len(l.input) && isDigit(l.peekChar()) {
				buf.WriteRune(l.advance())
			}
		}
		return Token{Type: TokenFloat, Value: buf.String(), Line: line, Column: col}, nil
	}

	// Check for ...
	if l.pos < len(l.input) && l.peekChar() == '.' {
		l.advance()
		if l.pos < len(l.input) && l.peekChar() == '.' {
			l.advance()
			return Token{Type: TokenSymbol, Value: "...", Line: line, Column: col}, nil
		}
		return Token{}, fmt.Errorf("%d:%d: unexpected '..'", line, col)
	}

	return Token{Type: TokenDot, Value: ".", Line: line, Column: col}, nil
}

// ---------------------------------------------------------------------------
// Symbols
// ---------------------------------------------------------------------------

func (l *Lexer) readSymbol(line, col int) (Token, error) {
	var buf strings.Builder
	buf.WriteRune(l.advance())
	for l.pos < len(l.input) && isSymbolSubsequent(l.peekChar()) {
		buf.WriteRune(l.advance())
	}
	return Token{Type: TokenSymbol, Value: buf.String(), Line: line, Column: col}, nil
}

// ---------------------------------------------------------------------------
// Character classification
// ---------------------------------------------------------------------------

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isLetter(r rune) bool {
	return unicode.IsLetter(r)
}

func isHexDigit(r rune) bool {
	return isDigit(r) || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}

func isRadixDigit(r rune, radix int) bool {
	switch radix {
	case 2:
		return r == '0' || r == '1'
	case 8:
		return r >= '0' && r <= '7'
	case 10:
		return isDigit(r)
	case 16:
		return isHexDigit(r)
	default:
		return false
	}
}

func isSymbolInitial(r rune) bool {
	if unicode.IsLetter(r) {
		return true
	}
	return strings.ContainsRune("!$%&*/:<=>?^_~", r)
}

func isSymbolSubsequent(r rune) bool {
	if isSymbolInitial(r) || isDigit(r) {
		return true
	}
	return strings.ContainsRune("+-.@", r)
}

func isDelimiter(r rune) bool {
	return unicode.IsSpace(r) || strings.ContainsRune("()[]\"';", r)
}

// AllTokens tokenizes the entire input and returns all tokens (excluding Eof).
func AllTokens(input string) ([]Token, error) {
	l := New(input)
	var tokens []Token
	for {
		tok, err := l.NextToken()
		if err != nil {
			return nil, err
		}
		if tok.Type == TokenEof {
			break
		}
		tokens = append(tokens, tok)
	}
	return tokens, nil
}

// Ensure isDelimiter is used (for documentation / future use).
var _ = isDelimiter
