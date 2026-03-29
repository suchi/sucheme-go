// Package evaluator - R7RS special forms: define-record-type, cond-expand,
// define-values, include, include-ci, and case.
package evaluator

import (
	"fmt"
	"os"
	"strings"
	"sync/atomic"

	"github.com/suchi/sucheme-go/pkg/env"
	"github.com/suchi/sucheme-go/pkg/parser"
	"github.com/suchi/sucheme-go/pkg/types"
)

// ---------------------------------------------------------------------------
// define-record-type
// ---------------------------------------------------------------------------

// recordTypeID is a globally unique counter for record type descriptors.
var recordTypeID uint64

// recordTypeDescriptor is a unique identity for a record type definition.
// Two define-record-type forms always produce distinct descriptors even if
// they share the same type name.
type recordTypeDescriptor struct {
	id   uint64
	name string
}

// newRecordTypeDescriptor creates a new unique record type descriptor.
func newRecordTypeDescriptor(name string) *recordTypeDescriptor {
	return &recordTypeDescriptor{
		id:   atomic.AddUint64(&recordTypeID, 1),
		name: name,
	}
}

// evalDefineRecordType implements R7RS define-record-type.
//
//	(define-record-type <name>
//	  (<constructor> <field-name> ...)
//	  <predicate>
//	  (<field-name> <accessor>) ...
//	  (<field-name> <accessor> <mutator>) ...)
func evalDefineRecordType(expr *types.Value, e *env.Env) (*evalResult, error) {
	args, err := listArgs(expr)
	if err != nil || len(args) < 3 {
		return nil, fmt.Errorf("define-record-type: requires at least name, constructor, and predicate")
	}

	// 1. Parse type name
	if args[0].Type != types.TypeSymbol {
		return nil, fmt.Errorf("define-record-type: name must be a symbol")
	}
	typeName := args[0].SymbolVal

	// Create a unique descriptor for this record type
	rtd := newRecordTypeDescriptor(typeName)

	// 2. Parse constructor: (<constructor-name> <field-name> ...)
	ctorParts, err := args[1].ToList()
	if err != nil || len(ctorParts) == 0 {
		return nil, fmt.Errorf("define-record-type: invalid constructor clause")
	}
	if ctorParts[0].Type != types.TypeSymbol {
		return nil, fmt.Errorf("define-record-type: constructor name must be a symbol")
	}
	ctorName := ctorParts[0].SymbolVal
	ctorFields := make([]string, len(ctorParts)-1)
	for i := 1; i < len(ctorParts); i++ {
		if ctorParts[i].Type != types.TypeSymbol {
			return nil, fmt.Errorf("define-record-type: constructor field name must be a symbol")
		}
		ctorFields[i-1] = ctorParts[i].SymbolVal
	}

	// 3. Parse predicate
	if args[2].Type != types.TypeSymbol {
		return nil, fmt.Errorf("define-record-type: predicate must be a symbol")
	}
	predName := args[2].SymbolVal

	// 4. Parse field specs
	type fieldSpec struct {
		name     string
		accessor string
		mutator  string
	}
	var fieldSpecs []fieldSpec
	seenFields := make(map[string]bool)

	for i := 3; i < len(args); i++ {
		parts, err := args[i].ToList()
		if err != nil || len(parts) < 2 || len(parts) > 3 {
			return nil, fmt.Errorf("define-record-type: invalid field spec")
		}
		if parts[0].Type != types.TypeSymbol {
			return nil, fmt.Errorf("define-record-type: field name must be a symbol")
		}
		if parts[1].Type != types.TypeSymbol {
			return nil, fmt.Errorf("define-record-type: accessor name must be a symbol")
		}
		fieldName := parts[0].SymbolVal
		if seenFields[fieldName] {
			return nil, fmt.Errorf("define-record-type: duplicate field name %q", fieldName)
		}
		seenFields[fieldName] = true

		fs := fieldSpec{
			name:     fieldName,
			accessor: parts[1].SymbolVal,
		}
		if len(parts) == 3 {
			if parts[2].Type != types.TypeSymbol {
				return nil, fmt.Errorf("define-record-type: mutator name must be a symbol")
			}
			fs.mutator = parts[2].SymbolVal
		}
		fieldSpecs = append(fieldSpecs, fs)
	}

	// Validate that constructor fields are in field specs
	for _, cf := range ctorFields {
		if !seenFields[cf] {
			return nil, fmt.Errorf("define-record-type: constructor field %q not in field specs", cf)
		}
	}

	// 5. Define constructor
	// Collect all field names so non-ctor fields get initialized to undefined.
	allFieldNames := make([]string, len(fieldSpecs))
	for i, fs := range fieldSpecs {
		allFieldNames[i] = fs.name
	}
	capturedCtorFields := ctorFields
	capturedRTD := rtd
	capturedAllFields := allFieldNames
	e.Define(ctorName, types.NewProcedure(ctorName, func(args []*types.Value) (*types.Value, error) {
		if len(args) != len(capturedCtorFields) {
			return nil, fmt.Errorf("expected %d arguments, got %d", len(capturedCtorFields), len(args))
		}
		fields := make(map[string]*types.Value)
		// Initialize all fields to undefined
		for _, name := range capturedAllFields {
			fields[name] = types.Undefined()
		}
		// Overwrite with constructor arguments
		for i, name := range capturedCtorFields {
			fields[name] = args[i]
		}
		rec := types.NewRecord(capturedRTD.name, fields)
		rec.RecordVal.Descriptor = capturedRTD
		return rec, nil
	}))

	// 6. Define predicate
	e.Define(predName, types.NewProcedure(predName, func(args []*types.Value) (*types.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument, got %d", len(args))
		}
		if args[0].Type == types.TypeRecord && args[0].RecordVal != nil && args[0].RecordVal.Descriptor == capturedRTD {
			return types.True, nil
		}
		return types.False, nil
	}))

	// 7. Define accessors and mutators
	for _, fs := range fieldSpecs {
		fieldName := fs.name
		e.Define(fs.accessor, types.NewProcedure(fs.accessor, func(args []*types.Value) (*types.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("expected 1 argument, got %d", len(args))
			}
			if args[0].Type != types.TypeRecord || args[0].RecordVal == nil || args[0].RecordVal.Descriptor != capturedRTD {
				return nil, fmt.Errorf("expected a %s record", capturedRTD.name)
			}
			val, ok := args[0].RecordVal.Fields[fieldName]
			if !ok {
				return nil, fmt.Errorf("field %s not found in record %s", fieldName, capturedRTD.name)
			}
			return val, nil
		}))

		if fs.mutator != "" {
			mutFieldName := fieldName
			e.Define(fs.mutator, types.NewProcedure(fs.mutator, func(args []*types.Value) (*types.Value, error) {
				if len(args) != 2 {
					return nil, fmt.Errorf("expected 2 arguments, got %d", len(args))
				}
				if args[0].Type != types.TypeRecord || args[0].RecordVal == nil || args[0].RecordVal.Descriptor != capturedRTD {
					return nil, fmt.Errorf("expected a %s record", capturedRTD.name)
				}
				args[0].RecordVal.Fields[mutFieldName] = args[1]
				return types.Void(), nil
			}))
		}
	}

	return resultVal(types.Void()), nil
}

