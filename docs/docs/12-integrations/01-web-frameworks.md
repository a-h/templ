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

### Gorilla/csrf
gorilla/csrf is a HTTP middleware library that provides cross-site request forgery (CSRF) protection. To integrate it with templ, start by creating a csrf component:

```go
templ Csrf() {
	<input type="hidden" name="gorilla.csrf.Token" value={ ctx.Value("gorilla.csrf.Token").(string) }/>
}
```

Then ensure you pass the request's context to the Render method:
```go
r := mux.NewRouter()
r.HandleFunc("/login", func (w http.ResponseWriter, r *http.Request) {
    login := Login(csrf.Token(r))
    err := login.Render(r.Context(), w)
    if err != nil {
        //handle error
    }
})
```

Finally in your template you add:
```go
templ Login() {
    <form method="post" action="/login">
        @Csrf()
        // other fields
    </form>
}
```

## Project scaffolding

- Gowebly - https://github.com/gowebly/gowebly
- Go-blueprint - https://github.com/Melkeydev/go-blueprint
- Slick - https://github.com/anthdm/slick

## Other templates

### `template/html`

See [Using with Go templates](../syntax-and-usage/using-with-go-templates)
