package gojand

import (
	"fmt"

	"github.com/dop251/goja"
)

// newNDModule creates a fresh "nd" namespace object bound to the given runtime.
func newNDModule(runtime *goja.Runtime) *goja.Object {
	obj := runtime.NewObject()

	mustSet := func(name string, fn func(goja.FunctionCall) goja.Value) {
		err := obj.Set(name, fn)
		if err != nil {
			panic(fmt.Sprintf("set nd.%s: %v", name, err))
		}
	}

	// Block querying
	mustSet("first_block", ndFirstBlock(runtime))
	mustSet("all_blocks", ndAllBlocks(runtime))
	mustSet("has_block", ndHasBlock(runtime))

	// Block manipulation
	mustSet("drop_blocks", ndDropBlocks(runtime))
	mustSet("dedupe_blocks", ndDedupeBlocks(runtime))
	mustSet("alter_blocks", ndAlterBlocks(runtime))
	mustSet("alter_first_block", ndAlterFirstBlock(runtime))
	mustSet("upsert_block", ndUpsertBlock(runtime))
	mustSet("add_or_replace_block", ndAddOrReplaceBlock(runtime))

	// DataMap helpers
	mustSet("get_data", ndGetData(runtime))
	mustSet("upsert_data", ndUpsertData(runtime))
	mustSet("data_defaults", ndDataDefaults(runtime))

	return obj
}

// requireBlockSlice extracts and validates a []any from the first argument.
// Nil and undefined are treated as an empty slice so that callers can pass
// optional document fields (e.g. doc.meta) without a nil guard.
func requireBlockSlice(runtime *goja.Runtime, val goja.Value, funcName string) []any {
	if val == nil || goja.IsUndefined(val) || goja.IsNull(val) {
		return []any{}
	}

	exported := val.Export()

	slice, ok := toSlice(exported)
	if !ok {
		panic(runtime.NewTypeError(
			"%s: first argument must be an array, got %T", funcName, exported))
	}

	return slice
}

// ndFirstBlock returns the first block matching the criteria, or null.
func ndFirstBlock(runtime *goja.Runtime) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 2 {
			panic(runtime.NewTypeError(
				"nd.first_block: expected 2 arguments, got %d", len(call.Arguments)))
		}

		blocks := requireBlockSlice(runtime, call.Arguments[0], "nd.first_block")

		match, err := toMatcher(runtime, call.Arguments[1])
		if err != nil {
			panic(runtime.NewTypeError("nd.first_block: %s", err))
		}

		for _, item := range blocks {
			m, ok := toMap(item)
			if !ok {
				continue
			}

			matched, err := match(m)
			if err != nil {
				panic(runtime.NewTypeError("%s", err))
			}

			if matched {
				return runtime.ToValue(m)
			}
		}

		return goja.Null()
	}
}

// ndAllBlocks returns all blocks matching the criteria.
func ndAllBlocks(runtime *goja.Runtime) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 2 {
			panic(runtime.NewTypeError(
				"nd.all_blocks: expected 2 arguments, got %d", len(call.Arguments)))
		}

		blocks := requireBlockSlice(runtime, call.Arguments[0], "nd.all_blocks")

		match, err := toMatcher(runtime, call.Arguments[1])
		if err != nil {
			panic(runtime.NewTypeError("nd.all_blocks: %s", err))
		}

		var result []any

		for _, item := range blocks {
			m, ok := toMap(item)
			if !ok {
				continue
			}

			matched, err := match(m)
			if err != nil {
				panic(runtime.NewTypeError("%s", err))
			}

			if matched {
				result = append(result, m)
			}
		}

		if result == nil {
			result = []any{}
		}

		return runtime.ToValue(result)
	}
}

// ndHasBlock returns true if any block matches the criteria.
func ndHasBlock(runtime *goja.Runtime) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 2 {
			panic(runtime.NewTypeError(
				"nd.has_block: expected 2 arguments, got %d", len(call.Arguments)))
		}

		blocks := requireBlockSlice(runtime, call.Arguments[0], "nd.has_block")

		match, err := toMatcher(runtime, call.Arguments[1])
		if err != nil {
			panic(runtime.NewTypeError("nd.has_block: %s", err))
		}

		for _, item := range blocks {
			m, ok := toMap(item)
			if !ok {
				continue
			}

			matched, err := match(m)
			if err != nil {
				panic(runtime.NewTypeError("%s", err))
			}

			if matched {
				return runtime.ToValue(true)
			}
		}

		return runtime.ToValue(false)
	}
}

// ndDropBlocks returns a new array with all matching blocks removed.
func ndDropBlocks(runtime *goja.Runtime) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 2 {
			panic(runtime.NewTypeError(
				"nd.drop_blocks: expected 2 arguments, got %d", len(call.Arguments)))
		}

		blocks := requireBlockSlice(runtime, call.Arguments[0], "nd.drop_blocks")

		match, err := toMatcher(runtime, call.Arguments[1])
		if err != nil {
			panic(runtime.NewTypeError("nd.drop_blocks: %s", err))
		}

		var result []any

		for _, item := range blocks {
			m, ok := toMap(item)
			if !ok {
				result = append(result, item)
				continue
			}

			matched, err := match(m)
			if err != nil {
				panic(runtime.NewTypeError("%s", err))
			}

			if !matched {
				result = append(result, item)
			}
		}

		if result == nil {
			result = []any{}
		}

		return runtime.ToValue(result)
	}
}

