# gojand

Go module for configurable NewsDoc document transformation using JavaScript scripts.

Uses [goja](https://github.com/dop251/goja) (pure-Go ES5.1+ engine) with optional [TypeScript](https://www.typescriptlang.org/) support via [esbuild](https://esbuild.github.io/). Scripts are compiled once and executed per-document with fresh runtime state.

## Usage

```go
script := `
function transform(doc) {
    // Remove internal metadata
    doc.meta = nd.drop_blocks(doc.meta, {type: "core/internal"});

    // Ensure a newsvalue block exists
    var default_nv = {type: "core/newsvalue", value: "1"};
    doc.meta = nd.upsert_block(doc.meta, {type: "core/newsvalue"}, default_nv, function(b) { return b; });

    // Prefix the title
    doc.title = "Published: " + doc.title;

    return doc;
}
`

tr, err := gojand.NewTransformer(script)
if err != nil {
    log.Fatal(err)
}

result, err := tr.Transform(ctx, doc)
```

### Options

```go
gojand.NewTransformer(script,
    gojand.WithGlobal("prefix", "BREAKING: "),
    gojand.WithFuncName("process"),        // default: "transform"
    gojand.WithTypeScript(),               // enable TS→JS transpilation
    gojand.WithPolicy("ugc", ugcPolicy),   // register bluemonday policy
)
```

## The `nd` Module

All functions operate on plain JS objects and arrays directly.

### Block Querying

| Function | Description |
|---|---|
| `nd.first_block(blocks, matcher)` | First matching block or null |
| `nd.all_blocks(blocks, matcher)` | Array of all matching blocks |
| `nd.has_block(blocks, matcher)` | Boolean: any block matches? |

### Block Manipulation

| Function | Description |
|---|---|
| `nd.drop_blocks(blocks, matcher)` | New array without matches |
| `nd.dedupe_blocks(blocks, matcher)` | Keep only first match |
| `nd.alter_blocks(blocks, matcher, fn)` | Apply fn to all matches |
| `nd.alter_first_block(blocks, matcher, fn)` | Apply fn to first match only |
| `nd.upsert_block(blocks, matcher, default, fn)` | Apply fn to match, or append default |
| `nd.add_or_replace_block(blocks, matcher, block)` | Replace first match or append |

### DataMap Helpers

| Function | Description |
|---|---|
| `nd.get_data(block, key [, default])` | Get value from block's data map |
| `nd.upsert_data(data, new_data)` | Merge objects (returns copy) |
| `nd.data_defaults(data, defaults)` | Fill missing/empty keys from defaults |

### Matchers

Matchers can be either:

- **Criteria object**: `{type: "core/text", rel: "main"}` — all fields must match
- **Function**: `function(b) { return b.type === "core/text"; }` — arbitrary predicate

## The `html` Module

| Function | Description |
|---|---|
| `html.encode(s)` | HTML entity encoding |
| `html.decode(s)` | HTML entity decoding |
| `html.strip(s)` | Strip all HTML tags |
| `html.strip_policy(s, name)` | Strip using a named bluemonday policy |

## Document Representation

Documents and blocks are converted to plain JS objects at the boundary. Field names match JSON tags (lowercase). Empty fields are omitted.

```
// Document object
{uuid: "...", type: "core/article", title: "...", content: [...], meta: [...], links: [...]}

// Block object
{type: "core/text", rel: "main", value: "...", data: {key: "val"}, content: [...]}
```

## TypeScript Support

Enable TypeScript with `WithTypeScript()`. The script is transpiled via esbuild before goja compilation:

```go
script := `
function transform(doc: any): any {
    doc.title = "TS: " + doc.title;
    return doc;
}
`

tr, err := gojand.NewTransformer(script, gojand.WithTypeScript())
```

## JavaScript Syntax Notes

Scripts run in an ES5.1+ environment (goja). Key syntax:

- Named functions: `function name(args) { ... }`
- Return values are explicit: `return doc;`
- Null check: `b.data != null` (not `!= nil`)
- Array length: `arr.length` (not `len(arr)`)
- String conversion: `String(n)` (not `string(n)`)
- Strict equality: `===` / `!==`
