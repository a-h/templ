# Internationalization

While Templ does not provide internationalization out of the box, it can be achieved with available internationalization libraries such as [ctxi18n](github.com/invopop/ctxi18n).

## ctxi18n

**ctxi18n** works well as it focuses on making i18n methods accessible through the application's context. Check out an example of using [ctxi18n](github.com/invopop/ctxi18n) over [here](https://github.com/a-h/templ/tree/main/examples/internationalization).

In the example, we create a HTTP server and add a middleware function that loads the requested language based on the provided URL parameter.

```go title="main.go"
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/a-h/templ/examples/internationalization/locales"
	"github.com/invopop/ctxi18n"
)

func formDefaultLangContext(ctx context.Context) (context.Context, error) {
  // Returns with the English translation loaded in context
  return ctxi18n.WithLocale(ctx, "en")
}

func langMiddleware(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    lang := "en" // Default language
    pathSegments := strings.Split(r.URL.Path, "/")
    if len(pathSegments) > 1 {
      lang = pathSegments[1]
    }

    // Tries to load the language provided as URL params
    ctx, err := ctxi18n.WithLocale(r.Context(), lang)
    if err != nil {
      ctx, _ = formDefaultLangContext(r.Context())
    }
    
    r = r.WithContext(ctx)
    next.ServeHTTP(w, r)
  })
}

func main() {
  ctxi18n.Load(locales.Content)
  muxWithLanguages := langMiddleware(mux)
  
  // Rest of your server initialisation follows...
}
```

This allows for easy translations based on your provided locale mappings, which can be provided by creating a `locales` module. Locale mapping files themselves look like this:

```yaml title="app.yaml"
en:
  welcome_message: "Welcome to our application!"
  contact_us: "Contact Us"
  about_us: "About Us"
  name_msg: "My name is %{name}"
es:
  welcome_message: "¡Bienvenido a nuestra aplicación!"
  contact_us: "Contáctanos"
  about_us: "Sobre nosotros"
  name_msg: "Mi nombre es %{name}"
# Your other mappings followed...
```

It is important to ensure that your language mappings begin with the language that you wish to map with, followed by key-value pairs. You can also create mappings which accept parameters, such as the `name_msg` key above.

By creating mappings which are loaded within the context, you can finally retrieve your content within your templ components like this:

```templ title="components.templ"
templ contactUsButton() {
  <button class="button is-warning">{ i18n.T(ctx, "contact_us") }</button>
}

templ aboutUsButton() {
  <button class="button is-info">{ i18n.T(ctx, "about_us") }</button>
}

// Followed by your other components...
```

Check out the documentation for [ctxi18n](github.com/invopop/ctxi18n)!
