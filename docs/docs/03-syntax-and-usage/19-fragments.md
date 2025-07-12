# Fragments

The `templ.Fragment` component can be used to render a subsection of a template, discarding all other output.

Fragments work well as an optimisation for HTMX, as discussed in https://htmx.org/essays/template-fragments/

## Define fragments

Define a fragment with `@templ.Fragment("name")`, where `"name"` is the identifier for the fragment.

```templ
templ Page() {
  <div>Page Header</div>
  @templ.Fragment("name") {
    <div>Content of the fragment</div>
  }
}
```

To avoid name clashes with other libraries, you can define a custom type for your package.

```templ
type nameFragmentKey struct {}
var Name = nameFragmentKey{}

templ Page() {
  <div>Page Header</div>
  @templ.Fragment(Name) {
    <div>Content of the fragment</div>
  }
}
```

## Use with HTTP

The most common use case for `Fragment` is to render only a specific part of the template to the HTML response, while discarding the rest of the output.

To render only the "name" fragment from the `Page` template, use the `templ.WithFragments("name")` option when creating the HTTP handler:

```go title="main.go"
handler := templ.Handler(Page(), templ.WithFragments("name"))
http.Handle("/", handler)
```

When the HTTP request is made, only the content of the specified fragment will be returned in the response:

```html title="output.html"
<div>Content of the fragment</div>
```

:::note
The whole of the template is rendered, so any function calls or logic in the template will still be executed, but only the specified fragment's output is sent to the client.
:::

If the `templ.WithFragments("name")` option is omitted, the whole page is rendered as normal.

```go title="main.go"
handler := templ.Handler(Page())
http.Handle("/", handler)
```

```html title="output.html"
<div>Page Header</div>
<div>Content of the fragment</div>
```

## Use outside of an HTTP handler

To use outside of an HTTP handler, e.g. when generating static content, you can render fragments with the `templ.RenderFragments` function.

```go
w := new(bytes.Buffer)
if err := templ.RenderFragments(context.Background(), w, fragmentPage, "name"); err != nil {
  t.Fatalf("failed to render: %v", err)
}

// <div>Content of the fragment</div>
html := w.String()
```

:::note
All fragments with matching identifiers will be rendered. If the fragment identifier isn't matched, no output will be produced.
:::

## Nested fragments

Fragments can be nested, allowing for complex structures to be defined and rendered selectively.

Given this example templ file:

```templ
templ Page() {
	@templ.Fragment("outer") {
		<div>Outer Fragment Start</div>
		@templ.Fragment("inner") {
			<div>Inner Fragment Content</div>
		}
		<div>Outer Fragment End</div>
	}
}
```

If the `outer` fragment is selected for rendering, then the `inner` fragment is also rendered.

## HTMX example

```templ title="main.templ"
package main

import (
  "fmt"
  "net/http"
  "strconv"
)

type PageState struct {
  Counter int
  Next    int
}

templ Page(state PageState) {
  <html>
    <head>
       <script src="https://cdn.jsdelivr.net/npm/htmx.org@2.0.6/dist/htmx.min.js"></script>
       <link rel="stylesheet" href="https://unpkg.com/missing.css@1.1.3/dist/missing.min.css"/>
    </head>
    <body>
      @templ.Fragment("buttonOnly") {
        <button hx-get={ fmt.Sprintf("/?counter=%d&template=buttonOnly", state.Next) } hx-swap="outerHTML">
          This Button Has Been Clicked { state.Counter } Times
        </button>
      }
    </body>
  </html>
}

// handleRequest does the work to execute the template (or fragment) and serve the result.
// It's mostly boilerplate, so don't get hung up on it.
func handleRequest(w http.ResponseWriter, r *http.Request) {
  // Collect state info to pass to the template.
  var state PageState
  state.Counter, _ = strconv.Atoi(r.URL.Query().Get("counter"))
  state.Next = state.Counter + 1

  // If the template querystring paramater is set, render the pecific fragment.
  var opts []func(*templ.ComponentHandler)
  if templateName := r.URL.Query().Get("template"); templateName != "" {
    opts = append(opts, templ.WithFragments(templateName))
  }

  // Render the template or fragment and serve it.
  templ.Handler(Page(state), opts...).ServeHTTP(w, r)
}

func main() {
  // Handle the template.
  http.HandleFunc("/", handleRequest)
  
  // Start the server.
  fmt.Println("Server is running at http://localhost:8080")
  http.ListenAndServe("localhost:8080", nil)
}
```

:::note
This was adapted from `benpate`'s Go stdlib example at https://gist.github.com/benpate/f92b77ea9b3a8503541eb4b9eb515d8a
:::
