package evaluator

import (
	"fmt"
	"os"
	"testing"

	"github.com/suchi/sucheme-go/pkg/builtin"
	"github.com/suchi/sucheme-go/pkg/env"
	"github.com/suchi/sucheme-go/pkg/types"
)

func setupEval() *env.Env {
	e := env.Global()
	builtin.RegisterBuiltins(e)
	RegisterApply(e)
	RegisterExceptionBuiltins(e)
	RegisterMapForEach(e)
	RegisterParameterBuiltins(e)
	RegisterPortBuiltins(e)
	RegisterContinuationBuiltins(e)
	RegisterLibraryBuiltins(e)
	return e
}

func evalHelper(t *testing.T, input string) *types.Value {
	t.Helper()
	e := setupEval()
	result, err := EvalString(input, e)
	if err != nil {
		t.Fatalf("unexpected error evaluating %q: %v", input, err)
	}
	return result
}

func evalWithEnv(t *testing.T, input string, e *env.Env) *types.Value {
	t.Helper()
	result, err := EvalString(input, e)
	if err != nil {
		t.Fatalf("unexpected error evaluating %q: %v", input, err)
	}
	return result
}

func evalErr(t *testing.T, input string) error {
	t.Helper()
	e := setupEval()
	_, err := EvalString(input, e)
	return err
}

// =========================================================================
// Self-evaluating forms
// =========================================================================

func TestSelfEvaluating(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"42", "42"},
		{"-7", "-7"},
		{"0", "0"},
		{"3.14", "3.14"},
		{"#t", "#t"},
		{"#f", "#f"},
		{`"hello"`, `"hello"`},
		{`""`, `""`},
		{`#\a`, `#\a`},
		{`#\space`, `#\space`},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			r := evalHelper(t, tt.input)
			if r.String() != tt.expect {
				t.Errorf("got %s, want %s", r.String(), tt.expect)
			}
		})
	}
}

// =========================================================================
// Quote
// =========================================================================

func TestQuote(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"'a", "a"},
		{"'42", "42"},
		{"'(1 2 3)", "(1 2 3)"},
		{"'()", "()"},
		{"'(a . b)", "(a . b)"},
		{"'(1 (2 3) 4)", "(1 (2 3) 4)"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			r := evalHelper(t, tt.input)
			if r.String() != tt.expect {
				t.Errorf("got %s, want %s", r.String(), tt.expect)
			}
		})
	}
}

// =========================================================================
// If
// =========================================================================

func TestIf(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"(if #t 1 2)", "1"},
		{"(if #f 1 2)", "2"},
		{"(if #f 1)", "#<void>"},
		// 0 is truthy in Scheme
		{"(if 0 1 2)", "1"},
		// Empty list is truthy in Scheme
		{"(if '() 1 2)", "1"},
		// String is truthy
		{`(if "" 1 2)`, "1"},
		// Nested if
		{"(if #t (if #f 1 2) 3)", "2"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			r := evalHelper(t, tt.input)
			if r.String() != tt.expect {
				t.Errorf("got %s, want %s", r.String(), tt.expect)
			}
		})
	}
}

// =========================================================================
// Define and set!
// =========================================================================

func TestDefine(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, "(define x 42)", e)
	r := evalWithEnv(t, "x", e)
	if r.IntVal != 42 {
		t.Errorf("expected 42, got %d", r.IntVal)
	}
	// Redefine
	evalWithEnv(t, "(define x 100)", e)
	r = evalWithEnv(t, "x", e)
	if r.IntVal != 100 {
		t.Errorf("expected 100, got %d", r.IntVal)
	}
}

func TestDefineFunction(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, "(define (square x) (* x x))", e)
	r := evalWithEnv(t, "(square 5)", e)
	if r.IntVal != 25 {
		t.Errorf("expected 25, got %d", r.IntVal)
	}
}