// ---------------------------------------------------------------------------
// cond-expand
// ---------------------------------------------------------------------------

// supportedFeatures lists the feature identifiers supported by this implementation.
var supportedFeatures = map[string]bool{
	"r7rs":         true,
	"sucheme":      true,
	"exact-closed": true,
	"ieee-float":   true,
}

// evalCondExpand implements R7RS cond-expand.
func evalCondExpand(expr *types.Value, e *env.Env) (*evalResult, error) {
	clauses, err := listArgs(expr)
	if err != nil {
		return nil, fmt.Errorf("cond-expand: invalid syntax")
	}
	if len(clauses) == 0 {
		return nil, fmt.Errorf("cond-expand: requires at least one clause")
	}

	for i, clause := range clauses {
		parts, err := clause.ToList()
		if err != nil || len(parts) == 0 {
			return nil, fmt.Errorf("cond-expand: invalid clause")
		}

		// else clause
		if parts[0].Type == types.TypeSymbol && parts[0].SymbolVal == "else" {
			if i != len(clauses)-1 {
				return nil, fmt.Errorf("cond-expand: else clause must be last")
			}
			if len(parts) < 2 {
				return resultVal(types.Void()), nil
			}
			return evalBodyTCO(parts[1:], e)
		}

		// Test the feature requirement
		if testFeatureRequirement(parts[0]) {
			if len(parts) < 2 {
				return resultVal(types.Void()), nil
			}
			return evalBodyTCO(parts[1:], e)
		}
	}

	return nil, fmt.Errorf("cond-expand: no matching clause")
}

