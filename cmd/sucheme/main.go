// Package main provides the CLI entry point for the sucheme Scheme interpreter.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/suchi/sucheme-go/pkg/builtin"
	"github.com/suchi/sucheme-go/pkg/env"
	"github.com/suchi/sucheme-go/pkg/evaluator"
	"github.com/suchi/sucheme-go/pkg/parser"
	"github.com/suchi/sucheme-go/pkg/types"
)

// Version is set at build time via -ldflags.
var Version = "dev"

func main() {
	var (
		evalExpr    string
		interactive bool
		showVersion bool
		quiet       bool
		loadFiles   stringSlice
	)

	flag.StringVar(&evalExpr, "e", "", "evaluate expression and print result")
	flag.BoolVar(&interactive, "i", false, "enter interactive REPL after loading files")
	flag.BoolVar(&showVersion, "version", false, "print version and exit")
	flag.BoolVar(&quiet, "q", false, "suppress REPL banner")
	flag.Var(&loadFiles, "l", "load a Scheme file before execution (can be repeated)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: sucheme [options] [file ...]\n\n")
		fmt.Fprintf(os.Stderr, "A R7RS Scheme interpreter.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  sucheme                    Start interactive REPL\n")
		fmt.Fprintf(os.Stderr, "  sucheme file.scm           Execute a Scheme file\n")
		fmt.Fprintf(os.Stderr, "  sucheme -e '(+ 1 2)'      Evaluate an expression\n")
		fmt.Fprintf(os.Stderr, "  sucheme -l lib.scm file.scm  Load lib, then run file\n")
		fmt.Fprintf(os.Stderr, "  sucheme -i file.scm        Run file, then enter REPL\n")
	}
	flag.Parse()

	if showVersion {
		fmt.Printf("sucheme %s\n", Version)
		os.Exit(0)
	}

	e := env.Global()
	builtin.RegisterBuiltins(e)
	evaluator.RegisterApply(e)
	evaluator.RegisterExceptionBuiltins(e)
	evaluator.RegisterMapForEach(e)
	registerDisplayWrite(e)

	// Load -l files
	for _, f := range loadFiles {
		if err := loadFile(f, e); err != nil {
			fmt.Fprintf(os.Stderr, "Error loading %s: %v\n", f, err)
			os.Exit(1)
		}
	}

	args := flag.Args()

	// -e: evaluate expression
	if evalExpr != "" {
		result, err := evaluator.EvalString(evalExpr, e)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if result != nil && result.Type != types.TypeVoid {
			fmt.Println(result.String())
		}
		if !interactive {
			return
		}
	}

	// File arguments: execute each
	if len(args) > 0 {
		for _, f := range args {
			if err := loadFile(f, e); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}
		if !interactive {
			return
		}
	}

	// REPL
	runREPL(e, quiet)
}

func loadFile(path string, e *env.Env) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot read file %s: %w", path, err)
	}
	exprs, err := parser.ParseAllString(string(data))
	if err != nil {
		return fmt.Errorf("parse error in %s: %w", path, err)
	}
	for _, expr := range exprs {
		_, err := evaluator.Eval(expr, e)
		if err != nil {
			return fmt.Errorf("error in %s: %w", path, err)
		}
	}
	return nil
}

func runREPL(e *env.Env, quiet bool) {
	if !quiet {
		fmt.Println("sucheme " + Version)
		fmt.Println("Type (exit) or Ctrl+D to quit.")
	}

	reader := bufio.NewReader(os.Stdin)
	var buf strings.Builder
	prompt := "sucheme> "

	for {
		fmt.Print(prompt)
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				if buf.Len() > 0 {
					fmt.Fprintln(os.Stderr, "\nError: unexpected end of input")
				} else {
					fmt.Println()
				}
				return
			}
			fmt.Fprintf(os.Stderr, "\nError reading input: %v\n", err)
			return
		}

		buf.WriteString(line)
		input := strings.TrimSpace(buf.String())

		if input == "" {
			buf.Reset()
			continue
		}

		// Check for balanced parentheses
		if !isComplete(input) {
			prompt = "... "
			continue
		}

		// Special REPL commands
		if input == "(exit)" || input == "(quit)" {
			return
		}

		result, err := evaluator.EvalString(input, e)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		} else if result != nil && result.Type != types.TypeVoid {
			fmt.Println(result.String())
		}

		buf.Reset()
		prompt = "sucheme> "
	}
}

