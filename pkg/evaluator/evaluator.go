// Package evaluator implements a tree-walking Scheme evaluator with trampoline-based TCO.
package evaluator

import (
	"fmt"

	"github.com/suchi/sucheme-go/pkg/builtin"
	"github.com/suchi/sucheme-go/pkg/env"
	"github.com/suchi/sucheme-go/pkg/parser"
	"github.com/suchi/sucheme-go/pkg/types"
)

// thunk represents a deferred evaluation for TCO.
type thunk struct {
	expr *types.Value
	env  *env.Env
}

// evalResult holds either a final value or a thunk to continue evaluation.
type evalResult struct {
	value *types.Value
	thunk *thunk
}

func resultVal(v *types.Value) *evalResult {
	return &evalResult{value: v}
}

func resultThunk(expr *types.Value, e *env.Env) *evalResult {
	return &evalResult{thunk: &thunk{expr: expr, env: e}}
}

// Eval evaluates a Scheme expression in the given environment.
func Eval(expr *types.Value, e *env.Env) (*types.Value, error) {
	result, err := evalTCO(expr, e)
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

func evalTCO(expr *types.Value, e *env.Env) (*evalResult, error) {
	switch expr.Type {
	case types.TypeInteger, types.TypeFloat, types.TypeBoolean, types.TypeString,
		types.TypeCharacter, types.TypeVector, types.TypeVoid, types.TypeEof,
		types.TypeBytevector:
		return resultVal(expr), nil

	case types.TypeSymbol:
		val, ok := e.Get(expr.SymbolVal)
		if !ok {
			return nil, fmt.Errorf("unbound variable: %s", expr.SymbolVal)
		}
		if val.Type == types.TypeUndefined {
			return nil, fmt.Errorf("uninitialized letrec binding: %s", expr.SymbolVal)
		}
		return resultVal(val), nil

	case types.TypeNil:
		return nil, fmt.Errorf("cannot evaluate empty list")

	case types.TypePair:
		return evalList(expr, e)

	default:
		return resultVal(expr), nil
	}
}

func evalList(expr *types.Value, e *env.Env) (*evalResult, error) {
	head := expr.PairVal.Car

	if head.Type == types.TypeSymbol {
		switch head.SymbolVal {
		case "quote":
			return evalQuote(expr)
		case "if":
			return evalIf(expr, e)
		case "cond":
			return evalCond(expr, e)
		case "define":
			return evalDefine(expr, e)
		case "set!":
			return evalSet(expr, e)
		case "lambda":
			return evalLambda(expr, e)
		case "begin":
			return evalBegin(expr, e)
		case "let":
			return evalLet(expr, e)
		case "let*":
			return evalLetStar(expr, e)
		case "letrec":
			return evalLetrec(expr, e)
		case "letrec*":
			return evalLetrecStar(expr, e)
		case "and":
			return evalAnd(expr, e)
		case "or":
			return evalOr(expr, e)
		case "when":
			return evalWhen(expr, e)
		case "unless":
			return evalUnless(expr, e)
		case "do":
			return evalDo(expr, e)
		case "quasiquote":
			return evalQuasiquote(expr, e)
		case "define-syntax":
			return evalDefineSyntax(expr, e)
		case "let-syntax":
			return evalLetSyntax(expr, e)
		case "letrec-syntax":
			return evalLetrecSyntax(expr, e)
		case "syntax-rules":
			return nil, fmt.Errorf("syntax-rules can only appear inside define-syntax, let-syntax, or letrec-syntax")
		case "guard":
			return evalGuard(expr, e)
		case "case-lambda":
			return evalCaseLambda(expr, e)
		case "parameterize":
			return evalParameterize(expr, e)
		case "define-library":
			return evalDefineLibrary(expr, e)
		case "import":
			return evalImport(expr, e)
		}

		// Check if head is a macro
		if val, ok := e.Get(head.SymbolVal); ok && val.Type == types.TypeSyntaxRules {
			expanded, err := expandMacro(val.SyntaxVal, expr)
			if err != nil {
				return nil, err
			}
			// Re-evaluate the expanded form (TCO)
			return resultThunk(expanded, e), nil
		}
	}

	return evalApplication(expr, e)
}

func listArgs(expr *types.Value) ([]*types.Value, error) {
	return expr.PairVal.Cdr.ToList()
}

func evalQuote(expr *types.Value) (*evalResult, error) {
	args, err := listArgs(expr)
	if err != nil || len(args) != 1 {
		return nil, fmt.Errorf("quote requires exactly one argument")
	}
	return resultVal(args[0]), nil
}

func evalIf(expr *types.Value, e *env.Env) (*evalResult, error) {
	args, err := listArgs(expr)
	if err != nil || (len(args) != 2 && len(args) != 3) {
		return nil, fmt.Errorf("if requires 2 or 3 arguments")
	}
	cond, err := Eval(args[0], e)
	if err != nil {
		return nil, err
	}
	if cond.IsTrue() {
		return resultThunk(args[1], e), nil
	}
	if len(args) == 3 {
		return resultThunk(args[2], e), nil
	}
	return resultVal(types.Void()), nil
}

func evalCond(expr *types.Value, e *env.Env) (*evalResult, error) {
	clauses, err := listArgs(expr)
	if err != nil {
		return nil, fmt.Errorf("invalid cond syntax")
	}
	for i, clause := range clauses {
		parts, err := clause.ToList()
		if err != nil || len(parts) == 0 {
			return nil, fmt.Errorf("invalid cond clause")
		}
		if parts[0].Type == types.TypeSymbol && parts[0].SymbolVal == "else" {
			if i != len(clauses)-1 {
				return nil, fmt.Errorf("else clause must be last in cond")
			}
			if len(parts) < 2 {
				return nil, fmt.Errorf("else clause in cond must have at least one expression")
			}
			return evalBodyTCO(parts[1:], e)
		}
		test, err := Eval(parts[0], e)
		if err != nil {
			return nil, err
		}
		if test.IsTrue() {
			if len(parts) == 3 && parts[1].Type == types.TypeSymbol && parts[1].SymbolVal == "=>" {
				proc, err := Eval(parts[2], e)
				if err != nil {
					return nil, err
				}
				return applyProc(proc, []*types.Value{test})
			}
			if len(parts) == 1 {
				return resultVal(test), nil
			}
			return evalBodyTCO(parts[1:], e)
		}
	}
	return resultVal(types.Void()), nil
}

func evalDefine(expr *types.Value, e *env.Env) (*evalResult, error) {
	args, err := listArgs(expr)
	if err != nil || len(args) < 2 {
		return nil, fmt.Errorf("invalid define syntax")
	}

	// (define (f params...) body...)
	if args[0].Type == types.TypePair {
		nameVal := args[0].PairVal.Car
		if nameVal.Type != types.TypeSymbol {
			return nil, fmt.Errorf("define: expected symbol")
		}
		paramList, err := args[0].PairVal.Cdr.ToList()
		if err != nil {
			// Check for dotted rest param
			params, rest, perr := parseParamsDotted(args[0].PairVal.Cdr)
			if perr != nil {
				return nil, fmt.Errorf("invalid parameter list")
			}
			lam := &types.Lambda{
				Params: types.LambdaParams{Fixed: params, Rest: rest, HasRest: rest != ""},
				Body:   args[1:],
				Env:    e,
				Name:   nameVal.SymbolVal,
			}
			e.Define(nameVal.SymbolVal, types.NewLambdaProcedure(nameVal.SymbolVal, lam))
			return resultVal(types.Void()), nil
		}
		params := make([]string, len(paramList))
		for i, p := range paramList {
			if p.Type != types.TypeSymbol {
				return nil, fmt.Errorf("parameter must be a symbol")
			}
			params[i] = p.SymbolVal
		}
		lam := &types.Lambda{
			Params: types.LambdaParams{Fixed: params},
			Body:   args[1:],
			Env:    e,
			Name:   nameVal.SymbolVal,
		}
		e.Define(nameVal.SymbolVal, types.NewLambdaProcedure(nameVal.SymbolVal, lam))
		return resultVal(types.Void()), nil
	}

	// (define x expr)
	if args[0].Type != types.TypeSymbol {
		return nil, fmt.Errorf("define: expected symbol")
	}
	if len(args) != 2 {
		return nil, fmt.Errorf("define: variable definition requires exactly one expression")
	}
	val, err := Eval(args[1], e)
	if err != nil {
		return nil, err
	}
	if val.Type == types.TypeProcedure && val.ProcVal != nil && !val.ProcVal.IsBuiltin && val.ProcVal.Name == "" {
		val.ProcVal.Name = args[0].SymbolVal
	}
	e.Define(args[0].SymbolVal, val)
	return resultVal(types.Void()), nil
}

func evalSet(expr *types.Value, e *env.Env) (*evalResult, error) {
	args, err := listArgs(expr)
	if err != nil || len(args) != 2 {
		return nil, fmt.Errorf("set! requires exactly 2 arguments")
	}
	if args[0].Type != types.TypeSymbol {
		return nil, fmt.Errorf("set!: expected symbol")
	}
	val, err := Eval(args[1], e)
	if err != nil {
		return nil, err
	}
	if err := e.Set(args[0].SymbolVal, val); err != nil {
		return nil, err
	}
	return resultVal(types.Void()), nil
}

func evalLambda(expr *types.Value, e *env.Env) (*evalResult, error) {
	args, err := listArgs(expr)
	if err != nil || len(args) < 2 {
		return nil, fmt.Errorf("lambda requires parameters and body")
	}
	params, rest, err := parseParams(args[0])
	if err != nil {
		return nil, err
	}
	lp := types.LambdaParams{Fixed: params, Rest: rest, HasRest: rest != ""}
	if len(params) == 0 && rest != "" {
		lp.IsRest = true
	}
	lam := &types.Lambda{
		Params: lp,
		Body:   args[1:],
		Env:    e,
	}
	return resultVal(types.NewLambdaProcedure("", lam)), nil
}

func parseParams(paramExpr *types.Value) ([]string, string, error) {
	if paramExpr.Type == types.TypeSymbol {
		return nil, paramExpr.SymbolVal, nil
	}
	if paramExpr.Type == types.TypeNil {
		return nil, "", nil
	}
	return parseParamsDotted(paramExpr)
}

func parseParamsDotted(paramExpr *types.Value) ([]string, string, error) {
	var params []string
	cur := paramExpr
	for cur.Type == types.TypePair {
		if cur.PairVal.Car.Type != types.TypeSymbol {
			return nil, "", fmt.Errorf("parameter must be a symbol")
		}
		params = append(params, cur.PairVal.Car.SymbolVal)
		cur = cur.PairVal.Cdr
	}
	if cur.Type == types.TypeSymbol {
		return params, cur.SymbolVal, nil
	}
	if cur.Type != types.TypeNil {
		return nil, "", fmt.Errorf("invalid parameter list")
	}
	return params, "", nil
}

func evalBegin(expr *types.Value, e *env.Env) (*evalResult, error) {
	args, err := listArgs(expr)
	if err != nil {
		return nil, fmt.Errorf("invalid begin syntax")
	}
	if len(args) == 0 {
		return resultVal(types.Void()), nil
	}
	return evalBodyTCO(args, e)
}

func evalLet(expr *types.Value, e *env.Env) (*evalResult, error) {
	args, err := listArgs(expr)
	if err != nil || len(args) < 2 {
		return nil, fmt.Errorf("invalid let syntax")
	}

	// Named let: (let name ((var init) ...) body ...)
	if args[0].Type == types.TypeSymbol {
		return evalNamedLet(args[0].SymbolVal, args[1:], e)
	}

	bindings, err := args[0].ToList()
	if err != nil {
		return nil, fmt.Errorf("invalid let bindings")
	}
	child := env.New(e)
	for _, b := range bindings {
		pair, err := b.ToList()
		if err != nil || len(pair) != 2 {
			return nil, fmt.Errorf("invalid let binding")
		}
		if pair[0].Type != types.TypeSymbol {
			return nil, fmt.Errorf("let: binding name must be a symbol")
		}
		val, err := Eval(pair[1], e)
		if err != nil {
			return nil, err
		}
		child.Define(pair[0].SymbolVal, val)
	}
	return evalBodyTCO(args[1:], child)
}

func evalNamedLet(name string, args []*types.Value, e *env.Env) (*evalResult, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("invalid named let syntax")
	}
	bindings, err := args[0].ToList()
	if err != nil {
		return nil, fmt.Errorf("invalid named let bindings")
	}
	params := make([]string, len(bindings))
	initVals := make([]*types.Value, len(bindings))
	for i, b := range bindings {
		pair, err := b.ToList()
		if err != nil || len(pair) != 2 {
			return nil, fmt.Errorf("invalid named let binding")
		}
		if pair[0].Type != types.TypeSymbol {
			return nil, fmt.Errorf("named let: binding name must be a symbol")
		}
		params[i] = pair[0].SymbolVal
		val, err := Eval(pair[1], e)
		if err != nil {
			return nil, err
		}
		initVals[i] = val
	}
	body := args[1:]
	lam := &types.Lambda{
		Params: types.LambdaParams{Fixed: params},
		Body:   body,
		Name:   name,
	}
	lambdaVal := types.NewLambdaProcedure(name, lam)
	letEnv := env.New(e)
	letEnv.Define(name, lambdaVal)
	lam.Env = letEnv
	return applyProc(lambdaVal, initVals)
}

