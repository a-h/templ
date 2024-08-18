package runtime

import "testing"

func TestGetBuilder(t *testing.T) {
	sb := GetBuilder()
	sb.WriteString("test")
	if sb.String() != "test" {
		t.Errorf("expected \"test\", got %q", sb.String())
	}
}
