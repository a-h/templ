# templ

* A strongly typed HTML templating language that compiles to Go code, and has great developer tooling.

## Getting started

* Install the `templ` command-line tool.
* Create a `*.templ` file containing a template.
* Run `templ generate` to create Go code from the template.

## Current state

This is beta software, the template language may still have breaking changes. There's no guarantees of stability or correctness at the moment.

If you're keen to see Go be practical for Web projects, see "Help needed" for where the project needs your help.

## Features

The language generates Go code, some sections of the template (e.g. `package`, `import`, `if`, `for` and `switch` statements) are output directly as Go expressions in the generated output, while HTML elements are converted to Go code that renders their output.

* `templ generate` generates Go code from `*.templ` files.
* `templ fmt` formats template files in the current directory tree.
* `templ lsp` provides a Language Server to support IDE integrations. The compile command generates a sourcemap which maps from the `*.templ` files to the compiled Go file. This enables the `templ` LSP to use the Go language `gopls` language server as is, providing a thin shim to do the source remapping. This is used to provide autocomplete for template variables and functions.

## Security

templ currently uses context unaware escaping, see https://github.com/a-h/templ/issues/6 for a proposal to add context-aware content escaping.

## Design

### Overview

* `*.templ` files are used to generate Go code to efficiently render the template at runtime.
* Go code is generated from template files using the `templ generate` command, while templates can be formatted with the `templ fmt` command. The `templ lsp` command provides an LSP (Language Server Protocol) server to enable autocomplete.
* Each `{% templ ComponentName(params Params) %}` section compiles into a function that creates a `templ.Component`.
* `templ.Component` is an interface with a single function - `Render(ctx context.Context, io.Writer) (err error)`. You can make a component entirely in Go code and interact with it via `templ`.
* `templ` aims for correctness, simplicity, developer experience and raw performance, in that order. The goal is to make writing Go web applications more practical, achievable and desirable.
* Provides minified HTML output only.
* Components can be composed into layouts.

### Package

Since `templ` files are as close to Go as possible, they start with a package expression.

```
{% package templ %}
```

### Importing packages

After the package expression, they might import other Go packages, just like Go files. There's no multi-line import statement, just a single import per line.

```
{% import "strings" %}
```

### Components

Once the package and import statements are done, we can define components using the `{% templ Name(params Params) %}` expression. The `templ` expressions are converted into Go functions when the `templ generate` command is executed.

```
{% templ AddressView(addr Address) %}
	<div>{%= addr.Address1 %}</div>
	<div>{%= addr.Address2 %}</div>
	<div>{%= addr.Address3 %}</div>
	<div>{%= addr.Address4 %}</div>
{% endtempl %}
```

Each `templ.Component` can contain HTML elements, strings, for loops, switch statements and references to other templates.

#### Referencing other components

Components can be referenced in the body of the template, and can pass data between then, for example, using the `AddressTemplate` from the `PersonTemplate`.

```
{% templ PersonTemplate(p Person) %}
	<div>
	    {% for _, v := range p.Addresses %}
		    {%! AddressTemplate(v) %}
	    {% endfor %}
	</div>
{% endtempl %}
```

It's also possible to create "higher order components" that compose other instances of `templ.Component` without passing data, or even knowing what the concrete type of the component will be ahead of time. So long as is implements `templ.Component`, it can be used.

For example, this template accepts 3 templates (header, footer, body) and renders all 3 of them in the expected order.

```
{% templ Layout(header, footer, body templ.Component) %}
	{%! header %}
	{%! body %}
	{%! footer %}
{% endtempl %}
```

#### Code-only components

It's possible to create a `templ.Component` entirely in Go code. Within `templ`, strings are automatically escaped to reduce the risk of cross-site-scripting attacks, but it's possible to create your own "Raw" component that bypasses this behaviour: 

```go
func Raw(s string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		_, err = io.WriteString(w, s)
		return
	})
}
```

Then call it in a template. So long as the `Raw` function is in scope, you can use it.

```
{%! Raw("<script>alert('xss vector');</script>") %}
```

### Elements

HTML elements look like HTML and you can write static attributes into them, just like with normal HTML. Don't worry about the spacing, the HTML will be minified when it's rendered.

```
<div id="address1">{%= addr.Address1 %}</div>
```

You can also have dynamic attributes that use template parameter, other Go variables that happen to be in scope, or call Go functions that return a string. Don't worry about HTML encoding element text and attribute values, that will be taken care of automatically.

```
<a href={%= p.URL %}>{%= strings.ToUpper(p.Name()) %}</a>
```

### Text

Text is rendered from Go expressions, which includes constant values:

```
{%= "this is a string" %}
```

Using the backtick format (single-line only):