// testFeatureRequirement tests a cond-expand feature requirement.
func testFeatureRequirement(req *types.Value) bool {
	// Simple feature identifier
	if req.Type == types.TypeSymbol {
		return supportedFeatures[req.SymbolVal]
	}

	// Compound requirement: (library <name>), (and ...), (or ...), (not ...)
	if req.Type != types.TypePair {
		return false
	}

	parts, err := req.ToList()
	if err != nil || len(parts) == 0 {
		return false
	}

	if parts[0].Type != types.TypeSymbol {
		return false
	}

	switch parts[0].SymbolVal {
	case "library":
		if len(parts) != 2 {
			return false
		}
		// Use parseLibraryName to get the registry key format (e.g. "scheme base")
		libName, err := parseLibraryName(parts[1])
		if err != nil {
			return false
		}
		_, ok := libraryRegistry[libName]
		return ok

	case "and":
		for _, sub := range parts[1:] {
			if !testFeatureRequirement(sub) {
				return false
			}
		}
		return true

	case "or":
		for _, sub := range parts[1:] {
			if testFeatureRequirement(sub) {
				return true
			}
		}
		return false

	case "not":
		if len(parts) != 2 {
			return false
		}
		return !testFeatureRequirement(parts[1])
	}

	return false
}

// ---------------------------------------------------------------------------
// define-values
// ---------------------------------------------------------------------------

// evalDefineValues implements R7RS define-values.
//
//	(define-values (x y z) (values 1 2 3))
func evalDefineValues(expr *types.Value, e *env.Env) (*evalResult, error) {
	args, err := listArgs(expr)
	if err != nil || len(args) != 2 {
		return nil, fmt.Errorf("define-values: requires formals and expression")
	}

	// Parse formals - could be a proper list or a single symbol (rest)
	var names []string
	var restName string
	if args[0].Type == types.TypeSymbol {
		// (define-values rest expr) - bind all values as a list
		restName = args[0].SymbolVal
	} else if args[0].Type == types.TypeNil {
		// (define-values () expr) - no bindings
		names = nil
	} else {
		// Parse as list, possibly dotted
		cur := args[0]
		for cur.Type == types.TypePair {
			if cur.PairVal.Car.Type != types.TypeSymbol {
				return nil, fmt.Errorf("define-values: formal must be a symbol")
			}
			names = append(names, cur.PairVal.Car.SymbolVal)
			cur = cur.PairVal.Cdr
		}
		if cur.Type == types.TypeSymbol {
			restName = cur.SymbolVal
		} else if cur.Type != types.TypeNil {
			return nil, fmt.Errorf("define-values: invalid formals")
		}
	}

	// Evaluate the expression
	val, err := Eval(args[1], e)
	if err != nil {
		return nil, err
	}

	// Extract values
	var vals []*types.Value
	if val.Type == types.TypeValues {
		vals = val.ValuesVal
	} else {
		// Single value
		vals = []*types.Value{val}
	}

	if restName != "" && len(names) == 0 {
		// (define-values rest expr) - bind all as list
		e.Define(restName, types.FromList(vals))
		return resultVal(types.Void()), nil
	}

	if restName != "" {
		// Dotted formals: (define-values (a b . rest) ...)
		if len(vals) < len(names) {
			return nil, fmt.Errorf("define-values: expected at least %d values, got %d", len(names), len(vals))
		}
		for i, name := range names {
			e.Define(name, vals[i])
		}
		e.Define(restName, types.FromList(vals[len(names):]))
	} else {
		if len(vals) != len(names) {
			return nil, fmt.Errorf("define-values: expected %d values, got %d", len(names), len(vals))
		}
		for i, name := range names {
			e.Define(name, vals[i])
		}
	}

	return resultVal(types.Void()), nil
}

