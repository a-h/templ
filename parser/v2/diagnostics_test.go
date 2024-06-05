package parser

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDiagnose(t *testing.T) {
	tests := []struct {
		name     string
		template string
		want     []Diagnostic
	}{
		{
			name: "no diagnostics",
			template: `
package main

import "github.com/a-h/templ"

templ template () {
	<p>Hello, World!</p>
}`,
			want: nil,
		},

		// useOfLegacyCallSyntaxDiagnoser

		{
			name: "useOfLegacyCallSyntaxDiagnoser: template root",
			template: `
package main

import "github.com/a-h/templ"

templ template () {
	{! templ.Raw("foo") }
}`,
			want: []Diagnostic{{
				Message: "`{! foo }` syntax is deprecated. Use `@foo` syntax instead. Run `templ fmt .` to fix all instances.",
				Range:   Range{Position{70, 6, 4}, Position{86, 6, 20}},
			}},
		},
		{
			name: "useOfLegacyCallSyntaxDiagnoser: in div",
			template: `
package main

import "github.com/a-h/templ"

templ template () {
	<div>
		{! templ.Raw("foo") }
	</div>
}`,
			want: []Diagnostic{{
				Message: "`{! foo }` syntax is deprecated. Use `@foo` syntax instead. Run `templ fmt .` to fix all instances.",
				Range:   Range{Position{78, 7, 5}, Position{94, 7, 21}},
			}},
		},
		{
			name: "useOfLegacyCallSyntaxDiagnoser: in if",
			template: `
package main

import "github.com/a-h/templ"

templ template () {
	if true {
		{! templ.Raw("foo") }
	}
}`,
			want: []Diagnostic{{
				Message: "`{! foo }` syntax is deprecated. Use `@foo` syntax instead. Run `templ fmt .` to fix all instances.",
				Range:   Range{Position{82, 7, 5}, Position{98, 7, 21}},
			}},
		},
		{
			name: "useOfLegacyCallSyntaxDiagnoser: in for",
			template: `
package main

import "github.com/a-h/templ"

templ template () {
	for i := range x {
		{! templ.Raw("foo") }
	}
}`,
			want: []Diagnostic{{
				Message: "`{! foo }` syntax is deprecated. Use `@foo` syntax instead. Run `templ fmt .` to fix all instances.",
				Range:   Range{Position{91, 7, 5}, Position{107, 7, 21}},
			}},
		},
		{
			name: "useOfLegacyCallSyntaxDiagnoser: in switch",
			template: `
package main

import "github.com/a-h/templ"

templ template () {
	switch x {
	case 1:
		{! templ.Raw("foo") }
	default:
		{! x }
	}
}`,
			want: []Diagnostic{
				{
					Message: "`{! foo }` syntax is deprecated. Use `@foo` syntax instead. Run `templ fmt .` to fix all instances.",
					Range:   Range{Position{92, 8, 5}, Position{108, 8, 21}},
				},
				{
					Message: "`{! foo }` syntax is deprecated. Use `@foo` syntax instead. Run `templ fmt .` to fix all instances.",
					Range:   Range{Position{126, 10, 5}, Position{127, 10, 6}},
				},
			},
		},
		{
			name: "useOfLegacyCallSyntaxDiagnoser: in block",
			template: `
package main

import "github.com/a-h/templ"

templ template () {
	@layout("Home") {
		{! templ.Raw("foo") }
	}
}`,
			want: []Diagnostic{{
				Message: "`{! foo }` syntax is deprecated. Use `@foo` syntax instead. Run `templ fmt .` to fix all instances.",
				Range:   Range{Position{90, 7, 5}, Position{106, 7, 21}},
			}},
		},
		{
			name: "voidElementWithChildrenDiagnoser: no diagnostics",
			template: `
package main

import "github.com/a-h/templ"

templ template () {
	<div>
		<input/>
	</div>
}`,
			want: nil,
		},
		{
			name: "voidElementWithChildrenDiagnoser: with diagnostics",
			template: `
package main

import "github.com/a-h/templ"

templ template () {
	<div>
	  <input>Child content</input>
	</div>
}`,
			want: []Diagnostic{{
				Message: "void element <input> should not have child content",
				Range:   Range{Position{77, 7, 4}, Position{82, 7, 9}},
			}},
		},
		{
			name: "templNotImportedDiagnoser: with diagnostics",
			template: `
package main

templ template () {
	<div></div>
}`,
			want: []Diagnostic{{
				Message: "no \"github.com/a-h/templ\" import found. Run `templ fmt .` to fix all instances.",
			}},
		},
		{
			name: "templNotImportedDiagnoser: with existing import",
			template: `
package main

import "fmt"

templ template () {
	<div></div>
}`,
			want: []Diagnostic{{
				Message: "no \"github.com/a-h/templ\" import found. Run `templ fmt .` to fix all instances.",
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tf, err := ParseString(tt.template)
			if err != nil {
				t.Fatalf("ParseTemplateFile() error = %v", err)
			}
			got, err := Diagnose(tf)
			if err != nil {
				t.Fatalf("Diagnose() error = %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Diagnose() mismatch (-got +want):\n%s", diff)
			}
		})
	}
}
