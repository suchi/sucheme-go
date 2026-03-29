package evaluator

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// =========================================================================
// define-record-type
// =========================================================================

func TestDefineRecordTypeBasic(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, `(define-record-type <point>
		(make-point x y)
		point?
		(x point-x)
		(y point-y))`, e)
	evalWithEnv(t, `(define p (make-point 1 2))`, e)

	tests := []struct {
		input  string
		expect string
	}{
		{`(point? p)`, `#t`},
		{`(point-x p)`, `1`},
		{`(point-y p)`, `2`},
		{`(point? 42)`, `#f`},
		{`(point? "hello")`, `#f`},
		{`(point? #t)`, `#f`},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			r := evalWithEnv(t, tt.input, e)
			if r.String() != tt.expect {
				t.Errorf("got %s, want %s", r.String(), tt.expect)
			}
		})
	}
}

func TestDefineRecordTypeMutator(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, `(define-record-type <mpoint>
		(make-mpoint x y)
		mpoint?
		(x mpoint-x set-mpoint-x!)
		(y mpoint-y set-mpoint-y!))`, e)
	evalWithEnv(t, `(define mp (make-mpoint 10 20))`, e)

	r := evalWithEnv(t, `(mpoint-x mp)`, e)
	if r.String() != "10" {
		t.Errorf("expected 10, got %s", r.String())
	}

	evalWithEnv(t, `(set-mpoint-x! mp 99)`, e)
	r = evalWithEnv(t, `(mpoint-x mp)`, e)
	if r.String() != "99" {
		t.Errorf("after mutation expected 99, got %s", r.String())
	}

	evalWithEnv(t, `(set-mpoint-y! mp 88)`, e)
	r = evalWithEnv(t, `(mpoint-y mp)`, e)
	if r.String() != "88" {
		t.Errorf("after mutation expected 88, got %s", r.String())
	}
}

func TestDefineRecordTypeZeroFields(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, `(define-record-type <unit>
		(make-unit)
		unit?)`, e)
	evalWithEnv(t, `(define u (make-unit))`, e)

	r := evalWithEnv(t, `(unit? u)`, e)
	if r.String() != "#t" {
		t.Errorf("expected #t, got %s", r.String())
	}
}

func TestDefineRecordTypeSingleField(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, `(define-record-type <box>
		(make-box value)
		box?
		(value box-value set-box-value!))`, e)
	evalWithEnv(t, `(define b (make-box 42))`, e)

	r := evalWithEnv(t, `(box-value b)`, e)
	if r.String() != "42" {
		t.Errorf("expected 42, got %s", r.String())
	}

	evalWithEnv(t, `(set-box-value! b "hello")`, e)
	r = evalWithEnv(t, `(box-value b)`, e)
	if r.String() != `"hello"` {
		t.Errorf("expected \"hello\", got %s", r.String())
	}
}

func TestDefineRecordTypeMultipleTypes(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, `(define-record-type <point>
		(make-point x y)
		point?
		(x point-x)
		(y point-y))`, e)
	evalWithEnv(t, `(define-record-type <color>
		(make-color r g b)
		color?
		(r color-r)
		(g color-g)
		(b color-b))`, e)
	evalWithEnv(t, `(define p (make-point 1 2))`, e)
	evalWithEnv(t, `(define c (make-color 255 128 0))`, e)

	tests := []struct {
		input  string
		expect string
	}{
		{`(point? p)`, `#t`},
		{`(color? c)`, `#t`},
		{`(point? c)`, `#f`},
		{`(color? p)`, `#f`},
		{`(color-r c)`, `255`},
		{`(color-g c)`, `128`},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			r := evalWithEnv(t, tt.input, e)
			if r.String() != tt.expect {
				t.Errorf("got %s, want %s", r.String(), tt.expect)
			}
		})
	}
}

func TestDefineRecordTypeConstructorWrongArgs(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, `(define-record-type <point>
		(make-point x y)
		point?
		(x point-x)
		(y point-y))`, e)

	_, err := EvalString(`(make-point 1)`, e)
	if err == nil {
		t.Error("expected error for wrong number of constructor args")
	}

	_, err = EvalString(`(make-point 1 2 3)`, e)
	if err == nil {
		t.Error("expected error for too many constructor args")
	}
}

func TestDefineRecordTypeAccessorWrongType(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, `(define-record-type <point>
		(make-point x y)
		point?
		(x point-x)
		(y point-y))`, e)

	_, err := EvalString(`(point-x 42)`, e)
	if err == nil {
		t.Error("expected error for accessor on non-record")
	}
}

