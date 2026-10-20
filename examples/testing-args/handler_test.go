package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUserHandlerPassesCorrectUser(t *testing.T) {
	getUser = func(id int) (User, error) {
		return User{ID: id, Name: "Alice"}, nil
	}

	argsMap := make(map[string]any)
	ctx := context.WithValue(context.Background(), "_templ_args_map", argsMap)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /users/{id}", UserPage)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/users/42", nil).WithContext(ctx)

	mux.ServeHTTP(w, r)

	u, ok := argsMap["u"].(User)
	if !ok {
		t.Fatal("template was not called or argument u was not captured")
	}
	if u.Name != "Alice" {
		t.Errorf("expected user name Alice, got %q", u.Name)
	}
}