func evalLetStar(expr *types.Value, e *env.Env) (*evalResult, error) {
	args, err := listArgs(expr)
	if err != nil || len(args) < 2 {
		return nil, fmt.Errorf("invalid let* syntax")
	}
	bindings, err := args[0].ToList()
	if err != nil {
		return nil, fmt.Errorf("invalid let* bindings")
	}
	child := env.New(e)
	for _, b := range bindings {
		pair, err := b.ToList()
		if err != nil || len(pair) != 2 {
			return nil, fmt.Errorf("invalid let* binding")
		}
		if pair[0].Type != types.TypeSymbol {
			return nil, fmt.Errorf("let*: binding name must be a symbol")
		}
		val, err := Eval(pair[1], child)
		if err != nil {
			return nil, err
		}
		child.Define(pair[0].SymbolVal, val)
	}
	return evalBodyTCO(args[1:], child)
}

func evalLetrec(expr *types.Value, e *env.Env) (*evalResult, error) {
	args, err := listArgs(expr)
	if err != nil || len(args) < 2 {
		return nil, fmt.Errorf("invalid letrec syntax")
	}
	bindings, err := args[0].ToList()
	if err != nil {
		return nil, fmt.Errorf("invalid letrec bindings")
	}
	child := env.New(e)
	names := make([]string, len(bindings))
	initExprs := make([]*types.Value, len(bindings))
	for i, b := range bindings {
		pair, err := b.ToList()
		if err != nil || len(pair) != 2 {
			return nil, fmt.Errorf("invalid letrec binding")
		}
		if pair[0].Type != types.TypeSymbol {
			return nil, fmt.Errorf("letrec: binding name must be a symbol")
		}
		names[i] = pair[0].SymbolVal
		initExprs[i] = pair[1]
		child.Define(names[i], types.Undefined())
	}
	for i, initExpr := range initExprs {
		val, err := Eval(initExpr, child)
		if err != nil {
			return nil, err
		}
		child.Define(names[i], val)
	}
	return evalBodyTCO(args[1:], child)
}

