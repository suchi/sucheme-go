// Package evaluator - macro.go implements R7RS syntax-rules macro system.
package evaluator

import (
	"fmt"

	"github.com/suchi/sucheme-go/pkg/env"
	"github.com/suchi/sucheme-go/pkg/types"
)

// compileSyntaxRules compiles a syntax-rules form into a SyntaxRules struct.
// args should be the arguments after "syntax-rules" keyword:
//   (literal ...) (pattern template) ...
//   ellipsis (literal ...) (pattern template) ...
func compileSyntaxRules(args []*types.Value, e *env.Env) (*types.SyntaxRules, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("invalid syntax-rules")
	}

	sr := &types.SyntaxRules{
		Ellipsis: "...",
		DefEnv:   e,
	}

	idx := 0

	// Optional custom ellipsis identifier
	if args[0].Type == types.TypeSymbol && args[0].SymbolVal != "..." {
		// Check if next arg is a list (literals) - if so, first arg is custom ellipsis
		if len(args) > 1 && (args[1].Type == types.TypeNil || args[1].Type == types.TypePair) {
			sr.Ellipsis = args[0].SymbolVal
			idx = 1
		}
	}

	// Literals list
	if idx >= len(args) {
		return nil, fmt.Errorf("syntax-rules: missing literals list")
	}
	lits, err := args[idx].ToList()
	if err != nil && args[idx].Type != types.TypeNil {
		return nil, fmt.Errorf("syntax-rules: invalid literals list")
	}
	for _, lit := range lits {
		if lit.Type != types.TypeSymbol {
			return nil, fmt.Errorf("syntax-rules: literal must be a symbol")
		}
		sr.Literals = append(sr.Literals, lit.SymbolVal)
	}
	idx++

	// Rules
	for i := idx; i < len(args); i++ {
		rule, err := args[i].ToList()
		if err != nil || len(rule) != 2 {
			return nil, fmt.Errorf("syntax-rules: each rule must be (pattern template)")
		}
		sr.Rules = append(sr.Rules, types.SyntaxRule{
			Pattern:  rule[0],
			Template: rule[1],
		})
	}

	if len(sr.Rules) == 0 {
		return nil, fmt.Errorf("syntax-rules: no rules provided")
	}

	return sr, nil
}

// expandMacro expands a macro application.
// It returns the expanded form, or an error if no rule matches.
func expandMacro(sr *types.SyntaxRules, form *types.Value) (*types.Value, error) {
	// Macro application must be a (non-empty) pair
	if form.Type != types.TypePair {
		return nil, fmt.Errorf("macro application: expected list")
	}

	for _, rule := range sr.Rules {
		// Top-level pattern must be a pair: (macro-name . args-pattern)
		if rule.Pattern.Type != types.TypePair {
			continue
		}

		// Match the rest of the pattern against the rest of the form,
		// using matchSingle to support dotted patterns at the top level.
		patArgs := rule.Pattern.PairVal.Cdr
		formArgs := form.PairVal.Cdr

		bindings := make(map[string]*macroBinding)
		if matchSingle(patArgs, formArgs, sr, bindings) {
			result, err := instantiateTemplate(rule.Template, bindings, sr)
			if err != nil {
				return nil, err
			}
			return result, nil
		}
	}

	return nil, fmt.Errorf("no matching pattern for macro %s", sr.Name)
}

// macroBinding holds a matched pattern variable's value.
type macroBinding struct {
	value  *types.Value   // for non-ellipsis match
	values []*types.Value // for ellipsis match (list of matches)
	depth  int            // nesting depth of ellipsis
}

// matchPattern matches a list of pattern elements against a list of form elements.
func matchPattern(patterns []*types.Value, forms []*types.Value, sr *types.SyntaxRules, bindings map[string]*macroBinding) bool {
	pi := 0
	fi := 0

	for pi < len(patterns) {
		pat := patterns[pi]

		// Check if next pattern element is ellipsis
		hasEllipsis := pi+1 < len(patterns) &&
			patterns[pi+1].Type == types.TypeSymbol &&
			patterns[pi+1].SymbolVal == sr.Ellipsis

		if hasEllipsis {
			// Match zero or more forms against pat
			// Determine how many trailing patterns there are after the ellipsis
			tailCount := len(patterns) - pi - 2
			available := len(forms) - fi - tailCount
			if available < 0 {
				return false
			}

			// Collect pattern variables in this sub-pattern
			vars := collectPatternVars(pat, sr)
			// Initialize ellipsis bindings
			for _, v := range vars {
				bindings[v] = &macroBinding{values: []*types.Value{}, depth: 1}
			}

			// Match each available form
			for i := 0; i < available; i++ {
				subBindings := make(map[string]*macroBinding)
				if !matchSingle(pat, forms[fi+i], sr, subBindings) {
					return false
				}
				// Collect matched values
				for _, v := range vars {
					if b, ok := subBindings[v]; ok {
						bindings[v].values = append(bindings[v].values, b.value)
					}
				}
			}
			fi += available
			pi += 2 // skip pattern and ellipsis
		} else {
			// Regular match
			if fi >= len(forms) {
				return false
			}
			if !matchSingle(pat, forms[fi], sr, bindings) {
				return false
			}
			pi++
			fi++
		}
	}

	return fi == len(forms)
}

