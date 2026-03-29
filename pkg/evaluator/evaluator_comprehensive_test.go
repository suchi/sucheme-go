package evaluator

import (
	"fmt"
	"testing"

	"github.com/suchi/sucheme-go/pkg/types"
)

// =========================================================================
// letrec*
// =========================================================================

func TestLetrecStar(t *testing.T) {
	r := evalHelper(t, `(letrec* ((x 1) (y (+ x 1))) y)`)
	if r.IntVal != 2 {
		t.Errorf("expected 2, got %s", r.String())
	}
}

func TestLetrecStarRecursive(t *testing.T) {
	r := evalHelper(t, `
		(letrec* ((fact (lambda (n) (if (= n 0) 1 (* n (fact (- n 1)))))))
			(fact 5))
	`)
	if r.IntVal != 120 {
		t.Errorf("expected 120, got %s", r.String())
	}
}

func TestLetrecStarMutualReference(t *testing.T) {
	r := evalHelper(t, `
		(letrec* ((x 1) (y x) (z (+ x y)))
			z)
	`)
	if r.IntVal != 2 {
		t.Errorf("expected 2, got %s", r.String())
	}
}

func TestLetrecStarEmpty(t *testing.T) {
	r := evalHelper(t, `(letrec* () 42)`)
	if r.IntVal != 42 {
		t.Errorf("expected 42, got %s", r.String())
	}
}

func TestLetrecStarInvalidSyntax(t *testing.T) {
	err := evalErr(t, `(letrec*)`)
	if err == nil {
		t.Error("expected error for invalid letrec* syntax")
	}
}

func TestLetrecStarInvalidBinding(t *testing.T) {
	err := evalErr(t, `(letrec* ((42 1)) 1)`)
	if err == nil {
		t.Error("expected error for non-symbol binding name")
	}
}

// =========================================================================
// apply
// =========================================================================

func TestApplyBasic(t *testing.T) {
	r := evalHelper(t, `(apply + '(1 2 3))`)
	if r.IntVal != 6 {
		t.Errorf("expected 6, got %s", r.String())
	}
}

func TestApplyWithLeadingArgs(t *testing.T) {
	r := evalHelper(t, `(apply + 1 2 '(3 4))`)
	if r.IntVal != 10 {
		t.Errorf("expected 10, got %s", r.String())
	}
}

func TestApplyLambda(t *testing.T) {
	r := evalHelper(t, `(apply (lambda (x y) (+ x y)) '(3 4))`)
	if r.IntVal != 7 {
		t.Errorf("expected 7, got %s", r.String())
	}
}

func TestApplyEmptyList(t *testing.T) {
	r := evalHelper(t, `(apply + '())`)
	if r.IntVal != 0 {
		t.Errorf("expected 0, got %s", r.String())
	}
}

func TestApplyNotProcedure(t *testing.T) {
	err := evalErr(t, `(apply 42 '(1 2))`)
	if err == nil {
		t.Error("expected error for apply with non-procedure")
	}
}

func TestApplyNotList(t *testing.T) {
	err := evalErr(t, `(apply + 42)`)
	if err == nil {
		t.Error("expected error for apply with non-list last arg")
	}
}

func TestApplyTooFewArgs(t *testing.T) {
	err := evalErr(t, `(apply +)`)
	if err == nil {
		t.Error("expected error for apply with too few args")
	}
}

// =========================================================================
// EvalAll
// =========================================================================

func TestEvalAllBasic(t *testing.T) {
	e := setupEval()
	exprs := []*types.Value{types.NewInteger(1), types.NewInteger(2), types.NewInteger(3)}
	result, err := EvalAll(exprs, e)
	if err != nil {
		t.Fatal(err)
	}
	if result.IntVal != 3 {
		t.Errorf("expected 3, got %s", result.String())
	}
}

func TestEvalAllEmpty(t *testing.T) {
	e := setupEval()
	result, err := EvalAll(nil, e)
	if err != nil {
		t.Fatal(err)
	}
	if result.Type != types.TypeVoid {
		t.Errorf("expected void for empty EvalAll, got %s", result.String())
	}
}

func TestEvalAllSingle(t *testing.T) {
	e := setupEval()
	exprs := []*types.Value{types.NewInteger(42)}
	result, err := EvalAll(exprs, e)
	if err != nil {
		t.Fatal(err)
	}
	if result.IntVal != 42 {
		t.Errorf("expected 42, got %s", result.String())
	}
}

// =========================================================================
// Quasiquote coverage for expandQQ
// =========================================================================

