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

### github.com/gorilla/csrf

`gorilla/csrf` is a HTTP middleware library that provides cross-site request forgery (CSRF) protection.

Follow the instructions at https://github.com/gorilla/csrf to add it to your project, by using the library as HTTP middleware.

```go title="main.go"
package main

import (
  "crypto/rand"
  "fmt"
  "net/http"
  "github.com/gorilla/csrf"
)

func mustGenerateCSRFKey() (key []byte) {
  key = make([]byte, 32)
  n, err := rand.Read(key)
  if err != nil {
    panic(err)
  }
  if n != 32 {
    panic("unable to read 32 bytes for CSRF key")
  }
  return
}

func main() {
  r := http.NewServeMux()
  r.Handle("/", templ.Handler(Form()))

  csrfMiddleware := csrf.Protect(mustGenerateCSRFKey())
  withCSRFProtection := csrfMiddleware(r)

  fmt.Println("Listening on localhost:8000")
  http.ListenAndServe("localhost:8000", withCSRFProtection)
}
```

Creating a `CSRF` templ component makes it easy to include the CSRF token in your forms.

```templ title="form.templ"
templ Form() {
  <h1>CSRF Example</h1>
  <form method="post" action="/">
    @CSRF()
    <div>
      If you inspect the HTML form, you will see a hidden field with the value: { ctx.Value("gorilla.csrf.Token").(string) }
    </div>
    <input type="submit" value="Submit with CSRF token"/>
  </form>
  <form method="post" action="/">
    <div>
      You can also submit the form without the CSRF token to validate that the CSRF protection is working.
    </div>
    <input type="submit" value="Submit without CSRF token"/>
  </form>
}

templ CSRF() {
  <input type="hidden" name="gorilla.csrf.Token" value={ ctx.Value("gorilla.csrf.Token").(string) }/>
}
```

## Project scaffolding

- Gowebly - https://github.com/gowebly/gowebly
- Go-blueprint - https://github.com/Melkeydev/go-blueprint
- Slick - https://github.com/anthdm/slick

## Other templates

### `template/html`

See [Using with Go templates](../syntax-and-usage/using-with-go-templates)