func TestDefineRecordTypeMutatorWrongType(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, `(define-record-type <mpoint>
		(make-mpoint x y)
		mpoint?
		(x mpoint-x set-mpoint-x!)
		(y mpoint-y set-mpoint-y!))`, e)

	_, err := EvalString(`(set-mpoint-x! 42 10)`, e)
	if err == nil {
		t.Error("expected error for mutator on non-record")
	}
}

func TestDefineRecordTypeStringRepresentation(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, `(define-record-type <point>
		(make-point x y)
		point?
		(x point-x)
		(y point-y))`, e)
	r := evalWithEnv(t, `(make-point 1 2)`, e)
	if r.String() != "#<<point>>" {
		t.Errorf("expected #<<point>>, got %s", r.String())
	}
}

func TestDefineRecordTypePartialConstructor(t *testing.T) {
	// Constructor that only takes some fields
	e := setupEval()
	evalWithEnv(t, `(define-record-type <named>
		(make-named name)
		named?
		(name named-name)
		(age named-age set-named-age!))`, e)
	evalWithEnv(t, `(define n (make-named "Alice"))`, e)

	r := evalWithEnv(t, `(named-name n)`, e)
	if r.String() != `"Alice"` {
		t.Errorf("expected \"Alice\", got %s", r.String())
	}

	// Age wasn't set by constructor, so accessor should fail
	_, err := EvalString(`(named-age n)`, e)
	if err == nil {
		t.Error("expected error for accessing unset field")
	}

	// Set age and access it
	evalWithEnv(t, `(set-named-age! n 30)`, e)
	r = evalWithEnv(t, `(named-age n)`, e)
	if r.String() != "30" {
		t.Errorf("expected 30, got %s", r.String())
	}
}

func TestDefineRecordTypeInvalidSyntax(t *testing.T) {
	tests := []string{
		`(define-record-type)`,
		`(define-record-type <point>)`,
		`(define-record-type <point> (make-point))`,
		`(define-record-type 42 (make-point x) point? (x point-x))`,
		`(define-record-type <point> 42 point? (x point-x))`,
		`(define-record-type <point> (make-point x) 42 (x point-x))`,
	}
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := EvalString(input, setupEval())
			if err == nil {
				t.Errorf("expected error for %s", input)
			}
		})
	}
}

// =========================================================================
// cond-expand
// =========================================================================