func evalLetrecStar(expr *types.Value, e *env.Env) (*evalResult, error) {
	args, err := listArgs(expr)
	if err != nil || len(args) < 2 {
		return nil, fmt.Errorf("invalid letrec* syntax")
	}
	bindings, err := args[0].ToList()
	if err != nil {
		return nil, fmt.Errorf("invalid letrec* bindings")
	}
	child := env.New(e)
	// Parse bindings once and reuse
	type binding struct {
		name string
		init *types.Value
	}
	parsed := make([]binding, len(bindings))
	for i, b := range bindings {
		pair, err := b.ToList()
		if err != nil || len(pair) != 2 {
			return nil, fmt.Errorf("invalid letrec* binding")
		}
		if pair[0].Type != types.TypeSymbol {
			return nil, fmt.Errorf("letrec*: binding name must be a symbol")
		}
		parsed[i] = binding{name: pair[0].SymbolVal, init: pair[1]}
		child.Define(pair[0].SymbolVal, types.Undefined())
	}
	for _, b := range parsed {
		val, err := Eval(b.init, child)
		if err != nil {
			return nil, err
		}
		child.Define(b.name, val)
	}
	return evalBodyTCO(args[1:], child)
}

func evalAnd(expr *types.Value, e *env.Env) (*evalResult, error) {
	args, err := listArgs(expr)
	if err != nil {
		return nil, fmt.Errorf("invalid and syntax")
	}
	if len(args) == 0 {
		return resultVal(types.NewBoolean(true)), nil
	}
	for i := 0; i < len(args)-1; i++ {
		val, err := Eval(args[i], e)
		if err != nil {
			return nil, err
		}
		if !val.IsTrue() {
			return resultVal(val), nil
		}
	}
	return resultThunk(args[len(args)-1], e), nil
}

