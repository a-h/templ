package main

import (
	"log"
	"net/http"
)

func main() {
	http.Handle("/posts", PostHandler{})
	http.ListenAndServe(":8000", nil)
}

type PostHandler struct{}

func (ph PostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	posts := []Post{
		{
			Name:   "templ",
			Author: "author",
		},
	}
	err := postsTemplate(posts).Render(r.Context(), w)
	if err != nil {
		log.Println("error", err)
	}
}

type Post struct {
	Name   string
	Author string
}