```
{%= `this is also a string` %}
```

Calling a function that returns a string:

```
{%= time.Now().String() %}
```

Or using a string parameter, or variable that's in scope.

```
{%= v.s %}
```

What you can't do, is write text directly between elements (e.g. `<div>Some text</div>`, because the parser would have to become more complex to support HTML entities and the various mistakes people make when they're doing that (bare ampersands etc.). Go strings support UTF-8 which is much easier, and the escaping rules are well known by Go programmers.

### If/Else

Templates can contain if/else statements that follow the same pattern as Go.

```
{% if p.Type == "test" %}
	<span>{%= "Test user" %}</span>
{% else %}
	<span>{%= "Not test user" %}</span>
{% endif %}
```

### For

Templates have the same loop behaviour as Go.

```
{% for _, v := range p.Addresses %}
	<li>{%= v.City %}</li>
{% endfor %}
```

### Switch/Case

Switch statements work in the same way as they do in Go. 

```
{% switch p.Type %}
	{% case "test" %}
		<span>{%= "Test user" %}</span>
	{% endcase %}
	{% case "admin" %}
		<span>{%= "Admin user" %}</span>
	{% endcase %}
	{% default %}
		<span>{%= "Unknown user" %}</span>
	{% enddefault %}
{% endswitch %}
```

## Full example

```templ
{% package templ %}

{% import "strings" %}

{% templ Layout(header, footer, body templ.Component) %}
	{%! header %}
	{%! body %}
	{%! footer %}
{% endtempl %}

{% templ AddressTemplate(addr Address) %}
	<div>{%= addr.Address1 %}</div>
	<div>{%= addr.Address2 %}</div>
	<div>{%= addr.Address3 %}</div>
	<div>{%= addr.Address4 %}</div>
{% endtempl %}

{% templ PersonTemplate(p Person) %}
	<div>
		<div>{%= p.Name() %}</div>
		<a href={%= p.URL %}>{%= strings.ToUpper(p.Name()) %}</a>
		<div>
			{% if p.Type == "test" %}
				<span>{%= "Test user" %}</span>
			{% else %}
				<span>{%= "Not test user" %}</span>
			{% endif %}
			{% for _, v := range p.Addresses %}
				{%! AddressTemplate(v) %}
			{% endfor %}
			{% switch p.Type %}
				{% case "test" %}
					<span>{%= "Test user" %}</span>
				{% endcase %}
				{% case "admin" %}
					<span>{%= "Admin user" %}</span>
				{% endcase %}
				{% default %}
					<span>{%= "Unknown user" %}</span>
				{% enddefault %}
			{% endswitch %}
		</div>
	</div>
{% endtempl %}
```

Will compile to Go code similar to the following (error handling removed for brevity):

```go
// Code generated by templ DO NOT EDIT.

package templ

import "github.com/a-h/templ"
import "context"
import "io"
import "strings"

func Layout(header, footer, body templ.Component) (t templ.Component) {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		err = header.Render(ctx, w)
		err = body.Render(ctx, w)
		err = footer.Render(ctx, w)
		return err
	})
}

func AddressTemplate(addr Address) (t templ.Component) {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		_, err = io.WriteString(w, "<div>")
		_, err = io.WriteString(w, templ.EscapeString(addr.Address1))
		_, err = io.WriteString(w, "</div>")
		_, err = io.WriteString(w, "<div>")
		_, err = io.WriteString(w, templ.EscapeString(addr.Address2))
		_, err = io.WriteString(w, "</div>")
		// Cut for brevity.
		return err
	})
}

func PersonTemplate(p Person) (t templ.Component) {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		_, err = io.WriteString(w, "<div>")
		_, err = io.WriteString(w, "<div>")
		_, err = io.WriteString(w, templ.EscapeString(p.Name()))
		_, err = io.WriteString(w, "</div>")
		_, err = io.WriteString(w, "<a")
		_, err = io.WriteString(w, " href=")
		_, err = io.WriteString(w, "\"")
		_, err = io.WriteString(w, templ.EscapeString(p.URL))
		_, err = io.WriteString(w, "\"")
		_, err = io.WriteString(w, ">")
		_, err = io.WriteString(w, templ.EscapeString(strings.ToUpper(p.Name())))
		_, err = io.WriteString(w, "</a>")
		_, err = io.WriteString(w, "<div>")
		if p.Type == "test" {
			_, err = io.WriteString(w, "<span>")
			_, err = io.WriteString(w, templ.EscapeString("Test user"))
			_, err = io.WriteString(w, "</span>")
		} else {
			_, err = io.WriteString(w, "<span>")
			_, err = io.WriteString(w, templ.EscapeString("Not test user"))
			_, err = io.WriteString(w, "</span>")
		}
		for _, v := range p.Addresses {
			err = AddressTemplate(v).Render(ctx, w)
		}
		switch p.Type {
		case "test":
			_, err = io.WriteString(w, "<span>")
			_, err = io.WriteString(w, templ.EscapeString("Test user"))
			_, err = io.WriteString(w, "</span>")
		case "admin":
			_, err = io.WriteString(w, "<span>")
			_, err = io.WriteString(w, templ.EscapeString("Admin user"))
			_, err = io.WriteString(w, "</span>")
		default:
		        _, err = io.WriteString(w, "<span>")
			_, err = io.WriteString(w, templ.EscapeString("Unknown user"))
			_, err = io.WriteString(w, "</span>")
		}
		_, err = io.WriteString(w, "</div>")
		_, err = io.WriteString(w, "</div>")
		return err
	})
}
```

