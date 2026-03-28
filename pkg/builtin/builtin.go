// Package builtin registers built-in procedures for the Scheme interpreter.
package builtin

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/suchi/sucheme-go/pkg/env"
	"github.com/suchi/sucheme-go/pkg/types"
)

// RegisterBuiltins registers all built-in procedures into the given environment.
func RegisterBuiltins(e *env.Env) {
	registerArithmetic(e)
	registerComparison(e)
	registerList(e)
	registerString(e)
	registerChar(e)
	registerVector(e)
	registerBytevector(e)
	registerTypePredicates(e)
	registerIO(e)
}

func registerArithmetic(e *env.Env) {
	e.Define("+", types.NewProcedure("+", func(args []*types.Value) (*types.Value, error) {
		var isum int64
		var fsum float64
		isFloat := false
		for _, a := range args {
			switch a.Type {
			case types.TypeInteger:
				if isFloat {
					fsum += float64(a.IntVal)
				} else {
					isum += a.IntVal
				}
			case types.TypeFloat:
				if !isFloat {
					fsum = float64(isum)
					isFloat = true
				}
				fsum += a.FloatVal
			default:
				return nil, fmt.Errorf("expected number, got %s", a.Type)
			}
		}
		if isFloat {
			return types.NewFloat(fsum), nil
		}
		return types.NewInteger(isum), nil
	}))

	e.Define("-", types.NewProcedure("-", func(args []*types.Value) (*types.Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("expected at least 1 argument")
		}
		if len(args) == 1 {
			switch args[0].Type {
			case types.TypeInteger:
				return types.NewInteger(-args[0].IntVal), nil
			case types.TypeFloat:
				return types.NewFloat(-args[0].FloatVal), nil
			default:
				return nil, fmt.Errorf("expected number, got %s", args[0].Type)
			}
		}
		isFloat := false
		var iresult int64
		var fresult float64
		switch args[0].Type {
		case types.TypeInteger:
			iresult = args[0].IntVal
		case types.TypeFloat:
			fresult = args[0].FloatVal
			isFloat = true
		default:
			return nil, fmt.Errorf("expected number, got %s", args[0].Type)
		}
		for _, a := range args[1:] {
			switch a.Type {
			case types.TypeInteger:
				if isFloat {
					fresult -= float64(a.IntVal)
				} else {
					iresult -= a.IntVal
				}
			case types.TypeFloat:
				if !isFloat {
					fresult = float64(iresult)
					isFloat = true
				}
				fresult -= a.FloatVal
			default:
				return nil, fmt.Errorf("expected number, got %s", a.Type)
			}
		}
		if isFloat {
			return types.NewFloat(fresult), nil
		}
		return types.NewInteger(iresult), nil
	}))

	e.Define("*", types.NewProcedure("*", func(args []*types.Value) (*types.Value, error) {
		var iprod int64 = 1
		var fprod float64 = 1
		isFloat := false
		for _, a := range args {
			switch a.Type {
			case types.TypeInteger:
				if isFloat {
					fprod *= float64(a.IntVal)
				} else {
					iprod *= a.IntVal
				}
			case types.TypeFloat:
				if !isFloat {
					fprod = float64(iprod)
					isFloat = true
				}
				fprod *= a.FloatVal
			default:
				return nil, fmt.Errorf("expected number, got %s", a.Type)
			}
		}
		if isFloat {
			return types.NewFloat(fprod), nil
		}
		return types.NewInteger(iprod), nil
	}))

	e.Define("/", types.NewProcedure("/", func(args []*types.Value) (*types.Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("expected at least 2 arguments")
		}
		isFloat := false
		var iresult int64
		var fresult float64
		switch args[0].Type {
		case types.TypeInteger:
			iresult = args[0].IntVal
		case types.TypeFloat:
			fresult = args[0].FloatVal
			isFloat = true
		default:
			return nil, fmt.Errorf("expected number, got %s", args[0].Type)
		}
		for _, a := range args[1:] {
			switch a.Type {
			case types.TypeInteger:
				if a.IntVal == 0 {
					return nil, fmt.Errorf("division by zero")
				}
				if isFloat {
					fresult /= float64(a.IntVal)
				} else {
					if iresult%a.IntVal == 0 {
						iresult /= a.IntVal
					} else {
						fresult = float64(iresult) / float64(a.IntVal)
						isFloat = true
					}
				}
			case types.TypeFloat:
				if a.FloatVal == 0 {
					return nil, fmt.Errorf("division by zero")
				}
				if !isFloat {
					fresult = float64(iresult)
					isFloat = true
				}
				fresult /= a.FloatVal
			default:
				return nil, fmt.Errorf("expected number, got %s", a.Type)
			}
		}
		if isFloat {
			return types.NewFloat(fresult), nil
		}
		return types.NewInteger(iresult), nil
	}))

	e.Define("modulo", types.NewProcedure("modulo", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("expected 2 arguments")
		}
		if args[0].Type != types.TypeInteger || args[1].Type != types.TypeInteger {
			return nil, fmt.Errorf("expected integers")
		}
		if args[1].IntVal == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		result := args[0].IntVal % args[1].IntVal
		// Scheme modulo follows sign of divisor
		if (result < 0 && args[1].IntVal > 0) || (result > 0 && args[1].IntVal < 0) {
			result += args[1].IntVal
		}
		return types.NewInteger(result), nil
	}))

	e.Define("quotient", types.NewProcedure("quotient", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 2 || args[0].Type != types.TypeInteger || args[1].Type != types.TypeInteger {
			return nil, fmt.Errorf("expected 2 integers")
		}
		if args[1].IntVal == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return types.NewInteger(args[0].IntVal / args[1].IntVal), nil
	}))

	e.Define("remainder", types.NewProcedure("remainder", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 2 || args[0].Type != types.TypeInteger || args[1].Type != types.TypeInteger {
			return nil, fmt.Errorf("expected 2 integers")
		}
		if args[1].IntVal == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return types.NewInteger(args[0].IntVal % args[1].IntVal), nil
	}))

	e.Define("abs", types.NewProcedure("abs", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		switch args[0].Type {
		case types.TypeInteger:
			v := args[0].IntVal
			if v == math.MinInt64 {
				return nil, fmt.Errorf("integer overflow: abs of minimum int64")
			}
			if v < 0 {
				v = -v
			}
			return types.NewInteger(v), nil
		case types.TypeFloat:
			v := args[0].FloatVal
			if v < 0 {
				v = -v
			}
			return types.NewFloat(v), nil
		default:
			return nil, fmt.Errorf("expected number, got %s", args[0].Type)
		}
	}))

	e.Define("min", types.NewProcedure("min", func(args []*types.Value) (*types.Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("expected at least 1 argument")
		}
		result := args[0]
		for _, a := range args[1:] {
			lt, err := numLt(a, result)
			if err != nil {
				return nil, err
			}
			if lt {
				result = a
			}
		}
		return result, nil
	}))

	e.Define("max", types.NewProcedure("max", func(args []*types.Value) (*types.Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("expected at least 1 argument")
		}
		result := args[0]
		for _, a := range args[1:] {
			lt, err := numLt(result, a)
			if err != nil {
				return nil, err
			}
			if lt {
				result = a
			}
		}
		return result, nil
	}))

	e.Define("floor", types.NewProcedure("floor", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		switch args[0].Type {
		case types.TypeInteger:
			return args[0], nil
		case types.TypeFloat:
			return types.NewFloat(math.Floor(args[0].FloatVal)), nil
		default:
			return nil, fmt.Errorf("expected number")
		}
	}))

	e.Define("ceiling", types.NewProcedure("ceiling", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		switch args[0].Type {
		case types.TypeInteger:
			return args[0], nil
		case types.TypeFloat:
			return types.NewFloat(math.Ceil(args[0].FloatVal)), nil
		default:
			return nil, fmt.Errorf("expected number")
		}
	}))

	e.Define("truncate", types.NewProcedure("truncate", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		switch args[0].Type {
		case types.TypeInteger:
			return args[0], nil
		case types.TypeFloat:
			return types.NewFloat(math.Trunc(args[0].FloatVal)), nil
		default:
			return nil, fmt.Errorf("expected number")
		}
	}))

	e.Define("round", types.NewProcedure("round", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		switch args[0].Type {
		case types.TypeInteger:
			return args[0], nil
		case types.TypeFloat:
			return types.NewFloat(math.RoundToEven(args[0].FloatVal)), nil
		default:
			return nil, fmt.Errorf("expected number")
		}
	}))

	e.Define("expt", types.NewProcedure("expt", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("expected 2 arguments")
		}
		// Integer exponentiation via repeated squaring
		if args[0].Type == types.TypeInteger && args[1].Type == types.TypeInteger && args[1].IntVal >= 0 {
			base := args[0].IntVal
			exp := args[1].IntVal
			result := int64(1)
			for exp > 0 {
				if exp%2 == 1 {
					result *= base
				}
				base *= base
				exp /= 2
			}
			return types.NewInteger(result), nil
		}
		base, berr := toFloat(args[0])
		if berr != nil {
			return nil, berr
		}
		exp, eerr := toFloat(args[1])
		if eerr != nil {
			return nil, eerr
		}
		return types.NewFloat(math.Pow(base, exp)), nil
	}))

	e.Define("sqrt", types.NewProcedure("sqrt", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		f, err := toFloat(args[0])
		if err != nil {
			return nil, err
		}
		return types.NewFloat(math.Sqrt(f)), nil
	}))

	e.Define("exact->inexact", types.NewProcedure("exact->inexact", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		if args[0].Type == types.TypeInteger {
			return types.NewFloat(float64(args[0].IntVal)), nil
		}
		return args[0], nil
	}))

	e.Define("inexact->exact", types.NewProcedure("inexact->exact", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		if args[0].Type == types.TypeFloat {
			f := args[0].FloatVal
			if f != math.Trunc(f) {
				return nil, fmt.Errorf("cannot convert non-integer float to exact: %g", f)
			}
			if f > float64(math.MaxInt64) || f < float64(math.MinInt64) {
				return nil, fmt.Errorf("float value out of int64 range")
			}
			return types.NewInteger(int64(f)), nil
		}
		return args[0], nil
	}))

	e.Define("zero?", types.NewProcedure("zero?", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		switch args[0].Type {
		case types.TypeInteger:
			return types.NewBoolean(args[0].IntVal == 0), nil
		case types.TypeFloat:
			return types.NewBoolean(args[0].FloatVal == 0), nil
		default:
			return nil, fmt.Errorf("expected number")
		}
	}))

	e.Define("positive?", types.NewProcedure("positive?", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		f, err := toFloat(args[0])
		if err != nil {
			return nil, err
		}
		return types.NewBoolean(f > 0), nil
	}))

	e.Define("negative?", types.NewProcedure("negative?", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		f, err := toFloat(args[0])
		if err != nil {
			return nil, err
		}
		return types.NewBoolean(f < 0), nil
	}))

	e.Define("odd?", types.NewProcedure("odd?", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 || args[0].Type != types.TypeInteger {
			return nil, fmt.Errorf("expected integer")
		}
		return types.NewBoolean(args[0].IntVal%2 != 0), nil
	}))

	e.Define("even?", types.NewProcedure("even?", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 || args[0].Type != types.TypeInteger {
			return nil, fmt.Errorf("expected integer")
		}
		return types.NewBoolean(args[0].IntVal%2 == 0), nil
	}))
}

