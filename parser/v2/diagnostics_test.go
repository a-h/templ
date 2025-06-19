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

templ template () {
	{! templ.Raw("foo") }
}`,
			want: []Diagnostic{{
				Message: "`{! foo }` syntax is deprecated. Use `@foo` syntax instead. Run `templ fmt .` to fix all instances.",
				Range:   Range{Position{39, 4, 4}, Position{55, 4, 20}},
			}},
		},
		{
			name: "useOfLegacyCallSyntaxDiagnoser: in div",
			template: `
package main

templ template () {
	<div>
		{! templ.Raw("foo") }
	</div>
}`,
			want: []Diagnostic{{
				Message: "`{! foo }` syntax is deprecated. Use `@foo` syntax instead. Run `templ fmt .` to fix all instances.",
				Range:   Range{Position{47, 5, 5}, Position{63, 5, 21}},
			}},
		},
		{
			name: "useOfLegacyCallSyntaxDiagnoser: in if",
			template: `
package main

templ template () {
	if true {
		{! templ.Raw("foo") }
	}
}`,
			want: []Diagnostic{{
				Message: "`{! foo }` syntax is deprecated. Use `@foo` syntax instead. Run `templ fmt .` to fix all instances.",
				Range:   Range{Position{51, 5, 5}, Position{67, 5, 21}},
			}},
		},
		{
			name: "useOfLegacyCallSyntaxDiagnoser: in for",
			template: `
package main

templ template () {
	for i := range x {
		{! templ.Raw("foo") }
	}
}`,
			want: []Diagnostic{{
				Message: "`{! foo }` syntax is deprecated. Use `@foo` syntax instead. Run `templ fmt .` to fix all instances.",
				Range:   Range{Position{60, 5, 5}, Position{76, 5, 21}},
			}},
		},
		{
			name: "useOfLegacyCallSyntaxDiagnoser: in switch",
			template: `
package main

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
					Range:   Range{Position{61, 6, 5}, Position{77, 6, 21}},
				},
				{
					Message: "`{! foo }` syntax is deprecated. Use `@foo` syntax instead. Run `templ fmt .` to fix all instances.",
					Range:   Range{Position{95, 8, 5}, Position{96, 8, 6}},
				},
			},
		},
		{
			name: "useOfLegacyCallSyntaxDiagnoser: in block",
			template: `
package main

templ template () {
	@layout("Home") {
		{! templ.Raw("foo") }
	}
}`,
			want: []Diagnostic{{
				Message: "`{! foo }` syntax is deprecated. Use `@foo` syntax instead. Run `templ fmt .` to fix all instances.",
				Range:   Range{Position{59, 5, 5}, Position{75, 5, 21}},
			}},
		},
		{
			name: "voidElementWithChildrenDiagnoser: no diagnostics",
			template: `
package main

templ template () {
	<div>
		<input/>
	</div>
}`,
			want: nil,
		},

		// missingComponentDiagnoser

		{
			name: "missingComponentDiagnoser: no diagnostics - no components",
			template: `
package main

templ template () {
	<div>Regular HTML</div>
}`,
			want: nil,
		},
		{
			name: "missingComponentDiagnoser: no diagnostics - valid component",
			template: `
package main

templ ValidComponent() {
	<div>Valid component</div>
}

templ template() {
	<ValidComponent />
}`,
			want: nil,
		},
		{
			name: "missingComponentDiagnoser: missing component",
			template: `
package main

templ template() {
	<MissingComponent />
}`,
			want: []Diagnostic{{
				Message: "Component MissingComponent not found",
				Range:   Range{Position{36, 4, 2}, Position{52, 4, 18}},
			}},
		},
		{
			name: "missingComponentDiagnoser: multiple missing components",
			template: `
package main

templ template() {
	<FirstMissing />
	<SecondMissing />
}`,
			want: []Diagnostic{
				{
					Message: "Component FirstMissing not found",
					Range:   Range{Position{36, 4, 2}, Position{48, 4, 14}},
				},
				{
					Message: "Component SecondMissing not found", 
					Range:   Range{Position{54, 5, 2}, Position{67, 5, 15}},
				},
			},
		},
		{
			name: "missingComponentDiagnoser: mixed valid and missing",
			template: `
package main

templ ValidComponent() {
	<div>Valid</div>
}

templ template() {
	<ValidComponent />
	<MissingComponent />
}`,
			want: []Diagnostic{{
				Message: "Component MissingComponent not found",
				Range:   Range{Position{102, 9, 2}, Position{118, 9, 18}},
			}},
		},
		{
			name: "missingComponentDiagnoser: skip package-qualified components",
			template: `
package main

templ template() {
	<pkg.ExternalComponent />
	<MissingLocal />
}`,
			want: []Diagnostic{{
				Message: "Component MissingLocal not found",
				Range:   Range{Position{63, 5, 2}, Position{75, 5, 14}},
			}},
		},
		{
			name: "missingComponentDiagnoser: skip struct method components",
			template: `
package main

templ template() {
	<variable.Method />
	<MissingLocal />
}`,
			want: []Diagnostic{{
				Message: "Component MissingLocal not found",
				Range:   Range{Position{57, 5, 2}, Position{69, 5, 14}},
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
