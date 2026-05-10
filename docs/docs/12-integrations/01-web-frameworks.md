# Web frameworks

Templ is framework agnostic but that does not mean it can not be used with Go frameworks and other tools.

Below are some examples of how to use templ with other Go libraries, frameworks and tools, and links to systems that have built-in templ support.

### Chi

See an example of using https://github.com/go-chi/chi with templ at:

https://github.com/a-h/templ/tree/main/examples/integration-chi

### Echo

See an example of using https://echo.labstack.com/ with templ at:

https://github.com/a-h/templ/tree/main/examples/integration-echo

### Gin

See an example of using https://github.com/gin-gonic/gin with templ at:

https://github.com/a-h/templ/tree/main/examples/integration-gin

### Go Fiber

See an example of using https://github.com/gofiber/fiber with templ at:

https://github.com/a-h/templ/tree/main/examples/integration-gofiber

### CSRF Protection

Go 1.25 includes built-in cross-site request forgery (CSRF) protection via `http.CrossOriginProtection`.

```go title="main.go"
package main

import (
  "net/http"
  "log"
)

func main() {
  r := http.NewServeMux()
  r.Handle("/", templ.Handler(Form()))

  // Configure CSRF protection with trusted origins.
  csrfProtection := http.NewCrossOriginProtection()
  if err := csrfProtection.AddTrustedOrigin("http://localhost:8000"); err != nil {
    log.Fatalf("failed to add trusted origin: %v", err)
  }

  http.ListenAndServe("localhost:8000", csrfProtection.Handler(r))
}
```

The built-in protection uses modern browser security headers (Sec-Fetch-Site) and does not require hidden form fields or tokens in your HTML:

```templ title="form.templ"
templ Form() {
  <h1>CSRF Example</h1>
  <form method="post" action="/">
    <div>
      This form is protected by the built-in CSRF protection which uses the Sec-Fetch-Site header.
    </div>
    <input type="submit" value="Submit"/>
  </form>
}
```

For applications requiring Go 1.24 or earlier, you can use the `github.com/gorilla/csrf` library instead.

## Project scaffolding

- Gowebly - https://github.com/gowebly/gowebly
- Go-blueprint - https://github.com/Melkeydev/go-blueprint
- Slick - https://github.com/anthdm/slick

## Other templates

### `template/html`

See [Using with Go templates](../syntax-and-usage/using-with-go-templates)
