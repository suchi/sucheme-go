//go:build js && wasm

// Package main provides the WebAssembly entry point for the sucheme interpreter.
package main

import (
	"fmt"
	"strings"
	"sync"
	"syscall/js"

	"github.com/suchi/sucheme-go/pkg/builtin"
	"github.com/suchi/sucheme-go/pkg/env"
	"github.com/suchi/sucheme-go/pkg/evaluator"
	"github.com/suchi/sucheme-go/pkg/types"
)

var (
	globalEnv *env.Env
	// outputBuf is the current output capture target.
	// I/O procedures always write to the current outputBuf.
	outputBuf *strings.Builder
	outputMu  sync.Mutex
)

func init() {
	globalEnv = env.Global()
	builtin.RegisterBuiltins(globalEnv)
	evaluator.RegisterApply(globalEnv)
	evaluator.RegisterExceptionBuiltins(globalEnv)
	evaluator.RegisterMapForEach(globalEnv)
	evaluator.RegisterParameterBuiltins(globalEnv)
	evaluator.RegisterPortBuiltins(globalEnv)
	evaluator.RegisterContinuationBuiltins(globalEnv)
	evaluator.RegisterLibraryBuiltins(globalEnv)
	registerWASMIO(globalEnv)
}

func main() {
	js.Global().Set("suchemeEval", js.FuncOf(evalJS))
	js.Global().Set("suchemeReset", js.FuncOf(resetJS))

	// Keep the Go program running
	select {}
}

func evalJS(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return map[string]interface{}{
			"result": "",
			"error":  "no input provided",
			"output": "",
		}
	}

	input := args[0].String()

	// Swap in a fresh output buffer for this evaluation
	outputMu.Lock()
	var buf strings.Builder
	outputBuf = &buf
	outputMu.Unlock()

	result, err := evaluator.EvalString(input, globalEnv)

	outputMu.Lock()
	captured := outputBuf.String()
	outputBuf = nil
	outputMu.Unlock()

	if err != nil {
		return map[string]interface{}{
			"result": "",
			"error":  err.Error(),
			"output": captured,
		}
	}

	resultStr := ""
	if result != nil && result.Type != types.TypeVoid {
		resultStr = result.String()
	}

	return map[string]interface{}{
		"result": resultStr,
		"output": captured,
		"error":  "",
	}
}

func resetJS(this js.Value, args []js.Value) interface{} {
	globalEnv = env.Global()
	builtin.RegisterBuiltins(globalEnv)
	evaluator.RegisterApply(globalEnv)
	evaluator.RegisterExceptionBuiltins(globalEnv)
	evaluator.RegisterMapForEach(globalEnv)
	evaluator.RegisterParameterBuiltins(globalEnv)
	evaluator.RegisterPortBuiltins(globalEnv)
	evaluator.RegisterContinuationBuiltins(globalEnv)
	evaluator.RegisterLibraryBuiltins(globalEnv)
	registerWASMIO(globalEnv)
	return nil
}

// registerWASMIO registers I/O procedures that write to the current outputBuf.
// These are defined once and always reference the current outputBuf pointer.
func registerWASMIO(e *env.Env) {
	e.Define("display", types.NewProcedure("display", func(args []*types.Value) (*types.Value, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("expected at least 1 argument")
		}
		outputMu.Lock()
		if outputBuf != nil {
			outputBuf.WriteString(args[0].Display())
		}
		outputMu.Unlock()
		return types.Void(), nil
	}))

	e.Define("write", types.NewProcedure("write", func(args []*types.Value) (*types.Value, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("expected at least 1 argument")
		}
		outputMu.Lock()
		if outputBuf != nil {
			outputBuf.WriteString(args[0].String())
		}
		outputMu.Unlock()
		return types.Void(), nil
	}))

	e.Define("newline", types.NewProcedure("newline", func(args []*types.Value) (*types.Value, error) {
		outputMu.Lock()
		if outputBuf != nil {
			outputBuf.WriteString("\n")
		}
		outputMu.Unlock()
		return types.Void(), nil
	}))
}
