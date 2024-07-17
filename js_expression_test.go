package templ

import (
	"testing"
)

func TestJsVar(t *testing.T) {
	functionName := "myJsFunction"
	params := []any{
		"StringValue",
		123,
		JsExpression("event"),
		JsExpression("1 + 2"),
	}

	expected := "myJsFunction(\"StringValue\",123,event,1 + 2)"
	result := SafeScriptInline(functionName, params...)

	if result != expected {
		t.Fatalf("Expected '%s' but got '%s'", expected, result)
	}
}
