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
	err := renderPosts(r.Context(), w, posts)
	if err != nil {
		log.Println("error", err)
	}
}

type Post struct {
	Name   string
	Author string
}
