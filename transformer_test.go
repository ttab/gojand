package gojand

import (
	"context"
	"testing"

	"github.com/microcosm-cc/bluemonday"
	"github.com/ttab/newsdoc"
)

func TestTransformerBasic(t *testing.T) {
	script := `
function transform(doc) {
	doc.title = "Modified: " + doc.title;
	return doc;
}
`

	tr, err := NewTransformer(script)
	if err != nil {
		t.Fatalf("NewTransformer: %v", err)
	}

	doc := newsdoc.Document{
		Type:  "core/article",
		Title: "Original Title",
	}

	got, err := tr.Transform(context.Background(), doc)
	if err != nil {
		t.Fatalf("Transform: %v", err)
	}

	if got.Title != "Modified: Original Title" {
		t.Errorf("expected 'Modified: Original Title', got %q", got.Title)
	}

	if got.Type != "core/article" {
		t.Errorf("expected type 'core/article', got %q", got.Type)
	}
}

func TestTransformerDropBlocks(t *testing.T) {
	script := `
function transform(doc) {
	doc.meta = nd.drop_blocks(doc.meta, {type: "core/internal"});
	return doc;
}
`

	tr, err := NewTransformer(script)
	if err != nil {
		t.Fatalf("NewTransformer: %v", err)
	}

	doc := newsdoc.Document{
		Type:  "core/article",
		Title: "Test",
		Meta: []newsdoc.Block{
			{Type: "core/newsvalue", Value: "3"},
			{Type: "core/internal", Data: newsdoc.DataMap{"source": "wire"}},
			{Type: "core/section", Title: "Sports"},
		},
	}

	got, err := tr.Transform(context.Background(), doc)
	if err != nil {
		t.Fatalf("Transform: %v", err)
	}

	if len(got.Meta) != 2 {
		t.Fatalf("expected 2 meta blocks, got %d", len(got.Meta))
	}

	if got.Meta[0].Type != "core/newsvalue" {
		t.Errorf("expected first meta to be core/newsvalue, got %q", got.Meta[0].Type)
	}

	if got.Meta[1].Type != "core/section" {
		t.Errorf("expected second meta to be core/section, got %q", got.Meta[1].Type)
	}
}

func TestTransformerMultiCall(t *testing.T) {
	script := `
function transform(doc) {
	doc.title = "Transformed";
	return doc;
}
`

	tr, err := NewTransformer(script)
	if err != nil {
		t.Fatalf("NewTransformer: %v", err)
	}

	for i := range 5 {
		doc := newsdoc.Document{
			Type:  "core/article",
			Title: "Original",
		}

		got, err := tr.Transform(context.Background(), doc)
		if err != nil {
			t.Fatalf("Transform call %d: %v", i, err)
		}

		if got.Title != "Transformed" {
			t.Errorf("call %d: expected 'Transformed', got %q", i, got.Title)
		}
	}
}

func TestTransformerWithGlobals(t *testing.T) {
	script := `
function transform(doc) {
	doc.title = prefix + doc.title;
	return doc;
}
`

	tr, err := NewTransformer(script,
		WithGlobal("prefix", "BREAKING: "))
	if err != nil {
		t.Fatalf("NewTransformer: %v", err)
	}

	doc := newsdoc.Document{
		Type:  "core/article",
		Title: "News Event",
	}

	got, err := tr.Transform(context.Background(), doc)
	if err != nil {
		t.Fatalf("Transform: %v", err)
	}

	if got.Title != "BREAKING: News Event" {
		t.Errorf("expected 'BREAKING: News Event', got %q", got.Title)
	}
}

func TestTransformerCustomFuncName(t *testing.T) {
	script := `
function process(doc) {
	doc.title = "Processed";
	return doc;
}
`

	tr, err := NewTransformer(script, WithFuncName("process"))
	if err != nil {
		t.Fatalf("NewTransformer: %v", err)
	}

	doc := newsdoc.Document{Title: "Original"}

	got, err := tr.Transform(context.Background(), doc)
	if err != nil {
		t.Fatalf("Transform: %v", err)
	}

	if got.Title != "Processed" {
		t.Errorf("expected 'Processed', got %q", got.Title)
	}
}

func TestTransformerCompileError(t *testing.T) {
	_, err := NewTransformer("this is not valid javascript }{}{")
	if err == nil {
		t.Fatal("expected compile error")
	}
}

func TestTransformerMissingFunction(t *testing.T) {
	script := `
var x = 42;
`

	tr, err := NewTransformer(script)
	if err != nil {
		t.Fatalf("NewTransformer: %v", err)
	}

	doc := newsdoc.Document{Title: "Test"}

	_, err = tr.Transform(context.Background(), doc)
	if err == nil {
		t.Fatal("expected error for missing transform function")
	}
}

func TestTransformerWrongReturnType(t *testing.T) {
	script := `
function transform(doc) {
	return "not an object";
}
`

	tr, err := NewTransformer(script)
	if err != nil {
		t.Fatalf("NewTransformer: %v", err)
	}

	doc := newsdoc.Document{Title: "Test"}

	_, err = tr.Transform(context.Background(), doc)
	if err == nil {
		t.Fatal("expected error for wrong return type")
	}
}

