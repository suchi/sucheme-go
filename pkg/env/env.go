// Package env implements the Scheme environment (scope chain) for variable bindings.
package env

import (
	"fmt"

	"github.com/suchi/sucheme-go/pkg/types"
)

// Env represents a Scheme environment (scope chain).
type Env struct {
	bindings map[string]*types.Value
	parent   *Env
}

// New creates a new environment with the given parent.
func New(parent *Env) *Env {
	return &Env{
		bindings: make(map[string]*types.Value),
		parent:   parent,
	}
}

// Global creates a new global (top-level) environment with no parent.
func Global() *Env {
	return New(nil)
}

// Get looks up a variable in this environment or any parent scope.
func (e *Env) Get(name string) (*types.Value, bool) {
	if val, ok := e.bindings[name]; ok {
		return val, true
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	return nil, false
}

// Set modifies an existing binding in the nearest scope where it exists (set! semantics).
// Returns an error if the variable is not found in any scope.
func (e *Env) Set(name string, val *types.Value) error {
	if _, ok := e.bindings[name]; ok {
		e.bindings[name] = val
		return nil
	}
	if e.parent != nil {
		return e.parent.Set(name, val)
	}
	return fmt.Errorf("unbound variable: %s", name)
}

// Define creates or updates a binding in the current (innermost) frame.
func (e *Env) Define(name string, val *types.Value) {
	e.bindings[name] = val
}

// Extend creates a child environment binding params to args.
// Returns an error if the number of params and args do not match.
func (e *Env) Extend(params []string, args []*types.Value) (*Env, error) {
	if len(params) != len(args) {
		return nil, fmt.Errorf("expected %d arguments, got %d", len(params), len(args))
	}
	child := New(e)
	for i, p := range params {
		child.bindings[p] = args[i]
	}
	return child, nil
}

// Parent returns the parent environment, or nil for the global scope.
func (e *Env) Parent() *Env {
	return e.parent
}

// Keys returns all variable names defined in this frame (not parents).
func (e *Env) Keys() []string {
	keys := make([]string, 0, len(e.bindings))
	for k := range e.bindings {
		keys = append(keys, k)
	}
	return keys
}
