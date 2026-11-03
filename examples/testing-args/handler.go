package main

import (
	"errors"
	"net/http"
	"strconv"
)

type User struct {
	ID   int
	Name string
}

var getUser = func(id int) (User, error) {
	return User{}, errors.New("user not found")
}

func UserPage(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	user, err := getUser(id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	userPage(user).Render(r.Context(), w)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /users/{id}", UserPage)
	http.ListenAndServe(":8080", mux)
}