func TestSet(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, "(define x 1)", e)
	evalWithEnv(t, "(set! x 2)", e)
	r := evalWithEnv(t, "x", e)
	if r.IntVal != 2 {
		t.Errorf("expected 2, got %d", r.IntVal)
	}
}

func TestSetUnbound(t *testing.T) {
	err := evalErr(t, "(set! nonexistent 1)")
	if err == nil {
		t.Error("expected error for set! on unbound variable")
	}
}

// =========================================================================
// Lambda
// =========================================================================

func TestLambdaBasic(t *testing.T) {
	r := evalHelper(t, "((lambda (x) x) 42)")
	if r.IntVal != 42 {
		t.Errorf("expected 42, got %d", r.IntVal)
	}
}

func TestLambdaMultiParam(t *testing.T) {
	r := evalHelper(t, "((lambda (x y) (+ x y)) 3 4)")
	if r.IntVal != 7 {
		t.Errorf("expected 7, got %d", r.IntVal)
	}
}

func TestLambdaClosure(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, "(define make-adder (lambda (n) (lambda (x) (+ n x))))", e)
	evalWithEnv(t, "(define add5 (make-adder 5))", e)
	r := evalWithEnv(t, "(add5 10)", e)
	if r.IntVal != 15 {
		t.Errorf("expected 15, got %d", r.IntVal)
	}
}

func TestLambdaVariadic(t *testing.T) {
	r := evalHelper(t, "((lambda args args) 1 2 3)")
	lst, _ := r.ToList()
	if len(lst) != 3 {
		t.Errorf("expected list of 3, got %d", len(lst))
	}
}

func TestLambdaDottedRest(t *testing.T) {
	r := evalHelper(t, "((lambda (x . rest) rest) 1 2 3)")
	lst, _ := r.ToList()
	if len(lst) != 2 || lst[0].IntVal != 2 || lst[1].IntVal != 3 {
		t.Errorf("expected (2 3), got %v", r)
	}
}

func TestLambdaWrongArity(t *testing.T) {
	err := evalErr(t, "((lambda (x y) x) 1)")
	if err == nil {
		t.Error("expected arity error")
	}
}

// =========================================================================
// Let / Let* / Letrec
// =========================================================================

func TestLet(t *testing.T) {
	r := evalHelper(t, "(let ((x 1) (y 2)) (+ x y))")
	if r.IntVal != 3 {
		t.Errorf("expected 3, got %d", r.IntVal)
	}
}

func TestLetStar(t *testing.T) {
	r := evalHelper(t, "(let* ((x 1) (y (+ x 1))) y)")
	if r.IntVal != 2 {
		t.Errorf("expected 2, got %d", r.IntVal)
	}
}

func TestLetrec(t *testing.T) {
	r := evalHelper(t, `
		(letrec ((fact (lambda (n) (if (= n 0) 1 (* n (fact (- n 1)))))))
			(fact 5))
	`)
	if r.IntVal != 120 {
		t.Errorf("expected 120, got %d", r.IntVal)
	}
}

func TestLetrecUninitialized(t *testing.T) {
	// Referencing a later binding during init should be an error (y is still undefined when x's init runs)
	err := evalErr(t, "(letrec ((x y) (y 1)) x)")
	if err == nil {
		t.Error("expected error for uninitialized letrec binding reference")
	}
}

func TestNamedLet(t *testing.T) {
	r := evalHelper(t, "(let loop ((i 0) (sum 0)) (if (= i 5) sum (loop (+ i 1) (+ sum i))))")
	if r.IntVal != 10 {
		t.Errorf("expected 10, got %d", r.IntVal)
	}
}

// =========================================================================
// Begin
// =========================================================================

func TestBegin(t *testing.T) {
	r := evalHelper(t, "(begin 1 2 3)")
	if r.IntVal != 3 {
		t.Errorf("expected 3, got %d", r.IntVal)
	}
}

