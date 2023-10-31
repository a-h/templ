# Contributing to templ

## Vision

Enable Go developers to build strongly typed, component-based HTML user interfaces with first-class developer tooling, and a short learning curve.

## Come up with a design and share it

Before starting work on any major pull requests or code changes, start a discussion at https://github.com/a-h/templ/discussions or raise an issue.

We don't want you to spend time on a PR or feature that ultimately doesn't get merged because it doesn't fit with the project goals, or the design doesn't work for some reason.

For issues, it really helps if you provide a reproduction repo, or can create a failing unit test to describe the behaviour.

In designs, we need to consider:

* Backwards compatibility - Not changing the public API between releases, introducing gradual deprecation - don't break people's code.
* Correctness over time - How can we reduce the risk of defects both now, and in future releases?
* Threat model - How could each change be used to inject vulnerabilities into web pages?
* Go version - We target the oldest supported version of Go as per https://go.dev/doc/devel/release
* Automatic migration - If we need to force through a change.
* Compile time vs runtime errors - Prefer compile time.
* Documentation - New features are only useful if people can understand the new feature, what would the documentation look like?
* Examples - How will we demonstrate the feature?

## Project structure

templ is structured into a few areas:

### Parser `./parser`

The parser directory currently contains both v1 and v2 parsers.

The v1 parser is not maintained, it's only used to migrate v1 code over to the v2 syntax.

The parser is responsible for parsing templ files into an object model. The types that make up the object model are in `types.go`. Automatic formatting of the types is tested in `types_test.go`.

A templ file is parsed into the `TemplateFile` struct object model.

```go
type TemplateFile struct {
	// Header contains comments or whitespace at the top of the file.
	Header []GoExpression
	// Package expression.
	Package Package
	// Nodes in the file.
	Nodes []TemplateFileNode
}
```

Parsers are individually tested using two types of unit test.

One test covers the successful parsing of text into an object. For example, the `HTMLCommentParser` test checks for successful patterns.

```go
func TestHTMLCommentParser(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected HTMLComment
	}{
		{
			name:  "comment - single line",
			input: `<!-- single line comment -->`,
			expected: HTMLComment{
				Contents: " single line comment ",
			},
		},
		{
			name:  "comment - no whitespace",
			input: `<!--no whitespace between sequence open and close-->`,
			expected: HTMLComment{
				Contents: "no whitespace between sequence open and close",
			},
		},
		{
			name: "comment - multiline",
			input: `<!-- multiline
								comment
					-->`,
			expected: HTMLComment{
				Contents: ` multiline
								comment
					`,
			},
		},
		{
			name:  "comment - with tag",
			input: `<!-- <p class="test">tag</p> -->`,
			expected: HTMLComment{
				Contents: ` <p class="test">tag</p> `,
			},
		},
		{
			name:  "comments can contain tags",
			input: `<!-- <div> hello world </div> -->`,
			expected: HTMLComment{
				Contents: ` <div> hello world </div> `,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			result, ok, err := htmlComment.Parse(input)
			if err != nil {
				t.Fatalf("parser error: %v", err)
			}
			if !ok {
				t.Fatalf("failed to parse at %d", input.Index())
			}
			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}
```

Alongside each success test, is a similar test to check that invalid syntax is detected.

```go
func TestHTMLCommentParserErrors(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected error
	}{
		{
			name:  "unclosed HTML comment",
			input: `<!-- unclosed HTML comment`,
			expected: parse.Error("expected end comment literal '-->' not found",
				parse.Position{
					Index: 26,
					Line:  0,
					Col:   26,
				}),
		},
		{
			name:  "comment in comment",
			input: `<!-- <-- other --> -->`,
			expected: parse.Error("comment contains invalid sequence '--'", parse.Position{
				Index: 8,
				Line:  0,
				Col:   8,
			}),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := parse.NewInput(tt.input)
			_, _, err := htmlComment.Parse(input)
			if diff := cmp.Diff(tt.expected, err); diff != "" {
				t.Error(diff)
			}
		})
	}
}
```

### Generator

The generator takes the object model and writes out Go code that produces the expected output. Any changes to Go code output by templ are made in this area.

Testing of the generator is carried out by creating a templ file, and a matching expected output file.

For example, `./generator/test-a-href` contains a templ file of:

```templ
package testahref

templ render() {
	<a href="javascript:alert(&#39;unaffected&#39;);">Ignored</a>
	<a href={ templ.URL("javascript:alert('should be sanitized')") }>Sanitized</a>
	<a href={ templ.SafeURL("javascript:alert('should not be sanitized')") }>Unsanitized</a>
}
```

It also contains an expected output file.

```html
<a href="javascript:alert(&#39;unaffected&#39;);">Ignored</a>
<a href="about:invalid#TemplFailedSanitizationURL">Sanitized</a>
<a href="javascript:alert(&#39;should not be sanitized&#39;)">Unsanitized</a>
```

These tests contribute towards the code coverage metrics by building an instrumented test CLI program. See the `test-cover` task in the `README.md` file.

### CLI

The command line interface for templ is used to generate Go code from templ files, format templ files, and run the LSP.

The code for this is at `./cmd/templ`.

Testing of the templ command line is done with unit tests to check the argument parsing.

The `templ generate` command is tested by generating templ files in the project, and testing that the expected output HTML is present.

### Runtime

The runtime is used by generated code, and by template authors, to serve template content over HTTP, and to carry out various operations.

It is in the root directory of the project at `./runtime.go`. The runtime is unit tested, as well as being tested as part of the `generate` tests.

### LSP

The LSP is structured within the command line interface, and proxies commands through to the `gopls` LSP.

### Docs

The docs are a Docusaurus project at `./docs`.

## Coding

### Build tasks

templ uses the `xc` task runner - https://github.com/joerdav/xc

If you run `xc` you can get see a list of the development tasks that can be run, or you can read the `README.md` file and see the `Tasks` section.

The most useful tasks for local development are:

* `install-snapshot` - this builds the templ CLI and installs it into `~/bin`. Ensure that this is in your path.
* `test` - this regenerates all templates, and runs the unit tests.
* `fmt` - run the `gofmt` tool to format all Go code.
* `lint` - run the same linting as run in the CI process.
* `docs-run` - run the Docusaurus documentation site.

### Commit messages

The project using https://www.conventionalcommits.org/en/v1.0.0/

Examples:

* `feat: support Go comments in templates, fixes #234"`

### Coding style

* Reduce nesting - i.e. prefer early returns over an `else` block, as per https://danp.net/posts/reducing-go-nesting/ or https://go.dev/doc/effective_go#if
* Use line breaks to separate "paragraphs" of code - don't use line breaks in between lines, or at the start/end of functions etc.
* Use the `fmt` and `lint` build tasks to format and lint your code before submitting a PR.

