package main

import (
	"fmt"
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
		attributeData := components.Data{
			Message: fmt.Sprintf("Hello, from the attribute data"),
		}
		scriptData := components.Data{
			Message: fmt.Sprintf("Hello, from the script data"),
		}
		templ.Handler(components.Page(attributeData, scriptData)).ServeHTTP(w, r)
	})

	fmt.Println("Listening on http://localhost:8080")
	http.ListenAndServe("localhost:8080", mux)
}
