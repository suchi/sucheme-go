// Package evaluator - exception.go implements R7RS exception handling.
package evaluator

import (
	"fmt"

	"github.com/suchi/sucheme-go/pkg/env"
	"github.com/suchi/sucheme-go/pkg/types"
)

// SchemeException is used to propagate Scheme exceptions via Go panic/recover.
type SchemeException struct {
	Value *types.Value
}

// exceptionHandlerStack is a thread-local stack of exception handlers.
// In a single-threaded interpreter, a global stack suffices.
var exceptionHandlerStack []*types.Value

// RegisterExceptionBuiltins registers raise, raise-continuable,
// with-exception-handler, and error-object predicates.
func RegisterExceptionBuiltins(e *env.Env) {
	e.Define("raise", types.NewProcedure("raise", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		raiseException(args[0])
		return nil, nil // unreachable
	}))

	e.Define("raise-continuable", types.NewProcedure("raise-continuable", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		// For now, raise-continuable behaves like raise
		raiseException(args[0])
		return nil, nil // unreachable
	}))

	e.Define("with-exception-handler", types.NewProcedure("with-exception-handler", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("expected 2 arguments")
		}
		handler := args[0]
		thunk := args[1]
		if handler.Type != types.TypeProcedure || thunk.Type != types.TypeProcedure {
			return nil, fmt.Errorf("expected two procedures")
		}
		return withExceptionHandler(handler, thunk)
	}))

	e.Define("error", types.NewProcedure("error", func(args []*types.Value) (*types.Value, error) {
		if len(args) == 0 {
			raiseException(types.NewString("error"))
		}
		msg := args[0].Display()
		if len(args) > 1 {
			irritants := make([]*types.Value, len(args)-1)
			copy(irritants, args[1:])
			// Create an error object as a list: (error-object message irritants...)
			errObj := types.FromList(append([]*types.Value{
				types.NewSymbol("error-object"),
				types.NewString(msg),
			}, irritants...))
			raiseException(errObj)
		}
		raiseException(types.NewString(msg))
		return nil, nil // unreachable
	}))

	e.Define("error-object?", types.NewProcedure("error-object?", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		if args[0].Type == types.TypePair && args[0].PairVal.Car.Type == types.TypeSymbol &&
			args[0].PairVal.Car.SymbolVal == "error-object" {
			return types.True, nil
		}
		return types.False, nil
	}))

	e.Define("error-object-message", types.NewProcedure("error-object-message", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		if args[0].Type == types.TypePair && args[0].PairVal.Car.Type == types.TypeSymbol &&
			args[0].PairVal.Car.SymbolVal == "error-object" {
			lst, _ := args[0].ToList()
			if len(lst) >= 2 {
				return lst[1], nil
			}
		}
		return types.NewString(args[0].Display()), nil
	}))

	e.Define("error-object-irritants", types.NewProcedure("error-object-irritants", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		if args[0].Type == types.TypePair && args[0].PairVal.Car.Type == types.TypeSymbol &&
			args[0].PairVal.Car.SymbolVal == "error-object" {
			lst, _ := args[0].ToList()
			if len(lst) >= 3 {
				return types.FromList(lst[2:]), nil
			}
		}
		return types.Nil(), nil
	}))
}

// raiseException raises a Scheme exception by panicking with a SchemeException.
func raiseException(val *types.Value) {
	panic(SchemeException{Value: val})
}

// withExceptionHandler installs a handler and runs a thunk.
func withExceptionHandler(handler, thunk *types.Value) (result *types.Value, err error) {
	exceptionHandlerStack = append(exceptionHandlerStack, handler)

	defer func() {
		exceptionHandlerStack = exceptionHandlerStack[:len(exceptionHandlerStack)-1]
		if r := recover(); r != nil {
			if exc, ok := r.(SchemeException); ok {
				// Call the handler with the exception value
				handlerResult, herr := callProc(handler, []*types.Value{exc.Value})
				if herr != nil {
					err = herr
					return
				}
				result = handlerResult
				return
			}
			// Re-panic for non-SchemeException panics
			panic(r)
		}
	}()

	result, err = callProc(thunk, []*types.Value{})
	return
}

// callProc calls a procedure (builtin or lambda) with arguments.
func callProc(proc *types.Value, args []*types.Value) (*types.Value, error) {
	if proc.ProcVal.IsBuiltin {
		return proc.ProcVal.Builtin(args)
	}
	// Lambda: use applyProc + trampoline
	result, err := applyProc(proc, args)
	if err != nil {
		return nil, err
	}
	for result.thunk != nil {
		result, err = evalTCO(result.thunk.expr, result.thunk.env)
		if err != nil {
			return nil, err
		}
	}
	return result.value, nil
}

