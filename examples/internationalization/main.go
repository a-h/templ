package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/a-h/templ/examples/internationalization/locales"
	"github.com/invopop/ctxi18n"
)

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
