package goexpression

import "testing"

var testStringExpression = `"this string expression" } 
<div>
  But afterwards, it keeps searching.
<div>

<div>
  But that's not right, we can stop searching. It won't find anything valid.
</div>

<div>
  Certainly not later in the file.
</div>

<div>
	It's going to try all the tokens.
  )}]@<+.
</div>

<div>
	It's going to try all the tokens.
  )}]@<+.
</div>

<div>
	It's going to try all the tokens.
  )}]@<+.
</div>

<div>
	It's going to try all the tokens.
  )}]@<+.
</div>
`

func BenchmarkExpression(b *testing.B) {
	// Baseline...
	// BenchmarkExpression-10              6484            184862 ns/op
	// Updated...
	// BenchmarkExpression-10           3942538               279.6 ns/op
	for n := 0; n < b.N; n++ {
		start, end, err := Expression(testStringExpression)
		if err != nil {
			b.Fatal(err)
		}
		if start != 0 || end != 24 {
			b.Fatalf("expected 0, 24, got %d, %d", start, end)
		}
	}
}

var testTemplExpression = `templates.CallMethod(map[string]any{
	"name": "this string expression",
})

<div>
  But afterwards, it keeps searching.
<div>

<div>
  But that's not right, we can stop searching. It won't find anything valid.
</div>

<div>
  Certainly not later in the file.
</div>

<div>
	It's going to try all the tokens.
  )}]@<+.
</div>

<div>
	It's going to try all the tokens.
  )}]@<+.
</div>

<div>
	It's going to try all the tokens.
  )}]@<+.
</div>

<div>
	It's going to try all the tokens.
  )}]@<+.
</div>
`

func BenchmarkTemplExpression(b *testing.B) {
	// BenchmarkTemplExpression-10         2694            431934 ns/op
	// Updated...
	// BenchmarkTemplExpression-10      1339399               897.6 ns/op
	for n := 0; n < b.N; n++ {
		start, end, err := TemplExpression(testTemplExpression)
		if err != nil {
			b.Fatal(err)
		}
		if start != 0 || end != 74 {
			b.Fatalf("expected 0, 74, got %d, %d", start, end)
		}
	}
}
