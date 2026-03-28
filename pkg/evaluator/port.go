// Package evaluator - port.go implements R7RS port/IO system.
package evaluator

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/suchi/sucheme-go/pkg/env"
	"github.com/suchi/sucheme-go/pkg/types"
)

// Port represents a Scheme port (input or output, textual).
type Port struct {
	Name     string
	IsInput  bool
	IsOutput bool
	Closed   bool
	Reader   *bufio.Reader
	Writer   *strings.Builder // for string output ports
	Src      io.Reader        // underlying reader (for close)
}

// NewInputStringPort creates an input port reading from a string.
func NewInputStringPort(s string) *Port {
	r := strings.NewReader(s)
	return &Port{
		Name:    "string-input",
		IsInput: true,
		Reader:  bufio.NewReader(r),
		Src:     r,
	}
}

// NewOutputStringPort creates an output port writing to a string buffer.
func NewOutputStringPort() *Port {
	return &Port{
		Name:     "string-output",
		IsOutput: true,
		Writer:   &strings.Builder{},
	}
}

// RegisterPortBuiltins registers port-related procedures.
func RegisterPortBuiltins(e *env.Env) {
	e.Define("open-input-string", types.NewProcedure("open-input-string", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 || args[0].Type != types.TypeString {
			return nil, fmt.Errorf("expected a string")
		}
		port := NewInputStringPort(args[0].StrVal)
		return makePortValue(port), nil
	}))

	e.Define("open-output-string", types.NewProcedure("open-output-string", func(args []*types.Value) (*types.Value, error) {
		port := NewOutputStringPort()
		return makePortValue(port), nil
	}))

	e.Define("get-output-string", types.NewProcedure("get-output-string", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		port, err := getPort(args[0])
		if err != nil {
			return nil, err
		}
		if !port.IsOutput || port.Writer == nil {
			return nil, fmt.Errorf("expected output string port")
		}
		return types.NewString(port.Writer.String()), nil
	}))

	e.Define("read-char", types.NewProcedure("read-char", func(args []*types.Value) (*types.Value, error) {
		port, err := getInputPort(args)
		if err != nil {
			return nil, err
		}
		r, _, err := port.Reader.ReadRune()
		if err == io.EOF {
			return types.Eof(), nil
		}
		if err != nil {
			return nil, fmt.Errorf("read-char: %v", err)
		}
		return types.NewChar(r), nil
	}))

	e.Define("peek-char", types.NewProcedure("peek-char", func(args []*types.Value) (*types.Value, error) {
		port, err := getInputPort(args)
		if err != nil {
			return nil, err
		}
		// Peek at least 1 byte
		bytes, err := port.Reader.Peek(1)
		if len(bytes) == 0 {
			return types.Eof(), nil
		}
		// Try to peek enough for a full rune
		if bytes[0] >= 0x80 {
			bytes, _ = port.Reader.Peek(4)
		}
		r, _ := decodeRune(bytes)
		return types.NewChar(r), nil
	}))

	e.Define("read-line", types.NewProcedure("read-line", func(args []*types.Value) (*types.Value, error) {
		port, err := getInputPort(args)
		if err != nil {
			return nil, err
		}
		line, err := port.Reader.ReadString('\n')
		if err == io.EOF {
			if line == "" {
				return types.Eof(), nil
			}
			return types.NewString(line), nil
		}
		if err != nil {
			return nil, fmt.Errorf("read-line: %v", err)
		}
		// Remove trailing newline
		if len(line) > 0 && line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
		}
		return types.NewString(line), nil
	}))

	e.Define("char-ready?", types.NewProcedure("char-ready?", func(args []*types.Value) (*types.Value, error) {
		port, err := getInputPort(args)
		if err != nil {
			return nil, err
		}
		return types.NewBoolean(port.Reader.Buffered() > 0), nil
	}))

	e.Define("write-char", types.NewProcedure("write-char", func(args []*types.Value) (*types.Value, error) {
		if len(args) < 1 || args[0].Type != types.TypeCharacter {
			return nil, fmt.Errorf("expected a character")
		}
		if len(args) >= 2 {
			port, err := getPort(args[1])
			if err != nil {
				return nil, err
			}
			if !port.IsOutput || port.Writer == nil {
				return nil, fmt.Errorf("expected output port")
			}
			port.Writer.WriteRune(args[0].CharVal)
		} else {
			fmt.Print(string(args[0].CharVal))
		}
		return types.Void(), nil
	}))

	e.Define("write-string", types.NewProcedure("write-string", func(args []*types.Value) (*types.Value, error) {
		if len(args) < 1 || args[0].Type != types.TypeString {
			return nil, fmt.Errorf("expected a string")
		}
		if len(args) >= 2 {
			port, err := getPort(args[1])
			if err != nil {
				return nil, err
			}
			if !port.IsOutput || port.Writer == nil {
				return nil, fmt.Errorf("expected output port")
			}
			port.Writer.WriteString(args[0].StrVal)
		} else {
			fmt.Print(args[0].StrVal)
		}
		return types.Void(), nil
	}))

	e.Define("port?", types.NewProcedure("port?", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		return types.NewBoolean(args[0].Type == types.TypePort), nil
	}))

	e.Define("input-port?", types.NewProcedure("input-port?", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		if args[0].Type != types.TypePort {
			return types.False, nil
		}
		port, err := getPort(args[0])
		if err != nil {
			return types.False, nil
		}
		return types.NewBoolean(port.IsInput), nil
	}))

	e.Define("output-port?", types.NewProcedure("output-port?", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		if args[0].Type != types.TypePort {
			return types.False, nil
		}
		port, err := getPort(args[0])
		if err != nil {
			return types.False, nil
		}
		return types.NewBoolean(port.IsOutput), nil
	}))

	e.Define("textual-port?", types.NewProcedure("textual-port?", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		// All ports in this implementation are textual
		return types.NewBoolean(args[0].Type == types.TypePort), nil
	}))

	e.Define("close-port", types.NewProcedure("close-port", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		port, err := getPort(args[0])
		if err != nil {
			return nil, err
		}
		port.Closed = true
		return types.Void(), nil
	}))

	e.Define("close-input-port", types.NewProcedure("close-input-port", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		port, err := getPort(args[0])
		if err != nil {
			return nil, err
		}
		if !port.IsInput {
			return nil, fmt.Errorf("expected input port")
		}
		port.Closed = true
		return types.Void(), nil
	}))

	e.Define("close-output-port", types.NewProcedure("close-output-port", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		port, err := getPort(args[0])
		if err != nil {
			return nil, err
		}
		if !port.IsOutput {
			return nil, fmt.Errorf("expected output port")
		}
		port.Closed = true
		return types.Void(), nil
	}))

	e.Define("eof-object", types.NewProcedure("eof-object", func(args []*types.Value) (*types.Value, error) {
		return types.Eof(), nil
	}))

	e.Define("eof-object?", types.NewProcedure("eof-object?", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		return types.NewBoolean(args[0].Type == types.TypeEof), nil
	}))
}

