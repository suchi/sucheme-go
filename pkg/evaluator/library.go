// Package evaluator - library.go implements R7RS library system.
//
// Libraries provide a module system for organizing Scheme code.
// (define-library <name> <declarations>...)
// (import <import-set>...)
package evaluator

import (
	"fmt"
	"strings"

	"github.com/suchi/sucheme-go/pkg/env"
	"github.com/suchi/sucheme-go/pkg/types"
)

// Library represents a Scheme library.
type Library struct {
	Name    string            // e.g. "scheme base"
	Env     *env.Env          // library's internal environment
	Exports map[string]string // exported name -> internal name
}

// libraryRegistry stores all defined libraries.
// Initialized by RegisterLibraryBuiltins or lazily on first use.
//
// NOTE: This is a global variable, not thread-safe. Same caveat as other
// global state in the evaluator (exceptionHandlerStack, parameterStack, etc.).
var libraryRegistry = make(map[string]*Library)

// RegisterLibraryBuiltins registers the library system.
func RegisterLibraryBuiltins(e *env.Env) {
	// Reset library registry to avoid leaking old environments
	libraryRegistry = make(map[string]*Library)

	// Register built-in libraries for the given environment
	registerSchemeBase(e)
}

// registerSchemeBase registers the (scheme base) library.
func registerSchemeBase(e *env.Env) {
	lib := &Library{
		Name:    "scheme base",
		Env:     e,
		Exports: make(map[string]string),
	}
	// (scheme base) exports everything in the given global environment
	for _, key := range e.Keys() {
		lib.Exports[key] = key
	}
	libraryRegistry["scheme base"] = lib
}

// evalDefineLibrary implements (define-library <name> <declarations>...)
// Declarations are collected first, then processed in order: export → import → begin.
// This ensures imports are available before begin bodies execute.
func evalDefineLibrary(expr *types.Value, e *env.Env) (*evalResult, error) {
	args, err := listArgs(expr)
	if err != nil || len(args) < 1 {
		return nil, fmt.Errorf("define-library: invalid syntax")
	}

	name, err := parseLibraryName(args[0])
	if err != nil {
		return nil, err
	}

	// Library environment has nil parent to enforce explicit imports
	lib := &Library{
		Name:    name,
		Env:     env.New(nil),
		Exports: make(map[string]string),
	}

	// Collect declarations by type
	var importDecls []*types.Value
	var beginDecls []*types.Value

	for _, decl := range args[1:] {
		if decl.Type != types.TypePair {
			return nil, fmt.Errorf("define-library: invalid declaration")
		}
		head := decl.PairVal.Car
		if head.Type != types.TypeSymbol {
			return nil, fmt.Errorf("define-library: expected declaration keyword")
		}

		switch head.SymbolVal {
		case "export":
			exportList, err := decl.PairVal.Cdr.ToList()
			if err != nil {
				return nil, fmt.Errorf("define-library: invalid export list")
			}
			for _, exp := range exportList {
				if exp.Type == types.TypeSymbol {
					lib.Exports[exp.SymbolVal] = exp.SymbolVal
				} else if exp.Type == types.TypePair {
					rename, err := exp.ToList()
					if err != nil || len(rename) != 3 {
						return nil, fmt.Errorf("define-library: invalid export spec")
					}
					if rename[0].Type != types.TypeSymbol || rename[0].SymbolVal != "rename" {
						return nil, fmt.Errorf("define-library: expected rename in export")
					}
					if rename[1].Type != types.TypeSymbol || rename[2].Type != types.TypeSymbol {
						return nil, fmt.Errorf("define-library: rename requires two symbols")
					}
					lib.Exports[rename[2].SymbolVal] = rename[1].SymbolVal
				} else {
					return nil, fmt.Errorf("define-library: invalid export spec: expected symbol or (rename ...)")
				}
			}

		case "begin":
			body, err := decl.PairVal.Cdr.ToList()
			if err != nil {
				return nil, fmt.Errorf("define-library: invalid begin body")
			}
			beginDecls = append(beginDecls, body...)

		case "import":
			importSets, err := decl.PairVal.Cdr.ToList()
			if err != nil {
				return nil, fmt.Errorf("define-library: invalid import")
			}
			importDecls = append(importDecls, importSets...)

		default:
			return nil, fmt.Errorf("define-library: unknown declaration: %s", head.SymbolVal)
		}
	}

	// Process in order: imports first, then begin bodies
	for _, impSet := range importDecls {
		if err := processImport(impSet, lib.Env); err != nil {
			return nil, fmt.Errorf("define-library %s: %w", name, err)
		}
	}

	for _, bodyExpr := range beginDecls {
		_, err := Eval(bodyExpr, lib.Env)
		if err != nil {
			return nil, fmt.Errorf("define-library %s: %w", name, err)
		}
	}

	libraryRegistry[name] = lib
	return resultVal(types.Void()), nil
}

// evalImport implements (import <import-set>...)
func evalImport(expr *types.Value, e *env.Env) (*evalResult, error) {
	args, err := listArgs(expr)
	if err != nil || len(args) == 0 {
		return nil, fmt.Errorf("import: expected at least one import set")
	}

	for _, impSet := range args {
		if err := processImport(impSet, e); err != nil {
			return nil, err
		}
	}

	return resultVal(types.Void()), nil
}

