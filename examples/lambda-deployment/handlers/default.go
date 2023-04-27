package handlers

import (
	"net/http"

	"github.com/a-h/templ/examples/lambda-deployment/components"
	"github.com/a-h/templ/examples/lambda-deployment/db"
	"github.com/a-h/templ/examples/lambda-deployment/session"
	"golang.org/x/exp/slog"
)

func New(log *slog.Logger, cs *db.CountStore) *DefaultHandler {
	return &DefaultHandler{
		Log:        log,
		CountStore: cs,
	}
}

type DefaultHandler struct {
	Log        *slog.Logger
	CountStore *db.CountStore
}

func (h *DefaultHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.Post(w, r)
		return
	}
	h.Get(w, r)
}

func (h *DefaultHandler) Get(w http.ResponseWriter, r *http.Request) {
	var props ViewProps
	var err error
	if props.GlobalCount, err = h.CountStore.Get(r.Context(), "global"); err != nil {
		h.Log.Error("failed to get global count", slog.Any("error", err))
		http.Error(w, "failed to get global count", http.StatusInternalServerError)
		return
	}
	if props.SessionCount, err = h.CountStore.Get(r.Context(), session.ID(r)); err != nil {
		h.Log.Error("failed to get session count", slog.Any("error", err))
		http.Error(w, "failed to get session count", http.StatusInternalServerError)
		return
	}
	h.View(w, r, props)
}

func (h *DefaultHandler) Post(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	var props ViewProps

	// Check to see if the global button was pressed.
	var err error
	if r.Form.Has("global") {
		if props.GlobalCount, err = h.CountStore.Increment(r.Context(), "global"); err != nil {
			h.Log.Error("failed to increment global count", slog.Any("error", err))
			http.Error(w, "failed to increment global count", http.StatusInternalServerError)
			return
		}
		if props.SessionCount, err = h.CountStore.Get(r.Context(), session.ID(r)); err != nil {
			h.Log.Error("failed to get session count", slog.Any("error", err))
			http.Error(w, "failed to get session count", http.StatusInternalServerError)
			return
		}
	}
	if r.Form.Has("session") {
		if props.GlobalCount, err = h.CountStore.Get(r.Context(), "global"); err != nil {
			h.Log.Error("failed to get global count", slog.Any("error", err))
			http.Error(w, "failed to get global count", http.StatusInternalServerError)
			return
		}
		if props.SessionCount, err = h.CountStore.Increment(r.Context(), session.ID(r)); err != nil {
			h.Log.Error("failed to increment session count", slog.Any("error", err))
			http.Error(w, "failed to increment session count", http.StatusInternalServerError)
			return
		}
	}

	// Display the view.
	h.View(w, r, props)
}

type ViewProps struct {
	GlobalCount  int
	SessionCount int
}

func (h *DefaultHandler) View(w http.ResponseWriter, r *http.Request, props ViewProps) {
	components.Page(props.GlobalCount, props.SessionCount).Render(r.Context(), w)
}
