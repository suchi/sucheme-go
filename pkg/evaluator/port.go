// Package evaluator - port.go implements R7RS port/IO system.
package evaluator

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/suchi/sucheme-go/pkg/builtin"
	"github.com/suchi/sucheme-go/pkg/env"
	"github.com/suchi/sucheme-go/pkg/parser"
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
	Src      io.Reader        // underlying reader for string input ports (not closable)
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

// RegisterPortBuiltins registers port-related procedures and initializes
// the default input/output ports so that current-input-port and
// current-output-port return usable port values.
func RegisterPortBuiltins(e *env.Env) {
	// Initialize default ports so current-input-port / current-output-port
	// return values with non-nil PortVal. Without this, port? and other
	// predicates would fail on the defaults.
	stdinPort := &Port{
		Name:    "stdin",
		IsInput: true,
		Reader:  bufio.NewReader(strings.NewReader("")),
	}
	builtin.SetDefaultInputPort(makePortValue(stdinPort))

	stdoutPort := &Port{
		Name:     "stdout",
		IsOutput: true,
		Writer:   &strings.Builder{},
	}
	builtin.SetDefaultOutputPort(makePortValue(stdoutPort))

	e.Define("open-input-string", types.NewProcedure("open-input-string", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 || args[0].Type != types.TypeString {
			return nil, fmt.Errorf("expected a string")
		}
		port := NewInputStringPort(args[0].StrVal)
		return makePortValue(port), nil
	}))

	e.Define("open-output-string", types.NewProcedure("open-output-string", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("expected no arguments")
		}
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
		r, _, readErr := port.Reader.ReadRune()
		if readErr == io.EOF {
			return types.Eof(), nil
		}
		if readErr != nil {
			return nil, readErr
		}
		return types.NewChar(r), nil
	}))

	e.Define("peek-char", types.NewProcedure("peek-char", func(args []*types.Value) (*types.Value, error) {
		port, err := getInputPort(args)
		if err != nil {
			return nil, err
		}
		// Read one rune then unread it to implement peek semantics safely for UTF-8
		r, _, readErr := port.Reader.ReadRune()
		if readErr == io.EOF {
			return types.Eof(), nil
		}
		if readErr != nil {
			return nil, readErr
		}
		if unreadErr := port.Reader.UnreadRune(); unreadErr != nil {
			return nil, unreadErr
		}
		return types.NewChar(r), nil
	}))

	e.Define("read-line", types.NewProcedure("read-line", func(args []*types.Value) (*types.Value, error) {
		port, err := getInputPort(args)
		if err != nil {
			return nil, err
		}
		line, readErr := port.Reader.ReadString('\n')
		if readErr == io.EOF {
			if line == "" {
				return types.Eof(), nil
			}
			return types.NewString(line), nil
		}
		if readErr != nil {
			return nil, readErr
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
		_, peekErr := port.Reader.Peek(1)
		if peekErr == nil {
			return types.True, nil
		}
		if peekErr == io.EOF {
			return types.False, nil
		}
		return nil, peekErr
	}))

	e.Define("write-char", types.NewProcedure("write-char", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("expected character and port")
		}
		if args[0].Type != types.TypeCharacter {
			return nil, fmt.Errorf("expected a character, got %s", args[0].Type)
		}
		port, err := getPort(args[1])
		if err != nil {
			return nil, err
		}
		if !port.IsOutput || port.Writer == nil {
			return nil, fmt.Errorf("expected output port")
		}
		if port.Closed {
			return nil, fmt.Errorf("port is closed")
		}
		port.Writer.WriteRune(args[0].CharVal)
		return types.Void(), nil
	}))

	e.Define("write-string", types.NewProcedure("write-string", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("expected string and port")
		}
		if args[0].Type != types.TypeString {
			return nil, fmt.Errorf("expected a string, got %s", args[0].Type)
		}
		port, err := getPort(args[1])
		if err != nil {
			return nil, err
		}
		if !port.IsOutput || port.Writer == nil {
			return nil, fmt.Errorf("expected output port")
		}
		if port.Closed {
			return nil, fmt.Errorf("port is closed")
		}
		port.Writer.WriteString(args[0].StrVal)
		return types.Void(), nil
	}))

	e.Define("port?", types.NewProcedure("port?", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		_, err := getPort(args[0])
		return types.NewBoolean(err == nil), nil
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
		_, err := getPort(args[0])
		return types.NewBoolean(err == nil), nil
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
		if len(args) != 0 {
			return nil, fmt.Errorf("expected 0 arguments")
		}
		return types.Eof(), nil
	}))

	e.Define("eof-object?", types.NewProcedure("eof-object?", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		return types.NewBoolean(args[0].Type == types.TypeEof), nil
	}))

	// read: R7RS (read) or (read port) - reads one S-expression from input port.
	// This overrides the builtin stub with a proper implementation using the parser.
	e.Define("read", types.NewProcedure("read", func(args []*types.Value) (*types.Value, error) {
		if len(args) > 1 {
			return nil, fmt.Errorf("read: expected 0 or 1 argument, got %d", len(args))
		}
		var port *Port
		if len(args) == 0 {
			return nil, fmt.Errorf("read: current-input-port is not available")
		}
		var err error
		port, err = getInputPort(args)
		if err != nil {
			return nil, fmt.Errorf("read: %w", err)
		}
		// Read available bytes from the port's reader
		var buf []byte
		for {
			b, readErr := port.Reader.ReadByte()
			if readErr == io.EOF {
				break
			}
			if readErr != nil {
				return nil, fmt.Errorf("read: %w", readErr)
			}
			buf = append(buf, b)
		}
		if len(buf) == 0 {
			return types.Eof(), nil
		}
		exprs, parseErr := parser.ParseAllString(string(buf))
		if parseErr != nil {
			return nil, fmt.Errorf("read: %w", parseErr)
		}
		if len(exprs) == 0 {
			return types.Eof(), nil
		}
		return exprs[0], nil
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

// getInputPort extracts an input port from args (exactly 1 port argument required).
func getInputPort(args []*types.Value) (*Port, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("expected 1 argument (port)")
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