func evalOr(expr *types.Value, e *env.Env) (*evalResult, error) {
	args, err := listArgs(expr)
	if err != nil {
		return nil, fmt.Errorf("invalid or syntax")
	}
	if len(args) == 0 {
		return resultVal(types.NewBoolean(false)), nil
	}
	for i := 0; i < len(args)-1; i++ {
		val, err := Eval(args[i], e)
		if err != nil {
			return nil, err
		}
		if val.IsTrue() {
			return resultVal(val), nil
		}
	}
	return resultThunk(args[len(args)-1], e), nil
}

func evalWhen(expr *types.Value, e *env.Env) (*evalResult, error) {
	args, err := listArgs(expr)
	if err != nil || len(args) < 2 {
		return nil, fmt.Errorf("when requires a test and body")
	}
	test, err := Eval(args[0], e)
	if err != nil {
		return nil, err
	}
	if test.IsTrue() {
		return evalBodyTCO(args[1:], e)
	}
	return resultVal(types.Void()), nil
}

func evalUnless(expr *types.Value, e *env.Env) (*evalResult, error) {
	args, err := listArgs(expr)
	if err != nil || len(args) < 2 {
		return nil, fmt.Errorf("unless requires a test and body")
	}
	test, err := Eval(args[0], e)
	if err != nil {
		return nil, err
	}
	if !test.IsTrue() {
		return evalBodyTCO(args[1:], e)
	}
	return resultVal(types.Void()), nil
}