func TestBeginWithSideEffects(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, "(define x 0)", e)
	evalWithEnv(t, "(begin (set! x 1) (set! x 2))", e)
	r := evalWithEnv(t, "x", e)
	if r.IntVal != 2 {
		t.Errorf("expected 2, got %d", r.IntVal)
	}
}

// =========================================================================
// Cond
// =========================================================================

func TestCond(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"(cond (#t 1))", "1"},
		{"(cond (#f 1) (#t 2))", "2"},
		{"(cond (#f 1) (else 3))", "3"},
		// => syntax
		{"(cond ((+ 1 1) => (lambda (x) (* x 10))))", "20"},
		// No match returns void
		{"(cond (#f 1))", "#<void>"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			r := evalHelper(t, tt.input)
			if r.String() != tt.expect {
				t.Errorf("got %s, want %s", r.String(), tt.expect)
			}
		})
	}
}

func TestCondElseNotLast(t *testing.T) {
	err := evalErr(t, "(cond (else 1) (#t 2))")
	if err == nil {
		t.Error("expected error: else must be last")
	}
}

func TestCondElseEmpty(t *testing.T) {
	err := evalErr(t, "(cond (else))")
	if err == nil {
		t.Error("expected error: else must have body")
	}
}

// =========================================================================
// And / Or
// =========================================================================

func TestAnd(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"(and)", "#t"},
		{"(and 1)", "1"},
		{"(and 1 2 3)", "3"},
		{"(and 1 #f 3)", "#f"},
		{"(and #f)", "#f"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			r := evalHelper(t, tt.input)
			if r.String() != tt.expect {
				t.Errorf("got %s, want %s", r.String(), tt.expect)
			}
		})
	}
}

func TestOr(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"(or)", "#f"},
		{"(or 1)", "1"},
		{"(or #f 2 3)", "2"},
		{"(or #f #f #f)", "#f"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			r := evalHelper(t, tt.input)
			if r.String() != tt.expect {
				t.Errorf("got %s, want %s", r.String(), tt.expect)
			}
		})
	}
}

// =========================================================================
// When / Unless
// =========================================================================

func TestWhen(t *testing.T) {
	r := evalHelper(t, "(when #t 1 2 3)")
	if r.IntVal != 3 {
		t.Errorf("expected 3, got %d", r.IntVal)
	}
	r = evalHelper(t, "(when #f 1 2 3)")
	if r.Type != types.TypeVoid {
		t.Errorf("expected void, got %s", r.String())
	}
}

func TestUnless(t *testing.T) {
	r := evalHelper(t, "(unless #f 1 2 3)")
	if r.IntVal != 3 {
		t.Errorf("expected 3, got %d", r.IntVal)
	}
	r = evalHelper(t, "(unless #t 1 2 3)")
	if r.Type != types.TypeVoid {
		t.Errorf("expected void, got %s", r.String())
	}
}

// =========================================================================
// Do
// =========================================================================

func TestDo(t *testing.T) {
	r := evalHelper(t, "(do ((i 0 (+ i 1))) ((= i 10) i))")
	if r.IntVal != 10 {
		t.Errorf("expected 10, got %d", r.IntVal)
	}
}

func TestDoWithBody(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, "(define sum 0)", e)
	evalWithEnv(t, "(do ((i 0 (+ i 1))) ((= i 5)) (set! sum (+ sum i)))", e)
	r := evalWithEnv(t, "sum", e)
	if r.IntVal != 10 {
		t.Errorf("expected 10, got %d", r.IntVal)
	}
}

// =========================================================================
// Quasiquote
// =========================================================================

func TestQuasiquote(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"`(1 2 3)", "(1 2 3)"},
		{"`(1 ,(+ 1 1) 3)", "(1 2 3)"},
		{"`(1 ,@(list 2 3) 4)", "(1 2 3 4)"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			r := evalHelper(t, tt.input)
			if r.String() != tt.expect {
				t.Errorf("got %s, want %s", r.String(), tt.expect)
			}
		})
	}
}

// =========================================================================
// TCO (Tail Call Optimization)
// =========================================================================