func TestCondExpandBasicFeature(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{`(cond-expand (r7rs 1))`, `1`},
		{`(cond-expand (sucheme 2))`, `2`},
		{`(cond-expand (exact-closed 3))`, `3`},
		{`(cond-expand (ieee-float 4))`, `4`},
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

func TestCondExpandElse(t *testing.T) {
	r := evalHelper(t, `(cond-expand (nonexistent-feature 1) (else 42))`)
	if r.String() != "42" {
		t.Errorf("expected 42, got %s", r.String())
	}
}

func TestCondExpandMultipleClauses(t *testing.T) {
	r := evalHelper(t, `(cond-expand (nonexistent 1) (r7rs 2) (else 3))`)
	if r.String() != "2" {
		t.Errorf("expected 2, got %s", r.String())
	}
}

func TestCondExpandLibrary(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{`(cond-expand ((library (scheme base)) 1) (else 0))`, `1`},
		{`(cond-expand ((library (scheme write)) 2) (else 0))`, `2`},
		{`(cond-expand ((library (scheme nonexistent)) 1) (else 0))`, `0`},
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

func TestCondExpandAnd(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{`(cond-expand ((and r7rs sucheme) 1) (else 0))`, `1`},
		{`(cond-expand ((and r7rs nonexistent) 1) (else 0))`, `0`},
		{`(cond-expand ((and) 1) (else 0))`, `1`},
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

func TestCondExpandOr(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{`(cond-expand ((or r7rs nonexistent) 1) (else 0))`, `1`},
		{`(cond-expand ((or nonexistent1 nonexistent2) 1) (else 0))`, `0`},
		{`(cond-expand ((or) 1) (else 0))`, `0`},
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

func TestCondExpandNot(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{`(cond-expand ((not nonexistent) 1) (else 0))`, `1`},
		{`(cond-expand ((not r7rs) 1) (else 0))`, `0`},
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

func TestCondExpandNestedCombinators(t *testing.T) {
	r := evalHelper(t, `(cond-expand ((and r7rs (not nonexistent) (or sucheme bogus)) 42) (else 0))`)
	if r.String() != "42" {
		t.Errorf("expected 42, got %s", r.String())
	}
}

func TestCondExpandMultipleExprs(t *testing.T) {
	e := setupEval()
	r := evalWithEnv(t, `(cond-expand (r7rs (define x 1) (define y 2) (+ x y)))`, e)
	if r.String() != "3" {
		t.Errorf("expected 3, got %s", r.String())
	}
}

func TestCondExpandNoMatch(t *testing.T) {
	_, err := EvalString(`(cond-expand (nonexistent 1))`, setupEval())
	if err == nil {
		t.Error("expected error when no clause matches")
	}
}

func TestCondExpandElseMustBeLast(t *testing.T) {
	_, err := EvalString(`(cond-expand (else 1) (r7rs 2))`, setupEval())
	if err == nil {
		t.Error("expected error for else not in last position")
	}
}

func TestCondExpandEmpty(t *testing.T) {
	_, err := EvalString(`(cond-expand)`, setupEval())
	if err == nil {
		t.Error("expected error for empty cond-expand")
	}
}

// =========================================================================
// define-values
// =========================================================================

func TestDefineValuesBasic(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, `(define-values (x y z) (values 1 2 3))`, e)

	tests := []struct {
		input  string
		expect string
	}{
		{`x`, `1`},
		{`y`, `2`},
		{`z`, `3`},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			r := evalWithEnv(t, tt.input, e)
			if r.String() != tt.expect {
				t.Errorf("got %s, want %s", r.String(), tt.expect)
			}
		})
	}
}

func TestDefineValuesSingle(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, `(define-values (x) (values 42))`, e)
	r := evalWithEnv(t, `x`, e)
	if r.String() != "42" {
		t.Errorf("expected 42, got %s", r.String())
	}
}

func TestDefineValuesSingleNonValues(t *testing.T) {
	// When the expression returns a single value (not wrapped in values)
	e := setupEval()
	evalWithEnv(t, `(define-values (x) 42)`, e)
	r := evalWithEnv(t, `x`, e)
	if r.String() != "42" {
		t.Errorf("expected 42, got %s", r.String())
	}
}

func TestDefineValuesEmpty(t *testing.T) {
	e := setupEval()
	// No bindings, just evaluate for side effects
	evalWithEnv(t, `(define-values () (values))`, e)
}

func TestDefineValuesRest(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, `(define-values rest (values 1 2 3))`, e)
	r := evalWithEnv(t, `rest`, e)
	if r.String() != "(1 2 3)" {
		t.Errorf("expected (1 2 3), got %s", r.String())
	}
}

func TestDefineValuesMismatch(t *testing.T) {
	_, err := EvalString(`(define-values (x y) (values 1))`, setupEval())
	if err == nil {
		t.Error("expected error for value count mismatch")
	}

	_, err = EvalString(`(define-values (x) (values 1 2))`, setupEval())
	if err == nil {
		t.Error("expected error for too many values")
	}
}

func TestDefineValuesInvalidFormals(t *testing.T) {
	_, err := EvalString(`(define-values (42) (values 1))`, setupEval())
	if err == nil {
		t.Error("expected error for non-symbol formal")
	}
}

func TestDefineValuesInvalidSyntax(t *testing.T) {
	_, err := EvalString(`(define-values)`, setupEval())
	if err == nil {
		t.Error("expected error for missing args")
	}

	_, err = EvalString(`(define-values (x))`, setupEval())
	if err == nil {
		t.Error("expected error for missing expression")
	}
}

func TestDefineValuesWithExpression(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, `(define (swap a b) (values b a))`, e)
	evalWithEnv(t, `(define-values (x y) (swap 1 2))`, e)
	rx := evalWithEnv(t, `x`, e)
	ry := evalWithEnv(t, `y`, e)
	if rx.String() != "2" || ry.String() != "1" {
		t.Errorf("expected 2 and 1, got %s and %s", rx.String(), ry.String())
	}
}

func TestDefineValuesTwoValues(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, `(define-values (a b) (values 10 20))`, e)
	r := evalWithEnv(t, `(+ a b)`, e)
	if r.String() != "30" {
		t.Errorf("expected 30, got %s", r.String())
	}
}

// =========================================================================
// include / include-ci
// =========================================================================