func registerComparison(e *env.Env) {
	e.Define("=", types.NewProcedure("=", func(args []*types.Value) (*types.Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("expected at least 2 arguments")
		}
		for i := 1; i < len(args); i++ {
			eq, err := numEq(args[i-1], args[i])
			if err != nil {
				return nil, err
			}
			if !eq {
				return types.False, nil
			}
		}
		return types.True, nil
	}))

	e.Define("<", types.NewProcedure("<", func(args []*types.Value) (*types.Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("expected at least 2 arguments")
		}
		for i := 1; i < len(args); i++ {
			lt, err := numLt(args[i-1], args[i])
			if err != nil {
				return nil, err
			}
			if !lt {
				return types.False, nil
			}
		}
		return types.True, nil
	}))

	e.Define(">", types.NewProcedure(">", func(args []*types.Value) (*types.Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("expected at least 2 arguments")
		}
		for i := 1; i < len(args); i++ {
			lt, err := numLt(args[i], args[i-1])
			if err != nil {
				return nil, err
			}
			if !lt {
				return types.False, nil
			}
		}
		return types.True, nil
	}))

	e.Define("<=", types.NewProcedure("<=", func(args []*types.Value) (*types.Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("expected at least 2 arguments")
		}
		for i := 1; i < len(args); i++ {
			lt, err := numLt(args[i], args[i-1])
			if err != nil {
				return nil, err
			}
			if lt {
				return types.False, nil
			}
		}
		return types.True, nil
	}))

	e.Define(">=", types.NewProcedure(">=", func(args []*types.Value) (*types.Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("expected at least 2 arguments")
		}
		for i := 1; i < len(args); i++ {
			lt, err := numLt(args[i-1], args[i])
			if err != nil {
				return nil, err
			}
			if lt {
				return types.False, nil
			}
		}
		return types.True, nil
	}))

	e.Define("not", types.NewProcedure("not", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		return types.NewBoolean(!args[0].IsTrue()), nil
	}))

	e.Define("eq?", types.NewProcedure("eq?", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("expected 2 arguments")
		}
		return types.NewBoolean(args[0].Eq(args[1])), nil
	}))

	e.Define("equal?", types.NewProcedure("equal?", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("expected 2 arguments")
		}
		return types.NewBoolean(args[0].Equal(args[1])), nil
	}))

	e.Define("eqv?", types.NewProcedure("eqv?", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("expected 2 arguments")
		}
		return types.NewBoolean(args[0].Eqv(args[1])), nil
	}))

	e.Define("boolean=?", types.NewProcedure("boolean=?", func(args []*types.Value) (*types.Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("expected at least 2 arguments")
		}
		for _, a := range args {
			if a.Type != types.TypeBoolean {
				return nil, fmt.Errorf("expected boolean, got %s", a.Type)
			}
		}
		for i := 1; i < len(args); i++ {
			if args[i].BoolVal != args[0].BoolVal {
				return types.False, nil
			}
		}
		return types.True, nil
	}))

	e.Define("assoc", types.NewProcedure("assoc", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("expected 2 arguments (custom comparator not yet supported)")
		}
		key := args[0]
		lst, err := args[1].ToList()
		if err != nil {
			return nil, fmt.Errorf("expected list")
		}
		for _, pair := range lst {
			if pair.Type != types.TypePair {
				return nil, fmt.Errorf("expected list of pairs")
			}
			if key.Equal(pair.PairVal.Car) {
				return pair, nil
			}
		}
		return types.False, nil
	}))

	e.Define("assv", types.NewProcedure("assv", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("expected 2 arguments")
		}
		key := args[0]
		lst, err := args[1].ToList()
		if err != nil {
			return nil, fmt.Errorf("expected list")
		}
		for _, pair := range lst {
			if pair.Type != types.TypePair {
				return nil, fmt.Errorf("expected list of pairs")
			}
			if key.Eqv(pair.PairVal.Car) {
				return pair, nil
			}
		}
		return types.False, nil
	}))

	e.Define("assq", types.NewProcedure("assq", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("expected 2 arguments")
		}
		key := args[0]
		lst, err := args[1].ToList()
		if err != nil {
			return nil, fmt.Errorf("expected list")
		}
		for _, pair := range lst {
			if pair.Type != types.TypePair {
				return nil, fmt.Errorf("expected list of pairs")
			}
			if key.Eq(pair.PairVal.Car) {
				return pair, nil
			}
		}
		return types.False, nil
	}))

	e.Define("member", types.NewProcedure("member", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("expected 2 arguments")
		}
		if args[1].Type != types.TypePair && args[1].Type != types.TypeNil {
			return nil, fmt.Errorf("expected list")
		}
		cur := args[1]
		for cur.Type == types.TypePair {
			if args[0].Equal(cur.PairVal.Car) {
				return cur, nil
			}
			cur = cur.PairVal.Cdr
		}
		return types.False, nil
	}))

	e.Define("memv", types.NewProcedure("memv", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("expected 2 arguments")
		}
		if args[1].Type != types.TypePair && args[1].Type != types.TypeNil {
			return nil, fmt.Errorf("expected list")
		}
		cur := args[1]
		for cur.Type == types.TypePair {
			if args[0].Eqv(cur.PairVal.Car) {
				return cur, nil
			}
			cur = cur.PairVal.Cdr
		}
		return types.False, nil
	}))

	e.Define("memq", types.NewProcedure("memq", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("expected 2 arguments")
		}
		if args[1].Type != types.TypePair && args[1].Type != types.TypeNil {
			return nil, fmt.Errorf("expected list")
		}
		cur := args[1]
		for cur.Type == types.TypePair {
			if args[0].Eq(cur.PairVal.Car) {
				return cur, nil
			}
			cur = cur.PairVal.Cdr
		}
		return types.False, nil
	}))
}

