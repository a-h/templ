package main

import (
	"fmt"
	"math/rand"
	"net/http"

	"github.com/a-h/templ"
	"github.com/a-h/templ/examples/typescript/components"
)

func main() {
	mux := http.NewServeMux()
	// Serve the JS bundle.
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	// Serve the page.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Create random server-side data.
		d := components.Data{
			Message: fmt.Sprintf("Hello, world! %d", rand.Intn(100)),
			Value:   42,
		}
		templ.Handler(components.Page(d)).ServeHTTP(w, r)
	})

	http.ListenAndServe("localhost:8080", mux)
}
