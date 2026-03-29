//go:build !js

// Package evaluator - include.go implements R7RS include and include-ci
// for non-WASM builds.
package evaluator

import (
	"fmt"
	"os"
	"strings"

	"github.com/suchi/sucheme-go/pkg/env"
	"github.com/suchi/sucheme-go/pkg/parser"
	"github.com/suchi/sucheme-go/pkg/types"
)

// evalInclude implements R7RS include and include-ci.
// For include-ci, identifiers in the source are folded to lowercase.
//
// NOTE: R7RS requires case-folding only for identifiers. This implementation
// folds the entire source including string/character literals, which is
// non-conformant. This is a known limitation.
//
// SECURITY: include reads files from the filesystem using os.ReadFile on
// arbitrary paths. In untrusted evaluation contexts, this could allow reading
// arbitrary host files. Consider restricting allowed paths in
// security-sensitive deployments.
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