func TestTransformerComplexScript(t *testing.T) {
	script := `
function transform(doc) {
	// Drop internal metadata
	doc.meta = nd.drop_blocks(doc.meta, {type: "core/internal"});

	// Ensure newsvalue exists
	var default_nv = {type: "core/newsvalue", value: "1"};
	doc.meta = nd.upsert_block(doc.meta, {type: "core/newsvalue"}, default_nv, function(b) { return b; });

	// Add prefix to title
	doc.title = "Published: " + doc.title;

	// Alter text blocks to add a role
	doc.content = nd.alter_blocks(doc.content, {type: "core/text"}, function(b) {
		b.role = "article-text";
		return b;
	});

	return doc;
}
`

	tr, err := NewTransformer(script)
	if err != nil {
		t.Fatalf("NewTransformer: %v", err)
	}

	doc := newsdoc.Document{
		Type:  "core/article",
		Title: "Breaking News",
		Content: []newsdoc.Block{
			{Type: "core/text", Rel: "main", Value: "paragraph 1"},
			{Type: "core/image", URI: "image://1"},
			{Type: "core/text", Rel: "sidebar", Value: "paragraph 2"},
		},
		Meta: []newsdoc.Block{
			{Type: "core/internal", Data: newsdoc.DataMap{"source": "wire"}},
			{Type: "core/newsvalue", Value: "3"},
		},
	}

	got, err := tr.Transform(context.Background(), doc)
	if err != nil {
		t.Fatalf("Transform: %v", err)
	}

	// Title should be prefixed.
	if got.Title != "Published: Breaking News" {
		t.Errorf("title: expected 'Published: Breaking News', got %q", got.Title)
	}

	// Internal meta should be removed.
	if len(got.Meta) != 1 {
		t.Fatalf("meta: expected 1 block, got %d", len(got.Meta))
	}

	if got.Meta[0].Type != "core/newsvalue" {
		t.Errorf("meta[0]: expected core/newsvalue, got %q", got.Meta[0].Type)
	}

	// Content text blocks should have role set.
	if len(got.Content) != 3 {
		t.Fatalf("content: expected 3 blocks, got %d", len(got.Content))
	}

	if got.Content[0].Role != "article-text" {
		t.Errorf("content[0].role: expected 'article-text', got %q", got.Content[0].Role)
	}

	if got.Content[1].Role != "" {
		t.Errorf("content[1].role: expected '' (image unchanged), got %q", got.Content[1].Role)
	}

	if got.Content[2].Role != "article-text" {
		t.Errorf("content[2].role: expected 'article-text', got %q", got.Content[2].Role)
	}
}

func TestTransformerFunctionMatcher(t *testing.T) {
	script := `
function transform(doc) {
	// Use function matcher to find blocks with data
	var with_data = nd.all_blocks(doc.meta, function(b) { return b.data != null; });
	doc.title = String(with_data.length) + " blocks with data";
	return doc;
}
`

	tr, err := NewTransformer(script)
	if err != nil {
		t.Fatalf("NewTransformer: %v", err)
	}

	doc := newsdoc.Document{
		Title: "Test",
		Meta: []newsdoc.Block{
			{Type: "core/newsvalue", Value: "3"},
			{Type: "core/internal", Data: newsdoc.DataMap{"source": "wire"}},
			{Type: "core/note", Data: newsdoc.DataMap{"text": "hi"}},
		},
	}

	got, err := tr.Transform(context.Background(), doc)
	if err != nil {
		t.Fatalf("Transform: %v", err)
	}

	if got.Title != "2 blocks with data" {
		t.Errorf("expected '2 blocks with data', got %q", got.Title)
	}
}

func TestTransformerHTMLStrip(t *testing.T) {
	script := `
function transform(doc) {
	doc.title = html.strip(doc.title);
	return doc;
}
`

	tr, err := NewTransformer(script)
	if err != nil {
		t.Fatalf("NewTransformer: %v", err)
	}

	doc := newsdoc.Document{
		Type:  "core/article",
		Title: "<b>Breaking</b> <script>xss</script>News",
	}

	got, err := tr.Transform(context.Background(), doc)
	if err != nil {
		t.Fatalf("Transform: %v", err)
	}

	if got.Title != "Breaking News" {
		t.Errorf("expected 'Breaking News', got %q", got.Title)
	}
}

func TestTransformerHTMLStripPolicy(t *testing.T) {
	script := `
function transform(doc) {
	doc.title = html.strip_policy(doc.title, "ugc");
	return doc;
}
`

	ugcPolicy := bluemonday.UGCPolicy()

	tr, err := NewTransformer(script, WithPolicy("ugc", ugcPolicy))
	if err != nil {
		t.Fatalf("NewTransformer: %v", err)
	}

	doc := newsdoc.Document{
		Type:  "core/article",
		Title: "<b>Bold</b> <script>xss</script>text",
	}

	got, err := tr.Transform(context.Background(), doc)
	if err != nil {
		t.Fatalf("Transform: %v", err)
	}

	// UGC policy keeps <b> but strips <script>.
	if got.Title != "<b>Bold</b> text" {
		t.Errorf("expected '<b>Bold</b> text', got %q", got.Title)
	}
}

func TestTransformerTypeScript(t *testing.T) {
	script := `
function transform(doc: any): any {
	doc.title = "TS: " + doc.title;
	return doc;
}
`

	tr, err := NewTransformer(script, WithTypeScript())
	if err != nil {
		t.Fatalf("NewTransformer: %v", err)
	}

	doc := newsdoc.Document{
		Type:  "core/article",
		Title: "Hello",
	}

	got, err := tr.Transform(context.Background(), doc)
	if err != nil {
		t.Fatalf("Transform: %v", err)
	}

	if got.Title != "TS: Hello" {
		t.Errorf("expected 'TS: Hello', got %q", got.Title)
	}
}
