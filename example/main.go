package main

import (
	"fmt"
	"net/http"

	"github.com/a-h/templ"
)

func main() {
	// Use a template that doesn't take parameters.
	http.Handle("/", templ.Handler(home()))

	// Use a template that accesses data or handles form posts.
	http.Handle("/posts", PostHandler{})

	// Start the server.
	fmt.Println("listening on http://localhost:8000")
	http.ListenAndServe("localhost:8000", nil)
}

type PostHandler struct{}

func (ph PostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get the posts from a database.
	postsToDisplay := []Post{{Name: "templ", Author: "author"}}

	// Render the template.
	templ.Handler(posts(postsToDisplay)).ServeHTTP(w, r)
}

type Post struct {
	Name   string
	Author string
}