func TestTCODeepRecursion(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, `(define (loop n) (if (= n 0) "done" (loop (- n 1))))`, e)
	r := evalWithEnv(t, "(loop 100000)", e)
	if r.StrVal != "done" {
		t.Errorf("expected \"done\", got %s", r.String())
	}
}

func TestTCOMutualRecursion(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, `(define (my-even? n) (if (= n 0) #t (my-odd? (- n 1))))`, e)
	evalWithEnv(t, `(define (my-odd? n) (if (= n 0) #f (my-even? (- n 1))))`, e)
	r := evalWithEnv(t, "(my-even? 10000)", e)
	if !r.BoolVal {
		t.Error("expected #t for (my-even? 10000)")
	}
}

// =========================================================================
// Error handling
// =========================================================================

func TestUnboundVariable(t *testing.T) {
	err := evalErr(t, "nonexistent")
	if err == nil {
		t.Error("expected unbound variable error")
	}
}

func TestNotAProcedure(t *testing.T) {
	err := evalErr(t, "(42 1 2)")
	if err == nil {
		t.Error("expected not-a-procedure error")
	}
}

func TestComparisonTypeError(t *testing.T) {
	err := evalErr(t, `(< 1 "a")`)
	if err == nil {
		t.Error("expected type error for (< 1 \"a\")")
	}
}

// =========================================================================
// Integration tests
// =========================================================================

func TestFactorial(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, "(define (fact n) (if (= n 0) 1 (* n (fact (- n 1)))))", e)
	r := evalWithEnv(t, "(fact 10)", e)
	if r.IntVal != 3628800 {
		t.Errorf("expected 3628800, got %d", r.IntVal)
	}
}

func TestFibonacci(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, `(define (fib n) (if (< n 2) n (+ (fib (- n 1)) (fib (- n 2)))))`, e)
	r := evalWithEnv(t, "(fib 10)", e)
	if r.IntVal != 55 {
		t.Errorf("expected 55, got %d", r.IntVal)
	}
}

func TestHigherOrderFunction(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, "(define (twice f x) (f (f x)))", e)
	evalWithEnv(t, "(define (inc x) (+ x 1))", e)
	r := evalWithEnv(t, "(twice inc 5)", e)
	if r.IntVal != 7 {
		t.Errorf("expected 7, got %d", r.IntVal)
	}
}

func TestAccumulator(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, `(define (make-counter)
		(let ((n 0))
			(lambda () (set! n (+ n 1)) n)))`, e)
	evalWithEnv(t, "(define c (make-counter))", e)
	evalWithEnv(t, "(c)", e)
	evalWithEnv(t, "(c)", e)
	r := evalWithEnv(t, "(c)", e)
	if r.IntVal != 3 {
		t.Errorf("expected 3, got %d", r.IntVal)
	}
}

func TestTailRecursiveFactorial(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, `(define (fact-iter n acc)
		(if (= n 0) acc (fact-iter (- n 1) (* n acc))))`, e)
	r := evalWithEnv(t, "(fact-iter 20 1)", e)
	if r.IntVal != 2432902008176640000 {
		t.Errorf("expected 2432902008176640000, got %d", r.IntVal)
	}
}

// --- R7RS special forms tests ---

func TestDefineRecordType(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, `(define-record-type <point>
		(make-point x y)
		point?
		(x point-x)
		(y point-y set-point-y!))`, e)

	// Constructor and accessors
	evalWithEnv(t, "(define p (make-point 3 4))", e)
	r := evalWithEnv(t, "(point-x p)", e)
	if r.IntVal != 3 {
		t.Errorf("expected x=3, got %d", r.IntVal)
	}
	r = evalWithEnv(t, "(point-y p)", e)
	if r.IntVal != 4 {
		t.Errorf("expected y=4, got %d", r.IntVal)
	}

	// Predicate
	r = evalWithEnv(t, "(point? p)", e)
	if !r.BoolVal {
		t.Error("expected point? to be #t")
	}
	r = evalWithEnv(t, "(point? 42)", e)
	if r.BoolVal {
		t.Error("expected point? on integer to be #f")
	}

	// Mutator
	evalWithEnv(t, "(set-point-y! p 10)", e)
	r = evalWithEnv(t, "(point-y p)", e)
	if r.IntVal != 10 {
		t.Errorf("expected y=10 after mutation, got %d", r.IntVal)
	}
}

