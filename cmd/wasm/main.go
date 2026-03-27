//go:build js && wasm

// Package main provides the WebAssembly entry point for the sucheme interpreter.
package main

import (
	"fmt"
	"strings"
	"syscall/js"

	"github.com/suchi/sucheme-go/pkg/builtin"
	"github.com/suchi/sucheme-go/pkg/env"
	"github.com/suchi/sucheme-go/pkg/evaluator"
	"github.com/suchi/sucheme-go/pkg/types"
)

var globalEnv *env.Env

func init() {
	globalEnv = env.Global()
	builtin.RegisterBuiltins(globalEnv)
	evaluator.RegisterApply(globalEnv)
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
			"error":  "no input provided",
			"output": "",
		}
	}

	input := args[0].String()
	var output strings.Builder

	// Override display/write to capture output
	globalEnv.Define("display", types.NewProcedure("display", func(args []*types.Value) (*types.Value, error) {
		if len(args) >= 1 {
			output.WriteString(args[0].Display())
		}
		return types.Void(), nil
	}))

	globalEnv.Define("write", types.NewProcedure("write", func(args []*types.Value) (*types.Value, error) {
		if len(args) >= 1 {
			output.WriteString(args[0].String())
		}
		return types.Void(), nil
	}))

	globalEnv.Define("newline", types.NewProcedure("newline", func(args []*types.Value) (*types.Value, error) {
		output.WriteString("\n")
		return types.Void(), nil
	}))

	result, err := evaluator.EvalString(input, globalEnv)
	if err != nil {
		return map[string]interface{}{
			"error":  err.Error(),
			"output": output.String(),
		}
	}

	resultStr := ""
	if result != nil && result.Type != types.TypeVoid {
		resultStr = result.String()
	}

	return map[string]interface{}{
		"result": resultStr,
		"output": output.String(),
		"error":  "",
	}
}

func resetJS(this js.Value, args []js.Value) interface{} {
	globalEnv = env.Global()
	builtin.RegisterBuiltins(globalEnv)
	evaluator.RegisterApply(globalEnv)
	registerWASMIO(globalEnv)
	return nil
}

func registerWASMIO(e *env.Env) {
	e.Define("display", types.NewProcedure("display", func(args []*types.Value) (*types.Value, error) {
		if len(args) >= 1 {
			fmt.Print(args[0].Display())
		}
		return types.Void(), nil
	}))

	e.Define("write", types.NewProcedure("write", func(args []*types.Value) (*types.Value, error) {
		if len(args) >= 1 {
			fmt.Print(args[0].String())
		}
		return types.Void(), nil
	}))

	e.Define("newline", types.NewProcedure("newline", func(args []*types.Value) (*types.Value, error) {
		fmt.Println()
		return types.Void(), nil
	}))
}
