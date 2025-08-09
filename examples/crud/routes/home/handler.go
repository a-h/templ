package home

import (
	"net/http"

	"github.com/a-h/templ/examples/crud/layout"
)

func NewHandler() http.Handler {
	return &Handler{}
}

type Handler struct {
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.Get(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	v := layout.Handler(View())
	v.ServeHTTP(w, r)
}