// processImport processes a single import set and adds bindings to the target environment.
func processImport(impSet *types.Value, target *env.Env) error {
	if impSet.Type == types.TypePair && impSet.PairVal.Car.Type == types.TypeSymbol {
		keyword := impSet.PairVal.Car.SymbolVal

		switch keyword {
		case "only":
			return processImportOnly(impSet, target)
		case "except":
			return processImportExcept(impSet, target)
		case "prefix":
			return processImportPrefix(impSet, target)
		case "rename":
			return processImportRename(impSet, target)
		}
	}

	name, err := parseLibraryName(impSet)
	if err != nil {
		return err
	}

	lib, ok := libraryRegistry[name]
	if !ok {
		return fmt.Errorf("import: unknown library: %s", name)
	}

	for extName, intName := range lib.Exports {
		val, found := lib.Env.Get(intName)
		if !found {
			return fmt.Errorf("import: library %s exports undefined identifier: %s", name, intName)
		}
		target.Define(extName, val)
	}

	return nil
}

// processImportOnly implements (only <import-set> <id>...)
func processImportOnly(impSet *types.Value, target *env.Env) error {
	args, err := impSet.PairVal.Cdr.ToList()
	if err != nil || len(args) < 2 {
		return fmt.Errorf("import only: invalid syntax")
	}

	temp := env.New(nil)
	if err := processImport(args[0], temp); err != nil {
		return err
	}

	for _, id := range args[1:] {
		if id.Type != types.TypeSymbol {
			return fmt.Errorf("import only: expected symbol")
		}
		val, found := temp.Get(id.SymbolVal)
		if !found {
			return fmt.Errorf("import only: %s not found in library", id.SymbolVal)
		}
		target.Define(id.SymbolVal, val)
	}
	return nil
}

// processImportExcept implements (except <import-set> <id>...)
func processImportExcept(impSet *types.Value, target *env.Env) error {
	args, err := impSet.PairVal.Cdr.ToList()
	if err != nil || len(args) < 2 {
		return fmt.Errorf("import except: invalid syntax")
	}

	excluded := make(map[string]bool)
	for _, id := range args[1:] {
		if id.Type != types.TypeSymbol {
			return fmt.Errorf("import except: expected symbol")
		}
		excluded[id.SymbolVal] = true
	}

	temp := env.New(nil)
	if err := processImport(args[0], temp); err != nil {
		return err
	}

	for _, k := range temp.Keys() {
		if !excluded[k] {
			val, _ := temp.Get(k)
			target.Define(k, val)
		}
	}
	return nil
}

// processImportPrefix implements (prefix <import-set> <prefix>)
func processImportPrefix(impSet *types.Value, target *env.Env) error {
	args, err := impSet.PairVal.Cdr.ToList()
	if err != nil || len(args) != 2 {
		return fmt.Errorf("import prefix: invalid syntax")
	}
	if args[1].Type != types.TypeSymbol {
		return fmt.Errorf("import prefix: expected symbol for prefix")
	}
	prefix := args[1].SymbolVal

	temp := env.New(nil)
	if err := processImport(args[0], temp); err != nil {
		return err
	}

	for _, k := range temp.Keys() {
		val, _ := temp.Get(k)
		target.Define(prefix+k, val)
	}
	return nil
}

// processImportRename implements (rename <import-set> (<old> <new>)...)
func processImportRename(impSet *types.Value, target *env.Env) error {
	args, err := impSet.PairVal.Cdr.ToList()
	if err != nil || len(args) < 2 {
		return fmt.Errorf("import rename: invalid syntax")
	}

	renames := make(map[string]string)
	for _, pair := range args[1:] {
		rename, err := pair.ToList()
		if err != nil || len(rename) != 2 {
			return fmt.Errorf("import rename: expected (old new) pair")
		}
		if rename[0].Type != types.TypeSymbol || rename[1].Type != types.TypeSymbol {
			return fmt.Errorf("import rename: expected symbols")
		}
		renames[rename[0].SymbolVal] = rename[1].SymbolVal
	}

	temp := env.New(nil)
	if err := processImport(args[0], temp); err != nil {
		return err
	}

	for _, k := range temp.Keys() {
		val, _ := temp.Get(k)
		newName := k
		if renamed, ok := renames[k]; ok {
			newName = renamed
		}
		target.Define(newName, val)
	}
	return nil
}

// parseLibraryName converts a library name form to a string.
func parseLibraryName(v *types.Value) (string, error) {
	parts, err := v.ToList()
	if err != nil {
		if v.Type == types.TypeSymbol {
			return v.SymbolVal, nil
		}
		return "", fmt.Errorf("invalid library name")
	}
	names := make([]string, len(parts))
	for i, p := range parts {
		if p.Type == types.TypeSymbol {
			names[i] = p.SymbolVal
		} else if p.Type == types.TypeInteger {
			names[i] = fmt.Sprintf("%d", p.IntVal)
		} else {
			return "", fmt.Errorf("invalid library name component")
		}
	}
	return strings.Join(names, " "), nil
}
