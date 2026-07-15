package gojand_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ttab/gojand"
)

const sessionScript = `
function transform(doc) {
	doc.title = "transformed";
	return doc;
}

let calls = 0;

function renditions(block) {
	calls++;

	if (block.type !== "core/image") {
		return null;
	}

	return [
		{ns: "mm", id: block.id, version: "0", variant: "preview", ext: "jpg", calls: calls},
	];
}

function boom() {
	throw new Error("boom");
}

function spin() {
	while (true) {}
}
`

func newSession(t *testing.T) *gojand.Session {
	t.Helper()

	tr, err := gojand.NewTransformer(sessionScript)
	if err != nil {
		t.Fatalf("NewTransformer: %v", err)
	}

	session, err := tr.NewSession(t.Context())
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	t.Cleanup(session.Close)

	return session
}

func TestSessionRepeatedCalls(t *testing.T) {
	session := newSession(t)

	blocks := []map[string]any{
		{"type": "core/image", "id": "img-1"},
		{"type": "core/text", "id": "txt-1"},
		{"type": "core/image", "id": "img-2"},
	}

	var results []any

	for _, block := range blocks {
		res, err := session.Call("renditions", block)
		if err != nil {
			t.Fatalf("call renditions: %v", err)
		}

		results = append(results, res)
	}

	if results[1] != nil {
		t.Errorf("expected null for non-image block, got %v", results[1])
	}

	first, ok := results[0].([]any)
	if !ok || len(first) != 1 {
		t.Fatalf("expected one rendition, got %v", results[0])
	}

	rend, ok := first[0].(map[string]any)
	if !ok || rend["id"] != "img-1" || rend["variant"] != "preview" {
		t.Errorf("unexpected rendition: %v", first[0])
	}

	// Script state persists across calls within one session: the third
	// call sees calls == 3.
	last, ok := results[2].([]any)
	if !ok || len(last) != 1 {
		t.Fatalf("expected one rendition, got %v", results[2])
	}

	lastRend, ok := last[0].(map[string]any)
	if !ok {
		t.Fatalf("expected a rendition object, got %T", last[0])
	}

	if lastRend["calls"] != int64(3) {
		t.Errorf("expected 3 calls recorded, got %v", lastRend["calls"])
	}
}

func TestSessionHasFunction(t *testing.T) {
	session := newSession(t)

	if !session.HasFunction("transform") {
		t.Error("expected transform to be defined")
	}

	if !session.HasFunction("renditions") {
		t.Error("expected renditions to be defined")
	}

	if session.HasFunction("nosuch") {
		t.Error("did not expect nosuch to be defined")
	}

	if session.HasFunction("calls") {
		t.Error("a non-function value should not count as a function")
	}
}

func TestSessionCallErrors(t *testing.T) {
	session := newSession(t)

	_, err := session.Call("nosuch", nil)
	if err == nil || !strings.Contains(err.Error(), "not defined") {
		t.Errorf("expected a not-defined error, got %v", err)
	}

	_, err = session.Call("boom", nil)
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Errorf("expected the thrown error, got %v", err)
	}

	// The session stays usable after a thrown error.
	_, err = session.Call("renditions", map[string]any{"type": "core/text"})
	if err != nil {
		t.Errorf("expected the session to survive a script error: %v", err)
	}
}

func TestSessionCancellation(t *testing.T) {
	tr, err := gojand.NewTransformer(sessionScript)
	if err != nil {
		t.Fatalf("NewTransformer: %v", err)
	}

	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()

	session, err := tr.NewSession(ctx)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	defer session.Close()

	_, err = session.Call("spin", nil)
	if err == nil {
		t.Fatal("expected an interrupt error")
	}
}

func TestSessionCloseIdempotent(t *testing.T) {
	session := newSession(t)

	session.Close()
	session.Close()
}