func toFloat(v *types.Value) (float64, error) {
	switch v.Type {
	case types.TypeInteger:
		return float64(v.IntVal), nil
	case types.TypeFloat:
		return v.FloatVal, nil
	default:
		return 0, fmt.Errorf("expected number, got %s", v.Type)
	}
}

func numEq(a, b *types.Value) (bool, error) {
	af, err := toFloat(a)
	if err != nil {
		return false, err
	}
	bf, err := toFloat(b)
	if err != nil {
		return false, err
	}
	return af == bf, nil
}

func numLt(a, b *types.Value) (bool, error) {
	af, err := toFloat(a)
	if err != nil {
		return false, err
	}
	bf, err := toFloat(b)
	if err != nil {
		return false, err
	}
	return af < bf, nil
}

func registerList(e *env.Env) {
	e.Define("cons", types.NewProcedure("cons", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("expected 2 arguments")
		}
		return types.Cons(args[0], args[1]), nil
	}))

	e.Define("car", types.NewProcedure("car", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 || args[0].Type != types.TypePair {
			return nil, fmt.Errorf("expected a pair")
		}
		return args[0].PairVal.Car, nil
	}))

	e.Define("cdr", types.NewProcedure("cdr", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 || args[0].Type != types.TypePair {
			return nil, fmt.Errorf("expected a pair")
		}
		return args[0].PairVal.Cdr, nil
	}))

	e.Define("list", types.NewProcedure("list", func(args []*types.Value) (*types.Value, error) {
		return types.FromList(args), nil
	}))

	e.Define("null?", types.NewProcedure("null?", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		return types.NewBoolean(args[0].Type == types.TypeNil), nil
	}))

	e.Define("pair?", types.NewProcedure("pair?", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		return types.NewBoolean(args[0].Type == types.TypePair), nil
	}))

	e.Define("list?", types.NewProcedure("list?", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		return types.NewBoolean(args[0].IsList()), nil
	}))

	e.Define("length", types.NewProcedure("length", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		n := args[0].ListLength()
		if n < 0 {
			return nil, fmt.Errorf("expected a proper list")
		}
		return types.NewInteger(int64(n)), nil
	}))

	e.Define("append", types.NewProcedure("append", func(args []*types.Value) (*types.Value, error) {
		if len(args) == 0 {
			return types.Nil(), nil
		}
		if len(args) == 1 {
			return args[0], nil
		}
		// Collect all but last into a flat slice, append last as tail
		var elems []*types.Value
		for i := 0; i < len(args)-1; i++ {
			lst, err := args[i].ToList()
			if err != nil {
				return nil, fmt.Errorf("expected a proper list")
			}
			elems = append(elems, lst...)
		}
		result := args[len(args)-1]
		for i := len(elems) - 1; i >= 0; i-- {
			result = types.Cons(elems[i], result)
		}
		return result, nil
	}))

	e.Define("reverse", types.NewProcedure("reverse", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		lst, err := args[0].ToList()
		if err != nil {
			return nil, fmt.Errorf("expected a proper list")
		}
		result := types.Nil()
		for _, v := range lst {
			result = types.Cons(v, result)
		}
		return result, nil
	}))

	e.Define("list-ref", types.NewProcedure("list-ref", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("expected 2 arguments")
		}
		if args[1].Type != types.TypeInteger {
			return nil, fmt.Errorf("expected integer index")
		}
		idx := args[1].IntVal
		if idx < 0 {
			return nil, fmt.Errorf("index must be non-negative, got %d", idx)
		}
		cur := args[0]
		for i := int64(0); i < idx; i++ {
			if cur.Type != types.TypePair {
				return nil, fmt.Errorf("index out of range")
			}
			cur = cur.PairVal.Cdr
		}
		if cur.Type != types.TypePair {
			return nil, fmt.Errorf("index out of range")
		}
		return cur.PairVal.Car, nil
	}))

	e.Define("list-tail", types.NewProcedure("list-tail", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("expected 2 arguments")
		}
		if args[1].Type != types.TypeInteger {
			return nil, fmt.Errorf("expected integer index")
		}
		if args[1].IntVal < 0 {
			return nil, fmt.Errorf("index must be non-negative, got %d", args[1].IntVal)
		}
		cur := args[0]
		for i := int64(0); i < args[1].IntVal; i++ {
			if cur.Type != types.TypePair {
				return nil, fmt.Errorf("index out of range")
			}
			cur = cur.PairVal.Cdr
		}
		return cur, nil
	}))

	e.Define("set-car!", types.NewProcedure("set-car!", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 2 || args[0].Type != types.TypePair {
			return nil, fmt.Errorf("expected a pair and a value")
		}
		args[0].PairVal.Car = args[1]
		return types.Void(), nil
	}))

	e.Define("set-cdr!", types.NewProcedure("set-cdr!", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 2 || args[0].Type != types.TypePair {
			return nil, fmt.Errorf("expected a pair and a value")
		}
		args[0].PairVal.Cdr = args[1]
		return types.Void(), nil
	}))

	e.Define("list-copy", types.NewProcedure("list-copy", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		// Proper list
		lst, err := args[0].ToList()
		if err == nil {
			cp := make([]*types.Value, len(lst))
			copy(cp, lst)
			return types.FromList(cp), nil
		}
		// Non-pair is an error
		if args[0].Type != types.TypePair {
			return nil, fmt.Errorf("expected list")
		}
		// Improper list: copy pair chain
		var head *types.Value
		var last *types.Value
		cur := args[0]
		for cur.Type == types.TypePair {
			newPair := types.Cons(cur.PairVal.Car, types.Nil())
			if head == nil {
				head = newPair
			} else {
				last.PairVal.Cdr = newPair
			}
			last = newPair
			cur = cur.PairVal.Cdr
		}
		if last != nil {
			last.PairVal.Cdr = cur
		}
		return head, nil
	}))

	// Note: apply is registered by the evaluator to support both builtins and lambdas.

	e.Define("values", types.NewProcedure("values", func(args []*types.Value) (*types.Value, error) {
		if len(args) == 1 {
			return args[0], nil
		}
		return types.NewValues(args), nil
	}))
}

