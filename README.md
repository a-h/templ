# templ

* A strongly typed HTML templating language that compiles to Go code, and has great developer tooling.

![vscode-autocomplete](https://user-images.githubusercontent.com/1029947/120372693-66b51000-c30f-11eb-8924-41a65616f620.gif)

## Getting started

* Install the `templ` command-line tool: `go install github.com/a-h/templ/cmd/templ@latest`
* Create a `*.templ` file containing a template.
* Run `templ generate` to create Go code from the template.

## Current state

This is beta software, and the template language may still have breaking changes. There's no guarantees of stability or correctness at the moment, but it has at least one production user.

If you're keen to see Go be practical for Web projects, see "Help needed" for where the project needs your help.

## Features

The language generates Go code, some sections of the template (e.g. `package`, `import`, `if`, `for` and `switch` statements) are output directly as Go expressions in the generated output, while HTML elements are converted to Go code that renders their output.

* `templ generate` generates Go code from `*.templ` files.
* `templ fmt` formats template files in the current directory tree.
* `templ lsp` provides a Language Server to support IDE integrations. The compile command generates a sourcemap which maps from the `*.templ` files to the compiled Go file. This enables the `templ` LSP to use the Go language `gopls` language server as is, providing a thin shim to do the source remapping. This is used to provide autocomplete for template variables and functions.

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

All elements must be balanced (have a start and and end tag, or be self-closing).

```
<div id="address1">{%= addr.Address1 %}</div>
```

You can also have dynamic attributes that use template parameters, other Go variables that happen to be in scope, or call Go functions that return a string. Don't worry about HTML encoding element text and attribute values, that will be taken care of automatically.

```
<a title={%= p.TitleText %}>{%= strings.ToUpper(p.Name()) %}</a>
```

Boolean attributes (see https://html.spec.whatwg.org/multipage/common-microsyntaxes.html#boolean-attributes) where the presence of an attribute name without a value means `true`, and the attribute name not being present means false are supported:

With constant values:

```
<hr noshade/>
```

To set boolean attributes using variables or template parameters, a question mark after the attribute name is used to denote that the attribute is boolean. In this example, the `noshade` attribute would be omitted from the output altogether:

```
<hr noshade?={%= false %} />
```

The `a` element's `href` attribute is treated differently. Templ expects you to provide a `templ.SafeURL`. A `templ.SafeURL` is a URL that is definitely safe to use (i.e. has come from a configuration system controlled by the developer), or has been through a sanitization process to filter out potential XSS attacks.

Templ provides a `templ.URL` function that sanitizes input URLs and checks that the protocol is http/https/mailto rather than `javascript` or another unexpected protocol.

```
<a href={%= templ.URL(p.URL) %}>{%= strings.ToUpper(p.Name()) %}</a>
```

### Text

Text is rendered from HTML included in the template itself, or by using Go expressions. No processing or conversion is applied to HTML included within the template, whereas Go string expressions are HTML encoded on output.

Plain HTML:

```html
<div>Plain HTML is allowed.</div>
```

Constant Go expressions:

```
<div>{%= "this is a string" %}</div>
```

The backtick constant expression (single-line only):

```
<div>{%= `this is also a string` %}</div>
```

Functions that return a string:

```
<div>{%= time.Now().String() %}</div>
```

A string parameter, or variable that's in scope:

```
<div>{%= v.s %}</div>
```

### CSS

Templ components can have CSS associated with them. CSS classes are created with the `css` template expression. CSS properties can be set to string variables or functions (e.g. `{%= red %}`). However, functions should be idempotent - i.e. return the same value every time.

All variable CSS values are passed through a value sanitizer to provide some protection against malicious data being added to CSS.

```
{% css className() %}
	background-color: #ffffff;
	color: {%= red %};
{% endcss %}
```

CSS class components can be used within templates.

```
{% templ Button(text string) %}
	<button class={%= templ.Classes(className(), templ.Class("other")) %} type="button">{%= text %}</button>
{% endtempl %}
```

The first time that the component is rendered in a HTTP request, it will render the CSS class to the output. The next time the same component is rendered, templ will skip rendering the CSS to the output because it is no longer required.

For example, if this template is rendered in a request:

```
{% templ TwoButtons() %}
	{%! Button("A") %}
	{%! Button("B") %}
{% endtempl %}
```

The output would contain one class. Note that the name contains a unique value is addition to the class name to reduce the likelihood of clashes. Don't rely on this name being consistent.

```html
<style type="text/css">.className_f179{background-color:#ffffff;color:#ff0000;}</style>
<button class="className_f179 other" type="button">A</button>
<button class="className_f179 other" type="button">B</button>`
```

#### CSS Middleware

If you want to provide a global stylesheet that includes this CSS to remove `<style>` tags from the output, you can use templ's CSS middleware, and register templ classes.

The middleware adds a HTTP route to the web server (`/styles/templ.css` by default) that renders the `text/css` classes that would otherwise be added to `<style>` tags when components are rendered. It's then your responsibility to add a `<link rel="stylesheet" href="/styles/templ.css">` to your HTML.

For example, to stop the `className` CSS class from being added to the output, the HTTP middleware can be used.

```go
c1 := className()
handler := NewCSSMiddleware(httpRoutes, c1)
http.ListenAndServe(":8000:, handler)
```

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
        "--log", "/Users/adrian/github.com/a-h/templ/cmd/templ/lspcmd/templ-log.txt", 
	"--goplsLog", "/Users/adrian/github.com/a-h/templ/cmd/templ/lspcmd/gopls-log.txt",
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
cd cmd/templ
go build
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

## Hot reload

For hot reload, you can use https://github.com/cosmtrek/air

For documentation on how to use it with templ see https://adrianhesketh.com/2021/05/28/templ-hot-reload-with-air/

### Help needed

The project is looking for help with:

* Adding features to the Language Server implementation, it just does autocomplete and error reporting the moment. It needs to be able to do definition and add imports automatically.
* Examples and testing of the tools.
* Writing a blog post that demonstrates using the tool to build a form-based Web application.
* Testing (including fuzzing), benchmarking and optimisation.
* An example of a web-based UI component library would be very useful, a more advanced version of the integration test suite, thatwould be a Go web server that runs the compiled `templ` file along with example JSON payloads that match the expected data structure types and renders the content - a UI playground. If it could do hot-reload, amazing.
* Low priority, but I'm thinking of developing a CSS-in-Go implementation to work in parallel. This might take the form of a pre-processor which would collect all "style" attributes of elements and automatically calculate a minimum set of CSS classes that could be created and applied to the elements - but a first pass could just be a way to define CSS classes in Go to allow the use of CSS variables.

Please get in touch if you're interested in building a feature as I don't want people to spend time on something that's already being worked on, or ends up being a waste of their time because it can't be integrated.

# Writing and examples

* https://adrianhesketh.com/2021/05/18/introducing-templ/
* https://adrianhesketh.com/2021/05/28/templ-hot-reload-with-air/
* https://adrianhesketh.com/2021/06/04/hotwired-go-with-templ/

## Security

templ will automatically escape content according to the following rules.

```
{% templ Example() %}
  <script type="text/javascript">
    {%= "will be HTML encoded using templ.Escape, which isn't JavaScript-aware, don't use templ to build scripts" %}
  </script>
  <div onClick={%= "will be HTML encoded using templ.Escape, but this isn't JavaScript aware, don't use user-controlled data here" %}>
    {%= "will be HTML encoded using templ.Escape" %}</div>  
  </div>
  <style type="text/css">
    {%= "will be escaped using templ.Escape, which isn't CSS-aware, don't use user-controlled data here" %}
  </style>
  <div style={%= "will be HTML encoded using templ.Escape, which isn't CSS-aware, don't use user controlled data here" %}</div>
  <div class={%= templ.CSSClasses(templ.Class("will not be escaped, because it's expected to be a constant value")) %}</div>
  <div>{%= "will be escaped using templ.Escape" %}</div>
  <a href="http://constants.example.com/are/not/sanitized">Text</a>
  <a href={%= templ.URL("will be sanitized by templ.URL to remove potential attacks") %}</div>
  <a href={%= templ.SafeURL("will not be sanitized by templ.URL") %}</div>
{% endtempl %}
```

CSS property names, and constant CSS property values are not sanitized or escaped.

```
{% css className() %}
	background-color: #ffffff;
{% endcss %}
```

CSS property values based on expressions are passed through `templ.SanitizeCSS` to replace potentially unsafe values with placeholders.

```
{% css className() %}
	color: {%= red %};
{% endcss %}
```

