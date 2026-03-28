package evaluator

import (
	"testing"

	"github.com/suchi/sucheme-go/pkg/types"
)

// Reuses setupEval, evalHelper, evalErr from evaluator_test.go

// =========================================================================
// call/cc basics
// =========================================================================

func TestCallCCSimple(t *testing.T) {
	r := evalHelper(t, `(call/cc (lambda (k) 42))`)
	if r.IntVal != 42 {
		t.Errorf("expected 42, got %s", r.String())
	}
}

func TestCallCCEscape(t *testing.T) {
	r := evalHelper(t, `(call/cc (lambda (k) (k 42) 99))`)
	if r.IntVal != 42 {
		t.Errorf("expected 42 (escaped), got %s", r.String())
	}
}

func TestCallCCInExpression(t *testing.T) {
	r := evalHelper(t, `(+ 1 (call/cc (lambda (k) (+ 2 (k 3)))))`)
	if r.IntVal != 4 {
		t.Errorf("expected 4 (1 + 3), got %s", r.String())
	}
}

func TestCallCCNoEscape(t *testing.T) {
	r := evalHelper(t, `(call/cc (lambda (k) (+ 1 2)))`)
	if r.IntVal != 3 {
		t.Errorf("expected 3, got %s", r.String())
	}
}

func TestCallCCWithDefine(t *testing.T) {
	r := evalHelper(t, `
		(define result #f)
		(call/cc (lambda (k)
			(set! result 42)
			(k #t)))
		result
	`)
	if r.IntVal != 42 {
		t.Errorf("expected 42, got %s", r.String())
	}
}

func TestCallCCNestedEscape(t *testing.T) {
	r := evalHelper(t, `
		(call/cc (lambda (outer)
			(call/cc (lambda (inner)
				(outer 42)))))
	`)
	if r.IntVal != 42 {
		t.Errorf("expected 42, got %s", r.String())
	}
}

func TestCallCCEarlyReturn(t *testing.T) {
	// Use call/cc for early return from a loop-like construct
	r := evalHelper(t, `
		(call/cc (lambda (return)
			(let loop ((i 0))
				(if (= i 5)
					(return i)
					(loop (+ i 1))))))
	`)
	if r.IntVal != 5 {
		t.Errorf("expected 5, got %s", r.String())
	}
}

func TestCallCCExpiredContinuation(t *testing.T) {
	// Continuation called after call/cc scope has exited should error, not panic
	err := evalErr(t, `
		(define saved-k #f)
		(call/cc (lambda (k) (set! saved-k k) 42))
		(saved-k 99)
	`)
	if err == nil {
		t.Error("expected error for expired continuation")
	}
}

func TestCallCCWithMultipleValues(t *testing.T) {
	r := evalHelper(t, `
		(call-with-values
			(lambda () (call/cc (lambda (k) (k 1 2 3))))
			(lambda (a b c) (+ a b c)))
	`)
	if r.IntVal != 6 {
		t.Errorf("expected 6, got %s", r.String())
	}
}

func TestCallCCError(t *testing.T) {
	err := evalErr(t, `(call/cc 42)`)
	if err == nil {
		t.Error("expected error for non-procedure argument")
	}
}

// =========================================================================
// call-with-current-continuation (long name)
// =========================================================================

func TestCallWithCurrentContinuation(t *testing.T) {
	r := evalHelper(t, `(call-with-current-continuation (lambda (k) (k 99)))`)
	if r.IntVal != 99 {
		t.Errorf("expected 99, got %s", r.String())
	}
}

// =========================================================================
// dynamic-wind
// =========================================================================