func registerTypePredicates(e *env.Env) {
	e.Define("number?", types.NewProcedure("number?", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		return types.NewBoolean(args[0].Type == types.TypeInteger || args[0].Type == types.TypeFloat), nil
	}))

	e.Define("integer?", types.NewProcedure("integer?", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		return types.NewBoolean(args[0].Type == types.TypeInteger), nil
	}))

	e.Define("boolean?", types.NewProcedure("boolean?", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		return types.NewBoolean(args[0].Type == types.TypeBoolean), nil
	}))

	e.Define("string?", types.NewProcedure("string?", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		return types.NewBoolean(args[0].Type == types.TypeString), nil
	}))

	e.Define("symbol?", types.NewProcedure("symbol?", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		return types.NewBoolean(args[0].Type == types.TypeSymbol), nil
	}))

	e.Define("procedure?", types.NewProcedure("procedure?", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		return types.NewBoolean(args[0].Type == types.TypeProcedure), nil
	}))

	e.Define("char?", types.NewProcedure("char?", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		return types.NewBoolean(args[0].Type == types.TypeCharacter), nil
	}))

	e.Define("vector?", types.NewProcedure("vector?", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		return types.NewBoolean(args[0].Type == types.TypeVector), nil
	}))

	e.Define("void?", types.NewProcedure("void?", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		return types.NewBoolean(args[0].Type == types.TypeVoid), nil
	}))
}