// ---------------------------------------------------------------------------
// include / include-ci
// ---------------------------------------------------------------------------

// evalInclude implements R7RS include and include-ci.
// For include-ci, identifiers in the source are folded to lowercase.
// R7RS specifies that include-ci applies case-folding to identifiers only,
// but as a practical approximation we fold the entire source since our lexer
// does not yet support a per-read case-fold mode. String/character literals
// that rely on case are rare in practice.
func evalInclude(expr *types.Value, e *env.Env, caseInsensitive bool) (*evalResult, error) {
	args, err := listArgs(expr)
	if err != nil || len(args) == 0 {
		return nil, fmt.Errorf("include: requires at least one filename")
	}

	var lastResult *types.Value = types.Void()

	for _, arg := range args {
		if arg.Type != types.TypeString {
			return nil, fmt.Errorf("include: filename must be a string")
		}
		filename := arg.StrVal

		data, err := os.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("include: %w", err)
		}

		source := string(data)
		if caseInsensitive {
			source = strings.ToLower(source)
		}

		exprs, err := parser.ParseAllString(source)
		if err != nil {
			return nil, fmt.Errorf("include: parse error in %s: %w", filename, err)
		}

		for _, ex := range exprs {
			lastResult, err = Eval(ex, e)
			if err != nil {
				return nil, fmt.Errorf("include: error in %s: %w", filename, err)
			}
		}
	}

	return resultVal(lastResult), nil
}

// ---------------------------------------------------------------------------
// case
// ---------------------------------------------------------------------------

// evalCase implements R7RS case expression.
//
//	(case <key>
//	  ((<datum> ...) <expr> ...)
//	  ((<datum> ...) => <proc>)
//	  (else <expr> ...)
//	  (else => <proc>))
func evalCase(expr *types.Value, e *env.Env) (*evalResult, error) {
	args, err := listArgs(expr)
	if err != nil || len(args) < 2 {
		return nil, fmt.Errorf("case: requires key expression and at least one clause")
	}

	// Evaluate the key expression
	key, err := Eval(args[0], e)
	if err != nil {
		return nil, err
	}

	clauses := args[1:]
	for i, clause := range clauses {
		parts, err := clause.ToList()
		if err != nil || len(parts) < 2 {
			return nil, fmt.Errorf("case: invalid clause")
		}

		// else clause
		if parts[0].Type == types.TypeSymbol && parts[0].SymbolVal == "else" {
			if i != len(clauses)-1 {
				return nil, fmt.Errorf("case: else clause must be last")
			}
			// (else => proc)
			if len(parts) == 3 && parts[1].Type == types.TypeSymbol && parts[1].SymbolVal == "=>" {
				proc, err := Eval(parts[2], e)
				if err != nil {
					return nil, err
				}
				return applyProc(proc, []*types.Value{key})
			}
			return evalBodyTCO(parts[1:], e)
		}

		// Normal clause: ((<datum> ...) body ...) or ((<datum> ...) => proc)
		datums, err := parts[0].ToList()
		if err != nil {
			return nil, fmt.Errorf("case: datum list must be a proper list")
		}

		matched := false
		for _, datum := range datums {
			if key.Eqv(datum) {
				matched = true
				break
			}
		}

		if matched {
			// ((<datum> ...) => proc)
			if len(parts) == 3 && parts[1].Type == types.TypeSymbol && parts[1].SymbolVal == "=>" {
				proc, err := Eval(parts[2], e)
				if err != nil {
					return nil, err
				}
				return applyProc(proc, []*types.Value{key})
			}
			return evalBodyTCO(parts[1:], e)
		}
	}

	return resultVal(types.Void()), nil
}
