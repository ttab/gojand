package gojand

import (
	"fmt"

	"github.com/ttab/newsdoc"
)

// DocumentToMap converts a newsdoc.Document to a plain map.
func DocumentToMap(doc newsdoc.Document) map[string]any {
	m := map[string]any{}

	setIfNonEmpty(m, "uuid", doc.UUID)
	setIfNonEmpty(m, "type", doc.Type)
	setIfNonEmpty(m, "uri", doc.URI)
	setIfNonEmpty(m, "url", doc.URL)
	setIfNonEmpty(m, "title", doc.Title)
	setIfNonEmpty(m, "language", doc.Language)

	if len(doc.Content) > 0 {
		m["content"] = blocksToSlice(doc.Content)
	}

	if len(doc.Meta) > 0 {
		m["meta"] = blocksToSlice(doc.Meta)
	}

	if len(doc.Links) > 0 {
		m["links"] = blocksToSlice(doc.Links)
	}

	return m
}

// MapToDocument converts a plain map back to a newsdoc.Document.
func MapToDocument(m map[string]any) (newsdoc.Document, error) {
	var doc newsdoc.Document

	doc.UUID = getString(m, "uuid")
	doc.Type = getString(m, "type")
	doc.URI = getString(m, "uri")
	doc.URL = getString(m, "url")
	doc.Title = getString(m, "title")
	doc.Language = getString(m, "language")

	var err error

	if v, ok := m["content"]; ok && v != nil {
		doc.Content, err = toBlocks(v)
		if err != nil {
			return doc, fmt.Errorf("convert document content: %w", err)
		}
	}

	if v, ok := m["meta"]; ok && v != nil {
		doc.Meta, err = toBlocks(v)
		if err != nil {
			return doc, fmt.Errorf("convert document meta: %w", err)
		}
	}

	if v, ok := m["links"]; ok && v != nil {
		doc.Links, err = toBlocks(v)
		if err != nil {
			return doc, fmt.Errorf("convert document links: %w", err)
		}
	}

	return doc, nil
}

// BlockToMap converts a newsdoc.Block to a plain map.
func BlockToMap(block newsdoc.Block) map[string]any {
	m := map[string]any{}

	setIfNonEmpty(m, "id", block.ID)
	setIfNonEmpty(m, "uuid", block.UUID)
	setIfNonEmpty(m, "uri", block.URI)
	setIfNonEmpty(m, "url", block.URL)
	setIfNonEmpty(m, "type", block.Type)
	setIfNonEmpty(m, "title", block.Title)
	setIfNonEmpty(m, "rel", block.Rel)
	setIfNonEmpty(m, "role", block.Role)
	setIfNonEmpty(m, "name", block.Name)
	setIfNonEmpty(m, "value", block.Value)
	setIfNonEmpty(m, "contenttype", block.Contenttype)
	setIfNonEmpty(m, "sensitivity", block.Sensitivity)

	if len(block.Data) > 0 {
		m["data"] = dataMapToPlain(block.Data)
	}

	if len(block.Links) > 0 {
		m["links"] = blocksToSlice(block.Links)
	}

	if len(block.Content) > 0 {
		m["content"] = blocksToSlice(block.Content)
	}

	if len(block.Meta) > 0 {
		m["meta"] = blocksToSlice(block.Meta)
	}

	return m
}

// MapToBlock converts a plain map back to a newsdoc.Block.
func MapToBlock(m map[string]any) (newsdoc.Block, error) {
	var block newsdoc.Block

	block.ID = getString(m, "id")
	block.UUID = getString(m, "uuid")
	block.URI = getString(m, "uri")
	block.URL = getString(m, "url")
	block.Type = getString(m, "type")
	block.Title = getString(m, "title")
	block.Rel = getString(m, "rel")
	block.Role = getString(m, "role")
	block.Name = getString(m, "name")
	block.Value = getString(m, "value")
	block.Contenttype = getString(m, "contenttype")
	block.Sensitivity = getString(m, "sensitivity")

	var err error

	if v, ok := m["data"]; ok && v != nil {
		block.Data, err = plainToDataMap(v)
		if err != nil {
			return block, fmt.Errorf("convert block data: %w", err)
		}
	}

	if v, ok := m["links"]; ok && v != nil {
		block.Links, err = toBlocks(v)
		if err != nil {
			return block, fmt.Errorf("convert block links: %w", err)
		}
	}

	if v, ok := m["content"]; ok && v != nil {
		block.Content, err = toBlocks(v)
		if err != nil {
			return block, fmt.Errorf("convert block content: %w", err)
		}
	}

	if v, ok := m["meta"]; ok && v != nil {
		block.Meta, err = toBlocks(v)
		if err != nil {
			return block, fmt.Errorf("convert block meta: %w", err)
		}
	}

	return block, nil
}

func blocksToSlice(blocks []newsdoc.Block) []any {
	items := make([]any, len(blocks))
	for i, b := range blocks {
		items[i] = BlockToMap(b)
	}

	return items
}

func toBlocks(v any) ([]newsdoc.Block, error) {
	slice, ok := toSlice(v)
	if !ok {
		return nil, fmt.Errorf("expected array, got %T", v)
	}

	blocks := make([]newsdoc.Block, len(slice))

	for i, item := range slice {
		m, ok := toMap(item)
		if !ok {
			return nil, fmt.Errorf(
				"expected object at index %d, got %T", i, item)
		}

		block, err := MapToBlock(m)
		if err != nil {
			return nil, fmt.Errorf("convert block at index %d: %w", i, err)
		}

		blocks[i] = block
	}

	return blocks, nil
}

func dataMapToPlain(dm newsdoc.DataMap) map[string]any {
	m := make(map[string]any, len(dm))
	for k, v := range dm {
		m[k] = v
	}

	return m
}

func plainToDataMap(v any) (newsdoc.DataMap, error) {
	m, ok := toMap(v)
	if !ok {
		return nil, fmt.Errorf("expected object, got %T", v)
	}

	dm := make(newsdoc.DataMap, len(m))

	for k, val := range m {
		s, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf(
				"expected string value for key %q, got %T", k, val)
		}

		dm[k] = s
	}

	return dm, nil
}

func getString(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}

	s, ok := v.(string)
	if !ok {
		return ""
	}

	return s
}

func setIfNonEmpty(m map[string]any, key, value string) {
	if value != "" {
		m[key] = value
	}
}

// toSlice coerces a goja export to []any. Goja may return []any or
// []interface{}, so we handle both.
func toSlice(v any) ([]any, bool) {
	switch s := v.(type) {
	case []any:
		return s, true
	case []map[string]any:
		out := make([]any, len(s))
		for i, item := range s {
			out[i] = item
		}

		return out, true
	default:
		return nil, false
	}
}

// toMap coerces a goja export to map[string]any.
func toMap(v any) (map[string]any, bool) {
	m, ok := v.(map[string]any)
	return m, ok
}
