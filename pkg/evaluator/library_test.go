package evaluator

import (
	"testing"
)

// Reuses setupEval, evalHelper, evalErr from evaluator_test.go

// =========================================================================
// define-library + import
// =========================================================================

func TestDefineLibraryBasic(t *testing.T) {
	e := setupEval()
	_, err := EvalString(`
		(define-library (test lib)
			(export double)
			(begin
				(define (double x) (* x 2))))
	`, e)
	if err != nil {
		t.Fatalf("define-library failed: %v", err)
	}

	_, err = EvalString(`(import (test lib))`, e)
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	result, err := EvalString(`(double 21)`, e)
	if err != nil {
		t.Fatalf("call failed: %v", err)
	}
	if result.IntVal != 42 {
		t.Errorf("expected 42, got %s", result.String())
	}
}

func TestDefineLibraryWithRenameExport(t *testing.T) {
	e := setupEval()
	_, err := EvalString(`
		(define-library (test rename)
			(export (rename internal-fn external-fn))
			(begin
				(define (internal-fn x) (+ x 100))))
	`, e)
	if err != nil {
		t.Fatalf("define-library failed: %v", err)
	}

	_, err = EvalString(`(import (test rename))`, e)
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	result, err := EvalString(`(external-fn 1)`, e)
	if err != nil {
		t.Fatalf("call failed: %v", err)
	}
	if result.IntVal != 101 {
		t.Errorf("expected 101, got %s", result.String())
	}
}

func TestImportOnly(t *testing.T) {
	e := setupEval()
	_, err := EvalString(`
		(define-library (test multi)
			(export a b c)
			(begin
				(define a 1) (define b 2) (define c 3)))
	`, e)
	if err != nil {
		t.Fatalf("define-library failed: %v", err)
	}

	_, err = EvalString(`(import (only (test multi) a c))`, e)
	if err != nil {
		t.Fatalf("import only failed: %v", err)
	}

	result, err := EvalString(`(+ a c)`, e)
	if err != nil {
		t.Fatalf("eval failed: %v", err)
	}
	if result.IntVal != 4 {
		t.Errorf("expected 4, got %s", result.String())
	}

	// b should not be imported
	_, err = EvalString(`b`, e)
	if err == nil {
		t.Error("expected error: b should not be imported")
	}
}

func TestImportExcept(t *testing.T) {
	e := setupEval()
	_, err := EvalString(`
		(define-library (test exc)
			(export x y z)
			(begin
				(define x 10) (define y 20) (define z 30)))
	`, e)
	if err != nil {
		t.Fatalf("define-library failed: %v", err)
	}

	_, err = EvalString(`(import (except (test exc) y))`, e)
	if err != nil {
		t.Fatalf("import except failed: %v", err)
	}

	result, err := EvalString(`(+ x z)`, e)
	if err != nil {
		t.Fatalf("eval failed: %v", err)
	}
	if result.IntVal != 40 {
		t.Errorf("expected 40, got %s", result.String())
	}
}

func TestImportPrefix(t *testing.T) {
	e := setupEval()
	_, err := EvalString(`
		(define-library (test pfx)
			(export val)
			(begin (define val 99)))
	`, e)
	if err != nil {
		t.Fatalf("define-library failed: %v", err)
	}

	_, err = EvalString(`(import (prefix (test pfx) my-))`, e)
	if err != nil {
		t.Fatalf("import prefix failed: %v", err)
	}

	result, err := EvalString(`my-val`, e)
	if err != nil {
		t.Fatalf("eval failed: %v", err)
	}
	if result.IntVal != 99 {
		t.Errorf("expected 99, got %s", result.String())
	}
}

func TestImportRename(t *testing.T) {
	e := setupEval()
	_, err := EvalString(`
		(define-library (test ren)
			(export orig)
			(begin (define orig 42)))
	`, e)
	if err != nil {
		t.Fatalf("define-library failed: %v", err)
	}

	_, err = EvalString(`(import (rename (test ren) (orig aliased)))`, e)
	if err != nil {
		t.Fatalf("import rename failed: %v", err)
	}

	result, err := EvalString(`aliased`, e)
	if err != nil {
		t.Fatalf("eval failed: %v", err)
	}
	if result.IntVal != 42 {
		t.Errorf("expected 42, got %s", result.String())
	}
}

func TestImportUnknownLibrary(t *testing.T) {
	e := setupEval()
	_, err := EvalString(`(import (nonexistent lib))`, e)
	if err == nil {
		t.Error("expected error for unknown library")
	}
}

func TestDefineLibraryInvalid(t *testing.T) {
	e := setupEval()
	_, err := EvalString(`(define-library)`, e)
	if err == nil {
		t.Error("expected error for empty define-library")
	}
}
