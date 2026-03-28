package evaluator

import (
	"testing"

	"github.com/suchi/sucheme-go/pkg/builtin"
	"github.com/suchi/sucheme-go/pkg/env"
	"github.com/suchi/sucheme-go/pkg/types"
)

func setupMacroEnv() *env.Env {
	e := env.Global()
	builtin.RegisterBuiltins(e)
	RegisterApply(e)
	return e
}

func evalMacro(t *testing.T, input string) *types.Value {
	t.Helper()
	e := setupMacroEnv()
	result, err := EvalString(input, e)
	if err != nil {
		t.Fatalf("unexpected error evaluating %q: %v", input, err)
	}
	return result
}

func evalMacroErr(t *testing.T, input string) error {
	t.Helper()
	e := setupMacroEnv()
	_, err := EvalString(input, e)
	return err
}

// =========================================================================
// define-syntax basic
// =========================================================================

func TestDefineSyntaxBasic(t *testing.T) {
	r := evalMacro(t, `
		(define-syntax my-if
			(syntax-rules ()
				((my-if test then else)
				 (cond (test then) (#t else)))))
		(my-if #t 1 2)
	`)
	if r.IntVal != 1 {
		t.Errorf("expected 1, got %s", r.String())
	}
}

func TestDefineSyntaxFalse(t *testing.T) {
	r := evalMacro(t, `
		(define-syntax my-if
			(syntax-rules ()
				((my-if test then else)
				 (cond (test then) (#t else)))))
		(my-if #f 1 2)
	`)
	if r.IntVal != 2 {
		t.Errorf("expected 2, got %s", r.String())
	}
}

// =========================================================================
// syntax-rules with ellipsis
// =========================================================================

func TestMacroEllipsis(t *testing.T) {
	r := evalMacro(t, `
		(define-syntax my-begin
			(syntax-rules ()
				((my-begin expr ...)
				 (let () expr ...))))
		(my-begin 1 2 3)
	`)
	if r.IntVal != 3 {
		t.Errorf("expected 3, got %s", r.String())
	}
}

func TestMacroEllipsisZero(t *testing.T) {
	r := evalMacro(t, `
		(define-syntax my-list
			(syntax-rules ()
				((my-list x ...)
				 (list x ...))))
		(my-list)
	`)
	if r.Type != types.TypeNil {
		t.Errorf("expected (), got %s", r.String())
	}
}

func TestMacroEllipsisMultiple(t *testing.T) {
	r := evalMacro(t, `
		(define-syntax my-list
			(syntax-rules ()
				((my-list x ...)
				 (list x ...))))
		(my-list 1 2 3)
	`)
	lst, _ := r.ToList()
	if len(lst) != 3 || lst[0].IntVal != 1 || lst[2].IntVal != 3 {
		t.Errorf("expected (1 2 3), got %s", r.String())
	}
}

func TestMacroWhen(t *testing.T) {
	r := evalMacro(t, `
		(define-syntax my-when
			(syntax-rules ()
				((my-when test body ...)
				 (if test (begin body ...)))))
		(my-when #t 1 2 3)
	`)
	if r.IntVal != 3 {
		t.Errorf("expected 3, got %s", r.String())
	}
}

func TestMacroWhenFalse(t *testing.T) {
	r := evalMacro(t, `
		(define-syntax my-when
			(syntax-rules ()
				((my-when test body ...)
				 (if test (begin body ...)))))
		(my-when #f 1 2 3)
	`)
	if r.Type != types.TypeVoid {
		t.Errorf("expected void, got %s", r.String())
	}
}

// =========================================================================
// Multiple rules
// =========================================================================

func TestMacroMultipleRules(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{`(define-syntax my-and
			(syntax-rules ()
				((my-and) #t)
				((my-and x) x)
				((my-and x y) (if x y #f))))
		  (my-and)`, "#t"},
		{`(define-syntax my-and
			(syntax-rules ()
				((my-and) #t)
				((my-and x) x)
				((my-and x y) (if x y #f))))
		  (my-and 42)`, "42"},
		{`(define-syntax my-and
			(syntax-rules ()
				((my-and) #t)
				((my-and x) x)
				((my-and x y) (if x y #f))))
		  (my-and #t 42)`, "42"},
		{`(define-syntax my-and
			(syntax-rules ()
				((my-and) #t)
				((my-and x) x)
				((my-and x y) (if x y #f))))
		  (my-and #f 42)`, "#f"},
	}
	for _, tt := range tests {
		t.Run(tt.expect, func(t *testing.T) {
			r := evalMacro(t, tt.input)
			if r.String() != tt.expect {
				t.Errorf("got %s, want %s", r.String(), tt.expect)
			}
		})
	}
}

