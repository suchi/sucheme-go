// Package parser converts a token stream into a Scheme value AST.
package parser

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/suchi/sucheme-go/pkg/lexer"
	"github.com/suchi/sucheme-go/pkg/types"
)

// Parser converts tokens into Scheme values.
type Parser struct {
	lexer *lexer.Lexer
}

// New creates a new Parser for the given input string.
func New(input string) *Parser {
	return &Parser{lexer: lexer.New(input)}
}

// NewFromLexer creates a parser from an existing lexer.
func NewFromLexer(l *lexer.Lexer) *Parser {
	return &Parser{lexer: l}
}

// Parse reads and returns the next expression from the input.
// Returns nil, nil at end of input.
func (p *Parser) Parse() (*types.Value, error) {
	tok, err := p.lexer.PeekToken()
	if err != nil {
		return nil, err
	}
	if tok.Type == lexer.TokenEof {
		return nil, nil
	}
	return p.parseExpr()
}

// ParseAll reads all expressions from the input.
func (p *Parser) ParseAll() ([]*types.Value, error) {
	var results []*types.Value
	for {
		tok, err := p.lexer.PeekToken()
		if err != nil {
			return nil, err
		}
		if tok.Type == lexer.TokenEof {
			break
		}
		val, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		results = append(results, val)
	}
	return results, nil
}

// ParseString is a convenience function to parse a single expression from a string.
func ParseString(input string) (*types.Value, error) {
	p := New(input)
	return p.Parse()
}

// ParseAll is a convenience function to parse all expressions from a string.
func ParseAllString(input string) ([]*types.Value, error) {
	p := New(input)
	return p.ParseAll()
}

func (p *Parser) parseExpr() (*types.Value, error) {
	tok, err := p.lexer.NextToken()
	if err != nil {
		return nil, err
	}

	switch tok.Type {
	case lexer.TokenEof:
		return nil, fmt.Errorf("unexpected end of input")

	case lexer.TokenInteger:
		return parseInteger(tok)

	case lexer.TokenFloat:
		return parseFloat(tok)

	case lexer.TokenString:
		return types.NewString(tok.Value), nil

	case lexer.TokenSymbol:
		// Handle datum comment
		if tok.Value == "#;" {
			// Skip the next datum
			_, err := p.parseExpr()
			if err != nil {
				return nil, fmt.Errorf("%d:%d: expected expression after #;: %w", tok.Line, tok.Column, err)
			}
			// Then parse and return the expression after the skipped datum
			return p.parseExpr()
		}
		return types.NewSymbol(tok.Value), nil

	case lexer.TokenBoolean:
		return types.NewBoolean(tok.Value == "#t"), nil

	case lexer.TokenCharacter:
		runes := []rune(tok.Value)
		if len(runes) == 0 {
			return types.NewChar(0), nil
		}
		return types.NewChar(runes[0]), nil

	case lexer.TokenLeftParen:
		return p.parseList(tok.Line, tok.Column, lexer.TokenRightParen)

	case lexer.TokenLeftBracket:
		return p.parseList(tok.Line, tok.Column, lexer.TokenRightBracket)

	case lexer.TokenHashLeftParen:
		return p.parseVector(tok.Line, tok.Column)

	case lexer.TokenHashU8LeftParen:
		return p.parseBytevector(tok.Line, tok.Column)

	case lexer.TokenQuote:
		return p.parseAbbrev("quote", tok.Line, tok.Column)

	case lexer.TokenQuasiquote:
		return p.parseAbbrev("quasiquote", tok.Line, tok.Column)

	case lexer.TokenUnquote:
		return p.parseAbbrev("unquote", tok.Line, tok.Column)

	case lexer.TokenUnquoteSplicing:
		return p.parseAbbrev("unquote-splicing", tok.Line, tok.Column)

	case lexer.TokenRightParen:
		return nil, fmt.Errorf("%d:%d: unexpected ')'", tok.Line, tok.Column)

	case lexer.TokenRightBracket:
		return nil, fmt.Errorf("%d:%d: unexpected ']'", tok.Line, tok.Column)

	case lexer.TokenDot:
		return nil, fmt.Errorf("%d:%d: unexpected '.'", tok.Line, tok.Column)

	default:
		return nil, fmt.Errorf("%d:%d: unexpected token: %s", tok.Line, tok.Column, tok.Value)
	}
}

