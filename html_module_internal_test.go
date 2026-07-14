package gojand

import (
	"testing"

	"github.com/dop251/goja"
	"github.com/microcosm-cc/bluemonday"
)

func htmlTestRuntime() *goja.Runtime {
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

func evalHTMLJS(t *testing.T, source string) goja.Value {
	t.Helper()

	runtime := htmlTestRuntime()

	v, err := runtime.RunString(source)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}

	return v
}

func TestHTMLEncode(t *testing.T) {
	result := evalHTMLJS(t, `html.encode("<b>Hello & World</b>")`)

	expected := "&lt;b&gt;Hello &amp; World&lt;/b&gt;"
	if result.Export() != expected {
		t.Errorf("expected %q, got %q", expected, result.Export())
	}
}

func TestHTMLEncodeAmpersand(t *testing.T) {
	result := evalHTMLJS(t, `html.encode("a & b & c")`)

	expected := "a &amp; b &amp; c"
	if result.Export() != expected {
		t.Errorf("expected %q, got %q", expected, result.Export())
	}
}

func TestHTMLDecode(t *testing.T) {
	result := evalHTMLJS(t, `html.decode("&lt;b&gt;Hello&lt;/b&gt;")`)

	expected := "<b>Hello</b>"
	if result.Export() != expected {
		t.Errorf("expected %q, got %q", expected, result.Export())
	}
}

func TestHTMLDecodeNumericEntities(t *testing.T) {
	result := evalHTMLJS(t, `html.decode("&#60;div&#62;")`)

	expected := "<div>"
	if result.Export() != expected {
		t.Errorf("expected %q, got %q", expected, result.Export())
	}
}

func TestHTMLEncodeDecodeRoundTrip(t *testing.T) {
	result := evalHTMLJS(t, `html.decode(html.encode("<p>Test & \"quotes\"</p>"))`)

	expected := "<p>Test & \"quotes\"</p>"
	if result.Export() != expected {
		t.Errorf("expected %q, got %q", expected, result.Export())
	}
}

func TestHTMLStripTags(t *testing.T) {
	result := evalHTMLJS(t, `html.strip("<b>bold</b> and <i>italic</i>")`)

	expected := "bold and italic"
	if result.Export() != expected {
		t.Errorf("expected %q, got %q", expected, result.Export())
	}
}

func TestHTMLStripScriptTags(t *testing.T) {
	result := evalHTMLJS(t, `html.strip("<script>alert('xss')</script>Safe text")`)

	expected := "Safe text"
	if result.Export() != expected {
		t.Errorf("expected %q, got %q", expected, result.Export())
	}
}

func TestHTMLStripPreservesEntities(t *testing.T) {
	result := evalHTMLJS(t, `html.strip("Tom &amp; Jerry")`)

	expected := "Tom &amp; Jerry"
	if result.Export() != expected {
		t.Errorf("expected %q, got %q", expected, result.Export())
	}
}

func TestHTMLStripPlainText(t *testing.T) {
	result := evalHTMLJS(t, `html.strip("no tags here")`)

	expected := "no tags here"
	if result.Export() != expected {
		t.Errorf("expected %q, got %q", expected, result.Export())
	}
}

func TestHTMLStripPolicyUGC(t *testing.T) {
	ugcPolicy := bluemonday.UGCPolicy()

	runtime := goja.New()

	err := runtime.Set("nd", newNDModule(runtime))
	if err != nil {
		t.Fatal(err)
	}

	err = runtime.Set("html", newHTMLModule(runtime, map[string]*bluemonday.Policy{
		"ugc": ugcPolicy,
	}))
	if err != nil {
		t.Fatal(err)
	}

	result, err := runtime.RunString(
		`html.strip_policy("<b>bold</b> <script>alert('xss')</script> <a href=\"https://example.com\">link</a>", "ugc")`)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}

	got, ok := result.Export().(string)
	if !ok {
		t.Fatalf("expected string, got %T", result.Export())
	}

	// UGC policy should preserve <b> and <a> but strip <script>.
	if got != `<b>bold</b>  <a href="https://example.com" rel="nofollow">link</a>` {
		t.Errorf("unexpected result: %q", got)
	}
}

func TestHTMLStripPolicyUnknown(t *testing.T) {
	runtime := goja.New()

	err := runtime.Set("html", newHTMLModule(runtime, map[string]*bluemonday.Policy{}))
	if err != nil {
		t.Fatal(err)
	}

	_, err = runtime.RunString(`html.strip_policy("text", "nonexistent")`)
	if err == nil {
		t.Fatal("expected error for unknown policy")
	}
}

func TestHTMLStripPolicyNilPolicies(t *testing.T) {
	runtime := goja.New()

	err := runtime.Set("html", newHTMLModule(runtime, nil))
	if err != nil {
		t.Fatal(err)
	}

	_, err = runtime.RunString(`html.strip_policy("text", "any")`)
	if err == nil {
		t.Fatal("expected error when no policies configured")
	}
}

func TestHTMLEncodeWrongArgCount(t *testing.T) {
	runtime := goja.New()

	err := runtime.Set("html", newHTMLModule(runtime, nil))
	if err != nil {
		t.Fatal(err)
	}

	_, err = runtime.RunString(`html.encode()`)
	if err == nil {
		t.Fatal("expected error for no arguments")
	}
}