func TestQuasiquoteNested(t *testing.T) {
	r := evalHelper(t, "`(a (b ,(+ 1 2)) c)")
	if r.String() != "(a (b 3) c)" {
		t.Errorf("expected (a (b 3) c), got %s", r.String())
	}
}

func TestQuasiquoteSplicing(t *testing.T) {
	r := evalHelper(t, "`(a ,@(list 1 2 3) b)")
	if r.String() != "(a 1 2 3 b)" {
		t.Errorf("expected (a 1 2 3 b), got %s", r.String())
	}
}

func TestQuasiquoteNoUnquote(t *testing.T) {
	r := evalHelper(t, "`(a b c)")
	if r.String() != "(a b c)" {
		t.Errorf("expected (a b c), got %s", r.String())
	}
}

func TestQuasiquoteAtom(t *testing.T) {
	r := evalHelper(t, "`42")
	if r.IntVal != 42 {
		t.Errorf("expected 42, got %s", r.String())
	}
}

func TestQuasiquoteSplicingEmpty(t *testing.T) {
	r := evalHelper(t, "`(a ,@'() b)")
	if r.String() != "(a b)" {
		t.Errorf("expected (a b), got %s", r.String())
	}
}

func TestQuasiquoteVector(t *testing.T) {
	r := evalHelper(t, "`#(1 ,(+ 1 1) 3)")
	if r.Type != types.TypeVector {
		t.Fatalf("expected vector, got %s", r.Type)
	}
	if r.VecVal[1].IntVal != 2 {
		t.Errorf("expected 2, got %s", r.VecVal[1].String())
	}
}

// =========================================================================
// define edge cases
// =========================================================================

func TestDefineFunctionWithDottedParams(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, "(define (f x . rest) (list x rest))", e)
	r := evalWithEnv(t, "(f 1 2 3)", e)
	lst, _ := r.ToList()
	if len(lst) != 2 || lst[0].IntVal != 1 {
		t.Errorf("expected (1 (2 3)), got %s", r.String())
	}
}

func TestDefineInvalidSyntax(t *testing.T) {
	err := evalErr(t, "(define)")
	if err == nil {
		t.Error("expected error for empty define")
	}
}

func TestDefineNonSymbol(t *testing.T) {
	err := evalErr(t, "(define 42 1)")
	if err == nil {
		t.Error("expected error for define with non-symbol")
	}
}

// =========================================================================
// set! edge cases
// =========================================================================

func TestSetInvalidSyntax(t *testing.T) {
	err := evalErr(t, "(set!)")
	if err == nil {
		t.Error("expected error for empty set!")
	}
}

func TestSetNonSymbol(t *testing.T) {
	err := evalErr(t, "(set! 42 1)")
	if err == nil {
		t.Error("expected error for set! with non-symbol")
	}
}

// =========================================================================
// lambda edge cases
// =========================================================================

func TestLambdaMultiBody(t *testing.T) {
	r := evalHelper(t, "((lambda (x) 1 2 x) 42)")
	if r.IntVal != 42 {
		t.Errorf("expected 42, got %s", r.String())
	}
}

// =========================================================================
// let edge cases
// =========================================================================

func TestLetEmpty(t *testing.T) {
	r := evalHelper(t, "(let () 42)")
	if r.IntVal != 42 {
		t.Errorf("expected 42, got %s", r.String())
	}
}

func TestLetStarEmpty(t *testing.T) {
	r := evalHelper(t, "(let* () 42)")
	if r.IntVal != 42 {
		t.Errorf("expected 42, got %s", r.String())
	}
}

func TestLetrecEmpty(t *testing.T) {
	r := evalHelper(t, "(letrec () 42)")
	if r.IntVal != 42 {
		t.Errorf("expected 42, got %s", r.String())
	}
}

func TestLetInvalidBindings(t *testing.T) {
	err := evalErr(t, "(let 42 1)")
	// named let with "42" as name will fail differently - just check it errors
	if err == nil {
		t.Error("expected error for invalid let bindings")
	}
}

func TestLetStarInvalidBindings(t *testing.T) {
	err := evalErr(t, "(let* 42 1)")
	if err == nil {
		t.Error("expected error for invalid let* bindings")
	}
}

// =========================================================================
// begin edge cases
// =========================================================================

func TestBeginEmpty(t *testing.T) {
	r := evalHelper(t, "(begin)")
	if r.Type != types.TypeVoid {
		t.Errorf("expected void for empty begin, got %s", r.String())
	}
}

func TestBeginSingle(t *testing.T) {
	r := evalHelper(t, "(begin 42)")
	if r.IntVal != 42 {
		t.Errorf("expected 42, got %s", r.String())
	}
}

