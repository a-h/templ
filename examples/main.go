package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/a-h/templ"
)

func main() {
	// Use a template that doesn't take parameters.
	http.Handle("/", templ.Handler(home()))

	// Use a template that accesses data or handles form posts.
	http.Handle("/posts", NewPostsHandler())

	// Start the server.
	fmt.Println("listening on http://localhost:8000")
	if err := http.ListenAndServe("localhost:8000", nil); err != nil {
		log.Printf("error listening: %v", err)
	}
}

func NewPostsHandler() PostsHandler {
	// Replace this in-memory function with a call to a database.
	postsGetter := func() (posts []Post, err error) {
		return []Post{{Name: "templ", Author: "author"}}, nil
	}
	return PostsHandler{
		GetPosts: postsGetter,
		Log:      log.Default(),
	}
}

type PostsHandler struct {
	Log      *log.Logger
	GetPosts func() ([]Post, error)
}

func (ph PostsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ps, err := ph.GetPosts()
	if err != nil {
		ph.Log.Printf("failed to get posts: %v", err)
		http.Error(w, "failed to retrieve posts", http.StatusInternalServerError)
		return
	}
	templ.Handler(posts(ps)).ServeHTTP(w, r)
}

type Post struct {
	Name   string
	Author string
}