// ndDedupeBlocks returns a new array keeping only the first match.
func ndDedupeBlocks(runtime *goja.Runtime) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 2 {
			panic(runtime.NewTypeError(
				"nd.dedupe_blocks: expected 2 arguments, got %d", len(call.Arguments)))
		}

		blocks := requireBlockSlice(runtime, call.Arguments[0], "nd.dedupe_blocks")

		match, err := toMatcher(runtime, call.Arguments[1])
		if err != nil {
			panic(runtime.NewTypeError("nd.dedupe_blocks: %s", err))
		}

		var result []any

		found := false

		for _, item := range blocks {
			m, ok := toMap(item)
			if !ok {
				result = append(result, item)
				continue
			}

			matched, err := match(m)
			if err != nil {
				panic(runtime.NewTypeError("%s", err))
			}

			if matched {
				if !found {
					found = true
					result = append(result, item)
				}
			} else {
				result = append(result, item)
			}
		}

		if result == nil {
			result = []any{}
		}

		return runtime.ToValue(result)
	}
}

// ndAlterBlocks applies fn to all matching blocks, returning a new array.
func ndAlterBlocks(runtime *goja.Runtime) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 3 {
			panic(runtime.NewTypeError(
				"nd.alter_blocks: expected 3 arguments, got %d", len(call.Arguments)))
		}

		blocks := requireBlockSlice(runtime, call.Arguments[0], "nd.alter_blocks")

		match, err := toMatcher(runtime, call.Arguments[1])
		if err != nil {
			panic(runtime.NewTypeError("nd.alter_blocks: %s", err))
		}

		fn, ok := goja.AssertFunction(call.Arguments[2])
		if !ok {
			panic(runtime.NewTypeError(
				"nd.alter_blocks: third argument must be callable"))
		}

		var result []any

		for _, item := range blocks {
			m, ok := toMap(item)
			if !ok {
				result = append(result, item)
				continue
			}

			matched, err := match(m)
			if err != nil {
				panic(runtime.NewTypeError("%s", err))
			}

			if matched {
				altered, err := fn(goja.Undefined(), runtime.ToValue(m))
				if err != nil {
					panic(runtime.NewTypeError(
						"nd.alter_blocks: call alter function: %s", err))
				}

				result = append(result, altered.Export())
			} else {
				result = append(result, item)
			}
		}

		if result == nil {
			result = []any{}
		}

		return runtime.ToValue(result)
	}
}

// ndAlterFirstBlock applies fn to only the first matching block.
func ndAlterFirstBlock(runtime *goja.Runtime) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 3 {
			panic(runtime.NewTypeError(
				"nd.alter_first_block: expected 3 arguments, got %d", len(call.Arguments)))
		}

		blocks := requireBlockSlice(runtime, call.Arguments[0], "nd.alter_first_block")

		match, err := toMatcher(runtime, call.Arguments[1])
		if err != nil {
			panic(runtime.NewTypeError("nd.alter_first_block: %s", err))
		}

		fn, ok := goja.AssertFunction(call.Arguments[2])
		if !ok {
			panic(runtime.NewTypeError(
				"nd.alter_first_block: third argument must be callable"))
		}

		var result []any

		altered := false

		for _, item := range blocks {
			if altered {
				result = append(result, item)
				continue
			}

			m, ok := toMap(item)
			if !ok {
				result = append(result, item)
				continue
			}

			matched, err := match(m)
			if err != nil {
				panic(runtime.NewTypeError("%s", err))
			}

			if matched {
				newItem, err := fn(goja.Undefined(), runtime.ToValue(m))
				if err != nil {
					panic(runtime.NewTypeError(
						"nd.alter_first_block: call alter function: %s", err))
				}

				result = append(result, newItem.Export())

				altered = true
			} else {
				result = append(result, item)
			}
		}

		if result == nil {
			result = []any{}
		}

		return runtime.ToValue(result)
	}
}

// ndUpsertBlock finds the first match and applies fn, or appends the default.
func ndUpsertBlock(runtime *goja.Runtime) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 4 {
			panic(runtime.NewTypeError(
				"nd.upsert_block: expected 4 arguments, got %d", len(call.Arguments)))
		}

		blocks := requireBlockSlice(runtime, call.Arguments[0], "nd.upsert_block")

		match, err := toMatcher(runtime, call.Arguments[1])
		if err != nil {
			panic(runtime.NewTypeError("nd.upsert_block: %s", err))
		}

		defaultBlock := call.Arguments[2].Export()

		fn, ok := goja.AssertFunction(call.Arguments[3])
		if !ok {
			panic(runtime.NewTypeError(
				"nd.upsert_block: fourth argument must be callable"))
		}

		var result []any

		found := false

		for _, item := range blocks {
			if found {
				result = append(result, item)
				continue
			}

			m, ok := toMap(item)
			if !ok {
				result = append(result, item)
				continue
			}

			matched, err := match(m)
			if err != nil {
				panic(runtime.NewTypeError("%s", err))
			}

			if matched {
				altered, err := fn(goja.Undefined(), runtime.ToValue(m))
				if err != nil {
					panic(runtime.NewTypeError(
						"nd.upsert_block: call upsert function: %s", err))
				}

				result = append(result, altered.Export())

				found = true
			} else {
				result = append(result, item)
			}
		}

		if !found {
			result = append(result, defaultBlock)
		}

		return runtime.ToValue(result)
	}
}