func evalDo(expr *types.Value, e *env.Env) (*evalResult, error) {
	args, err := listArgs(expr)
	if err != nil || len(args) < 2 {
		return nil, fmt.Errorf("invalid do syntax")
	}
	varClauses, err := args[0].ToList()
	if err != nil {
		return nil, fmt.Errorf("invalid do variable clauses")
	}
	testClause, err := args[1].ToList()
	if err != nil || len(testClause) == 0 {
		return nil, fmt.Errorf("invalid do test clause")
	}
	commands := args[2:]

	type doVar struct {
		name     string
		stepExpr *types.Value
	}
	vars := make([]doVar, len(varClauses))
	doEnv := env.New(e)
	for i, vc := range varClauses {
		parts, err := vc.ToList()
		if err != nil || len(parts) < 2 {
			return nil, fmt.Errorf("invalid do variable clause")
		}
		if parts[0].Type != types.TypeSymbol {
			return nil, fmt.Errorf("do: variable name must be a symbol")
		}
		vars[i].name = parts[0].SymbolVal
		initVal, err := Eval(parts[1], e)
		if err != nil {
			return nil, err
		}
		doEnv.Define(vars[i].name, initVal)
		if len(parts) >= 3 {
			vars[i].stepExpr = parts[2]
		}
	}

	for {
		testResult, err := Eval(testClause[0], doEnv)
		if err != nil {
			return nil, err
		}
		if testResult.IsTrue() {
			if len(testClause) > 1 {
				return evalBodyTCO(testClause[1:], doEnv)
			}
			return resultVal(types.Void()), nil
		}
		for _, cmd := range commands {
			_, err := Eval(cmd, doEnv)
			if err != nil {
				return nil, err
			}
		}
		newVals := make([]*types.Value, len(vars))
		for i, v := range vars {
			if v.stepExpr != nil {
				newVals[i], err = Eval(v.stepExpr, doEnv)
				if err != nil {
					return nil, err
				}
			} else {
				val, _ := doEnv.Get(v.name)
				newVals[i] = val
			}
		}
		for i, v := range vars {
			doEnv.Define(v.name, newVals[i])
		}
	}
}

func evalQuasiquote(expr *types.Value, e *env.Env) (*evalResult, error) {
	args, err := listArgs(expr)
	if err != nil || len(args) != 1 {
		return nil, fmt.Errorf("quasiquote requires exactly one argument")
	}
	result, err := expandQuasiquote(args[0], e)
	if err != nil {
		return nil, err
	}
	return resultVal(result), nil
}

// ---------------------------------------------------------------------------
// Macro special forms
// ---------------------------------------------------------------------------

func evalDefineSyntax(expr *types.Value, e *env.Env) (*evalResult, error) {
	args, err := listArgs(expr)
	if err != nil || len(args) != 2 {
		return nil, fmt.Errorf("define-syntax requires name and transformer")
	}
	if args[0].Type != types.TypeSymbol {
		return nil, fmt.Errorf("define-syntax: name must be a symbol")
	}
	name := args[0].SymbolVal

	// The transformer must be a syntax-rules form
	if args[1].Type != types.TypePair || args[1].PairVal.Car.Type != types.TypeSymbol ||
		args[1].PairVal.Car.SymbolVal != "syntax-rules" {
		return nil, fmt.Errorf("define-syntax: transformer must be syntax-rules")
	}

	srArgs, err := args[1].PairVal.Cdr.ToList()
	if err != nil {
		return nil, fmt.Errorf("define-syntax: invalid syntax-rules form")
	}
	sr, err := compileSyntaxRules(srArgs, e)
	if err != nil {
		return nil, fmt.Errorf("define-syntax %s: %w", name, err)
	}
	sr.Name = name

	e.Define(name, types.NewSyntaxRules(sr))
	return resultVal(types.Void()), nil
}