// matchSingle matches a single pattern against a single form.
func matchSingle(pat *types.Value, form *types.Value, sr *types.SyntaxRules, bindings map[string]*macroBinding) bool {
	switch pat.Type {
	case types.TypeSymbol:
		name := pat.SymbolVal
		// Underscore is wildcard
		if name == "_" {
			return true
		}
		// Check if it's a literal
		if isLiteral(name, sr) {
			return form.Type == types.TypeSymbol && form.SymbolVal == name
		}
		// Pattern variable - bind it, but check consistency with existing binding
		if existing, ok := bindings[name]; ok {
			return existing.depth == 0 && existing.value.Equal(form)
		}
		bindings[name] = &macroBinding{value: form, depth: 0}
		return true

	case types.TypePair:
		// Match list pattern
		if form.Type != types.TypePair && form.Type != types.TypeNil {
			return false
		}
		patElems, patTail := listAndTail(pat)
		formElems, formTail := listAndTail(form)

		// Check for dotted pattern tail
		if patTail != nil && patTail.Type != types.TypeNil {
			// Dotted pattern: (a b ... . rest)
			// Count how many fixed patterns (non-ellipsis) are in patElems
			// matchPattern handles ellipsis correctly, so pass all formElems
			// and let it determine how many to consume
			if !matchPattern(patElems, formElems, sr, bindings) {
				return false
			}
			// After matchPattern, the remaining forms (if any) + formTail go to patTail
			// matchPattern consumes exactly len(formElems) forms,
			// so the rest is just formTail
			rest := formTail
			if rest == nil {
				rest = types.Nil()
			}
			return matchSingle(patTail, rest, sr, bindings)
		}

		if !matchPattern(patElems, formElems, sr, bindings) {
			return false
		}
		// Both tails must be nil
		if patTail != nil && patTail.Type != types.TypeNil {
			return formTail != nil && formTail.Type != types.TypeNil &&
				matchSingle(patTail, formTail, sr, bindings)
		}
		return formTail == nil || formTail.Type == types.TypeNil

	case types.TypeNil:
		return form.Type == types.TypeNil

	default:
		// Literal datum (number, string, boolean, char)
		return pat.Equal(form)
	}
}

// listAndTail converts a pair chain to a list of elements and the tail.
func listAndTail(v *types.Value) ([]*types.Value, *types.Value) {
	var elems []*types.Value
	cur := v
	for cur.Type == types.TypePair {
		elems = append(elems, cur.PairVal.Car)
		cur = cur.PairVal.Cdr
	}
	return elems, cur
}

// collectPatternVars collects all unique pattern variable names in a pattern.
func collectPatternVars(pat *types.Value, sr *types.SyntaxRules) []string {
	var vars []string
	seen := make(map[string]bool)
	collectVarsHelper(pat, sr, &vars, seen)
	return vars
}

func collectVarsHelper(pat *types.Value, sr *types.SyntaxRules, vars *[]string, seen map[string]bool) {
	switch pat.Type {
	case types.TypeSymbol:
		name := pat.SymbolVal
		if name != "_" && name != sr.Ellipsis && !isLiteral(name, sr) && !seen[name] {
			*vars = append(*vars, name)
			seen[name] = true
		}
	case types.TypePair:
		elems, tail := listAndTail(pat)
		for _, e := range elems {
			if e.Type == types.TypeSymbol && e.SymbolVal == sr.Ellipsis {
				continue
			}
			collectVarsHelper(e, sr, vars, seen)
		}
		if tail != nil && tail.Type != types.TypeNil {
			collectVarsHelper(tail, sr, vars, seen)
		}
	}
}

// isLiteral checks if a name is in the literals list.
func isLiteral(name string, sr *types.SyntaxRules) bool {
	for _, lit := range sr.Literals {
		if lit == name {
			return true
		}
	}
	return false
}

// instantiateTemplate creates a new form from a template and bindings.
func instantiateTemplate(tmpl *types.Value, bindings map[string]*macroBinding, sr *types.SyntaxRules) (*types.Value, error) {
	return instantiate(tmpl, bindings, sr, 0)
}

