package gojand

import (
	"testing"

	"github.com/ttab/newsdoc"
)

func TestDocumentRoundTrip(t *testing.T) {
	doc := newsdoc.Document{
		UUID:     "550e8400-e29b-41d4-a716-446655440000",
		Type:     "core/article",
		URI:      "article://test/1",
		URL:      "https://example.com/article/1",
		Title:    "Test Article",
		Language: "en",
		Content: []newsdoc.Block{
			{
				Type: "core/text",
				Rel:  "main",
				Role: "body",
				Content: []newsdoc.Block{
					{
						Type:  "core/text",
						Value: "Hello, world!",
					},
				},
			},
			{
				Type:  "core/image",
				URI:   "image://test/1",
				Title: "Test Image",
				Data: newsdoc.DataMap{
					"width":  "800",
					"height": "600",
				},
			},
		},
		Meta: []newsdoc.Block{
			{
				Type:  "core/newsvalue",
				Value: "3",
				Data: newsdoc.DataMap{
					"score": "42",
				},
			},
			{
				Type: "core/internal",
				Data: newsdoc.DataMap{
					"source": "wire",
				},
			},
		},
		Links: []newsdoc.Block{
			{
				Type:  "core/section",
				Title: "Sports",
				UUID:  "660e8400-e29b-41d4-a716-446655440000",
			},
		},
	}

	m := DocumentToMap(doc)

	got, err := MapToDocument(m)
	if err != nil {
		t.Fatalf("MapToDocument: %v", err)
	}

	assertDocumentsEqual(t, doc, got)
}

func TestDocumentEmptyFields(t *testing.T) {
	doc := newsdoc.Document{
		Type:  "core/article",
		Title: "Minimal",
	}

	m := DocumentToMap(doc)

	// Verify that empty fields are omitted.
	if _, ok := m["uuid"]; ok {
		t.Error("expected uuid to be omitted")
	}

	if _, ok := m["content"]; ok {
		t.Error("expected content to be omitted")
	}

	got, err := MapToDocument(m)
	if err != nil {
		t.Fatalf("MapToDocument: %v", err)
	}

	assertDocumentsEqual(t, doc, got)
}

func TestBlockRoundTrip(t *testing.T) {
	block := newsdoc.Block{
		ID:          "block-1",
		UUID:        "770e8400-e29b-41d4-a716-446655440000",
		URI:         "block://test/1",
		URL:         "https://example.com/block/1",
		Type:        "core/text",
		Title:       "A Block",
		Rel:         "main",
		Role:        "body",
		Name:        "intro",
		Value:       "some value",
		Contenttype: "text/html",
		Sensitivity: "internal",
		Data: newsdoc.DataMap{
			"key1": "val1",
			"key2": "val2",
		},
		Links: []newsdoc.Block{
			{Type: "core/ref", UUID: "aaa"},
		},
		Content: []newsdoc.Block{
			{Type: "core/text", Value: "child"},
		},
		Meta: []newsdoc.Block{
			{Type: "core/note", Value: "note"},
		},
	}

	m := BlockToMap(block)

	got, err := MapToBlock(m)
	if err != nil {
		t.Fatalf("MapToBlock: %v", err)
	}

	assertBlocksEqual(t, "root", block, got)
}

func TestBlockEmptyFields(t *testing.T) {
	block := newsdoc.Block{
		Type: "core/text",
	}

	m := BlockToMap(block)

	// Only "type" should be set.
	if len(m) != 1 {
		t.Errorf("expected 1 key, got %d", len(m))
	}

	got, err := MapToBlock(m)
	if err != nil {
		t.Fatalf("MapToBlock: %v", err)
	}

	assertBlocksEqual(t, "root", block, got)
}

