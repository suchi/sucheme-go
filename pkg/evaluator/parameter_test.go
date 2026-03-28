package evaluator

import (
	"testing"

	"github.com/suchi/sucheme-go/pkg/types"
)

// Reuses setupEval, evalHelper, evalErr from evaluator_test.go

// =========================================================================
// make-parameter
// =========================================================================

func TestMakeParameterBasic(t *testing.T) {
	r := evalHelper(t, `
		(define p (make-parameter 42))
		(p)
	`)
	if r.IntVal != 42 {
		t.Errorf("expected 42, got %s", r.String())
	}
}

func TestMakeParameterWithConverter(t *testing.T) {
	r := evalHelper(t, `
		(define p (make-parameter "10" string->number))
		(p)
	`)
	if r.IntVal != 10 {
		t.Errorf("expected 10, got %s", r.String())
	}
}

// =========================================================================
// parameterize
// =========================================================================

func TestParameterizeBasic(t *testing.T) {
	r := evalHelper(t, `
		(define p (make-parameter 1))
		(parameterize ((p 2))
			(p))
	`)
	if r.IntVal != 2 {
		t.Errorf("expected 2, got %s", r.String())
	}
}

func TestParameterizeRestore(t *testing.T) {
	r := evalHelper(t, `
		(define p (make-parameter 1))
		(parameterize ((p 2))
			(p))
		(p)
	`)
	if r.IntVal != 1 {
		t.Errorf("expected 1 (restored), got %s", r.String())
	}
}

func TestParameterizeNested(t *testing.T) {
	r := evalHelper(t, `
		(define p (make-parameter 1))
		(parameterize ((p 2))
			(parameterize ((p 3))
				(p)))
	`)
	if r.IntVal != 3 {
		t.Errorf("expected 3, got %s", r.String())
	}
}

func TestParameterizeNestedRestore(t *testing.T) {
	r := evalHelper(t, `
		(define p (make-parameter 1))
		(define result '())
		(parameterize ((p 2))
			(set! result (cons (p) result))
			(parameterize ((p 3))
				(set! result (cons (p) result)))
			(set! result (cons (p) result)))
		(set! result (cons (p) result))
		(reverse result)
	`)
	lst, _ := r.ToList()
	if len(lst) != 4 || lst[0].IntVal != 2 || lst[1].IntVal != 3 || lst[2].IntVal != 2 || lst[3].IntVal != 1 {
		t.Errorf("expected (2 3 2 1), got %s", r.String())
	}
}

func TestParameterizeMultiple(t *testing.T) {
	r := evalHelper(t, `
		(define a (make-parameter 1))
		(define b (make-parameter 10))
		(parameterize ((a 2) (b 20))
			(+ (a) (b)))
	`)
	if r.IntVal != 22 {
		t.Errorf("expected 22, got %s", r.String())
	}
}

func TestParameterizeWithConverter(t *testing.T) {
	r := evalHelper(t, `
		(define p (make-parameter 0 (lambda (x) (* x 2))))
		(parameterize ((p 5))
			(p))
	`)
	if r.IntVal != 10 {
		t.Errorf("expected 10, got %s", r.String())
	}
}

func TestParameterizeError(t *testing.T) {
	err := evalErr(t, `(parameterize ((42 1)) 1)`)
	if err == nil {
		t.Error("expected error for non-parameter")
	}
}

// =========================================================================
// Port I/O
// =========================================================================

func TestOpenInputString(t *testing.T) {
	r := evalHelper(t, `
		(define p (open-input-string "abc"))
		(input-port? p)
	`)
	if !r.BoolVal {
		t.Error("expected #t for input-port?")
	}
}

func TestReadChar(t *testing.T) {
	r := evalHelper(t, `
		(define p (open-input-string "ab"))
		(list (read-char p) (read-char p))
	`)
	lst, _ := r.ToList()
	if len(lst) != 2 || lst[0].CharVal != 'a' || lst[1].CharVal != 'b' {
		t.Errorf("expected (#\\a #\\b), got %s", r.String())
	}
}

func TestReadCharEof(t *testing.T) {
	r := evalHelper(t, `
		(define p (open-input-string ""))
		(eof-object? (read-char p))
	`)
	if !r.BoolVal {
		t.Error("expected #t for eof-object?")
	}
}

func TestReadLine(t *testing.T) {
	r := evalHelper(t, `
		(define p (open-input-string "hello\nworld"))
		(list (read-line p) (read-line p))
	`)
	lst, _ := r.ToList()
	if len(lst) != 2 || lst[0].StrVal != "hello" || lst[1].StrVal != "world" {
		t.Errorf("expected (hello world), got %s", r.String())
	}
}

func TestOutputStringPort(t *testing.T) {
	r := evalHelper(t, `
		(define out (open-output-string))
		(write-string "hello" out)
		(write-char #\! out)
		(get-output-string out)
	`)
	if r.StrVal != "hello!" {
		t.Errorf("expected 'hello!', got %q", r.StrVal)
	}
}

func TestPortPredicates(t *testing.T) {
	tests := []struct {
		input  string
		expect bool
	}{
		{`(port? (open-input-string ""))`, true},
		{`(port? 42)`, false},
		{`(input-port? (open-input-string ""))`, true},
		{`(output-port? (open-output-string))`, true},
		{`(output-port? (open-input-string ""))`, false},
		{`(textual-port? (open-input-string ""))`, true},
	}
	for _, tt := range tests {
		r := evalHelper(t, tt.input)
		if r.BoolVal != tt.expect {
			t.Errorf("%s: got %v, want %v", tt.input, r.BoolVal, tt.expect)
		}
	}
}

func TestClosePort(t *testing.T) {
	r := evalHelper(t, `
		(define p (open-input-string "abc"))
		(close-port p)
	`)
	if r.Type != types.TypeVoid {
		t.Error("expected void")
	}
}

func TestEofObject(t *testing.T) {
	r := evalHelper(t, `(eof-object? (eof-object))`)
	if !r.BoolVal {
		t.Error("expected #t")
	}
}

func TestPeekChar(t *testing.T) {
	r := evalHelper(t, `
		(define p (open-input-string "ab"))
		(define c1 (peek-char p))
		(define c2 (read-char p))
		(eqv? c1 c2)
	`)
	if !r.BoolVal {
		t.Error("peek-char and read-char should return same char")
	}
}
