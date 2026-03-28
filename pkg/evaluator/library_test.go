package evaluator

import (
	"testing"

	"github.com/suchi/sucheme-go/pkg/env"
)

// Reuses setupEval, evalHelper, evalErr from evaluator_test.go

// =========================================================================
// define-library + import
// =========================================================================

func TestDefineLibraryBasic(t *testing.T) {
	e := setupEval()
	_, err := EvalString(`
		(define-library (test lib)
			(import (scheme base))
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
			(import (scheme base))
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
			(import (scheme base))
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
			(import (scheme base))
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
			(import (scheme base))
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
			(import (scheme base))
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

func TestImportSchemeBase(t *testing.T) {
	// Set up the standard env first so (scheme base) has bindings
	base := setupEval()

	// Now create a fresh empty env to verify import actually brings bindings
	fresh := env.New(nil)
	RegisterLibraryBuiltins(base) // re-register so scheme base points to base

	// Import (scheme base) into fresh env
	_, err := EvalString(`(import (scheme base))`, fresh)
	if err != nil {
		t.Fatalf("import (scheme base) failed: %v", err)
	}

	// Verify that + is now available in fresh env
	result, err := EvalString(`(+ 1 2)`, fresh)
	if err != nil {
		t.Fatalf("(+ 1 2) after import failed: %v", err)
	}
	if result.IntVal != 3 {
		t.Errorf("expected 3, got %s", result.String())
	}
}

func TestDefineLibraryUnknownDeclaration(t *testing.T) {
	e := setupEval()
	_, err := EvalString(`
		(define-library (test bad)
			(unknown-decl 1 2 3))
	`, e)
	if err == nil {
		t.Error("expected error for unknown declaration")
	}
}

func TestDefineLibraryInvalidExportSpec(t *testing.T) {
	e := setupEval()
	_, err := EvalString(`
		(define-library (test bad-export)
			(export 42)
			(begin (define x 1)))
	`, e)
	if err == nil {
		t.Error("expected error for invalid export spec (number)")
	}
}
