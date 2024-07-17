package templ

import (
	"testing"
)

func TestJsExpression(t *testing.T) {
	expected := "myJsFunction(\"StringValue\",123,event,1 + 2)"
	result := SafeScriptInline("myJsFunction", "StringValue", 123, JsExpression("event"), JsExpression("1 + 2"))

	if result != expected {
		t.Fatalf("TestJsExpression: Expected '%s' but got '%s'", expected, result)
	}
}