func registerString(e *env.Env) {
	e.Define("string-length", types.NewProcedure("string-length", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 || args[0].Type != types.TypeString {
			return nil, fmt.Errorf("expected a string")
		}
		return types.NewInteger(int64(len([]rune(args[0].StrVal)))), nil
	}))

	e.Define("string-ref", types.NewProcedure("string-ref", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 2 || args[0].Type != types.TypeString || args[1].Type != types.TypeInteger {
			return nil, fmt.Errorf("expected a string and an integer")
		}
		runes := []rune(args[0].StrVal)
		idx := args[1].IntVal
		if idx < 0 || idx >= int64(len(runes)) {
			return nil, fmt.Errorf("string index out of range: %d", idx)
		}
		return types.NewChar(runes[idx]), nil
	}))

	e.Define("string-append", types.NewProcedure("string-append", func(args []*types.Value) (*types.Value, error) {
		var sb strings.Builder
		for _, a := range args {
			if a.Type != types.TypeString {
				return nil, fmt.Errorf("expected string, got %s", a.Type)
			}
			sb.WriteString(a.StrVal)
		}
		return types.NewString(sb.String()), nil
	}))

	e.Define("substring", types.NewProcedure("substring", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 3 || args[0].Type != types.TypeString || args[1].Type != types.TypeInteger || args[2].Type != types.TypeInteger {
			return nil, fmt.Errorf("expected string, start, end")
		}
		runes := []rune(args[0].StrVal)
		start, end := args[1].IntVal, args[2].IntVal
		if start < 0 || end < start || end > int64(len(runes)) {
			return nil, fmt.Errorf("index out of range")
		}
		return types.NewString(string(runes[start:end])), nil
	}))

	e.Define("string->number", types.NewProcedure("string->number", func(args []*types.Value) (*types.Value, error) {
		if len(args) < 1 || len(args) > 2 {
			return nil, fmt.Errorf("expected 1 or 2 arguments, got %d", len(args))
		}
		if args[0].Type != types.TypeString {
			return nil, fmt.Errorf("expected a string")
		}
		if len(args) == 2 {
			// Radix argument not yet implemented; reject for now
			return nil, fmt.Errorf("radix argument not yet supported")
		}
		s := strings.TrimSpace(args[0].StrVal)
		if s == "" {
			return types.False, nil
		}
		// Try float if contains '.', 'e', 'E'
		if strings.ContainsAny(s, ".eE") {
			f, err := strconv.ParseFloat(s, 64)
			if err == nil {
				return types.NewFloat(f), nil
			}
		}
		// Try integer (strconv validates the entire string)
		n, err := strconv.ParseInt(s, 10, 64)
		if err == nil {
			return types.NewInteger(n), nil
		}
		return types.False, nil
	}))

	e.Define("number->string", types.NewProcedure("number->string", func(args []*types.Value) (*types.Value, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		if args[0].Type != types.TypeInteger && args[0].Type != types.TypeFloat {
			return nil, fmt.Errorf("expected number, got %s", args[0].Type)
		}
		return types.NewString(args[0].String()), nil
	}))

	e.Define("string->list", types.NewProcedure("string->list", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 || args[0].Type != types.TypeString {
			return nil, fmt.Errorf("expected a string")
		}
		runes := []rune(args[0].StrVal)
		vals := make([]*types.Value, len(runes))
		for i, r := range runes {
			vals[i] = types.NewChar(r)
		}
		return types.FromList(vals), nil
	}))

	e.Define("list->string", types.NewProcedure("list->string", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		lst, err := args[0].ToList()
		if err != nil {
			return nil, fmt.Errorf("expected a list of characters")
		}
		runes := make([]rune, len(lst))
		for i, v := range lst {
			if v.Type != types.TypeCharacter {
				return nil, fmt.Errorf("expected character, got %s", v.Type)
			}
			runes[i] = v.CharVal
		}
		return types.NewString(string(runes)), nil
	}))

	e.Define("string-upcase", types.NewProcedure("string-upcase", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 || args[0].Type != types.TypeString {
			return nil, fmt.Errorf("expected a string")
		}
		return types.NewString(strings.ToUpper(args[0].StrVal)), nil
	}))

	e.Define("string-downcase", types.NewProcedure("string-downcase", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 || args[0].Type != types.TypeString {
			return nil, fmt.Errorf("expected a string")
		}
		return types.NewString(strings.ToLower(args[0].StrVal)), nil
	}))

	e.Define("string=?", types.NewProcedure("string=?", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 2 || args[0].Type != types.TypeString || args[1].Type != types.TypeString {
			return nil, fmt.Errorf("expected two strings")
		}
		return types.NewBoolean(args[0].StrVal == args[1].StrVal), nil
	}))

	e.Define("string<?", types.NewProcedure("string<?", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 2 || args[0].Type != types.TypeString || args[1].Type != types.TypeString {
			return nil, fmt.Errorf("expected two strings")
		}
		return types.NewBoolean(args[0].StrVal < args[1].StrVal), nil
	}))

	e.Define("string>?", types.NewProcedure("string>?", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 2 || args[0].Type != types.TypeString || args[1].Type != types.TypeString {
			return nil, fmt.Errorf("expected two strings")
		}
		return types.NewBoolean(args[0].StrVal > args[1].StrVal), nil
	}))

	e.Define("string-copy", types.NewProcedure("string-copy", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 || args[0].Type != types.TypeString {
			return nil, fmt.Errorf("expected a string")
		}
		return types.NewString(args[0].StrVal), nil
	}))

	e.Define("make-string", types.NewProcedure("make-string", func(args []*types.Value) (*types.Value, error) {
		if len(args) < 1 || args[0].Type != types.TypeInteger {
			return nil, fmt.Errorf("expected integer length")
		}
		n := args[0].IntVal
		if n < 0 {
			return nil, fmt.Errorf("string length must be non-negative")
		}
		if n > int64(math.MaxInt32) {
			return nil, fmt.Errorf("string length too large")
		}
		length := int(n)
		ch := ' '
		if len(args) >= 2 {
			if args[1].Type != types.TypeCharacter {
				return nil, fmt.Errorf("expected character for fill, got %s", args[1].Type)
			}
			ch = args[1].CharVal
		}
		runes := make([]rune, length)
		for i := range runes {
			runes[i] = ch
		}
		return types.NewString(string(runes)), nil
	}))

	e.Define("symbol->string", types.NewProcedure("symbol->string", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 || args[0].Type != types.TypeSymbol {
			return nil, fmt.Errorf("expected a symbol")
		}
		return types.NewString(args[0].SymbolVal), nil
	}))

	e.Define("string->symbol", types.NewProcedure("string->symbol", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 || args[0].Type != types.TypeString {
			return nil, fmt.Errorf("expected a string")
		}
		return types.NewSymbol(args[0].StrVal), nil
	}))
}

