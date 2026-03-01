package gojand

import (
	"fmt"

	"github.com/dop251/goja"
)

// blockMatcher tests whether a block map matches some criteria.
type blockMatcher func(block map[string]any) (bool, error)

// toMatcher converts a goja value to a blockMatcher. It accepts either a
// criteria object (all specified fields must match) or a callable (arbitrary
// predicate returning a truthy value).
func toMatcher(runtime *goja.Runtime, val goja.Value) (blockMatcher, error) {
	if fn, ok := goja.AssertFunction(val); ok {
		return callableMatcher(runtime, fn), nil
	}

	obj := val.ToObject(runtime)
	if obj == nil {
		return nil, fmt.Errorf(
			"matcher must be an object or function, got %s", val.ExportType())
	}

	return criteriaMatcherFromObject(obj), nil
}

// criteriaMatcherFromObject creates a matcher that checks if all key/value
// pairs in the criteria object match the corresponding fields in a block map.
func criteriaMatcherFromObject(criteria *goja.Object) blockMatcher {
	type pair struct {
		key   string
		value string
	}

	keys := criteria.Keys()
	pairs := make([]pair, 0, len(keys))

	for _, k := range keys {
		v := criteria.Get(k)
		if v != nil && v != goja.Undefined() && v != goja.Null() {
			pairs = append(pairs, pair{key: k, value: v.String()})
		}
	}

	return func(block map[string]any) (bool, error) {
		for _, p := range pairs {
			v, ok := block[p.key]
			if !ok {
				return false, nil
			}

			s, ok := v.(string)
			if !ok || s != p.value {
				return false, nil
			}
		}

		return true, nil
	}
}

// callableMatcher creates a matcher that calls a JS function with the block
// map and returns whether the result is truthy.
func callableMatcher(runtime *goja.Runtime, fn goja.Callable) blockMatcher {
	return func(block map[string]any) (bool, error) {
		result, err := fn(goja.Undefined(), runtime.ToValue(block))
		if err != nil {
			return false, fmt.Errorf("call matcher function: %w", err)
		}

		return result.ToBoolean(), nil
	}
}
