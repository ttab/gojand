package gojand

import (
	"context"
	"fmt"
	"sync"

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
	session, err := t.NewSession(ctx)
	if err != nil {
		return nil, err
	}

	defer session.Close()

	return session.Call(t.funcName, arg)
}

// Session runs the compiled script in a single runtime so that the host
// can make repeated function calls without paying runtime setup per call.
// A Session is not safe for concurrent use; create one per goroutine.
type Session struct {
	runtime   *goja.Runtime
	done      chan struct{}
	closeOnce sync.Once
}

// NewSession creates a runtime with the modules and globals set up, runs
// the script's top-level code, and returns a session for calling the
// functions it defined. The context cancels or times out script execution
// within the session. Callers must Close the session when done.
func (t *Transformer) NewSession(ctx context.Context) (*Session, error) {
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

	// Set user globals. Maps and slices are converted to native JS
	// values so that scripts get real JS semantics for them.
	for k, v := range t.globals {
		err := runtime.Set(k, toJSValue(runtime, v))
		if err != nil {
			return nil, fmt.Errorf("set global %q: %w", k, err)
		}
	}

	session := Session{
		runtime: runtime,
		done:    make(chan struct{}),
	}

	// Context cancellation.
	go func() {
		select {
		case <-ctx.Done():
			runtime.Interrupt(ctx.Err())
		case <-session.done:
		}
	}()

	// Run top-level code to define functions.
	_, err = runtime.RunProgram(t.program)
	if err != nil {
		session.Close()

		return nil, fmt.Errorf("run script: %w", err)
	}

	return &session, nil
}

// HasFunction reports whether the script defines a function with the
// given name.
func (s *Session) HasFunction(name string) bool {
	fnVal := s.runtime.Get(name)
	if fnVal == nil || goja.IsUndefined(fnVal) {
		return false
	}

	_, ok := goja.AssertFunction(fnVal)

	return ok
}

// Call invokes a named function defined by the script. The argument is
// converted to a native JS value, and the result is exported back as
// plain Go values (map[string]any, []any, primitives).
func (s *Session) Call(name string, arg any) (any, error) {
	fnVal := s.runtime.Get(name)
	if fnVal == nil || goja.IsUndefined(fnVal) {
		return nil, fmt.Errorf("function %q is not defined", name)
	}

	fn, ok := goja.AssertFunction(fnVal)
	if !ok {
		return nil, fmt.Errorf(
			"expected %q to be a function", name)
	}

	result, err := fn(goja.Undefined(), toJSValue(s.runtime, arg))
	if err != nil {
		return nil, fmt.Errorf("call %q: %w", name, err)
	}

	return result.Export(), nil
}

// Close releases the session's context watcher. It is safe to call
// multiple times.
func (s *Session) Close() {
	s.closeOnce.Do(func() {
		close(s.done)
	})
}
