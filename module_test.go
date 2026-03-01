package gojand

import (
	"testing"

	"github.com/dop251/goja"
)

func testRuntime() *goja.Runtime {
	runtime := goja.New()

	err := runtime.Set("nd", newNDModule(runtime))
	if err != nil {
		panic(err)
	}

	err = runtime.Set("html", newHTMLModule(runtime, nil))
	if err != nil {
		panic(err)
	}

	return runtime
}

func evalJS(t *testing.T, source string) goja.Value {
	t.Helper()

	runtime := testRuntime()

	v, err := runtime.RunString(source)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}

	return v
}

func TestFirstBlock(t *testing.T) {
	result := evalJS(t, `
var blocks = [
	{type: "core/text", rel: "main"},
	{type: "core/image", rel: "photo"},
	{type: "core/text", rel: "sidebar"},
];
nd.first_block(blocks, {type: "core/text"});
`)

	m, ok := toMap(result.Export())
	if !ok {
		t.Fatalf("expected object, got %T", result.Export())
	}

	if getString(m, "rel") != "main" {
		t.Errorf("expected rel=main, got %q", getString(m, "rel"))
	}
}

func TestFirstBlockNotFound(t *testing.T) {
	result := evalJS(t, `
var blocks = [{type: "core/text"}];
nd.first_block(blocks, {type: "core/image"});
`)

	if !goja.IsNull(result) {
		t.Errorf("expected null, got %v", result.Export())
	}
}

func TestFirstBlockWithFunction(t *testing.T) {
	result := evalJS(t, `
var blocks = [
	{type: "core/text", rel: "main"},
	{type: "core/image", rel: "photo"},
];
nd.first_block(blocks, function(b) { return b.type === "core/image"; });
`)

	m, ok := toMap(result.Export())
	if !ok {
		t.Fatalf("expected object, got %T", result.Export())
	}

	if getString(m, "rel") != "photo" {
		t.Errorf("expected rel=photo, got %q", getString(m, "rel"))
	}
}

func TestAllBlocks(t *testing.T) {
	result := evalJS(t, `
var blocks = [
	{type: "core/text", rel: "main"},
	{type: "core/image", rel: "photo"},
	{type: "core/text", rel: "sidebar"},
];
nd.all_blocks(blocks, {type: "core/text"});
`)

	slice, ok := toSlice(result.Export())
	if !ok {
		t.Fatalf("expected array, got %T", result.Export())
	}

	if len(slice) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(slice))
	}
}

func TestHasBlock(t *testing.T) {
	result := evalJS(t, `
var blocks = [
	{type: "core/text"},
	{type: "core/image"},
];
nd.has_block(blocks, {type: "core/image"});
`)

	if result.Export() != true {
		t.Errorf("expected true, got %v", result.Export())
	}

	result = evalJS(t, `
var blocks = [{type: "core/text"}];
nd.has_block(blocks, {type: "core/image"});
`)

	if result.Export() != false {
		t.Errorf("expected false, got %v", result.Export())
	}
}

func TestDropBlocks(t *testing.T) {
	result := evalJS(t, `
var blocks = [
	{type: "core/text", rel: "main"},
	{type: "core/internal"},
	{type: "core/text", rel: "sidebar"},
	{type: "core/internal"},
];
nd.drop_blocks(blocks, {type: "core/internal"});
`)

	slice, ok := toSlice(result.Export())
	if !ok {
		t.Fatalf("expected array, got %T", result.Export())
	}

	if len(slice) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(slice))
	}
}

func TestDedupeBlocks(t *testing.T) {
	result := evalJS(t, `
var blocks = [
	{type: "core/section", title: "Sports"},
	{type: "core/section", title: "Politics"},
	{type: "core/text", rel: "main"},
];
nd.dedupe_blocks(blocks, {type: "core/section"});
`)

	slice, ok := toSlice(result.Export())
	if !ok {
		t.Fatalf("expected array, got %T", result.Export())
	}

	if len(slice) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(slice))
	}

	first, ok := toMap(slice[0])
	if !ok {
		t.Fatalf("expected object at 0, got %T", slice[0])
	}

	if getString(first, "title") != "Sports" {
		t.Errorf("expected first to be Sports, got %q", getString(first, "title"))
	}
}