// =========================================================================
// cond edge cases
// =========================================================================

func TestCondMultiBody(t *testing.T) {
	r := evalHelper(t, "(cond (#t 1 2 3))")
	if r.IntVal != 3 {
		t.Errorf("expected 3, got %s", r.String())
	}
}

// =========================================================================
// when/unless edge cases
// =========================================================================

func TestWhenMultiBody(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, "(define x 0)", e)
	evalWithEnv(t, "(when #t (set! x 1) (set! x 2))", e)
	r := evalWithEnv(t, "x", e)
	if r.IntVal != 2 {
		t.Errorf("expected 2, got %d", r.IntVal)
	}
}

func TestUnlessMultiBody(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, "(define x 0)", e)
	evalWithEnv(t, "(unless #f (set! x 1) (set! x 2))", e)
	r := evalWithEnv(t, "x", e)
	if r.IntVal != 2 {
		t.Errorf("expected 2, got %d", r.IntVal)
	}
}

// =========================================================================
// do edge cases
// =========================================================================

func TestDoWithMultipleVars(t *testing.T) {
	r := evalHelper(t, `
		(do ((i 0 (+ i 1))
			 (j 10 (- j 1)))
			((= i 5) (+ i j)))
	`)
	if r.IntVal != 10 {
		t.Errorf("expected 10, got %s", r.String())
	}
}

func TestDoNoBody(t *testing.T) {
	r := evalHelper(t, "(do ((i 0 (+ i 1))) ((= i 3) i))")
	if r.IntVal != 3 {
		t.Errorf("expected 3, got %s", r.String())
	}
}

func TestDoNoStepExpression(t *testing.T) {
	r := evalHelper(t, "(do ((i 0) (j 0 (+ j 1))) ((= j 3) i))")
	if r.IntVal != 0 {
		t.Errorf("expected 0, got %s", r.String())
	}
}

// =========================================================================
// case edge cases
// =========================================================================

func TestCaseNoMatch(t *testing.T) {
	r := evalHelper(t, `(case 99 ((1) 10) ((2) 20))`)
	if r.Type != types.TypeVoid {
		t.Errorf("expected void for no match, got %s", r.String())
	}
}

func TestCaseWithSymbols(t *testing.T) {
	r := evalHelper(t, `
		(define x 'b)
		(case x ((a) 1) ((b) 2) ((c) 3))
	`)
	if r.IntVal != 2 {
		t.Errorf("expected 2, got %s", r.String())
	}
}

// =========================================================================
// define-record-type edge cases
// =========================================================================

func TestRecordAccessorOnWrongType(t *testing.T) {
	err := evalErr(t, `
		(define-record-type <point> (make-point x y) point? (x point-x) (y point-y))
		(point-x 42)
	`)
	if err == nil {
		t.Error("expected error for accessor on non-record")
	}
}

// =========================================================================
// cond-expand edge cases
// =========================================================================

func TestCondExpandNoMatch(t *testing.T) {
	r := evalHelper(t, `(cond-expand (nonexistent1 1) (nonexistent2 2) (else 99))`)
	if r.IntVal != 99 {
		t.Errorf("expected 99, got %s", r.String())
	}
}

func TestCondExpandMultiBody(t *testing.T) {
	r := evalHelper(t, `(cond-expand (r7rs 1 2 3))`)
	if r.IntVal != 3 {
		t.Errorf("expected 3 (last of multi-body), got %s", r.String())
	}
}

// =========================================================================
// define-values edge cases
// =========================================================================

func TestDefineValuesMismatch(t *testing.T) {
	err := evalErr(t, `(define-values (a b c) (values 1 2))`)
	if err == nil {
		t.Error("expected error for arity mismatch in define-values")
	}
}

// =========================================================================
// TCO edge cases
// =========================================================================

func TestTCOInLet(t *testing.T) {
	count := 10000
	if testing.Short() {
		count = 1000
	}
	r := evalHelper(t, fmt.Sprintf(`
		(define (loop n)
			(let ((x n))
				(if (= x 0) "done" (loop (- x 1)))))
		(loop %d)
	`, count))
	if r.StrVal != "done" {
		t.Errorf("expected done, got %s", r.String())
	}
}

func TestTCOInCond(t *testing.T) {
	count := 10000
	if testing.Short() {
		count = 1000
	}
	r := evalHelper(t, fmt.Sprintf(`
		(define (loop n)
			(cond ((= n 0) "done")
				  (else (loop (- n 1)))))
		(loop %d)
	`, count))
	if r.StrVal != "done" {
		t.Errorf("expected done, got %s", r.String())
	}
}

