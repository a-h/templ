package templ

import (
	"testing"

	"github.com/a-h/lexical/input"
	"github.com/google/go-cmp/cmp"
)

func TestPackageParserErrors(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected parseError
	}{
		{
			name:  "unterminated package",
			input: "{% package",
			expected: newParseError(
				"package literal not terminated",
				Position{
					Index: 0,
					Line:  1,
					Col:   0,
				},
				Position{
					Index: 10,
					Line:  1,
					Col:   10,
				},
			),
		},
		{
			name:  "unterminated package, new line",
			input: "{% package \n%}",
			expected: newParseError(
				"package literal not terminated",
				Position{
					Index: 0,
					Line:  1,
					Col:   0,
				},
				Position{
					Index: 10,
					Line:  1,
					Col:   10,
				},
			),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			pi := input.NewFromString(tt.input)
			sril := NewSourceRangeToItemLookup()
			actual := newPackageParser(sril).Parse(pi)
			if actual.Success {
				t.Errorf("expected parsing to fail, but it succeeded")
			}
			if diff := cmp.Diff(tt.expected, actual.Error); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}

func TestPackageParser(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected interface{}
	}{
		{
			name:  "package: standard",
			input: `{% package templ %}`,
			expected: Package{
				Expression: "templ",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := input.NewFromString(tt.input)
			sril := NewSourceRangeToItemLookup()
			parser := newPackageParser(sril)
			result := parser.Parse(input)
			if result.Error != nil {
				t.Fatalf("paser error: %v", result.Error)
			}
			if !result.Success {
				t.Fatalf("failed to parse at %d", input.Index())
			}
			if diff := cmp.Diff(tt.expected, result.Item); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}

func TestPackageParserLocations(t *testing.T) {
	input := input.NewFromString(`{% package templ %}`)
	sril := NewSourceRangeToItemLookup()
	parser := newPackageParser(sril)

	result := parser.Parse(input)
	if result.Error != nil {
		t.Fatalf("paser error: %v", result.Error)
	}
	if !result.Success {
		t.Fatalf("failed to parse at %d", input.Index())
	}
	var expected Package
	expected = result.Item.(Package)

	t.Run("lookup by index", func(t *testing.T) {
		actualItemRange, ok := parser.SourceRangeToItemLookup.LookupByIndex(1)
		if !ok {
			t.Errorf("expected ok, got %v from %+v", ok, parser.SourceRangeToItemLookup)
		}
		actual := actualItemRange.Item.(Package)
		if expected != actual {
			t.Errorf("expected %v, got %v", expected, actual)
		}
	})
	t.Run("lookup by line, col", func(t *testing.T) {
		actualItemRange, ok := parser.SourceRangeToItemLookup.LookupByLineCol(1, 3)
		if !ok {
			t.Errorf("expected ok, got %v from %+v", ok, parser.SourceRangeToItemLookup)
		}
		actual := actualItemRange.Item.(Package)
		if expected != actual {
			t.Errorf("expected %v, got %v", expected, actual)
		}
	})
}