func TestAlterBlocks(t *testing.T) {
	result := evalJS(t, `
var blocks = [
	{type: "core/text", title: "Hello"},
	{type: "core/image", title: "Photo"},
	{type: "core/text", title: "World"},
];
nd.alter_blocks(blocks, {type: "core/text"}, function(b) {
	return {type: b.type, title: "Modified: " + b.title};
});
`)

	slice, ok := toSlice(result.Export())
	if !ok {
		t.Fatalf("expected array, got %T", result.Export())
	}

	if len(slice) != 3 {
		t.Fatalf("expected 3 blocks, got %d", len(slice))
	}

	first, _ := toMap(slice[0])
	if getString(first, "title") != "Modified: Hello" {
		t.Errorf("expected 'Modified: Hello', got %q", getString(first, "title"))
	}

	second, _ := toMap(slice[1])
	if getString(second, "title") != "Photo" {
		t.Errorf("expected 'Photo' (unchanged), got %q", getString(second, "title"))
	}

	third, _ := toMap(slice[2])
	if getString(third, "title") != "Modified: World" {
		t.Errorf("expected 'Modified: World', got %q", getString(third, "title"))
	}
}

func TestAlterFirstBlock(t *testing.T) {
	result := evalJS(t, `
var blocks = [
	{type: "core/text", title: "First"},
	{type: "core/text", title: "Second"},
];
nd.alter_first_block(blocks, {type: "core/text"}, function(b) {
	return {type: b.type, title: "Altered"};
});
`)

	slice, ok := toSlice(result.Export())
	if !ok {
		t.Fatalf("expected array, got %T", result.Export())
	}

	first, _ := toMap(slice[0])
	if getString(first, "title") != "Altered" {
		t.Errorf("expected 'Altered', got %q", getString(first, "title"))
	}

	second, _ := toMap(slice[1])
	if getString(second, "title") != "Second" {
		t.Errorf("expected 'Second' (unchanged), got %q", getString(second, "title"))
	}
}

func TestUpsertBlockExisting(t *testing.T) {
	result := evalJS(t, `
var blocks = [
	{type: "core/newsvalue", value: "3"},
	{type: "core/text"},
];
var default_block = {type: "core/newsvalue", value: "1"};
nd.upsert_block(blocks, {type: "core/newsvalue"}, default_block, function(b) {
	return {type: b.type, value: "5"};
});
`)

	slice, ok := toSlice(result.Export())
	if !ok {
		t.Fatalf("expected array, got %T", result.Export())
	}

	if len(slice) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(slice))
	}

	first, _ := toMap(slice[0])
	if getString(first, "value") != "5" {
		t.Errorf("expected value=5, got %q", getString(first, "value"))
	}
}

func TestUpsertBlockNew(t *testing.T) {
	result := evalJS(t, `
var blocks = [{type: "core/text"}];
var default_block = {type: "core/newsvalue", value: "1"};
nd.upsert_block(blocks, {type: "core/newsvalue"}, default_block, function(b) { return b; });
`)

	slice, ok := toSlice(result.Export())
	if !ok {
		t.Fatalf("expected array, got %T", result.Export())
	}

	if len(slice) != 2 {
		t.Fatalf("expected 2 blocks (original + default), got %d", len(slice))
	}

	last, _ := toMap(slice[1])
	if getString(last, "type") != "core/newsvalue" {
		t.Errorf("expected type=core/newsvalue, got %q", getString(last, "type"))
	}
}

func TestAddOrReplaceBlockReplace(t *testing.T) {
	result := evalJS(t, `
var blocks = [
	{type: "core/newsvalue", value: "3"},
	{type: "core/text"},
];
nd.add_or_replace_block(blocks, {type: "core/newsvalue"}, {type: "core/newsvalue", value: "5"});
`)

	slice, ok := toSlice(result.Export())
	if !ok {
		t.Fatalf("expected array, got %T", result.Export())
	}

	if len(slice) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(slice))
	}

	first, _ := toMap(slice[0])
	if getString(first, "value") != "5" {
		t.Errorf("expected value=5, got %q", getString(first, "value"))
	}
}

func TestAddOrReplaceBlockAdd(t *testing.T) {
	result := evalJS(t, `
var blocks = [{type: "core/text"}];
nd.add_or_replace_block(blocks, {type: "core/newsvalue"}, {type: "core/newsvalue", value: "1"});
`)

	slice, ok := toSlice(result.Export())
	if !ok {
		t.Fatalf("expected array, got %T", result.Export())
	}

	if len(slice) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(slice))
	}
}

