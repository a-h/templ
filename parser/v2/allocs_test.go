package parser

import (
	"testing"

	"github.com/a-h/parse"
)

func RunParserAllocTest[T any](t *testing.T, p parse.Parser[T], expectOK bool, maxAllocs int64, input string) {
	pi := parse.NewInput(input)
	actual := testing.AllocsPerRun(4, func() {
		pi.Seek(0)
		_, ok, err := p.Parse(pi)
		if err != nil {
			t.Fatalf("error parsing %T: %v", p, err)
		}
		if ok != expectOK {
			t.Fatalf("failed to parse %T", p)
		}
	})

	// Run the benchmark.
	if int64(actual) > maxAllocs {
		t.Fatalf("Expected allocs <= %d, got %d", maxAllocs, int64(actual))
	}
}