func registerChar(e *env.Env) {
	e.Define("char->integer", types.NewProcedure("char->integer", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 || args[0].Type != types.TypeCharacter {
			return nil, fmt.Errorf("expected a character")
		}
		return types.NewInteger(int64(args[0].CharVal)), nil
	}))

	e.Define("integer->char", types.NewProcedure("integer->char", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 || args[0].Type != types.TypeInteger {
			return nil, fmt.Errorf("expected an integer")
		}
		return types.NewChar(rune(args[0].IntVal)), nil
	}))

	e.Define("char-alphabetic?", types.NewProcedure("char-alphabetic?", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 || args[0].Type != types.TypeCharacter {
			return nil, fmt.Errorf("expected a character")
		}
		c := args[0].CharVal
		return types.NewBoolean((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')), nil
	}))

	e.Define("char-numeric?", types.NewProcedure("char-numeric?", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 || args[0].Type != types.TypeCharacter {
			return nil, fmt.Errorf("expected a character")
		}
		c := args[0].CharVal
		return types.NewBoolean(c >= '0' && c <= '9'), nil
	}))

	e.Define("char-whitespace?", types.NewProcedure("char-whitespace?", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 || args[0].Type != types.TypeCharacter {
			return nil, fmt.Errorf("expected a character")
		}
		c := args[0].CharVal
		return types.NewBoolean(c == ' ' || c == '\t' || c == '\n' || c == '\r'), nil
	}))

	e.Define("char-upper-case?", types.NewProcedure("char-upper-case?", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 || args[0].Type != types.TypeCharacter {
			return nil, fmt.Errorf("expected a character")
		}
		c := args[0].CharVal
		return types.NewBoolean(c >= 'A' && c <= 'Z'), nil
	}))

	e.Define("char-lower-case?", types.NewProcedure("char-lower-case?", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 || args[0].Type != types.TypeCharacter {
			return nil, fmt.Errorf("expected a character")
		}
		c := args[0].CharVal
		return types.NewBoolean(c >= 'a' && c <= 'z'), nil
	}))

	e.Define("char-upcase", types.NewProcedure("char-upcase", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 || args[0].Type != types.TypeCharacter {
			return nil, fmt.Errorf("expected a character")
		}
		c := args[0].CharVal
		if c >= 'a' && c <= 'z' {
			c = c - 'a' + 'A'
		}
		return types.NewChar(c), nil
	}))

	e.Define("char-downcase", types.NewProcedure("char-downcase", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 || args[0].Type != types.TypeCharacter {
			return nil, fmt.Errorf("expected a character")
		}
		c := args[0].CharVal
		if c >= 'A' && c <= 'Z' {
			c = c - 'A' + 'a'
		}
		return types.NewChar(c), nil
	}))

	e.Define("char=?", types.NewProcedure("char=?", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 2 || args[0].Type != types.TypeCharacter || args[1].Type != types.TypeCharacter {
			return nil, fmt.Errorf("expected two characters")
		}
		return types.NewBoolean(args[0].CharVal == args[1].CharVal), nil
	}))

	e.Define("char<?", types.NewProcedure("char<?", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 2 || args[0].Type != types.TypeCharacter || args[1].Type != types.TypeCharacter {
			return nil, fmt.Errorf("expected two characters")
		}
		return types.NewBoolean(args[0].CharVal < args[1].CharVal), nil
	}))

	e.Define("char>?", types.NewProcedure("char>?", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 2 || args[0].Type != types.TypeCharacter || args[1].Type != types.TypeCharacter {
			return nil, fmt.Errorf("expected two characters")
		}
		return types.NewBoolean(args[0].CharVal > args[1].CharVal), nil
	}))
}

