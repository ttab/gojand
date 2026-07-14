package gojand

import (
	"testing"

	"github.com/ttab/newsdoc"
)

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