func evalLetSyntax(expr *types.Value, e *env.Env) (*evalResult, error) {
	return evalLetSyntaxCommon(expr, e, false)
}

func evalLetrecSyntax(expr *types.Value, e *env.Env) (*evalResult, error) {
	return evalLetSyntaxCommon(expr, e, true)
}

func evalLetSyntaxCommon(expr *types.Value, e *env.Env, isRec bool) (*evalResult, error) {
	formName := "let-syntax"
	if isRec {
		formName = "letrec-syntax"
	}
	args, err := listArgs(expr)
	if err != nil || len(args) < 2 {
		return nil, fmt.Errorf("%s requires bindings and body", formName)
	}

	bindings, err := args[0].ToList()
	if err != nil {
		return nil, fmt.Errorf("%s: invalid bindings", formName)
	}

	child := env.New(e)

	for _, b := range bindings {
		pair, err := b.ToList()
		if err != nil || len(pair) != 2 {
			return nil, fmt.Errorf("%s: invalid binding", formName)
		}
		if pair[0].Type != types.TypeSymbol {
			return nil, fmt.Errorf("%s: binding name must be a symbol", formName)
		}
		name := pair[0].SymbolVal

		if pair[1].Type != types.TypePair || pair[1].PairVal.Car.Type != types.TypeSymbol ||
			pair[1].PairVal.Car.SymbolVal != "syntax-rules" {
			return nil, fmt.Errorf("%s: transformer must be syntax-rules", formName)
		}

		srArgs, err := pair[1].PairVal.Cdr.ToList()
		if err != nil {
			return nil, fmt.Errorf("%s: invalid syntax-rules form", formName)
		}
		defEnv := e
		if isRec {
			defEnv = child
		}
		sr, err := compileSyntaxRules(srArgs, defEnv)
		if err != nil {
			return nil, fmt.Errorf("%s %s: %w", formName, name, err)
		}
		sr.Name = name
		child.Define(name, types.NewSyntaxRules(sr))
	}

	return evalBodyTCO(args[1:], child)
}

func expandQuasiquote(expr *types.Value, e *env.Env) (*types.Value, error) {
	return expandQQ(expr, e, 1)
}

func expandQQ(expr *types.Value, e *env.Env, depth int) (*types.Value, error) {
	if expr.Type == types.TypePair {
		car := expr.PairVal.Car
		if car.Type == types.TypeSymbol && car.SymbolVal == "unquote" {
			if depth == 1 {
				args, err := expr.PairVal.Cdr.ToList()
				if err != nil || len(args) != 1 {
					return nil, fmt.Errorf("unquote requires exactly one argument")
				}
				return Eval(args[0], e)
			}
			// Nested unquote: decrease depth and recurse
			inner, err := expandQQ(expr.PairVal.Cdr, e, depth-1)
			if err != nil {
				return nil, err
			}
			return types.Cons(car, inner), nil
		}
		if car.Type == types.TypeSymbol && car.SymbolVal == "quasiquote" {
			// Nested quasiquote: increase depth
			inner, err := expandQQ(expr.PairVal.Cdr, e, depth+1)
			if err != nil {
				return nil, err
			}
			return types.Cons(car, inner), nil
		}
		// Handle unquote-splicing in list context
		var elems []*types.Value
		cur := expr
		for cur.Type == types.TypePair {
			item := cur.PairVal.Car
			if item.Type == types.TypePair && item.PairVal.Car.Type == types.TypeSymbol && item.PairVal.Car.SymbolVal == "unquote-splicing" {
				if depth == 1 {
					spliceArgs, err := item.PairVal.Cdr.ToList()
					if err != nil || len(spliceArgs) != 1 {
						return nil, fmt.Errorf("unquote-splicing requires exactly one argument")
					}
					val, err := Eval(spliceArgs[0], e)
					if err != nil {
						return nil, err
					}
					spliced, err := val.ToList()
					if err != nil {
						return nil, fmt.Errorf("unquote-splicing: expected list")
					}
					elems = append(elems, spliced...)
				} else {
					expanded, err := expandQQ(item, e, depth-1)
					if err != nil {
						return nil, err
					}
					elems = append(elems, expanded)
				}
			} else {
				expanded, err := expandQQ(item, e, depth)
				if err != nil {
					return nil, err
				}
				elems = append(elems, expanded)
			}
			cur = cur.PairVal.Cdr
		}
		tail := types.Nil()
		if cur.Type != types.TypeNil {
			var err error
			tail, err = expandQQ(cur, e, depth)
			if err != nil {
				return nil, err
			}
		}
		result := tail
		for i := len(elems) - 1; i >= 0; i-- {
			result = types.Cons(elems[i], result)
		}
		return result, nil
	}
	if expr.Type == types.TypeVector {
		var elems []*types.Value
		for _, el := range expr.VecVal {
			if el.Type == types.TypePair && el.PairVal.Car.Type == types.TypeSymbol && el.PairVal.Car.SymbolVal == "unquote-splicing" {
				if depth == 1 {
					spliceArgs, err := el.PairVal.Cdr.ToList()
					if err != nil || len(spliceArgs) != 1 {
						return nil, fmt.Errorf("unquote-splicing requires exactly one argument")
					}
					val, err := Eval(spliceArgs[0], e)
					if err != nil {
						return nil, err
					}
					spliced, err := val.ToList()
					if err != nil {
						return nil, fmt.Errorf("unquote-splicing in vector: expected list")
					}
					elems = append(elems, spliced...)
				} else {
					// Nested: preserve structure, decrease depth
					expanded, err := expandQQ(el, e, depth-1)
					if err != nil {
						return nil, err
					}
					elems = append(elems, expanded)
				}
			} else {
				expanded, err := expandQQ(el, e, depth)
				if err != nil {
					return nil, err
				}
				elems = append(elems, expanded)
			}
		}
		return types.NewVector(elems), nil
	}
	return expr, nil
}

