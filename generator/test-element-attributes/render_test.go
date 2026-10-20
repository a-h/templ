package testelementattributes

import (
	_ "embed"
	"os"
	"testing"

	"github.com/a-h/templ/generator/htmldiff"
)

//go:embed expected.html
var expected string

func Test(t *testing.T) {
	component := render(person{
		name:  "Luiz Bonfa",
		email: "luiz@example.com",
	})

	actual, diff, err := htmldiff.Diff(component, expected)
	if err != nil {
		t.Fatal(err)
	}
	if diff != "" {
		if err := os.WriteFile("actual.html", []byte(actual), 0644); err != nil {
			t.Errorf("failed to write actual.html: %v", err)
		}
		t.Error(diff)
	}
}

//go:embed expected_dynamic_keys.html
var expectedDynamicKeys string

func TestDynamicKeys(t *testing.T) {
	component := dynamicKeys("javascript:alert(document.domain)", "x onmouseover=alert(document.domain) y")

	actual, diff, err := htmldiff.Diff(component, expectedDynamicKeys)
	if err != nil {
		t.Fatal(err)
	}
	if diff != "" {
		if err := os.WriteFile("actual_dynamic_keys.html", []byte(actual), 0644); err != nil {
			t.Errorf("failed to write actual_dynamic_keys.html: %v", err)
		}
		t.Error(diff)
	}
}