func registerVector(e *env.Env) {
	e.Define("make-vector", types.NewProcedure("make-vector", func(args []*types.Value) (*types.Value, error) {
		if len(args) < 1 || args[0].Type != types.TypeInteger {
			return nil, fmt.Errorf("expected integer size")
		}
		n := args[0].IntVal
		if n < 0 {
			return nil, fmt.Errorf("vector size must be non-negative")
		}
		if n > int64(math.MaxInt32) {
			return nil, fmt.Errorf("vector size too large")
		}
		fill := types.NewInteger(0)
		if len(args) >= 2 {
			fill = args[1]
		}
		vec := make([]*types.Value, int(n))
		for i := range vec {
			vec[i] = fill
		}
		return types.NewVector(vec), nil
	}))

	e.Define("vector", types.NewProcedure("vector", func(args []*types.Value) (*types.Value, error) {
		return types.NewVector(args), nil
	}))

	e.Define("vector-length", types.NewProcedure("vector-length", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 || args[0].Type != types.TypeVector {
			return nil, fmt.Errorf("expected a vector")
		}
		return types.NewInteger(int64(len(args[0].VecVal))), nil
	}))

	e.Define("vector-ref", types.NewProcedure("vector-ref", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 2 || args[0].Type != types.TypeVector || args[1].Type != types.TypeInteger {
			return nil, fmt.Errorf("expected a vector and an integer")
		}
		idx := args[1].IntVal
		if idx < 0 || idx >= int64(len(args[0].VecVal)) {
			return nil, fmt.Errorf("vector index out of range: %d", idx)
		}
		return args[0].VecVal[idx], nil
	}))

	e.Define("vector-set!", types.NewProcedure("vector-set!", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 3 || args[0].Type != types.TypeVector || args[1].Type != types.TypeInteger {
			return nil, fmt.Errorf("expected a vector, index, and value")
		}
		idx := args[1].IntVal
		if idx < 0 || idx >= int64(len(args[0].VecVal)) {
			return nil, fmt.Errorf("vector index out of range: %d", idx)
		}
		args[0].VecVal[idx] = args[2]
		return types.Void(), nil
	}))

	e.Define("vector->list", types.NewProcedure("vector->list", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 || args[0].Type != types.TypeVector {
			return nil, fmt.Errorf("expected a vector")
		}
		return types.FromList(args[0].VecVal), nil
	}))

	e.Define("list->vector", types.NewProcedure("list->vector", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		lst, err := args[0].ToList()
		if err != nil {
			return nil, fmt.Errorf("expected a list")
		}
		return types.NewVector(lst), nil
	}))

	e.Define("vector-fill!", types.NewProcedure("vector-fill!", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 2 || args[0].Type != types.TypeVector {
			return nil, fmt.Errorf("expected a vector and a value")
		}
		for i := range args[0].VecVal {
			args[0].VecVal[i] = args[1]
		}
		return types.Void(), nil
	}))

	e.Define("vector-copy", types.NewProcedure("vector-copy", func(args []*types.Value) (*types.Value, error) {
		if len(args) < 1 || args[0].Type != types.TypeVector {
			return nil, fmt.Errorf("expected a vector")
		}
		cp := make([]*types.Value, len(args[0].VecVal))
		copy(cp, args[0].VecVal)
		return types.NewVector(cp), nil
	}))
}