// =========================================================================
// Literals in syntax-rules
// =========================================================================

func TestMacroLiterals(t *testing.T) {
	r := evalMacro(t, `
		(define-syntax my-cond
			(syntax-rules (else)
				((my-cond (else e ...))
				 (begin e ...))
				((my-cond (test e ...))
				 (if test (begin e ...)))))
		(my-cond (else 42))
	`)
	if r.IntVal != 42 {
		t.Errorf("expected 42, got %s", r.String())
	}
}

// =========================================================================
// let-syntax
// =========================================================================

func TestLetSyntax(t *testing.T) {
	r := evalMacro(t, `
		(let-syntax ((double (syntax-rules ()
			((double x) (+ x x)))))
			(double 5))
	`)
	if r.IntVal != 10 {
		t.Errorf("expected 10, got %s", r.String())
	}
}

func TestLetSyntaxScope(t *testing.T) {
	// double should not be visible outside let-syntax
	err := evalMacroErr(t, `
		(let-syntax ((double (syntax-rules ()
			((double x) (+ x x)))))
			(double 5))
		(double 3)
	`)
	if err == nil {
		t.Error("expected error: double should not be visible outside let-syntax")
	}
}

// =========================================================================
// letrec-syntax
// =========================================================================

func TestLetrecSyntax(t *testing.T) {
	r := evalMacro(t, `
		(letrec-syntax ((double (syntax-rules ()
			((double x) (+ x x)))))
			(double 7))
	`)
	if r.IntVal != 14 {
		t.Errorf("expected 14, got %s", r.String())
	}
}

// =========================================================================
// Nested patterns
// =========================================================================

func TestMacroDottedPattern(t *testing.T) {
	r := evalMacro(t, `
		(define-syntax my-apply
			(syntax-rules ()
				((my-apply f a . rest)
				 (f a))))
		(my-apply + 1 2 3)
	`)
	if r.IntVal != 1 {
		t.Errorf("expected 1, got %s", r.String())
	}
}

func TestMacroDottedPatternRest(t *testing.T) {
	r := evalMacro(t, `
		(define-syntax first-arg
			(syntax-rules ()
				((first-arg x . _) x)))
		(first-arg 42 99 100)
	`)
	if r.IntVal != 42 {
		t.Errorf("expected 42, got %s", r.String())
	}
}

func TestMacroNestedPattern(t *testing.T) {
	r := evalMacro(t, `
		(define-syntax swap!
			(syntax-rules ()
				((swap! a b)
				 (let ((tmp a))
					(set! a b)
					(set! b tmp)))))
		(define x 1)
		(define y 2)
		(swap! x y)
		(list x y)
	`)
	lst, _ := r.ToList()
	if len(lst) != 2 || lst[0].IntVal != 2 || lst[1].IntVal != 1 {
		t.Errorf("expected (2 1), got %s", r.String())
	}
}

// =========================================================================
// Error cases
// =========================================================================

func TestMacroNoMatch(t *testing.T) {
	err := evalMacroErr(t, `
		(define-syntax my-if
			(syntax-rules ()
				((my-if test then else)
				 (cond (test then) (#t else)))))
		(my-if #t 1)
	`)
	if err == nil {
		t.Error("expected error: no matching pattern")
	}
}

func TestMacroInvalidSyntaxRules(t *testing.T) {
	err := evalMacroErr(t, `(define-syntax bad (syntax-rules))`)
	if err == nil {
		t.Error("expected error: invalid syntax-rules")
	}
}

func TestSyntaxRulesNotTopLevel(t *testing.T) {
	err := evalMacroErr(t, `(syntax-rules () ((x) 1))`)
	if err == nil {
		t.Error("expected error: syntax-rules outside define-syntax")
	}
}

// =========================================================================
// Wildcard pattern _
// =========================================================================

func TestMacroWildcard(t *testing.T) {
	r := evalMacro(t, `
		(define-syntax first
			(syntax-rules ()
				((first a _) a)))
		(first 42 99)
	`)
	if r.IntVal != 42 {
		t.Errorf("expected 42, got %s", r.String())
	}
}

// =========================================================================
// Practical: while loop macro
// =========================================================================

func TestMacroWhileLoop(t *testing.T) {
	r := evalMacro(t, `
		(define-syntax my-while
			(syntax-rules ()
				((my-while test body ...)
				 (let loop ()
					(when test body ... (loop))))))
		(define counter 0)
		(my-while (< counter 5)
			(set! counter (+ counter 1)))
		counter
	`)
	if r.IntVal != 5 {
		t.Errorf("expected 5, got %s", r.String())
	}
}
