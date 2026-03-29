package env

import (
	"testing"

	"github.com/suchi/sucheme-go/pkg/types"
)

// =========================================================================
// Parent accessor
// =========================================================================

func TestParent(t *testing.T) {
	parent := Global()
	child := New(parent)
	if child.Parent() != parent {
		t.Error("Parent() should return the parent environment")
	}
}

func TestParentGlobalIsNil(t *testing.T) {
	global := Global()
	if global.Parent() != nil {
		t.Error("Global environment should have nil parent")
	}
}

// =========================================================================
// Variable shadowing chains
// =========================================================================

func TestTripleShadowing(t *testing.T) {
	root := Global()
	root.Define("x", types.NewInteger(1))
	child := New(root)
	child.Define("x", types.NewInteger(2))
	grandchild := New(child)
	grandchild.Define("x", types.NewInteger(3))

	val, _ := grandchild.Get("x")
	if val.IntVal != 3 {
		t.Errorf("expected 3 from grandchild, got %d", val.IntVal)
	}
	val, _ = child.Get("x")
	if val.IntVal != 2 {
		t.Errorf("expected 2 from child, got %d", val.IntVal)
	}
	val, _ = root.Get("x")
	if val.IntVal != 1 {
		t.Errorf("expected 1 from root, got %d", val.IntVal)
	}
}

// =========================================================================
// Set! through scope chain
// =========================================================================

func TestSetDeepScopeChain(t *testing.T) {
	root := Global()
	root.Define("x", types.NewInteger(0))
	child := New(root)
	grandchild := New(child)

	err := grandchild.Set("x", types.NewInteger(99))
	if err != nil {
		t.Fatal(err)
	}
	val, _ := root.Get("x")
	if val.IntVal != 99 {
		t.Errorf("expected 99, got %d", val.IntVal)
	}
}

func TestSetShadowedVariable(t *testing.T) {
	root := Global()
	root.Define("x", types.NewInteger(1))
	child := New(root)
	child.Define("x", types.NewInteger(2))

	err := child.Set("x", types.NewInteger(99))
	if err != nil {
		t.Fatal(err)
	}
	val, _ := child.Get("x")
	if val.IntVal != 99 {
		t.Errorf("expected 99 in child, got %d", val.IntVal)
	}
	val, _ = root.Get("x")
	if val.IntVal != 1 {
		t.Errorf("expected 1 in root, got %d", val.IntVal)
	}
}

func TestSetUnboundDeep(t *testing.T) {
	root := Global()
	child := New(root)
	grandchild := New(child)
	err := grandchild.Set("nonexistent", types.NewInteger(1))
	if err == nil {
		t.Error("expected error for set! on unbound variable in deep scope")
	}
}

// =========================================================================
// Extend edge cases
// =========================================================================

func TestExtendEmpty(t *testing.T) {
	e := Global()
	child, err := e.Extend(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(child.Keys()) != 0 {
		t.Error("expected empty keys for extend with no bindings")
	}
}

func TestExtendSingle(t *testing.T) {
	e := Global()
	child, err := e.Extend([]string{"x"}, []*types.Value{types.NewInteger(42)})
	if err != nil {
		t.Fatal(err)
	}
	val, ok := child.Get("x")
	if !ok || val.IntVal != 42 {
		t.Error("expected x=42")
	}
}

func TestExtendArityMismatchMore(t *testing.T) {
	e := Global()
	_, err := e.Extend([]string{"a"}, []*types.Value{types.NewInteger(1), types.NewInteger(2)})
	if err == nil {
		t.Error("expected arity mismatch error when values > params")
	}
}

// =========================================================================
// Keys with parent
// =========================================================================

func TestKeysDoesNotIncludeParent(t *testing.T) {
	parent := Global()
	parent.Define("a", types.NewInteger(1))
	child := New(parent)
	child.Define("b", types.NewInteger(2))

	keys := child.Keys()
	if len(keys) != 1 || keys[0] != "b" {
		t.Errorf("Keys() should only include own bindings, got %v", keys)
	}
}

func TestKeysEmpty(t *testing.T) {
	e := Global()
	keys := e.Keys()
	if len(keys) != 0 {
		t.Errorf("expected empty keys, got %v", keys)
	}
}

// =========================================================================
// Define various types
// =========================================================================

func TestDefineVariousTypes(t *testing.T) {
	e := Global()
	e.Define("int", types.NewInteger(42))
	e.Define("float", types.NewFloat(3.14))
	e.Define("str", types.NewString("hello"))
	e.Define("bool", types.True)
	e.Define("nil", types.Nil())

	val, _ := e.Get("int")
	if val.IntVal != 42 {
		t.Error("int mismatch")
	}
	val, _ = e.Get("float")
	if val.FloatVal != 3.14 {
		t.Error("float mismatch")
	}
	val, _ = e.Get("str")
	if val.StrVal != "hello" {
		t.Error("string mismatch")
	}
	val, _ = e.Get("bool")
	if !val.BoolVal {
		t.Error("bool mismatch")
	}
	val, _ = e.Get("nil")
	if val.Type != types.TypeNil {
		t.Error("nil mismatch")
	}
}

// =========================================================================
// Many bindings
// =========================================================================

func TestManyBindings(t *testing.T) {
	e := Global()
	for i := 0; i < 1000; i++ {
		name := "var" + types.NewInteger(int64(i)).String()
		e.Define(name, types.NewInteger(int64(i)))
	}
	for i := 0; i < 1000; i++ {
		name := "var" + types.NewInteger(int64(i)).String()
		val, ok := e.Get(name)
		if !ok || val.IntVal != int64(i) {
			t.Errorf("expected %d for %s, got %v", i, name, val)
		}
	}
}
