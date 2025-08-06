package runtime

import (
	"errors"
	"testing"
)

func TestScriptContent(t *testing.T) {
	tests := []struct {
		name                         string
		input                        any
		expectedInsideStringLiteral  string
		expectedOutsideStringLiteral string
	}{
		{
			name:                         "string",
			input:                        "hello",
			expectedInsideStringLiteral:  `hello`,
			expectedOutsideStringLiteral: `"hello"`,
		},
		{
			name:                         "string with single quotes",
			input:                        "hello 'world'",
			expectedInsideStringLiteral:  `hello \u0027world\u0027`,
			expectedOutsideStringLiteral: `"hello 'world'"`,
		},
		{
			name:                         "string with double quotes",
			input:                        "hello \"world\"",
			expectedInsideStringLiteral:  `hello \u0022world\u0022`,
			expectedOutsideStringLiteral: `"hello \"world\""`,
		},
		{
			name:                         "string with backticks",
			input:                        "hello `world`",
			expectedInsideStringLiteral:  `hello \u0060world\u0060`,
			expectedOutsideStringLiteral: "\"hello `world`\"",
		},
		{
			name:                         "int",
			input:                        1,
			expectedInsideStringLiteral:  `1`,
			expectedOutsideStringLiteral: `1`,
		},
		{
			name:                         "float",
			input:                        1.1,
			expectedInsideStringLiteral:  `1.1`,
			expectedOutsideStringLiteral: `1.1`,
		},
		{
			name:                         "bool",
			input:                        true,
			expectedInsideStringLiteral:  `true`,
			expectedOutsideStringLiteral: `true`,
		},
		{
			name:                         "array",
			input:                        []int{1, 2, 3},
			expectedInsideStringLiteral:  `[1,2,3]`,
			expectedOutsideStringLiteral: `[1,2,3]`,
		},
		{
			name:                         "object",
			input:                        struct{ Name string }{"Alice"},
			expectedInsideStringLiteral:  `{\u0022Name\u0022:\u0022Alice\u0022}`,
			expectedOutsideStringLiteral: `{"Name":"Alice"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			t.Run("inside string literal", func(t *testing.T) {
				actualInsideStringLiteral, err := ScriptContentInsideStringLiteral(tt.input)
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				if actualInsideStringLiteral != tt.expectedInsideStringLiteral {
					t.Errorf("expected:\n%s\ngot:\n%s", tt.expectedInsideStringLiteral, actualInsideStringLiteral)
				}
			})
			t.Run("outside string literal", func(t *testing.T) {
				actualOutsideStringLiteral, err := ScriptContentOutsideStringLiteral(tt.input)
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				if actualOutsideStringLiteral != tt.expectedOutsideStringLiteral {
					t.Errorf("expected:\n%s\ngot:\n%s", tt.expectedOutsideStringLiteral, actualOutsideStringLiteral)
				}
			})
		})
	}
}

func TestScriptContentErrors(t *testing.T) {
	t.Run("inside string literal", func(t *testing.T) {
		_, err := ScriptContentInsideStringLiteral("s", errors.New("error"))
		if err == nil {
			t.Fatal("expected error")
		}
		if err.Error() != "error" {
			t.Errorf("expected error: %s", err)
		}
	})
	t.Run("outside string literal", func(t *testing.T) {
		_, err := ScriptContentOutsideStringLiteral("s", errors.New("error"))
		if err == nil {
			t.Fatal("expected error")
		}
		if err.Error() != "error" {
			t.Errorf("expected error: %s", err)
		}
	})
	t.Run("unmarshal error", func(t *testing.T) {
		_, err := ScriptContentInsideStringLiteral(func() string {
			return "hello"
		})
		if err == nil {
			t.Fatalf("expected unmarshal error, but got %v", err)
		}
	})
}
