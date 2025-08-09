# Using with `html/template`

Templ components can be used with the Go standard library [`html/template`](https://pkg.go.dev/html/template) package.

## Using `html/template` in a templ component

To use an existing `html/template` in a templ component, use the `templ.FromGoHTML` function.

```templ title="component.templ"
package testgotemplates

import "html/template"

var goTemplate = template.Must(template.New("example").Parse("<div>{{ . }}</div>"))

templ Example() {
	<!DOCTYPE html>
	<html>
		<body>
			@templ.FromGoHTML(goTemplate, "Hello, World!")
		</body>
	</html>
}
```

```go title="main.go"
func main() {
	Example.Render(context.Background(), os.Stdout)
}
```

```html title="Output"
<!DOCTYPE html>
<html>
	<body>
		<div>Hello, World!</div>
	</body>
</html>
```

## Using a templ component with	`html/template`

To use a templ component within a `html/template`, use the `templ.ToGoHTML` function to render the component into a `template.HTML value`.

```templ title="component.html"
package testgotemplates

import "html/template"

var example = template.Must(template.New("example").Parse(`<!DOCTYPE html>
<html>
	<body>
		{{ . }}
	</body>
</html>
`))

templ greeting() {
	<div>Hello, World!</div>
}
```

```go title="main.go"
func main() {
	// Create the templ component.
	templComponent := greeting()

	// Render the templ component to a `template.HTML` value.
	html, err := templ.ToGoHTML(context.Background(), templComponent)
	if err != nil {
		t.Fatalf("failed to convert to html: %v", err)
	}

	// Use the `template.HTML` value within the text/html template.
	err = example.Execute(os.Stdout, html)
	if err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}
}
```

```html title="Output"
<!DOCTYPE html>
<html>
	<body>
		<div>Hello, World!</div>
	</body>
</html>
```