func evalApplication(expr *types.Value, e *env.Env) (*evalResult, error) {
	op, err := Eval(expr.PairVal.Car, e)
	if err != nil {
		return nil, err
	}
	argExprs, err := expr.PairVal.Cdr.ToList()
	if err != nil {
		return nil, fmt.Errorf("invalid argument list")
	}
	args := make([]*types.Value, len(argExprs))
	for i, argExpr := range argExprs {
		args[i], err = Eval(argExpr, e)
		if err != nil {
			return nil, err
		}
	}
	return applyProc(op, args)
}

func applyProc(proc *types.Value, args []*types.Value) (*evalResult, error) {
	if proc.Type != types.TypeProcedure || proc.ProcVal == nil {
		return nil, fmt.Errorf("not a procedure: %s", proc.String())
	}

	p := proc.ProcVal
	if p.IsBuiltin {
		result, err := p.Builtin(args)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", p.Name, err)
		}
		return resultVal(result), nil
	}

	// Lambda application
	lam := p.Lambda
	closureEnv, ok := lam.Env.(*env.Env)
	if !ok {
		return nil, fmt.Errorf("invalid closure environment")
	}

	var callEnv *env.Env
	if lam.Params.HasRest || lam.Params.IsRest {
		if lam.Params.IsRest {
			// (lambda args body)
			callEnv = env.New(closureEnv)
			callEnv.Define(lam.Params.Rest, types.FromList(args))
		} else {
			// (lambda (a b . rest) body)
			nFixed := len(lam.Params.Fixed)
			if len(args) < nFixed {
				name := lam.Name
				if name == "" {
					name = "anonymous"
				}
				return nil, fmt.Errorf("%s: expected at least %d arguments, got %d", name, nFixed, len(args))
			}
			var err error
			callEnv, err = closureEnv.Extend(lam.Params.Fixed, args[:nFixed])
			if err != nil {
				return nil, err
			}
			callEnv.Define(lam.Params.Rest, types.FromList(args[nFixed:]))
		}
	} else {
		var err error
		callEnv, err = closureEnv.Extend(lam.Params.Fixed, args)
		if err != nil {
			name := lam.Name
			if name == "" {
				name = "anonymous"
			}
			return nil, fmt.Errorf("%s: %w", name, err)
		}
	}

	if len(lam.Body) == 0 {
		return resultVal(types.Void()), nil
	}
	for i := 0; i < len(lam.Body)-1; i++ {
		_, err := Eval(lam.Body[i], callEnv)
		if err != nil {
			return nil, err
		}
	}
	return resultThunk(lam.Body[len(lam.Body)-1], callEnv), nil
}

func evalBodyTCO(body []*types.Value, e *env.Env) (*evalResult, error) {
	if len(body) == 0 {
		return resultVal(types.Void()), nil
	}
	for i := 0; i < len(body)-1; i++ {
		_, err := Eval(body[i], e)
		if err != nil {
			return nil, err
		}
	}
	return resultThunk(body[len(body)-1], e), nil
}