func TestTCOInBegin(t *testing.T) {
	count := 10000
	if testing.Short() {
		count = 1000
	}
	e := setupEval()
	evalWithEnv(t, `(define (loop n)
		(begin
			(if (= n 0) "done" (loop (- n 1)))))`, e)
	r := evalWithEnv(t, fmt.Sprintf("(loop %d)", count), e)
	if r.StrVal != "done" {
		t.Errorf("expected done, got %s", r.String())
	}
}

// =========================================================================
// map/for-each/filter edge cases
// =========================================================================

func TestMapSingle(t *testing.T) {
	r := evalHelper(t, `(map (lambda (x) (* x 2)) '(1 2 3))`)
	lst, _ := r.ToList()
	if len(lst) != 3 || lst[0].IntVal != 2 || lst[1].IntVal != 4 || lst[2].IntVal != 6 {
		t.Errorf("expected (2 4 6), got %s", r.String())
	}
}

func TestMapNotProcedure(t *testing.T) {
	err := evalErr(t, `(map 42 '(1 2 3))`)
	if err == nil {
		t.Error("expected error for map with non-procedure")
	}
}

func TestForEachNotProcedure(t *testing.T) {
	err := evalErr(t, `(for-each 42 '(1 2 3))`)
	if err == nil {
		t.Error("expected error for for-each with non-procedure")
	}
}

func TestFilterAll(t *testing.T) {
	r := evalHelper(t, `(filter (lambda (x) #t) '(1 2 3))`)
	lst, _ := r.ToList()
	if len(lst) != 3 {
		t.Errorf("expected (1 2 3), got %s", r.String())
	}
}

func TestFilterNone(t *testing.T) {
	r := evalHelper(t, `(filter (lambda (x) #f) '(1 2 3))`)
	if r.Type != types.TypeNil {
		t.Errorf("expected (), got %s", r.String())
	}
}

// =========================================================================
// EvalString edge cases
// =========================================================================

func TestEvalStringEmpty(t *testing.T) {
	e := setupEval()
	r, err := EvalString("", e)
	if err != nil {
		t.Fatal(err)
	}
	if r.Type != types.TypeVoid {
		t.Errorf("expected void for empty input, got %s", r.String())
	}
}

func TestEvalStringParseError(t *testing.T) {
	e := setupEval()
	_, err := EvalString("(", e)
	if err == nil {
		t.Error("expected parse error")
	}
}

func TestEvalStringMultipleExpressions(t *testing.T) {
	e := setupEval()
	r, err := EvalString("(define x 1) (define y 2) (+ x y)", e)
	if err != nil {
		t.Fatal(err)
	}
	if r.IntVal != 3 {
		t.Errorf("expected 3, got %s", r.String())
	}
}

// =========================================================================
// Practical integration tests
// =========================================================================

func TestMapFilter(t *testing.T) {
	r := evalHelper(t, `
		(define (even? n) (= (modulo n 2) 0))
		(map (lambda (x) (* x x))
			 (filter even? '(1 2 3 4 5 6)))
	`)
	lst, _ := r.ToList()
	if len(lst) != 3 || lst[0].IntVal != 4 || lst[1].IntVal != 16 || lst[2].IntVal != 36 {
		t.Errorf("expected (4 16 36), got %s", r.String())
	}
}

func TestYCombinator(t *testing.T) {
	r := evalHelper(t, `
		(define Y
			(lambda (f)
				((lambda (x) (f (lambda (v) ((x x) v))))
				 (lambda (x) (f (lambda (v) ((x x) v)))))))
		(define fact
			(Y (lambda (self)
				(lambda (n)
					(if (= n 0) 1 (* n (self (- n 1))))))))
		(fact 5)
	`)
	if r.IntVal != 120 {
		t.Errorf("expected 120, got %s", r.String())
	}
}

func TestAccumulateWithFold(t *testing.T) {
	r := evalHelper(t, `
		(define (foldl f init lst)
			(if (null? lst) init
				(foldl f (f (car lst) init) (cdr lst))))
		(foldl + 0 '(1 2 3 4 5))
	`)
	if r.IntVal != 15 {
		t.Errorf("expected 15, got %s", r.String())
	}
}

func TestClosureMutation(t *testing.T) {
	r := evalHelper(t, `
		(define (make-box val)
			(define (get) val)
			(define (set-val! new-val) (set! val new-val))
			(list get set-val!))
		(define box (make-box 10))
		(define get (car box))
		(define set-box! (car (cdr box)))
		(set-box! 42)
		(get)
	`)
	if r.IntVal != 42 {
		t.Errorf("expected 42, got %s", r.String())
	}
}