func TestDefineRecordTypeDistinctTypes(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, `(define-record-type <a> (make-a x) a? (x a-x))`, e)
	evalWithEnv(t, `(define-record-type <a> (make-a2 x) a2? (x a2-x))`, e)
	evalWithEnv(t, "(define v1 (make-a 1))", e)
	evalWithEnv(t, "(define v2 (make-a2 2))", e)
	r := evalWithEnv(t, "(a? v1)", e)
	if !r.BoolVal {
		t.Error("expected a? v1 = #t")
	}
	r = evalWithEnv(t, "(a? v2)", e)
	if r.BoolVal {
		t.Error("expected a? v2 = #f (different type)")
	}
}

func TestDefineRecordTypeUninitializedField(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, `(define-record-type <pair>
		(make-pair a)
		pair-rec?
		(a pair-a)
		(b pair-b set-pair-b!))`, e)
	evalWithEnv(t, "(define p (make-pair 1))", e)
	// b should be undefined
	r := evalWithEnv(t, "(pair-b p)", e)
	if r.Type != types.TypeUndefined {
		t.Errorf("expected undefined for uninitialized field, got %s", r.Type)
	}
}

func TestCase(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`(case 1 ((1) 10) ((2) 20))`, 10},
		{`(case 2 ((1) 10) ((2) 20))`, 20},
		{`(case 3 ((1) 10) ((2) 20) (else 30))`, 30},
		{`(case 1 ((1 2 3) 100) (else 0))`, 100},
		{`(case 2 ((1 2 3) 100) (else 0))`, 100},
	}
	for _, tt := range tests {
		r := evalHelper(t, tt.input)
		if r.IntVal != tt.expected {
			t.Errorf("%s: expected %d, got %d", tt.input, tt.expected, r.IntVal)
		}
	}
}

func TestCaseArrow(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, `(define (double x) (* x 2))`, e)
	r := evalWithEnv(t, `(case 5 ((5) => double) (else 0))`, e)
	if r.IntVal != 10 {
		t.Errorf("expected 10 from => syntax, got %d", r.IntVal)
	}
}

func TestCondExpand(t *testing.T) {
	r := evalHelper(t, `(cond-expand (r7rs 1) (else 2))`)
	if r.IntVal != 1 {
		t.Errorf("expected 1 for r7rs feature, got %d", r.IntVal)
	}
	r = evalHelper(t, `(cond-expand (sucheme 42) (else 0))`)
	if r.IntVal != 42 {
		t.Errorf("expected 42 for sucheme feature, got %d", r.IntVal)
	}
	r = evalHelper(t, `(cond-expand (nonexistent 1) (else 99))`)
	if r.IntVal != 99 {
		t.Errorf("expected 99 for else, got %d", r.IntVal)
	}
}

func TestCondExpandAnd(t *testing.T) {
	r := evalHelper(t, `(cond-expand ((and r7rs sucheme) 1) (else 0))`)
	if r.IntVal != 1 {
		t.Errorf("expected 1 for (and r7rs sucheme), got %d", r.IntVal)
	}
	r = evalHelper(t, `(cond-expand ((and r7rs nonexistent) 1) (else 0))`)
	if r.IntVal != 0 {
		t.Errorf("expected 0 for (and r7rs nonexistent), got %d", r.IntVal)
	}
}

func TestCondExpandOr(t *testing.T) {
	r := evalHelper(t, `(cond-expand ((or nonexistent sucheme) 1) (else 0))`)
	if r.IntVal != 1 {
		t.Errorf("expected 1 for (or nonexistent sucheme), got %d", r.IntVal)
	}
}