# IDE Support

## vim / neovim

A vim / neovim plugin is available from https://github.com/Joe-Davidson1802/templ.vim which adds syntax highlighting.

Neovim 5 supports Language Servers directly. For the moment, I'm using https://github.com/neoclide/coc.nvim to test the language server after using Joe-Davidson1802's plugin to set the language type:

```json
{
  "languageserver": {
    "templ": {
      "command": "templ",
      "args": ["lsp"],
      "filetypes": ["templ"]
    }
}
```

To add extensive debug information, you can include additional args to the LSP, like this:

```json
{
  "languageserver": {
    "templ": {
      "command": "templ",
      "args": ["lsp",
        "--log", "/Users/adrian/github.com/a-h/templ/cmd/lspcmd/templ-log.txt", 
	"--goplsLog", "/Users/adrian/github.com/a-h/templ/cmd/lspcmd/gopls-log.txt",
	"--goplsRPCTrace", "true"
      ],
      "filetypes": ["templ"]
    }
}
```

## vscode

There's a VS Code extension, just make sure you've already installed templ and that it's on your path. 

* https://marketplace.visualstudio.com/items?itemName=a-h.templ
* https://github.com/a-h/templ-vscode

# Development

## Local builds

To build a local version you can use the `go build` tool:

```
cd cmd
go build -o templ
```

## Testing

Unit tests use the `go test` tool:

```
go test ./...
```

## Release testing

This project uses https://github.com/goreleaser/goreleaser to build the command line binary and deploy it to Github. You will need to install this to test releases.

```
make build-snapshot
```

The binaries are created by me and signed by my GPG key. You can verify with my key https://adrianhesketh.com/a-h.gpg

# Inspiration

Doesn't this look like a lot like https://github.com/valyala/quicktemplate ?

Yes, yes it does. I looked at the landscape of Go templating languages before I started writing code and my initial plan was to improve the IDE support of quicktemplate, see https://github.com/valyala/quicktemplate/issues/80

The package author didn't respond (hey, we're all busy), and looking through the code, I realised that it would be hard to modify what's there to have the concept of source mapping, mostly because there's no internal object model of the language, it reads and emits code in one go.

It's also a really feature rich project, with all sorts of formatters, and support for various languages (JSON etc.), so I borrowed some syntax ideas, but left the code. If `valyala` is up for it, I'd be happy to help integrate the ideas from here. I just want Go to have a templating language with great IDE support.

### Help needed

The project is looking for help with:

* Testing the `fmt` tool, and updating the formatter so that inline elements aren't separated onto newlines.
* Adding features to the Language Server implementation, it just does autocomplete the moment. It needs to be able to do definition and add imports automatically.
* Writing a VS Code plugin that uses the LSP support.
* Examples and testing of the tools.
* Adding a `hot` option to the compiler that recompiles the `*.templ` files when they change on disk. This could be achieved by documenting and making it easy to use external tools such as `ag`, ripgrep (`rg`) and `entr` in the short term.
* Writing documentation of the components.
* Writing a blog post that demonstrates using the tool to build a form-based Web application.
* Testing (including fuzzing), benchmarking and optimisation.
* An example of a web-based UI component library would be very useful, a more advanced version of the integration test suite, thatwould be a Go web server that runs the compiled `templ` file along with example JSON payloads that match the expected data structure types and renders the content - a UI playground. If it could do hot-reload, amazing.
* Low priority, but I'm thinking of developing a CSS-in-Go implementation to work in parallel. This might take the form of a pre-processor which would collect all "style" attributes of elements and automatically calculate a minimum set of CSS classes that could be created and applied to the elements - but a first pass could just be a way to define CSS classes in Go to allow the use of CSS variables.

Please get in touch if you're interested in building a feature as I don't want people to spend time on something that's already being worked on, or ends up being a waste of their time because it can't be integrated.
