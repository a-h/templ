package templ

import (
	"testing"
)

func TestJSExpression(t *testing.T) {
	expected := "myJSFunction(\"StringValue\",123,event,1 + 2)"
	actual := SafeScriptInline("myJSFunction", "StringValue", 123, JSExpression("event"), JSExpression("1 + 2"))

	if actual != expected {
		t.Fatalf("TestJSExpression: expected %q, got %q", expected, actual)
	}
}
