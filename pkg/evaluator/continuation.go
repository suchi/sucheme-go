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
//
// In this escape-only continuation implementation, WindEntry and windStack
// record the nesting structure of dynamic-wind calls. The stack contents
// are not inspected when moving between dynamic extents (e.g. when invoking
// a captured continuation), so before/after thunks are not automatically
// re-run across such transitions.
//
// For full re-entrant continuations, this stack would be used to compute
// which before/after thunks need to be run when entering/leaving dynamic
// extents relative to a captured continuation.
type WindEntry struct {
	Before *types.Value // thunk
	After  *types.Value // thunk
}

// windStack tracks dynamic-wind entries in a LIFO manner.
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

	// Liveness flag: becomes false when callWithCC exits
	alive := true
	defer func() { alive = false }()

	// Create the continuation object (as a procedure)
	cont := types.NewProcedure("continuation", func(contArgs []*types.Value) (*types.Value, error) {
		if !alive {
			return nil, fmt.Errorf("continuation is no longer valid (escape-only)")
		}
		var v *types.Value
		switch len(contArgs) {
		case 0:
			v = types.NewValues([]*types.Value{})
		case 1:
			v = contArgs[0]
		default:
			v = types.NewValues(contArgs)
		}
		panic(ContinuationEscape{Value: v, Tag: tag})
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

		// Recover any panic from the body (including ContinuationEscape)
		var recovered interface{}
		if r := recover(); r != nil {
			recovered = r
		}

		// Always run the after thunk
		_, afterErr := callProc(after, []*types.Value{})

		if recovered != nil {
			if afterErr != nil {
				// Body panicked and after also failed: suppress panic, return after's error
				result = nil
				err = afterErr
				return
			}
			// After succeeded: re-panic the original value
			panic(recovered)
		}

		// No panic in body. Prefer after's error even if body also errored.
		if afterErr != nil {
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