func registerBytevector(e *env.Env) {
	e.Define("make-bytevector", types.NewProcedure("make-bytevector", func(args []*types.Value) (*types.Value, error) {
		if len(args) < 1 || args[0].Type != types.TypeInteger {
			return nil, fmt.Errorf("expected integer size")
		}
		n64 := args[0].IntVal
		if n64 < 0 {
			return nil, fmt.Errorf("bytevector size must be non-negative")
		}
		if n64 > int64(math.MaxInt32) {
			return nil, fmt.Errorf("bytevector size too large")
		}
		n := int(n64)
		fill := byte(0)
		if len(args) >= 2 {
			if args[1].Type != types.TypeInteger {
				return nil, fmt.Errorf("expected integer fill value")
			}
			fillVal := args[1].IntVal
			if fillVal < 0 || fillVal > 255 {
				return nil, fmt.Errorf("fill value out of range: %d (expected 0..255)", fillVal)
			}
			fill = byte(fillVal)
		}
		bv := make([]byte, n)
		for i := range bv {
			bv[i] = fill
		}
		return types.NewBytevector(bv), nil
	}))

	e.Define("bytevector-length", types.NewProcedure("bytevector-length", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 || args[0].Type != types.TypeBytevector {
			return nil, fmt.Errorf("expected a bytevector")
		}
		return types.NewInteger(int64(len(args[0].ByteVec))), nil
	}))

	e.Define("bytevector-u8-ref", types.NewProcedure("bytevector-u8-ref", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 2 || args[0].Type != types.TypeBytevector || args[1].Type != types.TypeInteger {
			return nil, fmt.Errorf("expected a bytevector and an integer")
		}
		idx := args[1].IntVal
		if idx < 0 || idx >= int64(len(args[0].ByteVec)) {
			return nil, fmt.Errorf("bytevector index out of range: %d", idx)
		}
		return types.NewInteger(int64(args[0].ByteVec[idx])), nil
	}))

	e.Define("bytevector-u8-set!", types.NewProcedure("bytevector-u8-set!", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 3 || args[0].Type != types.TypeBytevector || args[1].Type != types.TypeInteger || args[2].Type != types.TypeInteger {
			return nil, fmt.Errorf("expected bytevector, index, and byte value")
		}
		idx := args[1].IntVal
		if idx < 0 || idx >= int64(len(args[0].ByteVec)) {
			return nil, fmt.Errorf("bytevector index out of range: %d", idx)
		}
		val := args[2].IntVal
		if val < 0 || val > 255 {
			return nil, fmt.Errorf("byte value out of range: %d (expected 0..255)", val)
		}
		args[0].ByteVec[idx] = byte(val)
		return types.Void(), nil
	}))

	e.Define("bytevector-copy", types.NewProcedure("bytevector-copy", func(args []*types.Value) (*types.Value, error) {
		if len(args) < 1 || args[0].Type != types.TypeBytevector {
			return nil, fmt.Errorf("expected a bytevector")
		}
		cp := make([]byte, len(args[0].ByteVec))
		copy(cp, args[0].ByteVec)
		return types.NewBytevector(cp), nil
	}))

	e.Define("bytevector-append", types.NewProcedure("bytevector-append", func(args []*types.Value) (*types.Value, error) {
		var result []byte
		for _, a := range args {
			if a.Type != types.TypeBytevector {
				return nil, fmt.Errorf("expected bytevector, got %s", a.Type)
			}
			result = append(result, a.ByteVec...)
		}
		return types.NewBytevector(result), nil
	}))

	e.Define("bytevector?", types.NewProcedure("bytevector?", func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		return types.NewBoolean(args[0].Type == types.TypeBytevector), nil
	}))
}

func registerIO(e *env.Env) {
	e.Define("display", types.NewProcedure("display", func(args []*types.Value) (*types.Value, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("expected at least 1 argument")
		}
		// No-op for now, returns void
		return types.Void(), nil
	}))

	e.Define("newline", types.NewProcedure("newline", func(args []*types.Value) (*types.Value, error) {
		return types.Void(), nil
	}))

	e.Define("error", types.NewProcedure("error", func(args []*types.Value) (*types.Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("error")
		}
		msg := args[0].Display()
		return nil, fmt.Errorf("%s", msg)
	}))

	e.Define("write", types.NewProcedure("write", func(args []*types.Value) (*types.Value, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("expected at least 1 argument")
		}
		return types.Void(), nil
	}))
}