func instantiate(tmpl *types.Value, bindings map[string]*macroBinding, sr *types.SyntaxRules, depth int) (*types.Value, error) {
	switch tmpl.Type {
	case types.TypeSymbol:
		name := tmpl.SymbolVal
		if b, ok := bindings[name]; ok {
			if b.depth == 0 {
				return b.value, nil
			}
			// Ellipsis variable at wrong depth
			return nil, fmt.Errorf("template: variable %s used outside of ellipsis context", name)
		}
		return tmpl, nil

	case types.TypePair:
		elems, tail := listAndTail(tmpl)
		var result []*types.Value

		for i := 0; i < len(elems); i++ {
			elem := elems[i]
			// Check if next element is ellipsis
			hasEllipsis := i+1 < len(elems) &&
				elems[i+1].Type == types.TypeSymbol &&
				elems[i+1].SymbolVal == sr.Ellipsis

			if hasEllipsis {
				// Find ellipsis-bound variables in this sub-template
				vars := collectTemplateVars(elem, bindings, sr)
				if len(vars) == 0 {
					return nil, fmt.Errorf("template: ellipsis with no pattern variables")
				}

				// Determine repetition count from the first ellipsis-bound variable
				count := -1
				for _, v := range vars {
					if b, ok := bindings[v]; ok && b.depth > 0 {
						if count == -1 {
							count = len(b.values)
						} else if len(b.values) != count {
							return nil, fmt.Errorf("template: inconsistent ellipsis counts")
						}
					}
				}
				if count == -1 {
					count = 0
				}

				// Instantiate the sub-template for each repetition
				for j := 0; j < count; j++ {
					subBindings := make(map[string]*macroBinding)
					for k, v := range bindings {
						subBindings[k] = v
					}
					for _, v := range vars {
						if b, ok := bindings[v]; ok && b.depth > 0 && j < len(b.values) {
							subBindings[v] = &macroBinding{value: b.values[j], depth: 0}
						}
					}
					expanded, err := instantiate(elem, subBindings, sr, depth+1)
					if err != nil {
						return nil, err
					}
					result = append(result, expanded)
				}
				i++ // skip the ellipsis
			} else {
				expanded, err := instantiate(elem, bindings, sr, depth)
				if err != nil {
					return nil, err
				}
				result = append(result, expanded)
			}
		}

		// Reconstruct list with tail
		var resultVal *types.Value
		if tail != nil && tail.Type != types.TypeNil {
			var err error
			resultVal, err = instantiate(tail, bindings, sr, depth)
			if err != nil {
				return nil, err
			}
		} else {
			resultVal = types.Nil()
		}
		for i := len(result) - 1; i >= 0; i-- {
			resultVal = types.Cons(result[i], resultVal)
		}
		return resultVal, nil

	case types.TypeVector:
		var result []*types.Value
		for i := 0; i < len(tmpl.VecVal); i++ {
			elem := tmpl.VecVal[i]
			hasEllipsis := i+1 < len(tmpl.VecVal) &&
				tmpl.VecVal[i+1].Type == types.TypeSymbol &&
				tmpl.VecVal[i+1].SymbolVal == sr.Ellipsis

			if hasEllipsis {
				vars := collectTemplateVars(elem, bindings, sr)
				count := -1
				for _, v := range vars {
					if b, ok := bindings[v]; ok && b.depth > 0 {
						cur := len(b.values)
						if count < 0 {
							count = cur
						} else if count != cur {
							return nil, fmt.Errorf("inconsistent ellipsis variable lengths in vector template")
						}
					}
				}
				if count < 0 {
					count = 0
				}
				for j := 0; j < count; j++ {
					subBindings := make(map[string]*macroBinding)
					for k, v := range bindings {
						subBindings[k] = v
					}
					for _, v := range vars {
						if b, ok := bindings[v]; ok && b.depth > 0 && j < len(b.values) {
							subBindings[v] = &macroBinding{value: b.values[j], depth: 0}
						}
					}
					expanded, err := instantiate(elem, subBindings, sr, depth+1)
					if err != nil {
						return nil, err
					}
					result = append(result, expanded)
				}
				i++
			} else {
				expanded, err := instantiate(elem, bindings, sr, depth)
				if err != nil {
					return nil, err
				}
				result = append(result, expanded)
			}
		}
		return types.NewVector(result), nil

	default:
		return tmpl, nil
	}
}

// collectTemplateVars finds all pattern variables with ellipsis bindings used in a template.
func collectTemplateVars(tmpl *types.Value, bindings map[string]*macroBinding, sr *types.SyntaxRules) []string {
	var vars []string
	seen := make(map[string]bool)
	collectTemplateVarsHelper(tmpl, bindings, sr, &vars, seen)
	return vars
}

func collectTemplateVarsHelper(tmpl *types.Value, bindings map[string]*macroBinding, sr *types.SyntaxRules, vars *[]string, seen map[string]bool) {
	switch tmpl.Type {
	case types.TypeSymbol:
		name := tmpl.SymbolVal
		if b, ok := bindings[name]; ok && b.depth > 0 && !seen[name] {
			*vars = append(*vars, name)
			seen[name] = true
		}
	case types.TypePair:
		elems, tail := listAndTail(tmpl)
		for _, e := range elems {
			if e.Type == types.TypeSymbol && e.SymbolVal == sr.Ellipsis {
				continue
			}
			collectTemplateVarsHelper(e, bindings, sr, vars, seen)
		}
		if tail != nil && tail.Type != types.TypeNil {
			collectTemplateVarsHelper(tail, bindings, sr, vars, seen)
		}
	case types.TypeVector:
		for _, e := range tmpl.VecVal {
			collectTemplateVarsHelper(e, bindings, sr, vars, seen)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
