// Package evaluator - continuation.go implements call/cc and dynamic-wind.
//
// This implementation uses panic/recover for escape continuations.
// Re-entrant continuations (calling a captured continuation after the
// call/cc has returned) are not supported in this version.
// Full re-entrant continuations would require a CPS-transformed evaluator.
package evaluator

import (
	"fmt"

	"github.com/suchi/sucheme-go/pkg/env"
	"github.com/suchi/sucheme-go/pkg/types"
)

// ContinuationEscape is panic'd to escape to the call/cc capture point.
type ContinuationEscape struct {
	Value *types.Value
	Tag   uint64
}

var continuationCounter uint64

// WindEntry represents a before/after pair for dynamic-wind.
type WindEntry struct {
	Before *types.Value // thunk
	After  *types.Value // thunk
}

// windStack tracks dynamic-wind entries.
//
// NOTE: This stack is not thread-safe. It assumes a single-threaded evaluator.
var windStack []*WindEntry

// RegisterContinuationBuiltins registers call/cc, dynamic-wind, and related procedures.
func RegisterContinuationBuiltins(e *env.Env) {
	e.Define("call-with-current-continuation", types.NewProcedure("call-with-current-continuation",
		func(args []*types.Value) (*types.Value, error) {
			return callWithCC(args)
		}))

	e.Define("call/cc", types.NewProcedure("call/cc",
		func(args []*types.Value) (*types.Value, error) {
			return callWithCC(args)
		}))

	e.Define("dynamic-wind", types.NewProcedure("dynamic-wind",
		func(args []*types.Value) (*types.Value, error) {
			if len(args) != 3 {
				return nil, fmt.Errorf("expected 3 arguments")
			}
			for i, a := range args {
				if a.Type != types.TypeProcedure {
					return nil, fmt.Errorf("argument %d: expected procedure", i+1)
				}
			}
			return dynamicWind(args[0], args[1], args[2])
		}))

	e.Define("call-with-values", types.NewProcedure("call-with-values",
		func(args []*types.Value) (*types.Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("expected 2 arguments")
			}
			if args[0].Type != types.TypeProcedure || args[1].Type != types.TypeProcedure {
				return nil, fmt.Errorf("expected two procedures")
			}
			return callWithValues(args[0], args[1])
		}))
}

// callWithCC implements call-with-current-continuation.
// This is an escape-only implementation using panic/recover.
func callWithCC(args []*types.Value) (result *types.Value, err error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("expected 1 argument")
	}
	proc := args[0]
	if proc.Type != types.TypeProcedure {
		return nil, fmt.Errorf("expected a procedure")
	}

	continuationCounter++
	tag := continuationCounter

	// Create the continuation object (as a procedure)
	cont := types.NewProcedure("continuation", func(contArgs []*types.Value) (*types.Value, error) {
		if len(contArgs) != 1 {
			return nil, fmt.Errorf("continuation expects 1 argument")
		}
		// Escape to the call/cc capture point
		panic(ContinuationEscape{Value: contArgs[0], Tag: tag})
	})

	// Set up recover to catch the escape
	defer func() {
		if r := recover(); r != nil {
			if ce, ok := r.(ContinuationEscape); ok && ce.Tag == tag {
				result = ce.Value
				err = nil
				return
			}
			// Not our continuation or not a ContinuationEscape, re-panic
			panic(r)
		}
	}()

	// Call proc with the continuation
	result, err = callProc(proc, []*types.Value{cont})
	return
}

// dynamicWind implements (dynamic-wind before thunk after).
func dynamicWind(before, thunk, after *types.Value) (result *types.Value, err error) {
	// Call before thunk
	_, err = callProc(before, []*types.Value{})
	if err != nil {
		return nil, err
	}

	// Push wind entry
	entry := &WindEntry{Before: before, After: after}
	windStack = append(windStack, entry)

	// Call body thunk, ensuring after is called even on panic
	defer func() {
		windStack = windStack[:len(windStack)-1]
		_, afterErr := callProc(after, []*types.Value{})
		if err == nil && afterErr != nil {
			err = afterErr
		}
	}()

	result, err = callProc(thunk, []*types.Value{})
	return
}

// callWithValues implements (call-with-values producer consumer).
func callWithValues(producer, consumer *types.Value) (*types.Value, error) {
	produced, err := callProc(producer, []*types.Value{})
	if err != nil {
		return nil, err
	}

	// Check if produced is a multiple values object
	if produced.Type == types.TypeValues {
		return callProc(consumer, produced.ValuesVal)
	}

	// Single value
	return callProc(consumer, []*types.Value{produced})
}
