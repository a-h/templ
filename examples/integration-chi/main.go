package main

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()
	r.Get("/", templ.Handler(Home()).ServeHTTP)
	http.ListenAndServe(":3000", r)
}