// makePortValue creates a Value wrapping a Port.
func makePortValue(p *Port) *types.Value {
	v := &types.Value{Type: types.TypePort}
	v.PortVal = p
	return v
}

// getPort extracts a Port from a Value.
func getPort(v *types.Value) (*Port, error) {
	if v.Type != types.TypePort || v.PortVal == nil {
		return nil, fmt.Errorf("expected a port")
	}
	p, ok := v.PortVal.(*Port)
	if !ok {
		return nil, fmt.Errorf("invalid port value")
	}
	return p, nil
}

// getInputPort extracts an input port from args (optional port argument).
func getInputPort(args []*types.Value) (*Port, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("no port specified and current-input-port not yet implemented")
	}
	port, err := getPort(args[0])
	if err != nil {
		return nil, err
	}
	if !port.IsInput {
		return nil, fmt.Errorf("expected input port")
	}
	if port.Closed {
		return nil, fmt.Errorf("port is closed")
	}
	return port, nil
}

// decodeRune decodes the first rune from a byte slice.
func decodeRune(b []byte) (rune, int) {
	if len(b) == 0 {
		return 0, 0
	}
	// Simple UTF-8 decode
	c := rune(b[0])
	if c < 0x80 {
		return c, 1
	}
	// Multi-byte: use string conversion
	s := string(b)
	for _, r := range s {
		return r, len(string(r))
	}
	return 0, 0
}
