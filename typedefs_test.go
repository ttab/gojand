package gojand_test

import (
	"testing"

	"github.com/ttab/gojand"
)

func TestTypeDefinitions(t *testing.T) {
	td := gojand.TypeDefinitions()
	if td == "" {
		t.Fatal("TypeDefinitions() returned an empty string")
	}

	if len(td) < 100 {
		t.Fatalf("TypeDefinitions() suspiciously short: %d bytes", len(td))
	}
}