func TestDynamicWindBasic(t *testing.T) {
	r := evalHelper(t, `
		(define log '())
		(dynamic-wind
			(lambda () (set! log (cons 'before log)))
			(lambda () (set! log (cons 'during log)) 42)
			(lambda () (set! log (cons 'after log))))
		(reverse log)
	`)
	lst, _ := r.ToList()
	if len(lst) != 3 {
		t.Errorf("expected 3 entries, got %s", r.String())
	}
	if lst[0].SymbolVal != "before" || lst[1].SymbolVal != "during" || lst[2].SymbolVal != "after" {
		t.Errorf("expected (before during after), got %s", r.String())
	}
}

func TestDynamicWindResult(t *testing.T) {
	r := evalHelper(t, `
		(dynamic-wind
			(lambda () #f)
			(lambda () 42)
			(lambda () #f))
	`)
	if r.IntVal != 42 {
		t.Errorf("expected 42, got %s", r.String())
	}
}

func TestDynamicWindWithEscape(t *testing.T) {
	r := evalHelper(t, `
		(define after-called #f)
		(call/cc (lambda (k)
			(dynamic-wind
				(lambda () #f)
				(lambda () (k 42))
				(lambda () (set! after-called #t)))))
		after-called
	`)
	if !r.BoolVal {
		t.Error("after thunk should be called on escape")
	}
}

func TestDynamicWindError(t *testing.T) {
	err := evalErr(t, `(dynamic-wind 1 2 3)`)
	if err == nil {
		t.Error("expected error for non-procedure arguments")
	}
}

// =========================================================================
// call-with-values
// =========================================================================

func TestCallWithValuesSingle(t *testing.T) {
	r := evalHelper(t, `(call-with-values (lambda () 42) (lambda (x) x))`)
	if r.IntVal != 42 {
		t.Errorf("expected 42, got %s", r.String())
	}
}

func TestCallWithValuesMultiple(t *testing.T) {
	r := evalHelper(t, `(call-with-values (lambda () (values 1 2 3)) (lambda (a b c) (+ a b c)))`)
	if r.IntVal != 6 {
		t.Errorf("expected 6, got %s", r.String())
	}
}

func TestCallWithValuesError(t *testing.T) {
	err := evalErr(t, `(call-with-values 42 (lambda (x) x))`)
	if err == nil {
		t.Error("expected error for non-procedure")
	}
}

// =========================================================================
// Practical uses
// =========================================================================

func TestCallCCFindFirst(t *testing.T) {
	// Use call/cc to implement early exit from a search
	r := evalHelper(t, `
		(define (find-first pred lst)
			(call/cc (lambda (return)
				(for-each (lambda (x)
					(when (pred x) (return x)))
					lst)
				#f)))
		(find-first (lambda (x) (> x 3)) '(1 2 3 4 5))
	`)
	if r.IntVal != 4 {
		t.Errorf("expected 4, got %s", r.String())
	}
}

func TestCallCCFindFirstNotFound(t *testing.T) {
	r := evalHelper(t, `
		(define (find-first pred lst)
			(call/cc (lambda (return)
				(for-each (lambda (x)
					(when (pred x) (return x)))
					lst)
				#f)))
		(find-first (lambda (x) (> x 10)) '(1 2 3))
	`)
	if r.Type != types.TypeBoolean || r.BoolVal {
		t.Errorf("expected #f, got %s", r.String())
	}
}

func TestCallCCZeroValues(t *testing.T) {
	// Continuation called with 0 args should produce empty values
	r := evalHelper(t, `
		(call-with-values
			(lambda () (call/cc (lambda (k) (k))))
			(lambda () 42))
	`)
	if r.IntVal != 42 {
		t.Errorf("expected 42, got %s", r.String())
	}
}

func TestValuesBasic(t *testing.T) {
	r := evalHelper(t, `(values 1)`)
	if r.IntVal != 1 {
		t.Errorf("expected 1, got %s", r.String())
	}
}

func TestValuesMultiple(t *testing.T) {
	r := evalHelper(t, `(values 1 2 3)`)
	if r.Type != types.TypeValues || len(r.ValuesVal) != 3 {
		t.Errorf("expected 3 values, got %s", r.String())
	}
}