func TestCondExpandNot(t *testing.T) {
	r := evalHelper(t, `(cond-expand ((not nonexistent) 1) (else 0))`)
	if r.IntVal != 1 {
		t.Errorf("expected 1 for (not nonexistent), got %d", r.IntVal)
	}
}

func TestDefineValues(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, `(define-values (a b c) (values 1 2 3))`, e)
	r := evalWithEnv(t, "a", e)
	if r.IntVal != 1 {
		t.Errorf("expected a=1, got %d", r.IntVal)
	}
	r = evalWithEnv(t, "b", e)
	if r.IntVal != 2 {
		t.Errorf("expected b=2, got %d", r.IntVal)
	}
	r = evalWithEnv(t, "c", e)
	if r.IntVal != 3 {
		t.Errorf("expected c=3, got %d", r.IntVal)
	}
}

func TestDefineValuesSingle(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, `(define-values (x) 42)`, e)
	r := evalWithEnv(t, "x", e)
	if r.IntVal != 42 {
		t.Errorf("expected x=42, got %d", r.IntVal)
	}
}

func TestDefineValuesRest(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, `(define-values rest (values 1 2 3))`, e)
	r := evalWithEnv(t, "(car rest)", e)
	if r.IntVal != 1 {
		t.Errorf("expected (car rest)=1, got %d", r.IntVal)
	}
}

func TestAssocCustomComparator(t *testing.T) {
	e := setupEval()
	r := evalWithEnv(t, `(assoc 2.0 '((1 . a) (2 . b) (3 . c)) =)`, e)
	if r.Type != types.TypePair {
		t.Errorf("expected pair, got %s", r.Type)
	}
}

func TestMemberCustomComparator(t *testing.T) {
	e := setupEval()
	r := evalWithEnv(t, `(member 2.0 '(1 2 3) =)`, e)
	if r.Type != types.TypePair {
		t.Errorf("expected pair, got %s", r.Type)
	}
}

func TestStringSet(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, `(define s (string-copy "hello"))`, e)
	evalWithEnv(t, `(string-set! s 0 #\H)`, e)
	r := evalWithEnv(t, "s", e)
	if r.StrVal != "Hello" {
		t.Errorf("expected Hello, got %s", r.StrVal)
	}
}

// =========================================================================
// Port and I/O tests
// =========================================================================

func TestCurrentInputPortReturnsPort(t *testing.T) {
	e := setupEval()
	r := evalWithEnv(t, "(current-input-port)", e)
	if r.Type != types.TypePort {
		t.Errorf("expected port type, got %s", r.Type)
	}
}

func TestCurrentOutputPortReturnsPort(t *testing.T) {
	e := setupEval()
	r := evalWithEnv(t, "(current-output-port)", e)
	if r.Type != types.TypePort {
		t.Errorf("expected port type, got %s", r.Type)
	}
}

func TestCurrentInputPortArityError(t *testing.T) {
	err := evalErr(t, "(current-input-port 1)")
	if err == nil {
		t.Errorf("expected error for (current-input-port 1)")
	}
}

func TestCurrentOutputPortArityError(t *testing.T) {
	err := evalErr(t, "(current-output-port 1)")
	if err == nil {
		t.Errorf("expected error for (current-output-port 1)")
	}
}

func TestReadFromStringPort(t *testing.T) {
	e := setupEval()
	r := evalWithEnv(t, `(read (open-input-string "42"))`, e)
	if r.Type != types.TypeInteger || r.IntVal != 42 {
		t.Errorf("expected 42, got %v", r)
	}
}

func TestReadFromStringPortSymbol(t *testing.T) {
	e := setupEval()
	r := evalWithEnv(t, `(read (open-input-string "hello"))`, e)
	if r.Type != types.TypeSymbol || r.SymbolVal != "hello" {
		t.Errorf("expected symbol hello, got %v", r)
	}
}