// EvalString parses and evaluates Scheme source code.
// RegisterApply registers the apply procedure in the environment.
// This must be called after builtin.RegisterBuiltins since apply needs
// access to the evaluator's applyProc to support both builtins and lambdas.
// It also sets builtin.EvalApply so builtin functions can call user procedures.
func RegisterApply(e *env.Env) {
	// Set EvalApply so that builtin functions (assoc, member with comparators)
	// can invoke user-defined procedures.
	builtin.EvalApply = func(proc *types.Value, args []*types.Value) (*types.Value, error) {
		return callProc(proc, args)
	}
	e.Define("apply", types.NewProcedure("apply", func(args []*types.Value) (*types.Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("expected at least 2 arguments")
		}
		proc := args[0]
		if proc.Type != types.TypeProcedure {
			return nil, fmt.Errorf("expected a procedure, got %s", proc.Type)
		}
		last := args[len(args)-1]
		lst, err := last.ToList()
		if err != nil {
			return nil, fmt.Errorf("last argument must be a list")
		}
		allArgs := make([]*types.Value, 0, len(args)-2+len(lst))
		allArgs = append(allArgs, args[1:len(args)-1]...)
		allArgs = append(allArgs, lst...)
		result, err := applyProc(proc, allArgs)
		if err != nil {
			return nil, err
		}
		// Trampoline the result
		for result.thunk != nil {
			result, err = evalTCO(result.thunk.expr, result.thunk.env)
			if err != nil {
				return nil, err
			}
		}
		return result.value, nil
	}))
}

// RegisterMapForEach registers map, for-each, and filter that work with both builtins and lambdas.
func RegisterMapForEach(e *env.Env) {
	e.Define("map", types.NewProcedure("map", func(args []*types.Value) (*types.Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("expected at least 2 arguments")
		}
		proc := args[0]
		if proc.Type != types.TypeProcedure {
			return nil, fmt.Errorf("expected a procedure")
		}
		lst, err := args[1].ToList()
		if err != nil {
			return nil, fmt.Errorf("expected a list")
		}
		results := make([]*types.Value, len(lst))
		for i, el := range lst {
			r, err := callProc(proc, []*types.Value{el})
			if err != nil {
				return nil, err
			}
			results[i] = r
		}
		return types.FromList(results), nil
	}))

	e.Define("for-each", types.NewProcedure("for-each", func(args []*types.Value) (*types.Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("expected at least 2 arguments")
		}
		proc := args[0]
		if proc.Type != types.TypeProcedure {
			return nil, fmt.Errorf("expected a procedure")
		}
		lst, err := args[1].ToList()
		if err != nil {
			return nil, fmt.Errorf("expected a list")
		}
		for _, el := range lst {
			_, err := callProc(proc, []*types.Value{el})
			if err != nil {
				return nil, err
			}
		}
		return types.Void(), nil
	}))

	e.Define("filter", types.NewProcedure("filter", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("expected 2 arguments")
		}
		proc := args[0]
		if proc.Type != types.TypeProcedure {
			return nil, fmt.Errorf("expected a procedure")
		}
		lst, err := args[1].ToList()
		if err != nil {
			return nil, fmt.Errorf("expected a list")
		}
		var results []*types.Value
		for _, el := range lst {
			r, err := callProc(proc, []*types.Value{el})
			if err != nil {
				return nil, err
			}
			if r.IsTrue() {
				results = append(results, el)
			}
		}
		return types.FromList(results), nil
	}))
}

// EvalString parses and evaluates Scheme source code.
// It catches unhandled SchemeExceptions and converts them to Go errors.
func EvalString(input string, e *env.Env) (result *types.Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			if exc, ok := r.(SchemeException); ok {
				err = fmt.Errorf("unhandled exception: %s", exc.Value.Display())
				result = nil
				return
			}
			panic(r) // re-panic non-SchemeException
		}
	}()

	exprs, parseErr := parser.ParseAllString(input)
	if parseErr != nil {
		return nil, parseErr
	}
	for _, expr := range exprs {
		result, err = Eval(expr, e)
		if err != nil {
			return nil, err
		}
	}
	if result == nil {
		return types.Void(), nil
	}
	return result, nil
}

// EvalAll evaluates multiple expressions in order, returning the last result.
func EvalAll(exprs []*types.Value, e *env.Env) (*types.Value, error) {
	var result *types.Value
	var err error
	for _, expr := range exprs {
		result, err = Eval(expr, e)
		if err != nil {
			return nil, err
		}
	}
	if result == nil {
		return types.Void(), nil
	}
	return result, nil
}
