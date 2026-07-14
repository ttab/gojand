package gojand

import (
	"github.com/dop251/goja"
)

// toJSValue recursively converts plain Go maps and slices into native
// JavaScript objects and arrays in the runtime.
//
// Passing Go values directly to the runtime makes goja wrap them as live
// views onto the Go data. That mostly works, but reading a nested slice
// through an object property wraps a copy of the Go slice header, so JS
// array mutations that change the length — push, pop, splice — are
// silently lost while index assignments still work. Scripts must get
// native JS values so that arrays behave like arrays.
func toJSValue(runtime *goja.Runtime, v any) goja.Value {
	switch val := v.(type) {
	case map[string]any:
		obj := runtime.NewObject()

		for k, item := range val {
			err := obj.Set(k, toJSValue(runtime, item))
			if err != nil {
				panic(runtime.NewTypeError(
					"set property %q: %v", k, err))
			}
		}

		return obj
	case []any:
		items := make([]any, len(val))
		for i, item := range val {
			items[i] = toJSValue(runtime, item)
		}

		return runtime.NewArray(items...)
	case []map[string]any:
		items := make([]any, len(val))
		for i, item := range val {
			items[i] = toJSValue(runtime, item)
		}

		return runtime.NewArray(items...)
	default:
		return runtime.ToValue(v)
	}
}
