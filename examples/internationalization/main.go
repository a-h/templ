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
	return ctxi18n.WithLocale(ctx, "en")
}

func langMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := "en" // Default language
		pathSegments := strings.Split(r.URL.Path, "/")
		if len(pathSegments) > 1 {
			lang = pathSegments[1]
		}

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
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		component := page()
		component.Render(r.Context(), w)
	})

	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	muxWithLanguages := langMiddleware(mux)

	fmt.Println("listening on :8080")
	if err := http.ListenAndServe("127.0.0.1:8080", muxWithLanguages); err != nil {
		log.Printf("error listening: %v", err)
	}
}
