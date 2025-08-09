package delete

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
	case http.MethodPost:
		h.Post(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func NewModel(name string) Model {
	return Model{
		Name: name,
	}
}

type Model struct {
	Name string
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	// Read the ID from the URL.
	id := r.PathValue("id")
	if id == "" {
		http.Redirect(w, r, "/contacts", http.StatusSeeOther)
		return
	}
	// Get the existing contact from the database.
	contact, ok, err := h.DB.Get(r.Context(), id)
	if err != nil {
		h.Log.Error("Failed to get contact", slog.String("id", id), slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Redirect(w, r, "/contacts", http.StatusSeeOther)
		return
	}
	h.DisplayForm(w, r, NewModel(contact.Name))
}

func (h *Handler) DisplayForm(w http.ResponseWriter, r *http.Request, m Model) {
	layout.Handler(View(m)).ServeHTTP(w, r)
}

func (h *Handler) Post(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Redirect(w, r, "/contacts", http.StatusSeeOther)
		return
	}

	// Delete the contact from the database.
	err := h.DB.Delete(r.Context(), id)
	if err != nil {
		h.Log.Error("Failed to delete contact", slog.String("id", id), slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect back to the contact list.
	http.Redirect(w, r, "/contacts", http.StatusSeeOther)
}
