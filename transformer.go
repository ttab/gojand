package gojand

import (
	"context"
	"fmt"

	"github.com/dop251/goja"
	"github.com/evanw/esbuild/pkg/api"
	"github.com/ttab/newsdoc"
)

// Transformer compiles a script once and can transform documents repeatedly.
// It is safe for concurrent use: each Transform call creates a fresh goja
// Runtime.
type Transformer struct {
	program  *goja.Program
	funcName string
	globals  map[string]any
	cfg      *config
}

// NewTransformer compiles the given script and returns a Transformer. The
// script must define a function (default name "transform") that takes a
// document object and returns a document object.
//
// If WithTypeScript() was set, the script is transpiled from TypeScript to
// JavaScript via esbuild before compilation.
func NewTransformer(script string, opts ...Option) (*Transformer, error) {
	cfg := newConfig(opts)

	src := script

	if cfg.typescript {
		result := api.Transform(src, api.TransformOptions{
			Loader: api.LoaderTS,
		})

		if len(result.Errors) > 0 {
			return nil, fmt.Errorf("transpile TypeScript: %s", result.Errors[0].Text)
		}

		src = string(result.Code)
	}

	program, err := goja.Compile("", src, true)
	if err != nil {
		return nil, fmt.Errorf("compile script: %w", err)
	}

	return &Transformer{
		program:  program,
		funcName: cfg.funcName,
		globals:  cfg.globals,
		cfg:      cfg,
	}, nil
}

// Transform converts a document to a map, runs the script's transform
// function, and converts the result back to a Document.
func (t *Transformer) Transform(ctx context.Context, doc newsdoc.Document) (newsdoc.Document, error) {
	docMap := DocumentToMap(doc)

	result, err := t.callTransform(ctx, docMap)
	if err != nil {
		return newsdoc.Document{}, err
	}

	resultMap, ok := toMap(result)
	if !ok {
		return newsdoc.Document{}, fmt.Errorf(
			"transform function must return an object, got %T", result)
	}

	return MapToDocument(resultMap)
}

func (t *Transformer) callTransform(ctx context.Context, arg any) (any, error) {
	runtime := goja.New()

	// Set up modules.
	err := runtime.Set("nd", newNDModule(runtime))
	if err != nil {
		return nil, fmt.Errorf("set nd module: %w", err)
	}

	err = runtime.Set("html", newHTMLModule(runtime, t.cfg.policies))
	if err != nil {
		return nil, fmt.Errorf("set html module: %w", err)
	}

	// Set user globals.
	for k, v := range t.globals {
		err := runtime.Set(k, v)
		if err != nil {
			return nil, fmt.Errorf("set global %q: %w", k, err)
		}
	}

	// Context cancellation.
	done := make(chan struct{})
	defer close(done)

	go func() {
		select {
		case <-ctx.Done():
			runtime.Interrupt(ctx.Err())
		case <-done:
		}
	}()

	// Run top-level code to define functions.
	_, err = runtime.RunProgram(t.program)
	if err != nil {
		return nil, fmt.Errorf("run script: %w", err)
	}

	// Get the transform function.
	fnVal := runtime.Get(t.funcName)
	if fnVal == nil || goja.IsUndefined(fnVal) {
		return nil, fmt.Errorf("function %q is not defined", t.funcName)
	}

	fn, ok := goja.AssertFunction(fnVal)
	if !ok {
		return nil, fmt.Errorf(
			"expected %q to be a function", t.funcName)
	}

	// Call the function with the document map.
	result, err := fn(goja.Undefined(), runtime.ToValue(arg))
	if err != nil {
		return nil, fmt.Errorf("call %q: %w", t.funcName, err)
	}

	return result.Export(), nil
}