func TestIncludeBasic(t *testing.T) {
	// Create a temp file
	dir := t.TempDir()
	path := filepath.Join(dir, "test.scm")
	if err := os.WriteFile(path, []byte("(define included-x 42)"), 0644); err != nil {
		t.Fatal(err)
	}

	e := setupEval()
	evalWithEnv(t, fmt.Sprintf(`(include %q)`, path), e)
	r := evalWithEnv(t, `included-x`, e)
	if r.String() != "42" {
		t.Errorf("expected 42, got %s", r.String())
	}
}

func TestIncludeMultipleFiles(t *testing.T) {
	dir := t.TempDir()
	path1 := filepath.Join(dir, "a.scm")
	path2 := filepath.Join(dir, "b.scm")
	os.WriteFile(path1, []byte("(define inc-a 1)"), 0644)
	os.WriteFile(path2, []byte("(define inc-b 2)"), 0644)

	e := setupEval()
	evalWithEnv(t, fmt.Sprintf(`(include %q %q)`, path1, path2), e)
	ra := evalWithEnv(t, `inc-a`, e)
	rb := evalWithEnv(t, `inc-b`, e)
	if ra.String() != "1" || rb.String() != "2" {
		t.Errorf("expected 1 and 2, got %s and %s", ra.String(), rb.String())
	}
}

func TestIncludeMultipleExprs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "multi.scm")
	os.WriteFile(path, []byte("(define inc-p 10)\n(define inc-q 20)"), 0644)

	e := setupEval()
	evalWithEnv(t, fmt.Sprintf(`(include %q)`, path), e)
	rp := evalWithEnv(t, `inc-p`, e)
	rq := evalWithEnv(t, `inc-q`, e)
	if rp.String() != "10" || rq.String() != "20" {
		t.Errorf("expected 10 and 20, got %s and %s", rp.String(), rq.String())
	}
}

func TestIncludeFileNotFound(t *testing.T) {
	_, err := EvalString(`(include "/nonexistent/file.scm")`, setupEval())
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestIncludeNonStringArg(t *testing.T) {
	_, err := EvalString(`(include 42)`, setupEval())
	if err == nil {
		t.Error("expected error for non-string filename")
	}
}

func TestIncludeNoArgs(t *testing.T) {
	_, err := EvalString(`(include)`, setupEval())
	if err == nil {
		t.Error("expected error for no args")
	}
}

func TestIncludeCiBasic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "upper.scm")
	// Write uppercase identifiers - include-ci should lowercase them
	os.WriteFile(path, []byte("(DEFINE UPPER-VAR 99)"), 0644)

	e := setupEval()
	evalWithEnv(t, fmt.Sprintf(`(include-ci %q)`, path), e)
	r := evalWithEnv(t, `upper-var`, e)
	if r.String() != "99" {
		t.Errorf("expected 99, got %s", r.String())
	}
}

func TestIncludeParseError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.scm")
	os.WriteFile(path, []byte("(define x"), 0644)

	_, err := EvalString(fmt.Sprintf(`(include %q)`, path), setupEval())
	if err == nil {
		t.Error("expected parse error for malformed file")
	}
}

func TestIncludeRuntimeError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "err.scm")
	os.WriteFile(path, []byte("(+ 1 \"not a number\")"), 0644)

	_, err := EvalString(fmt.Sprintf(`(include %q)`, path), setupEval())
	if err == nil {
		t.Error("expected runtime error from included file")
	}
}

func TestIncludeEmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.scm")
	os.WriteFile(path, []byte(""), 0644)

	// Should succeed and return void
	r := evalHelper(t, fmt.Sprintf(`(include %q)`, path))
	if r.String() != "#<void>" {
		t.Errorf("expected #<void>, got %s", r.String())
	}
}

// =========================================================================
// case
// =========================================================================

