// Package evaluator - parameter.go implements R7RS parameters (make-parameter, parameterize).
package evaluator

import (
	"fmt"

	"github.com/suchi/sucheme-go/pkg/env"
	"github.com/suchi/sucheme-go/pkg/types"
)

// Parameter represents a dynamically-bound parameter object.
type Parameter struct {
	value     *types.Value
	converter func(*types.Value) (*types.Value, error) // optional converter
}

// parameterStack tracks dynamic parameter bindings.
// Each entry maps parameter identity to its current value.
//
// NOTE: This stack is not thread-safe. It assumes a single-threaded evaluator
// and must not be shared across concurrent evaluators, similar to the
// exception handler stack in exception.go.
var parameterStack []map[*Parameter]*types.Value

// RegisterParameterBuiltins registers make-parameter and related procedures.
func RegisterParameterBuiltins(e *env.Env) {
	e.Define("make-parameter", types.NewProcedure("make-parameter", func(args []*types.Value) (*types.Value, error) {
		if len(args) < 1 || len(args) > 2 {
			return nil, fmt.Errorf("expected 1 or 2 arguments")
		}

		param := &Parameter{value: args[0]}

		if len(args) == 2 {
			if args[1].Type != types.TypeProcedure {
				return nil, fmt.Errorf("expected procedure for converter")
			}
			converter := args[1]
			param.converter = func(v *types.Value) (*types.Value, error) {
				return callProc(converter, []*types.Value{v})
			}
			// Apply converter to initial value
			converted, err := param.converter(args[0])
			if err != nil {
				return nil, err
			}
			param.value = converted
		}

		// A parameter object is a procedure: 0 args = get, 1 arg = set
		proc := types.NewProcedure("parameter", func(callArgs []*types.Value) (*types.Value, error) {
			if len(callArgs) == 0 {
				// Look up in dynamic parameter stack first
				for i := len(parameterStack) - 1; i >= 0; i-- {
					if v, ok := parameterStack[i][param]; ok {
						return v, nil
					}
				}
				return param.value, nil
			}
			if len(callArgs) == 1 {
				val := callArgs[0]
				if param.converter != nil {
					var err error
					val, err = param.converter(val)
					if err != nil {
						return nil, err
					}
				}
				// Set in current dynamic frame if exists, else set base value
				for i := len(parameterStack) - 1; i >= 0; i-- {
					if _, ok := parameterStack[i][param]; ok {
						parameterStack[i][param] = val
						return types.Void(), nil
					}
				}
				param.value = val
				return types.Void(), nil
			}
			return nil, fmt.Errorf("expected 0 or 1 arguments")
		})

		// Store the Parameter pointer in the procedure for parameterize to find
		proc.ProcVal.ParamRef = param

		return proc, nil
	}))
}

// evalParameterize implements (parameterize ((param val) ...) body ...)
func evalParameterize(expr *types.Value, e *env.Env) (*evalResult, error) {
	args, err := listArgs(expr)
	if err != nil || len(args) < 2 {
		return nil, fmt.Errorf("parameterize requires bindings and body")
	}

	bindings, err := args[0].ToList()
	if err != nil {
		return nil, fmt.Errorf("parameterize: invalid bindings")
	}

	// Build parameter frame
	frame := make(map[*Parameter]*types.Value)

	for _, b := range bindings {
		pair, err := b.ToList()
		if err != nil || len(pair) != 2 {
			return nil, fmt.Errorf("parameterize: each binding must be (param expr)")
		}

		// Evaluate the parameter expression to get the parameter object
		paramVal, err := Eval(pair[0], e)
		if err != nil {
			return nil, err
		}
		if paramVal.Type != types.TypeProcedure || paramVal.ProcVal.ParamRef == nil {
			return nil, fmt.Errorf("parameterize: expected parameter object")
		}
		param, ok := paramVal.ProcVal.ParamRef.(*Parameter)
		if !ok {
			return nil, fmt.Errorf("parameterize: invalid parameter object")
		}

		// Evaluate the value expression
		val, err := Eval(pair[1], e)
		if err != nil {
			return nil, err
		}

		// Apply converter if present
		if param.converter != nil {
			val, err = param.converter(val)
			if err != nil {
				return nil, err
			}
		}

		frame[param] = val
	}

	// Push frame, evaluate body, ensure frame is popped even on panic
	parameterStack = append(parameterStack, frame)
	defer func() {
		parameterStack = parameterStack[:len(parameterStack)-1]
	}()

	var result *types.Value
	for _, bodyExpr := range args[1:] {
		var evalErr error
		result, evalErr = Eval(bodyExpr, e)
		if evalErr != nil {
			return nil, evalErr
		}
	}

	return resultVal(result), nil
}