func (p *Parser) parseList(line, col int, closer lexer.TokenType) (*types.Value, error) {
	var elements []*types.Value

	for {
		tok, err := p.lexer.PeekToken()
		if err != nil {
			return nil, err
		}

		if tok.Type == lexer.TokenEof {
			return nil, fmt.Errorf("%d:%d: unterminated list", line, col)
		}

		if tok.Type == closer {
			p.lexer.NextToken() // consume closer
			return types.FromList(elements), nil
		}

		if tok.Type == lexer.TokenDot {
			p.lexer.NextToken() // consume dot
			if len(elements) == 0 {
				return nil, fmt.Errorf("%d:%d: unexpected '.' at start of list", tok.Line, tok.Column)
			}
			// Parse the cdr
			cdr, err := p.parseExpr()
			if err != nil {
				return nil, err
			}

			// Expect closer
			closerTok, err := p.lexer.NextToken()
			if err != nil {
				return nil, err
			}
			if closerTok.Type != closer {
				return nil, fmt.Errorf("%d:%d: expected closing delimiter after dotted pair", closerTok.Line, closerTok.Column)
			}

			// Build the dotted list
			result := cdr
			for i := len(elements) - 1; i >= 0; i-- {
				result = types.Cons(elements[i], result)
			}
			return result, nil
		}

		elem, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		elements = append(elements, elem)
	}
}

func (p *Parser) parseVector(line, col int) (*types.Value, error) {
	var elements []*types.Value
	for {
		tok, err := p.lexer.PeekToken()
		if err != nil {
			return nil, err
		}
		if tok.Type == lexer.TokenEof {
			return nil, fmt.Errorf("%d:%d: unterminated vector", line, col)
		}
		if tok.Type == lexer.TokenRightParen {
			p.lexer.NextToken()
			return types.NewVector(elements), nil
		}
		elem, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		elements = append(elements, elem)
	}
}

func (p *Parser) parseBytevector(line, col int) (*types.Value, error) {
	var bytes []byte
	for {
		tok, err := p.lexer.PeekToken()
		if err != nil {
			return nil, err
		}
		if tok.Type == lexer.TokenEof {
			return nil, fmt.Errorf("%d:%d: unterminated bytevector", line, col)
		}
		if tok.Type == lexer.TokenRightParen {
			p.lexer.NextToken()
			return types.NewBytevector(bytes), nil
		}
		elem, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if elem.Type != types.TypeInteger {
			return nil, fmt.Errorf("%d:%d: bytevector elements must be integers", line, col)
		}
		if elem.IntVal < 0 || elem.IntVal > 255 {
			return nil, fmt.Errorf("%d:%d: bytevector element out of range (0-255): %d", line, col, elem.IntVal)
		}
		bytes = append(bytes, byte(elem.IntVal))
	}
}

func (p *Parser) parseAbbrev(name string, line, col int) (*types.Value, error) {
	expr, err := p.parseExpr()
	if err != nil {
		return nil, fmt.Errorf("%d:%d: expected expression after %s: %w", line, col, name, err)
	}
	return types.FromList([]*types.Value{types.NewSymbol(name), expr}), nil
}

// ---------------------------------------------------------------------------
// Number parsing
// ---------------------------------------------------------------------------

func parseInteger(tok lexer.Token) (*types.Value, error) {
	s := tok.Value

	// Handle radix prefixes
	if len(s) >= 2 && s[0] == '#' {
		prefix := s[1]
		numStr := s[2:]
		var base int
		switch prefix {
		case 'b', 'B':
			base = 2
		case 'o', 'O':
			base = 8
		case 'd', 'D':
			base = 10
		case 'x', 'X':
			base = 16
		default:
			return nil, fmt.Errorf("%d:%d: unknown radix prefix: #%c", tok.Line, tok.Column, prefix)
		}
		n, err := strconv.ParseInt(numStr, base, 64)
		if err != nil {
			return nil, fmt.Errorf("%d:%d: invalid integer: %s: %w", tok.Line, tok.Column, s, err)
		}
		return types.NewInteger(n), nil
	}

	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%d:%d: invalid integer: %s: %w", tok.Line, tok.Column, s, err)
	}
	return types.NewInteger(n), nil
}

func parseFloat(tok lexer.Token) (*types.Value, error) {
	s := tok.Value

	// Handle special float values
	switch strings.ToLower(s) {
	case "+inf.0":
		return types.NewFloat(math.Inf(1)), nil
	case "-inf.0":
		return types.NewFloat(math.Inf(-1)), nil
	case "+nan.0":
		return types.NewFloat(math.NaN()), nil
	}

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil, fmt.Errorf("%d:%d: invalid float: %s: %w", tok.Line, tok.Column, s, err)
	}
	return types.NewFloat(f), nil
}
