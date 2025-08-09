package contacts

import (
	"log/slog"
	"net/http"

	"github.com/a-h/templ/examples/crud/db"
	"github.com/a-h/templ/examples/crud/layout"
)

func NewHandler(log *slog.Logger, db *db.DB) http.Handler {
	return &Handler{
		Log: log,
		DB:  db,
	}
}

type Handler struct {
	Log *slog.Logger
	DB  *db.DB
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
	contacts, err := h.DB.List(r.Context())
	if err != nil {
		h.Log.Error("Failed to list contacts", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	v := layout.Handler(View(contacts))
	v.ServeHTTP(w, r)
}