func TestReadFromStringPortEOF(t *testing.T) {
	e := setupEval()
	r := evalWithEnv(t, `(read (open-input-string ""))`, e)
	if r.Type != types.TypeEof {
		t.Errorf("expected eof, got %s", r.Type)
	}
}

func TestReadNoArgReturnsEof(t *testing.T) {
	// (read) with no args uses current-input-port which is an empty stdin port
	e := setupEval()
	r := evalWithEnv(t, "(read)", e)
	if r.Type != types.TypeEof {
		t.Errorf("expected eof from (read) on empty stdin, got %s", r.Type)
	}
}

func TestReadArityError(t *testing.T) {
	err := evalErr(t, `(read (open-input-string "a") (open-input-string "b"))`)
	if err == nil {
		t.Errorf("expected error for (read) with too many arguments")
	}
}

func TestCurrentInputPortIsPort(t *testing.T) {
	e := setupEval()
	r := evalWithEnv(t, "(port? (current-input-port))", e)
	if r.Type != types.TypeBoolean || !r.BoolVal {
		t.Errorf("expected (port? (current-input-port)) to be #t, got %v", r)
	}
}

func TestCurrentOutputPortIsPort(t *testing.T) {
	e := setupEval()
	r := evalWithEnv(t, "(port? (current-output-port))", e)
	if r.Type != types.TypeBoolean || !r.BoolVal {
		t.Errorf("expected (port? (current-output-port)) to be #t, got %v", r)
	}
}

func TestReadFromStringPortList(t *testing.T) {
	e := setupEval()
	r := evalWithEnv(t, `(read (open-input-string "(1 2 3)"))`, e)
	if r.Type != types.TypePair {
		t.Errorf("expected pair, got %s", r.Type)
	}
}

// =========================================================================
// include / include-ci tests
// =========================================================================

func TestInclude(t *testing.T) {
	// Create a temp file with Scheme code
	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/test-include.scm"
	if err := os.WriteFile(tmpFile, []byte("(define include-test-var 42)"), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	e := setupEval()
	evalWithEnv(t, fmt.Sprintf(`(include %q)`, tmpFile), e)
	r := evalWithEnv(t, "include-test-var", e)
	if r.Type != types.TypeInteger || r.IntVal != 42 {
		t.Errorf("expected 42, got %v", r)
	}
}

func TestIncludeMultipleExpressions(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/test-multi.scm"
	if err := os.WriteFile(tmpFile, []byte("(define a 10) (define b 20)"), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	e := setupEval()
	evalWithEnv(t, fmt.Sprintf(`(include %q)`, tmpFile), e)
	r := evalWithEnv(t, "(+ a b)", e)
	if r.Type != types.TypeInteger || r.IntVal != 30 {
		t.Errorf("expected 30, got %v", r)
	}
}

func TestIncludeFileNotFound(t *testing.T) {
	err := evalErr(t, `(include "/nonexistent/file.scm")`)
	if err == nil {
		t.Errorf("expected error for include of nonexistent file")
	}
}

func TestIncludeCi(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/test-ci.scm"
	// include-ci should fold identifiers to lowercase
	if err := os.WriteFile(tmpFile, []byte("(define CI-TEST-VAR 99)"), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	e := setupEval()
	evalWithEnv(t, fmt.Sprintf(`(include-ci %q)`, tmpFile), e)
	// After case-folding, CI-TEST-VAR becomes ci-test-var
	r := evalWithEnv(t, "ci-test-var", e)
	if r.Type != types.TypeInteger || r.IntVal != 99 {
		t.Errorf("expected 99, got %v", r)
	}
}

func TestDefineRecordTypeDuplicateCtorField(t *testing.T) {
	err := evalErr(t, `(define-record-type <dup> (make-dup x x) dup? (x dup-x))`)
	if err == nil {
		t.Errorf("expected error for duplicate constructor field name")
	}
}
