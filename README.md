![templ](https://github.com/a-h/templ/raw/main/templ.png)

## A HTML templating language for Go that has great developer tooling.

![templ](https://user-images.githubusercontent.com/1029947/171962961-38aec64d-eac3-4166-8cb6-e7337c907bae.gif)

## Getting started

* Install the `templ` command-line tool: `go install github.com/a-h/templ/cmd/templ@latest`
* Initialize a new Go project with `go mod`, e.g. `go mod init example`.
* Create the `example.templ` and `main.go` files shown below.
* Run `templ generate` followed by `go run *.go` to create Go code from the template and run the web server.

### example.templ

```html
package main

import "fmt"
import "time"

templ headerTemplate(name string) {
	<header data-testid="headerTemplate">
		<h1>{ name }</h1>
	</header>
}

templ footerTemplate() {
	<footer data-testid="footerTemplate">
		<div>&copy; { fmt.Sprintf("%d", time.Now().Year()) }</div>
	</footer>
}

templ navTemplate() {
	<nav data-testid="navTemplate">
		<ul>
			<li><a href="/">Home</a></li>
			<li><a href="/posts">Posts</a></li>
		</ul>
	</nav>
}

templ layout(name string, content templ.Component) {
	<html>
		<head><title>{ name }</title></head>
		<body>
			{! headerTemplate(name) }
			{! navTemplate() }
			<main>
				{! content }
			</main>
		</body>
		{! footerTemplate() }
	</html>
}

templ homeTemplate() {
	<div data-testid="homeTemplate">Welcome to my website.</div>
}

templ postsTemplate(posts []Post) {
	<div data-testid="postsTemplate">
		for _, p := range posts {
			<div data-testid="postsTemplatePost">
				<div data-testid="postsTemplatePostName">{ p.Name }</div>
				<div data-testid="postsTemplatePostAuthor">{ p.Author }</div>
			</div>
		}
	</div>
}

templ home() {
	{! layout("Home", homeTemplate()) }
}

templ posts(posts []Post) {
	{! layout("Posts", postsTemplate(posts)) }
}
```

### main.go

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/a-h/templ"
)

func main() {
	// Use a template that doesn't take parameters.
	http.Handle("/", templ.Handler(home()))

	// Use a template that accesses data or handles form posts.
	http.Handle("/posts", PostHandler{})

	// Start the server.
	fmt.Println("listening on http://localhost:8000")
	http.ListenAndServe("localhost:8000", nil)
}

type PostHandler struct{}

func (ph PostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get the posts from a database.
	postsToDisplay := []Post{{Name: "templ", Author: "author"}}

	// Render the template.
	templ.Handler(posts(postsToDisplay)).ServeHTTP(w, r)
}

type Post struct {
	Name   string
	Author string
}
```

## Current state

This is beta software, and the template language may still have breaking changes. There's no guarantees of stability or correctness at the moment, but it has at least one production user.

If you're keen to see Go be practical for Web projects, see "Help needed" for where the project needs your help.

## Features

The language generates Go code, some sections of the template (e.g. `package`, `import`, `if`, `for` and `switch` statements) are output directly as Go expressions in the generated output, while HTML elements are converted to Go code that renders their output.

* `templ generate` generates Go code from `*.templ` files.
* `templ fmt` formats template files (`templ fmt .` for everything in the current directory and subdirectories, `templ fmt` to format stdin and output to stdout.)
* `templ lsp` provides a Language Server to support IDE integrations. The compile command generates a sourcemap which maps from the `*.templ` files to the compiled Go file. This enables the `templ` LSP to use the Go language `gopls` language server as is, providing a thin shim to do the source remapping. This is used to provide autocomplete for template variables and functions.
* Storybook support, see https://adrianhesketh.com/2021/10/23/using-storybook-with-go-frontends/

## Template files

Template files end with a `.templ` extension and combine Go code with HTML-like expressions.

### Package

Since `templ` files are as close to Go as possible, they start with a package expression.

```go
package templ
```

### Importing packages

After the package expression, you can import other Go packages, just like Go files.

```go
import "strings"
```

### Adding functions

Outside of the `templ` statement, you can use any Go code you like.

### Components

Once the package and import statements are done, we can start a component using the `templ Name(params Params) {` expression. The `templ` expressions are converted into Go functions when the `templ generate` command is executed.

```html
templ AddressView(addr Address) {
	<div>{ addr.Address1 }</div>
	<div>{ addr.Address2 }</div>
	<div>{ addr.Address3 }</div>
	<div>{ addr.Address4 }</div>
}
```

Each `templ.Component` can contain HTML elements, strings, for loops, switch statements and references to other templates.

#### Referencing other components

Components can be referenced in the body of the template, and can pass data between then, for example, using the `AddressTemplate` from the `PersonTemplate`.

```html
templ PersonTemplate(p Person) {
	<div>
	    for _, v := range p.Addresses {
		    {! AddressTemplate(v) }
	    }
	</div>
}
```

It's also possible to create "higher order components" that compose other instances of `templ.Component` without passing data, or even knowing what the concrete type of the component will be ahead of time. So long as is implements `templ.Component`, it can be used.

For example, this template accepts 3 templates (header, footer, body) and renders all 3 of them in the expected order.

```html
templ Layout(header, footer, body templ.Component) {
	{! header }
	{! body }
	{! footer }
}
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

```html
{! Raw("<script>alert('xss vector');</script>") }
```

For larger scripts you want to embed, you should create a code component that writes the constant to the output writer using the embed feature of Go - see https://pkg.go.dev/embed for more information.

```go
func EmbeddedScript(s string) Component {
	return ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		_, err = io.WriteString(w, "<script>")
		if err != nil {
			return
		}
		//go:embed script.js
		var b []byte
		_, err = w.Write(b)
		if err != nil {
			return
		}
		_, err = io.WriteString(w, "</script>")
		return
	})
}
```

### Elements

HTML elements look like HTML and you can write static attributes into them, just like with normal HTML. Don't worry about the spacing, the HTML will be minified when it's rendered.

All elements must be balanced (have a start and and end tag, or be self-closing).

```html
<div id="address1">{ addr.Address1 }</div>
```

You can also have dynamic attributes that use template parameters, other Go variables that happen to be in scope, or call Go functions that return a string. Don't worry about HTML encoding element text and attribute values, that will be taken care of automatically.

```html
<a title={ p.TitleText }>{ strings.ToUpper(p.Name()) }</a>
```

Boolean attributes (see https://html.spec.whatwg.org/multipage/common-microsyntaxes.html#boolean-attributes) where the presence of an attribute name without a value means `true`, and the attribute name not being present means false are supported:

With constant values:

```html
<hr noshade/>
```

To set boolean attributes using variables or template parameters, a question mark after the attribute name is used to denote that the attribute is boolean. In this example, the `noshade` attribute would be omitted from the output altogether:

```html
<hr noshade?={ false } />
```

The `a` element's `href` attribute is treated differently. Templ expects you to provide a `templ.SafeURL`. A `templ.SafeURL` is a URL that is definitely safe to use (i.e. has come from a configuration system controlled by the developer), or has been through a sanitization process to filter out potential XSS attacks.

Templ provides a `templ.URL` function that sanitizes input URLs and checks that the protocol is http/https/mailto rather than `javascript` or another unexpected protocol.

```html
<a href={ templ.URL(p.URL) }>{ strings.ToUpper(p.Name()) }</a>
```

### Text

Text is rendered from HTML included in the template itself, or by using Go expressions. No processing or conversion is applied to HTML included within the template, whereas Go string expressions are HTML encoded on output.

Plain HTML:

```html
<div>Plain HTML is allowed.</div>
```

Constant Go expressions:

```html
<div>{ "this is a string" }</div>
```

The backtick constant expression:

```html
<div>{ `this is also a string` }</div>
```

Functions that return a string:

```html
<div>{ time.Now().String() }</div>
```

A string parameter, or variable that's in scope:

```html
<div>{ v.s }</div>
```

templ will look for Go code. If, for some reason, you need start a sentence with `for`, `switch` or another Go statement, you can use `<>` and `</>` to encapsulate raw HTML.

```html
<div>
	<>
	for x := 0; x < 100; x ++ {
	}
	</>
</div>
```

### onClick etc. handlers

`onClick` and other `on*` handlers have special behaviour, they expect a reference to a `script` template.

```go
package testscriptusage

script withParameters(a string, b string, c int) {
	console.log(a, b, c);
}

script withoutParameters() {
	alert("hello");
}

templ Button(text string) {
	<button onClick={ withParameters("test", text, 123) } onMouseover={ withoutParameters() } type="button">{ text }</button>
}
```

Rendering the button with `A` as the text input, would render the following HTML. Note that the function names are modified to reduce the likelihood of namespace collisions.

```html
<script type="text/javascript">function __templ_withParameters_rnd(a, b, c){console.log(a, b, c);}function __templ_withoutParameters_rnd(){alert("hello");}</script>
<button onClick="__templ_withParameters_rnd(&#34;test&#34;,&#34;A&#34;,123)" onMouseover="__templ_withoutParameters_rnd()" type="button">A</button>
```

### CSS

Templ components can have CSS associated with them. CSS classes are created with the `css` template expression. CSS properties can be set to string variables or functions (e.g. `{ red }`). However, functions should be idempotent - i.e. return the same value every time.

All variable CSS values are passed through a value sanitizer to provide some protection against malicious data being added to CSS.

```css
css className() {
	background-color: #ffffff;
	color: { red };
}
```

CSS class components can be used within templates.

```html
templ Button(text string) {
	<button class={ templ.Classes(className(), templ.Class("other")) } type="button">{ text }</button>
}
```

The first time that the component is rendered in a HTTP request, it will render the CSS class to the output. The next time the same component is rendered, templ will skip rendering the CSS to the output because it is no longer required.

For example, if this template is rendered in a request:

```html
templ TwoButtons() {
	{! Button("A") }
	{! Button("B") }
}
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

```html
if p.Type == "test" {
	<span>{ "Test user" }</span>
} else {
	<span>{ "Not test user" }</span>
}
```

### For

Templates have the same loop behaviour as Go.

```html
for _, v := range p.Addresses {
	<li>{ v.City }</li>
}
```

### Switch/Case

Switch statements work in the same way as they do in Go. 

```html
switch p.Type {
	case "test":
		<span>{ "Test user" }</span>
	case "admin"
		<span>{ "Admin user" }</span>
	default:
		<span>{ "Unknown user" }</span>
}
```

## Full example

```html
package templ

import "strings"

templ Layout(header, footer, body templ.Component) {
	{! header }
	{! body }
	{! footer }
}

templ AddressTemplate(addr Address) {
	<div>{ addr.Address1 }</div>
	<div>{ addr.Address2 }</div>
	<div>{ addr.Address3 }</div>
	<div>{ addr.Address4 }</div>
}

templ PersonTemplate(p Person) {
	<div>
		<div>{ p.Name() }</div>
		<a href={ p.URL }>{ strings.ToUpper(p.Name()) }</a>
		<div>
			if p.Type == "test" {
				<span>{ "Test user" }</span>
			} else {
				<span>{ "Not test user" }</span>
			}
			for _, v := range p.Addresses {
				{! AddressTemplate(v) }
			}
			switch p.Type {
				case "test":
					<span>{ "Test user" }</span>
				case "admin":
					<span>{ "Admin user" }</span>
				default:
					<span>{ "Unknown user" }</span>
			}
		</div>
	</div>
}
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

## vscode

There's a VS Code extension, just make sure you've already installed templ and that it's on your path. 

* https://marketplace.visualstudio.com/items?itemName=a-h.templ
* https://github.com/a-h/templ-vscode

## Neovim 5

A vim / neovim plugin is available from https://github.com/Joe-Davidson1802/templ.vim which adds syntax highlighting.

To enable the built-in Language Server support of Neovim 5.x add the following code to your `.vimrc` prior to calling `setup` on the language servers, e.g.:

```lua
-- Add templ configuration.
local configs = require'lspconfig/configs'
if not nvim_lsp.templ then
  configs.templ = {
    default_config = {
      cmd = {"templ", "lsp"},
      filetypes = {'templ'},
      root_dir = nvim_lsp.util.root_pattern("go.mod", ".git"),
      settings = {},
    };
  }
end

-- Use a loop to conveniently call 'setup' on multiple servers and
-- map buffer local keybindings when the language server attaches
local servers = { 'gopls', 'ccls', 'cmake', 'tsserver', 'templ' }
for _, lsp in ipairs(servers) do
  nvim_lsp[lsp].setup {
    on_attach = on_attach,
    flags = {
      debounce_text_changes = 150,
    },
  }
end
```

## vim / neovim 4.x

A vim / neovim plugin is available from https://github.com/Joe-Davidson1802/templ.vim which adds syntax highlighting.

https://github.com/neoclide/coc.nvim can be used to run the language server after using Joe-Davidson1802's plugin to set the language type:

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

## Tasks

### nix-develop

Run a Nix shell that contains everything required to build templ.

```sh
nix develop --impure
```

### build

Build a local version.

```sh
cd cmd/templ
go build
```

### install-snapshot

Build and install to ~/bin

```sh
rm cmd/templ/lspcmd/*.txt || true
cd cmd/templ && go build -o ~/bin/templ
```

### build-snapshot

Use goreleaser to build the command line binary using goreleaser.

```sh
goreleaser build --snapshot --rm-dist
```

### generate

Run templ generate using local version.

```sh
go run ./cmd/templ generate
```

### test

Run Go tests.

```sh
go run ./cmd/templ generate && go test ./...
```

### test-cover

Run Go tests.

```sh
# Create test profile directories.
mkdir -p coverage/generate
mkdir -p coverage/unit
# Build the test binary.
go build -cover -o ./coverage/templ-cover ./cmd/templ
# Run the covered generate command.
GOCOVERDIR=coverage/generate ./coverage/templ-cover generate
# Run the unit tests.
go test -cover ./... -args -test.gocoverdir="$PWD/coverage/unit"
# Display the combined percentage.
go tool covdata percent -i=./coverage/generate,./coverage/unit
# Generate a text coverage profile for tooling to use.
go tool covdata textfmt -i=./coverage/generate,./coverage/unit -o coverage.out
```

### lint

```sh
docker run --rm -v $(pwd):/app -w /app golangci/golangci-lint:v1.52.2 golangci-lint run -v
```

### release

Create production build with goreleaser.

```sh
if [ "${GITHUB_TOKEN}" == "" ]; then echo "No github token, run:"; echo "export GITHUB_TOKEN=`pass github.com/goreleaser_access_token`"; exit 1; fi
./push-tag.sh
goreleaser --clean
```

### docs-run

Run the development server.

Directory: docs

```
npm run start
```

### docs-build

Build production docs site.

Directory: docs

```
npm run build
```

### docker-build

Build a Docker container with a full development environment and Neovim setup for testing the LSP.

```
docker build -t templ:latest .
```

### docker-run

Run a Docker development container in the current directory.

```
docker run -p 7474:7474 -v `pwd`:/templ -it --rm templ:latest
```

# Code signing

The binaries are created by me and signed by my GPG key. You can verify with my key https://adrianhesketh.com/a-h.gpg

## Hot reload

For hot reload, you can use https://github.com/cosmtrek/air

For documentation on how to use it with templ see https://adrianhesketh.com/2021/05/28/templ-hot-reload-with-air/

# Writing and examples

* https://adrianhesketh.com/2021/05/18/introducing-templ/
* https://adrianhesketh.com/2021/05/28/templ-hot-reload-with-air/
* https://adrianhesketh.com/2021/06/04/hotwired-go-with-templ/
* https://adrianhesketh.com/2021/10/17/testing-templ-html-rendering-with-goquery/
* https://adrianhesketh.com/2021/10/23/using-storybook-with-go-frontends/

## Security

templ is designed to prevent user provided data from being used to inject vulnerabilities.

`<script>` and `<style>` tags could allow user data to inject vulnerabilities, so variables are not permitted in these sections.

```html
templ Example() {
  <script type="text/javascript">
    function showAlert() {
      alert("hello");
    }
  </script>
  <style type="text/css">
    /* Only CSS is allowed */
  </style>
}
```

`onClick` attributes, and other `on*` attributes are used to execute JavaScript. To prevent user data from being unescapted, `on*` attributes accept a `templ.ComponentScript`.

```html
script onClickHandler(msg string) {
  alert(msg);
}

templ Example(msg string) {
  <div onClick={ onClickHandler(msg) }>
    { "will be HTML encoded using templ.Escape" }
  </div>
}
```

Style attributes cannot be expressions, only constants, to avoid escaping vulnerabilities. templ style templates (`css className()`) should be used instead.

```html
templ Example() {
  <div style={ "will throw an error" }</div>
}
```

Class names are escaped unless bypassed.

```html
templ Example() {
  <div class={ templ.CSSClasses(templ.Class("unsafe</style&gt;-will-sanitized"), templ.SafeClass("sanitization bypassed")) }</div>
}
```

```html
templ Example() {
  <div>Node text is not modified at all.</div>
  <div>{ "will be escaped using templ.Escape" }</div>
}
```

`href` attributes must be a `templ.SafeURL` and are sanitized to remove JavaScript URLs unless bypassed.

```html
templ Example() {
  <a href="http://constants.example.com/are/not/sanitized">Text</a>
  <a href={ templ.URL("will be sanitized by templ.URL to remove potential attacks") }</a>
  <a href={ templ.SafeURL("will not be sanitized by templ.URL") }</a>
}
```

Within css blocks, property names, and constant CSS property values are not sanitized or escaped.

```css
css className() {
	background-color: #ffffff;
}
```

CSS property values based on expressions are passed through `templ.SanitizeCSS` to replace potentially unsafe values with placeholders.

```css
css className() {
	color: { red };
}
```