func TestDataMapRoundTrip(t *testing.T) {
	dm := newsdoc.DataMap{
		"width":  "800",
		"height": "600",
		"format": "jpeg",
	}

	m := dataMapToPlain(dm)

	got, err := plainToDataMap(m)
	if err != nil {
		t.Fatalf("plainToDataMap: %v", err)
	}

	if len(got) != len(dm) {
		t.Fatalf("expected %d entries, got %d", len(dm), len(got))
	}

	for k, v := range dm {
		if got[k] != v {
			t.Errorf("key %q: expected %q, got %q", k, v, got[k])
		}
	}
}

func TestDataMapInvalidValue(t *testing.T) {
	m := map[string]any{
		"ok":  "fine",
		"bad": 42,
	}

	_, err := plainToDataMap(m)
	if err == nil {
		t.Fatal("expected error for non-string value")
	}
}

func TestToBlocksInvalidType(t *testing.T) {
	_, err := toBlocks("not a slice")
	if err == nil {
		t.Fatal("expected error for non-slice")
	}
}

func TestToBlocksInvalidElement(t *testing.T) {
	_, err := toBlocks([]any{"not a map"})
	if err == nil {
		t.Fatal("expected error for non-map element")
	}
}

func assertDocumentsEqual(t *testing.T, want, got newsdoc.Document) {
	t.Helper()

	if want.UUID != got.UUID {
		t.Errorf("UUID: want %q, got %q", want.UUID, got.UUID)
	}

	if want.Type != got.Type {
		t.Errorf("Type: want %q, got %q", want.Type, got.Type)
	}

	if want.URI != got.URI {
		t.Errorf("URI: want %q, got %q", want.URI, got.URI)
	}

	if want.URL != got.URL {
		t.Errorf("URL: want %q, got %q", want.URL, got.URL)
	}

	if want.Title != got.Title {
		t.Errorf("Title: want %q, got %q", want.Title, got.Title)
	}

	if want.Language != got.Language {
		t.Errorf("Language: want %q, got %q", want.Language, got.Language)
	}

	assertBlockSlicesEqual(t, "Content", want.Content, got.Content)
	assertBlockSlicesEqual(t, "Meta", want.Meta, got.Meta)
	assertBlockSlicesEqual(t, "Links", want.Links, got.Links)
}

func assertBlockSlicesEqual(t *testing.T, name string, want, got []newsdoc.Block) {
	t.Helper()

	if len(want) != len(got) {
		t.Fatalf("%s: want %d blocks, got %d", name, len(want), len(got))
	}

	for i := range want {
		assertBlocksEqual(t, name, want[i], got[i])
	}
}

func assertBlocksEqual(t *testing.T, ctx string, want, got newsdoc.Block) {
	t.Helper()

	check := func(field, w, g string) {
		if w != g {
			t.Errorf("%s.%s: want %q, got %q", ctx, field, w, g)
		}
	}

	check("ID", want.ID, got.ID)
	check("UUID", want.UUID, got.UUID)
	check("URI", want.URI, got.URI)
	check("URL", want.URL, got.URL)
	check("Type", want.Type, got.Type)
	check("Title", want.Title, got.Title)
	check("Rel", want.Rel, got.Rel)
	check("Role", want.Role, got.Role)
	check("Name", want.Name, got.Name)
	check("Value", want.Value, got.Value)
	check("Contenttype", want.Contenttype, got.Contenttype)
	check("Sensitivity", want.Sensitivity, got.Sensitivity)

	if len(want.Data) != len(got.Data) {
		t.Errorf("%s.Data: want %d entries, got %d", ctx, len(want.Data), len(got.Data))
	} else {
		for k, v := range want.Data {
			if got.Data[k] != v {
				t.Errorf("%s.Data[%q]: want %q, got %q", ctx, k, v, got.Data[k])
			}
		}
	}

	assertBlockSlicesEqual(t, ctx+".Links", want.Links, got.Links)
	assertBlockSlicesEqual(t, ctx+".Content", want.Content, got.Content)
	assertBlockSlicesEqual(t, ctx+".Meta", want.Meta, got.Meta)
}