// ndAddOrReplaceBlock replaces the first match or appends the block.
func ndAddOrReplaceBlock(runtime *goja.Runtime) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 3 {
			panic(runtime.NewTypeError(
				"nd.add_or_replace_block: expected 3 arguments, got %d", len(call.Arguments)))
		}

		blocks := requireBlockSlice(runtime, call.Arguments[0], "nd.add_or_replace_block")

		match, err := toMatcher(runtime, call.Arguments[1])
		if err != nil {
			panic(runtime.NewTypeError("nd.add_or_replace_block: %s", err))
		}

		newBlock := call.Arguments[2].Export()

		var result []any

		replaced := false

		for _, item := range blocks {
			if replaced {
				result = append(result, item)
				continue
			}

			m, ok := toMap(item)
			if !ok {
				result = append(result, item)
				continue
			}

			matched, err := match(m)
			if err != nil {
				panic(runtime.NewTypeError("%s", err))
			}

			if matched {
				result = append(result, newBlock)
				replaced = true
			} else {
				result = append(result, item)
			}
		}

		if !replaced {
			result = append(result, newBlock)
		}

		return runtime.ToValue(result)
	}
}

// ndGetData returns a value from a block's data map. If the key is missing the
// optional default value is returned, or an empty string if no default was
// provided.
func ndGetData(runtime *goja.Runtime) func(goja.FunctionCall) goja.Value {
	emptyString := runtime.ToValue("")

	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 || len(call.Arguments) > 3 {
			panic(runtime.NewTypeError(
				"nd.get_data: expected 2 or 3 arguments, got %d", len(call.Arguments)))
		}

		block := call.Arguments[0].Export()

		blockMap, ok := toMap(block)
		if !ok {
			panic(runtime.NewTypeError(
				"nd.get_data: first argument must be an object, got %T", block))
		}

		key := call.Arguments[1].String()

		fallback := emptyString
		if len(call.Arguments) == 3 {
			fallback = call.Arguments[2]
		}

		dataVal, ok := blockMap["data"]
		if !ok || dataVal == nil {
			return fallback
		}

		dataMap, ok := toMap(dataVal)
		if !ok {
			return fallback
		}

		v, ok := dataMap[key]
		if !ok || v == nil {
			return fallback
		}

		return runtime.ToValue(v)
	}
}

// ndUpsertData merges new_data into data, returning a new map (no mutation).
func ndUpsertData(runtime *goja.Runtime) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 2 {
			panic(runtime.NewTypeError(
				"nd.upsert_data: expected 2 arguments, got %d", len(call.Arguments)))
		}

		baseExport := call.Arguments[0].Export()

		base, ok := toMap(baseExport)
		if !ok {
			panic(runtime.NewTypeError(
				"nd.upsert_data: first argument must be an object, got %T", baseExport))
		}

		updatesExport := call.Arguments[1].Export()

		updates, ok := toMap(updatesExport)
		if !ok {
			panic(runtime.NewTypeError(
				"nd.upsert_data: second argument must be an object, got %T", updatesExport))
		}

		result := make(map[string]any, len(base)+len(updates))
		for k, v := range base {
			result[k] = v
		}

		for k, v := range updates {
			result[k] = v
		}

		return runtime.ToValue(result)
	}
}

// ndDataDefaults fills in missing or empty keys from defaults.
func ndDataDefaults(runtime *goja.Runtime) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 2 {
			panic(runtime.NewTypeError(
				"nd.data_defaults: expected 2 arguments, got %d", len(call.Arguments)))
		}

		dataExport := call.Arguments[0].Export()

		data, ok := toMap(dataExport)
		if !ok {
			panic(runtime.NewTypeError(
				"nd.data_defaults: first argument must be an object, got %T", dataExport))
		}

		defaultsExport := call.Arguments[1].Export()

		defaults, ok := toMap(defaultsExport)
		if !ok {
			panic(runtime.NewTypeError(
				"nd.data_defaults: second argument must be an object, got %T", defaultsExport))
		}

		result := make(map[string]any, len(data)+len(defaults))
		for k, v := range data {
			result[k] = v
		}

		for k, v := range defaults {
			existing, exists := result[k]
			if !exists {
				result[k] = v
				continue
			}

			// Also fill in empty strings.
			if s, isStr := existing.(string); isStr && s == "" {
				result[k] = v
			}
		}

		return runtime.ToValue(result)
	}
}
