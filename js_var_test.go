package templ

import (
	"testing"
)

func TestJsVar(t *testing.T) {
	functionName := "myJsFunction"
	params := []any{
		"StringValue",
		123,
		JsVar("event"),
	}

	expected := "myJsFunction(\"StringValue\",123,event)"
	result := SafeScriptInline(functionName, params...)

	if result != expected {
		t.Fatalf("Expected '%s' but got '%s'", expected, result)
	}
}
