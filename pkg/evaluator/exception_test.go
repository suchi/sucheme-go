package evaluator

import (
	"testing"

	"github.com/suchi/sucheme-go/pkg/types"
)

// Reuses setupEval, evalHelper, evalErr from evaluator_test.go

// =========================================================================
// raise and with-exception-handler
// =========================================================================

func TestRaiseContinuableWithHandler(t *testing.T) {
	r := evalHelper(t, `
		(with-exception-handler
			(lambda (e) (string-append "caught: " e))
			(lambda () (raise-continuable "test")))
	`)
	if r.StrVal != "caught: test" {
		t.Errorf("expected 'caught: test', got %s", r.String())
	}
}

func TestRaiseContinuableNumber(t *testing.T) {
	r := evalHelper(t, `
		(with-exception-handler
			(lambda (e) (+ e 100))
			(lambda () (raise-continuable 42)))
	`)
	if r.IntVal != 142 {
		t.Errorf("expected 142, got %s", r.String())
	}
}

func TestNoException(t *testing.T) {
	r := evalHelper(t, `
		(with-exception-handler
			(lambda (e) "error")
			(lambda () 42))
	`)
	if r.IntVal != 42 {
		t.Errorf("expected 42, got %s", r.String())
	}
}

func TestRaiseUnhandled(t *testing.T) {
	// raise without handler should return an error (not panic)
	err := evalErr(t, `(raise "boom")`)
	if err == nil {
		t.Error("expected error from unhandled raise")
	}
}

// =========================================================================
// guard
// =========================================================================

func TestGuardBasic(t *testing.T) {
	r := evalHelper(t, `
		(guard (exn
			((string? exn) (string-append "got: " exn)))
			(raise "hello"))
	`)
	if r.StrVal != "got: hello" {
		t.Errorf("expected 'got: hello', got %s", r.String())
	}
}

func TestGuardNoException(t *testing.T) {
	r := evalHelper(t, `
		(guard (exn
			((string? exn) "error"))
			42)
	`)
	if r.IntVal != 42 {
		t.Errorf("expected 42, got %s", r.String())
	}
}

func TestGuardElse(t *testing.T) {
	r := evalHelper(t, `
		(guard (exn
			(else "default"))
			(raise 99))
	`)
	if r.StrVal != "default" {
		t.Errorf("expected 'default', got %s", r.String())
	}
}

func TestGuardMultipleClauses(t *testing.T) {
	r := evalHelper(t, `
		(guard (exn
			((number? exn) (+ exn 1))
			((string? exn) exn))
			(raise 10))
	`)
	if r.IntVal != 11 {
		t.Errorf("expected 11, got %s", r.String())
	}
}

func TestGuardArrowSyntax(t *testing.T) {
	// => receives the test result (#t from number?), not the exception value
	r := evalHelper(t, `
		(guard (exn
			((number? exn) => (lambda (x) (if x "yes" "no"))))
			(raise 5))
	`)
	if r.StrVal != "yes" {
		t.Errorf("expected 'yes', got %s", r.String())
	}
}

func TestGuardGoError(t *testing.T) {
	// Go error (e.g. from unbound variable) should be caught by guard
	r := evalHelper(t, `
		(guard (exn
			(else "caught"))
			(+ 1 "a"))
	`)
	if r.StrVal != "caught" {
		t.Errorf("expected 'caught', got %s", r.String())
	}
}

// =========================================================================
// error procedure
// =========================================================================

func TestErrorProcedure(t *testing.T) {
	r := evalHelper(t, `
		(guard (exn
			((error-object? exn) (error-object-message exn)))
			(error "test message" 1 2))
	`)
	if r.StrVal != "test message" {
		t.Errorf("expected 'test message', got %s", r.String())
	}
}

func TestErrorObjectIrritants(t *testing.T) {
	r := evalHelper(t, `
		(guard (exn
			((error-object? exn) (error-object-irritants exn)))
			(error "msg" 1 2 3))
	`)
	lst, _ := r.ToList()
	if len(lst) != 3 || lst[0].IntVal != 1 {
		t.Errorf("expected (1 2 3), got %s", r.String())
	}
}

// =========================================================================
// case-lambda
// =========================================================================

func TestCaseLambdaZeroArgs(t *testing.T) {
	r := evalHelper(t, `
		(define f (case-lambda
			(() 0)
			((x) x)))
		(f)
	`)
	if r.IntVal != 0 {
		t.Errorf("expected 0, got %s", r.String())
	}
}

func TestCaseLambdaOneArg(t *testing.T) {
	r := evalHelper(t, `
		(define f (case-lambda
			(() 0)
			((x) (* x 10))))
		(f 5)
	`)
	if r.IntVal != 50 {
		t.Errorf("expected 50, got %s", r.String())
	}
}

func TestCaseLambdaTwoArgs(t *testing.T) {
	r := evalHelper(t, `
		(define f (case-lambda
			(() 0)
			((x) x)
			((x y) (+ x y))))
		(f 3 4)
	`)
	if r.IntVal != 7 {
		t.Errorf("expected 7, got %s", r.String())
	}
}

func TestCaseLambdaRest(t *testing.T) {
	r := evalHelper(t, `
		(define f (case-lambda
			(() 0)
			(args (length args))))
		(f 1 2 3)
	`)
	if r.IntVal != 3 {
		t.Errorf("expected 3, got %s", r.String())
	}
}

func TestCaseLambdaNoMatch(t *testing.T) {
	err := evalErr(t, `
		(define f (case-lambda
			((x) x)
			((x y) (+ x y))))
		(f)
	`)
	if err == nil {
		t.Error("expected error: no matching clause")
	}
}

// =========================================================================
// map, for-each, filter with lambdas
// =========================================================================

func TestMapWithLambda(t *testing.T) {
	r := evalHelper(t, `(map (lambda (x) (* x x)) '(1 2 3 4 5))`)
	lst, _ := r.ToList()
	if len(lst) != 5 || lst[0].IntVal != 1 || lst[4].IntVal != 25 {
		t.Errorf("expected (1 4 9 16 25), got %s", r.String())
	}
}

func TestMapWithBuiltin(t *testing.T) {
	r := evalHelper(t, `(map car '((1 2) (3 4) (5 6)))`)
	lst, _ := r.ToList()
	if len(lst) != 3 || lst[0].IntVal != 1 {
		t.Errorf("expected (1 3 5), got %s", r.String())
	}
}

func TestForEach(t *testing.T) {
	r := evalHelper(t, `
		(define sum 0)
		(for-each (lambda (x) (set! sum (+ sum x))) '(1 2 3))
		sum
	`)
	if r.IntVal != 6 {
		t.Errorf("expected 6, got %d", r.IntVal)
	}
}

func TestFilterWithLambda(t *testing.T) {
	r := evalHelper(t, `(filter (lambda (x) (> x 3)) '(1 2 3 4 5))`)
	lst, _ := r.ToList()
	if len(lst) != 2 || lst[0].IntVal != 4 {
		t.Errorf("expected (4 5), got %s", r.String())
	}
}

func TestFilterEmpty(t *testing.T) {
	r := evalHelper(t, `(filter (lambda (x) #f) '(1 2 3))`)
	if r.Type != types.TypeNil {
		t.Errorf("expected (), got %s", r.String())
	}
}

func TestMapEmpty(t *testing.T) {
	r := evalHelper(t, `(map (lambda (x) x) '())`)
	if r.Type != types.TypeNil {
		t.Errorf("expected (), got %s", r.String())
	}
}