// isComplete checks if the input is a complete expression (not awaiting more input).
// Returns true when parentheses/brackets are balanced or over-closed,
// respecting strings, line comments, and #|...|# block comments.
func isComplete(input string) bool {
	runes := []rune(input)
	depth := 0
	inString := false
	escaped := false
	inLineComment := false
	blockCommentDepth := 0

	for i := 0; i < len(runes); i++ {
		ch := runes[i]

		// Block comment handling (nestable)
		if blockCommentDepth > 0 {
			if ch == '|' && i+1 < len(runes) && runes[i+1] == '#' {
				blockCommentDepth--
				i++ // skip '#'
			} else if ch == '#' && i+1 < len(runes) && runes[i+1] == '|' {
				blockCommentDepth++
				i++ // skip '|'
			}
			continue
		}

		if inLineComment {
			if ch == '\n' {
				inLineComment = false
			}
			continue
		}
		if escaped {
			escaped = false
			continue
		}
		if inString {
			if ch == '\\' {
				escaped = true
			} else if ch == '"' {
				inString = false
			}
			continue
		}

		// Check for block comment start
		if ch == '#' && i+1 < len(runes) && runes[i+1] == '|' {
			blockCommentDepth++
			i++ // skip '|'
			continue
		}

		// Check for datum comment (#;) - just skip the '#;' prefix,
		// the datum itself will be counted normally for balance purposes
		if ch == '#' && i+1 < len(runes) && runes[i+1] == ';' {
			i++ // skip ';'
			continue
		}

		// Check for character literal (#\x) - skip the character
		// to avoid counting #\( or #\) as parentheses
		if ch == '#' && i+1 < len(runes) && runes[i+1] == '\\' {
			i += 2 // skip '#' and '\'
			if i >= len(runes) {
				// Input ends with #\ — incomplete
				return false
			}
			// Always consume at least one character (the literal char itself)
			i++
			// Then skip remaining characters of named character literals
			// (e.g., #\space, #\newline, #\x41)
			for i < len(runes) {
				c := runes[i]
				if c == ' ' || c == '\t' || c == '\n' || c == '(' ||
					c == ')' || c == '[' || c == ']' || c == '"' || c == ';' {
					break
				}
				i++
			}
			i-- // will be incremented by for loop
			continue
		}

		switch ch {
		case ';':
			inLineComment = true
		case '"':
			inString = true
		case '(', '[':
			depth++
		case ')', ']':
			depth--
		}
	}

	return depth <= 0 && !inString && blockCommentDepth == 0
}

// registerDisplayWrite registers display and write with actual I/O.
func registerDisplayWrite(e *env.Env) {
	e.Define("display", types.NewProcedure("display", func(args []*types.Value) (*types.Value, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("expected at least 1 argument")
		}
		fmt.Print(args[0].Display())
		return types.Void(), nil
	}))

	e.Define("write", types.NewProcedure("write", func(args []*types.Value) (*types.Value, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("expected at least 1 argument")
		}
		fmt.Print(args[0].String())
		return types.Void(), nil
	}))

	e.Define("newline", types.NewProcedure("newline", func(args []*types.Value) (*types.Value, error) {
		fmt.Println()
		return types.Void(), nil
	}))

	e.Define("print", types.NewProcedure("print", func(args []*types.Value) (*types.Value, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("expected at least 1 argument")
		}
		fmt.Print(args[0].String())
		fmt.Println()
		return types.Void(), nil
	}))
}

// stringSlice implements flag.Value for repeatable -l flags.
type stringSlice []string

func (s *stringSlice) String() string { return strings.Join(*s, ", ") }
func (s *stringSlice) Set(val string) error {
	*s = append(*s, val)
	return nil
}