func TestCaseBasic(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{`(case 1 ((1) 'one) ((2) 'two))`, `one`},
		{`(case 2 ((1) 'one) ((2) 'two))`, `two`},
		{`(case 3 ((1) 'one) ((2) 'two))`, `#<void>`},
		{`(case 1 ((1 2 3) 'small) ((4 5 6) 'large))`, `small`},
		{`(case 5 ((1 2 3) 'small) ((4 5 6) 'large))`, `large`},
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

func TestCaseElse(t *testing.T) {
	r := evalHelper(t, `(case 99 ((1) 'one) (else 'other))`)
	if r.String() != "other" {
		t.Errorf("expected other, got %s", r.String())
	}
}

func TestCaseArrow(t *testing.T) {
	// (case key ((datum) => proc))
	r := evalHelper(t, `(case (+ 1 1)
		((1) => (lambda (x) (* x 10)))
		((2) => (lambda (x) (* x 20)))
		(else 'none))`)
	if r.String() != "40" {
		t.Errorf("expected 40, got %s", r.String())
	}
}

func TestCaseElseArrow(t *testing.T) {
	r := evalHelper(t, `(case 99
		((1) 'one)
		(else => (lambda (x) (+ x 1))))`)
	if r.String() != "100" {
		t.Errorf("expected 100, got %s", r.String())
	}
}

func TestCaseSymbols(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{`(case 'a ((a) 1) ((b) 2))`, `1`},
		{`(case 'b ((a) 1) ((b) 2))`, `2`},
		{`(case 'c ((a) 1) ((b) 2) (else 3))`, `3`},
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

func TestCaseCharacters(t *testing.T) {
	r := evalHelper(t, `(case #\a ((#\a #\b) 'found) (else 'not-found))`)
	if r.String() != "found" {
		t.Errorf("expected found, got %s", r.String())
	}
}

func TestCaseMultipleBody(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, `(define case-result 0)`, e)
	r := evalWithEnv(t, `(case 1 ((1) (set! case-result 10) (+ case-result 5)))`, e)
	if r.String() != "15" {
		t.Errorf("expected 15, got %s", r.String())
	}
}

func TestCaseEqvComparison(t *testing.T) {
	// eqv? should match numbers by value
	tests := []struct {
		input  string
		expect string
	}{
		{`(case 1 ((1) 'match) (else 'no))`, `match`},
		{`(case #t ((#t) 'match) (else 'no))`, `match`},
		{`(case #f ((#f) 'match) (else 'no))`, `match`},
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

func TestCaseElseMustBeLast(t *testing.T) {
	_, err := EvalString(`(case 1 (else 'a) ((1) 'b))`, setupEval())
	if err == nil {
		t.Error("expected error for else not in last position")
	}
}

func TestCaseInvalidSyntax(t *testing.T) {
	tests := []string{
		`(case)`,
		`(case 1)`,
	}
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := EvalString(input, setupEval())
			if err == nil {
				t.Errorf("expected error for %s", input)
			}
		})
	}
}

func TestCaseKeyExpression(t *testing.T) {
	r := evalHelper(t, `(case (* 2 3) ((5) 'five) ((6) 'six) (else 'other))`)
	if r.String() != "six" {
		t.Errorf("expected six, got %s", r.String())
	}
}

func TestCaseNoMatchNoElse(t *testing.T) {
	r := evalHelper(t, `(case 99 ((1) 'one) ((2) 'two))`)
	if r.String() != "#<void>" {
		t.Errorf("expected #<void>, got %s", r.String())
	}
}

// =========================================================================
// Integration tests
// =========================================================================

func TestRecordWithCase(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, `(define-record-type <shape>
		(make-shape kind size)
		shape?
		(kind shape-kind)
		(size shape-size))`, e)
	evalWithEnv(t, `(define s (make-shape 'circle 10))`, e)
	r := evalWithEnv(t, `(case (shape-kind s)
		((circle) (shape-size s))
		((square) (* (shape-size s) (shape-size s)))
		(else 0))`, e)
	if r.String() != "10" {
		t.Errorf("expected 10, got %s", r.String())
	}
}

func TestCondExpandWithDefineValues(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, `(cond-expand
		(r7rs (define-values (a b) (values 100 200))))`, e)
	r := evalWithEnv(t, `(+ a b)`, e)
	if r.String() != "300" {
		t.Errorf("expected 300, got %s", r.String())
	}
}

func TestRecordInList(t *testing.T) {
	e := setupEval()
	evalWithEnv(t, `(define-record-type <pair-record>
		(make-pr first second)
		pr?
		(first pr-first)
		(second pr-second))`, e)
	evalWithEnv(t, `(define records (list (make-pr 1 2) (make-pr 3 4)))`, e)
	r := evalWithEnv(t, `(pr-first (car records))`, e)
	if r.String() != "1" {
		t.Errorf("expected 1, got %s", r.String())
	}
	r = evalWithEnv(t, `(pr-second (car (cdr records)))`, e)
	if r.String() != "4" {
		t.Errorf("expected 4, got %s", r.String())
	}
}
