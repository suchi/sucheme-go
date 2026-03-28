// Package evaluator - exception.go implements R7RS exception handling.
package evaluator

import (
	"fmt"

	"github.com/suchi/sucheme-go/pkg/env"
	"github.com/suchi/sucheme-go/pkg/types"
)

// SchemeException is used to propagate Scheme exceptions via Go panic/recover.
type SchemeException struct {
	Value       *types.Value
	Continuable bool // true for raise-continuable, false for raise
}

// exceptionHandlerStack is a global stack of exception handlers.
// It is safe only for single-threaded use; concurrent evaluators must not share it.
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
		raiseContinuable(args[0])
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
			return nil, fmt.Errorf("expected at least 1 argument")
		}
		if args[0].Type != types.TypeString {
			return nil, fmt.Errorf("expected string message, got %s", args[0].Type)
		}
		msg := args[0].StrVal
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
		if args[0].Type != types.TypePair || args[0].PairVal.Car.Type != types.TypeSymbol ||
			args[0].PairVal.Car.SymbolVal != "error-object" {
			return nil, fmt.Errorf("expected error object")
		}
		lst, err := args[0].ToList()
		if err != nil || len(lst) < 2 {
			return nil, fmt.Errorf("invalid error object")
		}
		return lst[1], nil
	}))

	e.Define("error-object-irritants", types.NewProcedure("error-object-irritants", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		if args[0].Type != types.TypePair || args[0].PairVal.Car.Type != types.TypeSymbol ||
			args[0].PairVal.Car.SymbolVal != "error-object" {
			return nil, fmt.Errorf("expected error object")
		}
		lst, err := args[0].ToList()
		if err != nil {
			return nil, fmt.Errorf("invalid error object")
		}
		if len(lst) >= 3 {
			return types.FromList(lst[2:]), nil
		}
		return types.Nil(), nil
	}))
}

// raiseException raises a non-continuable Scheme exception.
func raiseException(val *types.Value) {
	panic(SchemeException{Value: val, Continuable: false})
}

// raiseContinuable raises a continuable Scheme exception.
func raiseContinuable(val *types.Value) {
	panic(SchemeException{Value: val, Continuable: true})
}

// withExceptionHandler installs a handler and runs a thunk.
// For non-continuable raise, the handler's return value causes a re-raise.
func withExceptionHandler(handler, thunk *types.Value) (result *types.Value, err error) {
	exceptionHandlerStack = append(exceptionHandlerStack, handler)

	defer func() {
		exceptionHandlerStack = exceptionHandlerStack[:len(exceptionHandlerStack)-1]
		if r := recover(); r != nil {
			if exc, ok := r.(SchemeException); ok {
				handlerResult, herr := callProc(handler, []*types.Value{exc.Value})
				if herr != nil {
					err = herr
					return
				}
				if exc.Continuable {
					// raise-continuable: handler return value is the result
					result = handlerResult
				} else {
					// raise (non-continuable): handler returned, which is an error
					raiseException(types.NewString("handler returned from non-continuable exception"))
				}
				return
			}
			panic(r)
		}
	}()

	result, err = callProc(thunk, []*types.Value{})
	return
}

// callProc calls a procedure (builtin or lambda) with arguments,
// using applyProc for consistent error formatting and TCO.
func callProc(proc *types.Value, args []*types.Value) (*types.Value, error) {
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
// Each clause is pre-validated and stored as a Lambda for TCO support.
func evalCaseLambda(expr *types.Value, e *env.Env) (*evalResult, error) {
	args, err := listArgs(expr)
	if err != nil || len(args) == 0 {
		return nil, fmt.Errorf("invalid case-lambda syntax")
	}

	// Pre-validate and compile all clauses into Lambdas
	var lambdas []*types.Lambda
	for _, clause := range args {
		parts, err := clause.ToList()
		if err != nil || len(parts) < 2 {
			return nil, fmt.Errorf("case-lambda: each clause must have formals and body")
		}
		formals := parts[0]
		body := parts[1:]

		lam := &types.Lambda{
			Body: body,
			Env:  e,
		}

		if formals.Type == types.TypeSymbol {
			// (rest body ...) - matches any number
			lam.Params = types.LambdaParams{IsRest: true, Rest: formals.SymbolVal}
		} else if formals.Type == types.TypeNil {
			// () - matches 0 args
			lam.Params = types.LambdaParams{Fixed: []string{}}
		} else {
			// (a b c) or (a b . rest)
			paramList, err := formals.ToList()
			if err != nil {
				// Dotted: (a b . rest)
				fixed, rest, perr := parseParamsDotted(formals)
				if perr != nil {
					return nil, fmt.Errorf("case-lambda: invalid formals")
				}
				lam.Params = types.LambdaParams{Fixed: fixed, Rest: rest, HasRest: rest != ""}
			} else {
				names := make([]string, len(paramList))
				for i, p := range paramList {
					if p.Type != types.TypeSymbol {
						return nil, fmt.Errorf("case-lambda: formal must be a symbol, got %s", p.Type)
					}
					names[i] = p.SymbolVal
				}
				lam.Params = types.LambdaParams{Fixed: names}
			}
		}
		lambdas = append(lambdas, lam)
	}

	// Create a procedure that dispatches based on argument count
	proc := types.NewProcedure("case-lambda", func(callArgs []*types.Value) (*types.Value, error) {
		for _, lam := range lambdas {
			if lam.Params.IsRest {
				// Matches any number of args
				lamProc := types.NewLambdaProcedure("case-lambda", lam)
				return callProc(lamProc, callArgs)
			}
			if lam.Params.HasRest {
				// (a b . rest) - at least len(Fixed) args
				if len(callArgs) >= len(lam.Params.Fixed) {
					lamProc := types.NewLambdaProcedure("case-lambda", lam)
					return callProc(lamProc, callArgs)
				}
				continue
			}
			// Fixed args only
			if len(callArgs) == len(lam.Params.Fixed) {
				lamProc := types.NewLambdaProcedure("case-lambda", lam)
				return callProc(lamProc, callArgs)
			}
		}
		return nil, fmt.Errorf("case-lambda: no matching clause for %d arguments", len(callArgs))
	})

	return resultVal(proc), nil
}
