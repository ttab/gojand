package gojand

import _ "embed"

//go:embed gojand.d.ts
var typeDefinitions string

// TypeDefinitions returns the TypeScript type definitions for the gojand
// scripting environment. Host applications can write this to disk so that
// script authors get editor support when writing transformation scripts.
func TypeDefinitions() string {
	return typeDefinitions
}