// evalGuard implements (guard (var clause ...) body ...)
func evalGuard(expr *types.Value, e *env.Env) (*evalResult, error) {
	args, err := listArgs(expr)
	if err != nil || len(args) < 2 {
		return nil, fmt.Errorf("invalid guard syntax")
	}

	// First arg: (var clause ...)
	guardSpec, err := args[0].ToList()
	if err != nil || len(guardSpec) < 2 {
		return nil, fmt.Errorf("guard: invalid clause specification")
	}
	if guardSpec[0].Type != types.TypeSymbol {
		return nil, fmt.Errorf("guard: variable must be a symbol")
	}
	varName := guardSpec[0].SymbolVal
	clauses := guardSpec[1:]
	body := args[1:]

	// Evaluate body with exception handling
	var bodyResult *types.Value
	var caughtException *types.Value

	func() {
		defer func() {
			if r := recover(); r != nil {
				if exc, ok := r.(SchemeException); ok {
					caughtException = exc.Value
					return
				}
				panic(r)
			}
		}()

		var evalErr error
		for _, bodyExpr := range body {
			bodyResult, evalErr = Eval(bodyExpr, e)
			if evalErr != nil {
				// Convert Go errors to scheme exceptions
				caughtException = types.NewString(evalErr.Error())
				return
			}
		}
	}()

	if caughtException == nil {
		return resultVal(bodyResult), nil
	}

	// Exception caught: evaluate clauses with var bound to exception
	guardEnv := env.New(e)
	guardEnv.Define(varName, caughtException)

	for i, clause := range clauses {
		clauseParts, err := clause.ToList()
		if err != nil || len(clauseParts) == 0 {
			return nil, fmt.Errorf("guard: invalid clause")
		}

		// Check for else
		if clauseParts[0].Type == types.TypeSymbol && clauseParts[0].SymbolVal == "else" {
			if i != len(clauses)-1 {
				return nil, fmt.Errorf("guard: else must be last clause")
			}
			if len(clauseParts) < 2 {
				return nil, fmt.Errorf("guard: else clause must have a body")
			}
			return evalBodyTCO(clauseParts[1:], guardEnv)
		}

		// Evaluate test
		test, err := Eval(clauseParts[0], guardEnv)
		if err != nil {
			return nil, err
		}
		if test.IsTrue() {
			if len(clauseParts) == 1 {
				return resultVal(test), nil
			}
			// => syntax
			if len(clauseParts) == 3 && clauseParts[1].Type == types.TypeSymbol && clauseParts[1].SymbolVal == "=>" {
				proc, err := Eval(clauseParts[2], guardEnv)
				if err != nil {
					return nil, err
				}
				return applyProc(proc, []*types.Value{test})
			}
			return evalBodyTCO(clauseParts[1:], guardEnv)
		}
	}

	// No clause matched: re-raise the exception
	raiseException(caughtException)
	return nil, nil // unreachable
}

// evalCaseLambda implements (case-lambda (formals body ...) ...)
func evalCaseLambda(expr *types.Value, e *env.Env) (*evalResult, error) {
	args, err := listArgs(expr)
	if err != nil || len(args) == 0 {
		return nil, fmt.Errorf("invalid case-lambda syntax")
	}

	// Store the clauses and environment
	type caseLambdaClause struct {
		params *types.Value
		body   []*types.Value
	}

	var clauses []caseLambdaClause
	for _, clause := range args {
		parts, err := clause.ToList()
		if err != nil || len(parts) < 2 {
			return nil, fmt.Errorf("case-lambda: each clause must have formals and body")
		}
		clauses = append(clauses, caseLambdaClause{
			params: parts[0],
			body:   parts[1:],
		})
	}

	// Create a procedure that dispatches based on argument count
	proc := types.NewProcedure("case-lambda", func(callArgs []*types.Value) (*types.Value, error) {
		for _, clause := range clauses {
			params := clause.params

			// Check if this clause's formals match the argument count
			if params.Type == types.TypeSymbol {
				// (rest-param body ...) - matches any number of args
				callEnv := env.New(e)
				callEnv.Define(params.SymbolVal, types.FromList(callArgs))
				var result *types.Value
				var err error
				for _, bodyExpr := range clause.body {
					result, err = Eval(bodyExpr, callEnv)
					if err != nil {
						return nil, err
					}
				}
				return result, nil
			}

			if params.Type == types.TypeNil {
				// () - matches 0 args
				if len(callArgs) == 0 {
					callEnv := env.New(e)
					var result *types.Value
					var err error
					for _, bodyExpr := range clause.body {
						result, err = Eval(bodyExpr, callEnv)
						if err != nil {
							return nil, err
						}
					}
					return result, nil
				}
				continue
			}

			// (a b c ...) - fixed params, possibly with rest
			paramList, err := params.ToList()
			if err != nil {
				// Dotted: (a b . rest)
				fixed, rest, perr := parseParamsDotted(params)
				if perr != nil {
					continue
				}
				if len(callArgs) >= len(fixed) {
					callEnv := env.New(e)
					for i, p := range fixed {
						callEnv.Define(p, callArgs[i])
					}
					if rest != "" {
						callEnv.Define(rest, types.FromList(callArgs[len(fixed):]))
					}
					var result *types.Value
					for _, bodyExpr := range clause.body {
						result, err = Eval(bodyExpr, callEnv)
						if err != nil {
							return nil, err
						}
					}
					return result, nil
				}
				continue
			}

			// Proper list params
			if len(callArgs) == len(paramList) {
				callEnv := env.New(e)
				for i, p := range paramList {
					if p.Type != types.TypeSymbol {
						break
					}
					callEnv.Define(p.SymbolVal, callArgs[i])
				}
				var result *types.Value
				for _, bodyExpr := range clause.body {
					result, err = Eval(bodyExpr, callEnv)
					if err != nil {
						return nil, err
					}
				}
				return result, nil
			}
		}
		return nil, fmt.Errorf("case-lambda: no matching clause for %d arguments", len(callArgs))
	})

	return resultVal(proc), nil
}
