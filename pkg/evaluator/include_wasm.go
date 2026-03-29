//go:build js

// Package evaluator - include_wasm.go disables include/include-ci in WASM builds.
// Filesystem access via os.ReadFile is not available in WASM environments,
// so include is not supported.
package evaluator

import (
	"fmt"

	"github.com/suchi/sucheme-go/pkg/env"
	"github.com/suchi/sucheme-go/pkg/types"
)

// evalInclude returns an error in WASM builds because filesystem access
// is not available.
func evalInclude(expr *types.Value, e *env.Env, caseInsensitive bool) (*evalResult, error) {
	return nil, fmt.Errorf("include: not supported in WASM builds")
}