func TestGetData(t *testing.T) {
	result := evalJS(t, `
var block = {type: "core/image", data: {width: "800", height: "600"}};
nd.get_data(block, "width");
`)

	if result.Export() != "800" {
		t.Errorf("expected '800', got %v", result.Export())
	}
}

func TestGetDataMissing(t *testing.T) {
	result := evalJS(t, `
var block = {type: "core/image", data: {width: "800"}};
nd.get_data(block, "format");
`)

	if !goja.IsNull(result) {
		t.Errorf("expected null, got %v", result.Export())
	}
}

func TestGetDataDefault(t *testing.T) {
	result := evalJS(t, `
var block = {type: "core/image", data: {width: "800"}};
nd.get_data(block, "format", "jpeg");
`)

	if result.Export() != "jpeg" {
		t.Errorf("expected 'jpeg', got %v", result.Export())
	}
}

func TestGetDataNoDataMap(t *testing.T) {
	result := evalJS(t, `
var block = {type: "core/text"};
nd.get_data(block, "key", "default");
`)

	if result.Export() != "default" {
		t.Errorf("expected 'default', got %v", result.Export())
	}
}

func TestUpsertData(t *testing.T) {
	result := evalJS(t, `
var data = {width: "800", height: "600"};
nd.upsert_data(data, {height: "400", format: "png"});
`)

	m, ok := toMap(result.Export())
	if !ok {
		t.Fatalf("expected object, got %T", result.Export())
	}

	if getString(m, "width") != "800" {
		t.Errorf("width: expected '800', got %q", getString(m, "width"))
	}

	if getString(m, "height") != "400" {
		t.Errorf("height: expected '400', got %q", getString(m, "height"))
	}

	if getString(m, "format") != "png" {
		t.Errorf("format: expected 'png', got %q", getString(m, "format"))
	}
}

func TestDataDefaults(t *testing.T) {
	result := evalJS(t, `
var data = {width: "800", format: ""};
nd.data_defaults(data, {width: "100", height: "100", format: "jpeg"});
`)

	m, ok := toMap(result.Export())
	if !ok {
		t.Fatalf("expected object, got %T", result.Export())
	}

	if getString(m, "width") != "800" {
		t.Errorf("width: expected '800' (kept), got %q", getString(m, "width"))
	}

	if getString(m, "height") != "100" {
		t.Errorf("height: expected '100' (default), got %q", getString(m, "height"))
	}

	if getString(m, "format") != "jpeg" {
		t.Errorf("format: expected 'jpeg' (default for empty), got %q", getString(m, "format"))
	}
}

func TestUndefinedBlockSlice(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{
			name: "first_block with undefined",
			code: `nd.first_block(undefined, {type: "core/text"})`,
		},
		{
			name: "all_blocks with undefined",
			code: `nd.all_blocks(undefined, {type: "core/text"})`,
		},
		{
			name: "has_block with undefined",
			code: `nd.has_block(undefined, {type: "core/text"})`,
		},
		{
			name: "drop_blocks with undefined",
			code: `nd.drop_blocks(undefined, {type: "core/text"})`,
		},
		{
			name: "dedupe_blocks with undefined",
			code: `nd.dedupe_blocks(undefined, {type: "core/text"})`,
		},
		{
			name: "upsert_block with undefined appends default",
			code: `
var result = nd.upsert_block(undefined, {type: "core/newsvalue"},
    {type: "core/newsvalue", value: "1"},
    function(b) { return b; });
if (result.length !== 1) throw new Error("expected 1 block, got " + result.length);
result;
`,
		},
		{
			name: "add_or_replace_block with undefined appends",
			code: `
var result = nd.add_or_replace_block(undefined, {type: "core/newsvalue"},
    {type: "core/newsvalue", value: "1"});
if (result.length !== 1) throw new Error("expected 1 block, got " + result.length);
result;
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evalJS(t, tt.code)
		})
	}
}

func TestMultiFieldCriteria(t *testing.T) {
	result := evalJS(t, `
var blocks = [
	{type: "core/text", rel: "main"},
	{type: "core/text", rel: "sidebar"},
	{type: "core/image", rel: "main"},
];
nd.first_block(blocks, {type: "core/text", rel: "sidebar"});
`)

	m, ok := toMap(result.Export())
	if !ok {
		t.Fatalf("expected object, got %T", result.Export())
	}

	if getString(m, "rel") != "sidebar" {
		t.Errorf("expected rel=sidebar, got %q", getString(m, "rel"))
	}
}
