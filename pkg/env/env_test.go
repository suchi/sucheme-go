package env

import (
	"testing"

	"github.com/suchi/sucheme-go/pkg/types"
)

func TestDefineAndGet(t *testing.T) {
	e := Global()
	e.Define("x", types.NewInteger(42))

	val, ok := e.Get("x")
	if !ok {
		t.Fatal("expected to find x")
	}
	if val.IntVal != 42 {
		t.Errorf("expected 42, got %v", val.IntVal)
	}
}

func TestGetUnbound(t *testing.T) {
	e := Global()
	_, ok := e.Get("nonexistent")
	if ok {
		t.Error("expected not to find nonexistent variable")
	}
}

func TestSetExisting(t *testing.T) {
	e := Global()
	e.Define("x", types.NewInteger(1))
	err := e.Set("x", types.NewInteger(2))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, _ := e.Get("x")
	if val.IntVal != 2 {
		t.Errorf("expected 2, got %v", val.IntVal)
	}
}

func TestSetUnbound(t *testing.T) {
	e := Global()
	err := e.Set("x", types.NewInteger(1))
	if err == nil {
		t.Error("expected error for set! on unbound variable")
	}
}

func TestNestedScopes(t *testing.T) {
	parent := Global()
	parent.Define("x", types.NewInteger(10))

	child := New(parent)
	child.Define("y", types.NewInteger(20))

	val, ok := child.Get("x")
	if !ok || val.IntVal != 10 {
		t.Error("child should see parent's x")
	}
	val, ok = child.Get("y")
	if !ok || val.IntVal != 20 {
		t.Error("child should see its own y")
	}
	_, ok = parent.Get("y")
	if ok {
		t.Error("parent should not see child's y")
	}
}

func TestSetInParentScope(t *testing.T) {
	parent := Global()
	parent.Define("x", types.NewInteger(10))

	child := New(parent)
	err := child.Set("x", types.NewInteger(99))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, _ := parent.Get("x")
	if val.IntVal != 99 {
		t.Errorf("expected 99, got %v", val.IntVal)
	}
}

func TestDefineShadows(t *testing.T) {
	parent := Global()
	parent.Define("x", types.NewInteger(10))

	child := New(parent)
	child.Define("x", types.NewInteger(20))

	val, _ := child.Get("x")
	if val.IntVal != 20 {
		t.Errorf("expected 20, got %v", val.IntVal)
	}
	val, _ = parent.Get("x")
	if val.IntVal != 10 {
		t.Errorf("expected 10, got %v", val.IntVal)
	}
}

func TestExtend(t *testing.T) {
	e := Global()
	e.Define("global", types.NewInteger(100))

	child, err := e.Extend([]string{"a", "b"}, []*types.Value{types.NewInteger(1), types.NewInteger(2)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, ok := child.Get("a")
	if !ok || val.IntVal != 1 {
		t.Error("expected a=1")
	}
	val, ok = child.Get("b")
	if !ok || val.IntVal != 2 {
		t.Error("expected b=2")
	}
	val, ok = child.Get("global")
	if !ok || val.IntVal != 100 {
		t.Error("expected global=100 visible in child")
	}
}

func TestExtendArityMismatch(t *testing.T) {
	e := Global()
	_, err := e.Extend([]string{"a", "b"}, []*types.Value{types.NewInteger(1)})
	if err == nil {
		t.Error("expected arity mismatch error")
	}
}

func TestDeeplyNestedScopes(t *testing.T) {
	root := Global()
	root.Define("root", types.NewInteger(0))
	current := root
	for i := 1; i <= 100; i++ {
		current = New(current)
	}
	val, ok := current.Get("root")
	if !ok || val.IntVal != 0 {
		t.Error("should find root binding through 100 levels")
	}
}

func TestDefineOverwrite(t *testing.T) {
	e := Global()
	e.Define("x", types.NewInteger(1))
	e.Define("x", types.NewInteger(2))
	val, _ := e.Get("x")
	if val.IntVal != 2 {
		t.Errorf("expected 2 after redefine, got %v", val.IntVal)
	}
}

func TestKeys(t *testing.T) {
	e := Global()
	e.Define("a", types.NewInteger(1))
	e.Define("b", types.NewInteger(2))
	keys := e.Keys()
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
	found := map[string]bool{}
	for _, k := range keys {
		found[k] = true
	}
	if !found["a"] || !found["b"] {
		t.Error("expected keys a and b")
	}
}
