# Content security policy

## Nonces

In templ [script templates](/syntax-and-usage/script-templates#script-templates) are rendered as inline `<script>` tags. This means by default they are not compatible with a strict CSP.

Nonces can be used to allow templ to bypass a strict CSP. They are randomly generated string that should be re-generated on every request.

The `templ.WithNonce` function can be used to provide a nonce for templ to use when rendering a script.

```templ title="templates.templ"
package main

import "context"
import "os"

script onLoad() {
    alert("Hello, world!")
}

templ template() {
    @onLoad()
}

func main() {
    nonce := generateSecurelyRandomString()
    ctx := templ.WithNonce(context.Background(), nonce)
    if err := template().Render(ctx, os.Stdout); err != nil {
        panic(err)
    }
}
```
```go title="main.go"
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	mux := http.NewServeMux()

	// Handle template.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        nonce := generateSecurelyRandomString()
        rw.Header().Set("Content-Security-Policy", fmt.Sprintf("script-src: 'nonce-%s'", nonce))
        ctx := templ.WithNonce(context.Background(), nonce)
        if err := template().Render(ctx, os.Stdout); err != nil {
            http.Error(w, "failed to render", 500)
        }
	})

	// Start the server.
	fmt.Println("listening on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Printf("error listening: %v", err)
	}
}

```

This would render the following:

```html
<script type="text/javascript" nonce="..randomly generated nonce">function __templ_onLoad_5a85(){alert("Hello, world!")}</script>
<script type="text/javascript" nonce="..randomly generated nonce">__templ_onLoad_5a85()</script>
```

## Nonce Middleware

Generate and apply nonces in a middleware to remove repeated code in handlers:

```go title="main.go"
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func nonceMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nonce := generateSecurelyRandomString()
		ctx := templ.WithNonce(context.Background(), nonce)
		w.Header().Add("Content-Security-Policy", fmt.Sprintf("script-src 'nonce-%s'", nonce))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func main() {
	mux := http.NewServeMux()

	// Handle template.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := template().Render(ctx, os.Stdout); err != nil {
			http.Error(w, "failed to render", 500)
		}
	})

	// Start the server.
	fmt.Println("listening on :8080")
	// Apply middlewares.
	mux = nonceMiddleware(mux)
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Printf("error listening: %v", err)
	}
}
```
