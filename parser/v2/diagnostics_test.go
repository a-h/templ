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

		// textCallDiagnoser and futureTextCallDiagnoser

		{
			name: "text call diagnosers: email addresses produce no diagnostics",
			template: `
package main

templ template () {
	<p>user@example.com</p>
}`,
			want: nil,
		},
		{
			name: "text call diagnosers: component calls after whitespace produce no diagnostics",
			template: `
package main

templ template () {
	<p>Created @relativeTime(created)</p>
}`,
			want: nil,
		},
		{
			name: "text call diagnosers: @ without an identifier produces no diagnostics",
			template: `
package main

templ template () {
	<p>(@)</p>
}`,
			want: nil,
		},
		{
			name: "text call diagnosers: call after punctuation warns that it renders as text and that parsing will change",
			template: `
package main

templ template () {
	<p>Created (@relativeTime(created))</p>
}`,
			want: []Diagnostic{
				{
					Message: "`@relativeTime(` is rendered as text, not as a component call, because `@` does not follow whitespace. Place whitespace before `@` to call the component.",
					Range:   Range{Position{48, 4, 13}, Position{61, 4, 26}},
				},
				{
					Message: "A future templ release will parse `@relativeTime` as a component call, because `@` that does not follow a letter, digit, or underscore will start a component call. Use `{ \"@\" }` to render a literal `@`.",
					Range:   Range{Position{48, 4, 13}, Position{61, 4, 26}},
				},
			},
		},
		{
			name: "text call diagnosers: package-qualified call after punctuation warns for the qualified name",
			template: `
package main

templ template () {
	<p>(@components.Icon())</p>
}`,
			want: []Diagnostic{
				{
					Message: "`@components.Icon(` is rendered as text, not as a component call, because `@` does not follow whitespace. Place whitespace before `@` to call the component.",
					Range:   Range{Position{40, 4, 5}, Position{56, 4, 21}},
				},
				{
					Message: "A future templ release will parse `@components.Icon` as a component call, because `@` that does not follow a letter, digit, or underscore will start a component call. Use `{ \"@\" }` to render a literal `@`.",
					Range:   Range{Position{40, 4, 5}, Position{56, 4, 21}},
				},
			},
		},
		{
			name: "text call diagnosers: call after an expression and multi-byte text warns at the @ position",
			template: `
package main

templ template () {
	<p>Version { "v1" } · Created { created } (@snapshotRelativeTime(created))</p>
}`,
			want: []Diagnostic{
				{
					Message: "`@snapshotRelativeTime(` is rendered as text, not as a component call, because `@` does not follow whitespace. Place whitespace before `@` to call the component.",
					Range:   Range{Position{80, 4, 45}, Position{101, 4, 66}},
				},
				{
					Message: "A future templ release will parse `@snapshotRelativeTime` as a component call, because `@` that does not follow a letter, digit, or underscore will start a component call. Use `{ \"@\" }` to render a literal `@`.",
					Range:   Range{Position{80, 4, 45}, Position{101, 4, 66}},
				},
			},
		},
		{
			name: "text call diagnosers: handle after punctuation warns only that parsing will change",
			template: `
package main

templ template () {
	<p>Follow "@handle"</p>
}`,
			want: []Diagnostic{
				{
					Message: "A future templ release will parse `@handle` as a component call, because `@` that does not follow a letter, digit, or underscore will start a component call. Use `{ \"@\" }` to render a literal `@`.",
					Range:   Range{Position{47, 4, 12}, Position{54, 4, 19}},
				},
			},
		},
		{
			name: "text call diagnosers: trailing full stop is excluded from the name and each @ warns in order",
			template: `
package main

templ template () {
	<p>(@a) (@b.</p>
}`,
			want: []Diagnostic{
				{
					Message: "A future templ release will parse `@a` as a component call, because `@` that does not follow a letter, digit, or underscore will start a component call. Use `{ \"@\" }` to render a literal `@`.",
					Range:   Range{Position{40, 4, 5}, Position{42, 4, 7}},
				},
				{
					Message: "A future templ release will parse `@b` as a component call, because `@` that does not follow a letter, digit, or underscore will start a component call. Use `{ \"@\" }` to render a literal `@`.",
					Range:   Range{Position{45, 4, 10}, Position{47, 4, 12}},
				},
			},
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
