package gojand

import (
	"fmt"
	"html"

	"github.com/dop251/goja"
	"github.com/microcosm-cc/bluemonday"
)

var strictPolicy = bluemonday.StrictPolicy()

// newHTMLModule creates a fresh "html" namespace object bound to the given
// runtime.
func newHTMLModule(runtime *goja.Runtime, policies map[string]*bluemonday.Policy) *goja.Object {
	obj := runtime.NewObject()

	mustSet := func(name string, fn func(goja.FunctionCall) goja.Value) {
		err := obj.Set(name, fn)
		if err != nil {
			panic(fmt.Sprintf("set html.%s: %v", name, err))
		}
	}

	mustSet("encode", htmlEncode(runtime))
	mustSet("decode", htmlDecode(runtime))
	mustSet("strip", htmlStrip(runtime))
	mustSet("strip_policy", htmlStripPolicy(runtime, policies))

	return obj
}

// htmlEncode wraps html.EscapeString.
func htmlEncode(runtime *goja.Runtime) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			panic(runtime.NewTypeError(
				"html.encode: expected 1 argument, got %d", len(call.Arguments)))
		}

		s := call.Arguments[0].String()

		return runtime.ToValue(html.EscapeString(s))
	}
}

// htmlDecode wraps html.UnescapeString.
func htmlDecode(runtime *goja.Runtime) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			panic(runtime.NewTypeError(
				"html.decode: expected 1 argument, got %d", len(call.Arguments)))
		}

		s := call.Arguments[0].String()

		return runtime.ToValue(html.UnescapeString(s))
	}
}

// htmlStrip uses a strict bluemonday policy to strip all HTML tags.
func htmlStrip(runtime *goja.Runtime) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			panic(runtime.NewTypeError(
				"html.strip: expected 1 argument, got %d", len(call.Arguments)))
		}

		s := call.Arguments[0].String()

		return runtime.ToValue(strictPolicy.Sanitize(s))
	}
}

// htmlStripPolicy sanitizes using a named policy.
func htmlStripPolicy(runtime *goja.Runtime, policies map[string]*bluemonday.Policy) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 2 {
			panic(runtime.NewTypeError(
				"html.strip_policy: expected 2 arguments, got %d", len(call.Arguments)))
		}

		s := call.Arguments[0].String()
		name := call.Arguments[1].String()

		if policies == nil {
			panic(runtime.NewTypeError(
				"html.strip_policy: no policies configured"))
		}

		policy, exists := policies[name]
		if !exists {
			panic(runtime.NewTypeError(
				"html.strip_policy: unknown policy %q", name))
		}

		return runtime.ToValue(policy.Sanitize(s))
	}
}
