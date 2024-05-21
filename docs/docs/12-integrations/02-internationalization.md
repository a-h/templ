# Internationalization

templ can be used with 3rd party internationalization libraries.

## ctxi18n

https://github.com/invopop/ctxi18n uses the context package to load strings based on the selected locale.

An example is available at https://github.com/a-h/templ/tree/main/examples/internationalization

### Storing translations

Translations are stored in YAML files, according to the language.

```yaml title="locales/en/en.yaml"
en:
  hello: "Hello"
  select_language: "Select Language"
```

### Selecting the language

HTTP middleware selects the language to load based on the URL path, `/en`, `/de`, etc.

```go title="main.go"
func newLanguageMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := "en" // Default language
		pathSegments := strings.Split(r.URL.Path, "/")
		if len(pathSegments) > 1 {
			lang = pathSegments[1]
		}
		ctx, err := ctxi18n.WithLocale(r.Context(), lang)
		if err != nil {
			log.Printf("error setting locale: %v", err)
			http.Error(w, "error setting locale", http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
```

### Using the middleware

The `ctxi18n.Load` function is used to load the translations, and the middleware is used to set the language.

```go title="main.go"
func main() {
	if err := ctxi18n.Load(locales.Content); err != nil {
		log.Fatalf("error loading locales: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", templ.Handler(page()))

	withLanguageMiddleware := newLanguageMiddleware(mux)

	log.Println("listening on :8080")
	if err := http.ListenAndServe("127.0.0.1:8080", withLanguageMiddleware); err != nil {
		log.Printf("error listening: %v", err)
	}
}
```

### Fetching translations in templates

Translations are fetched using the `i18n.T` function, passing the implicit context that's available in all templ components, and the key for the translation.

```templ
package main

import (
	"github.com/invopop/ctxi18n/i18n"
)

templ page() {
	<html>
		<head>
			<meta charset="UTF-8"/>
			<meta name="viewport" content="width=device-width, initial-scale=1.0"/>
			<title>{ i18n.T(ctx, "hello") }</title>
		</head>
		<body>
			<h1>{ i18n.T(ctx, "hello") }</h1>
			<h2>{ i18n.T(ctx, "select_language") }</h2>
			<ul>
				<li><a href="/en">English</a></li>
				<li><a href="/de">Deutsch</a></li>
				<li><a href="/zh-cn">中文</a></li>
			</ul>
		</body>
	</html>
}
```
