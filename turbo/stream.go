package turbo

import (
	"context"
	"net/http"
	"strings"

	"github.com/a-h/templ"
)

// Append adds an append action to the output stream.
func Append(w http.ResponseWriter, target string, template templ.Component) error {
	return AppendWithContext(context.Background(), w, target, template)
}

// AppendWithContext adds an append action to the output stream.
func AppendWithContext(ctx context.Context, w http.ResponseWriter, target string, template templ.Component) error {
	w.Header().Set("Content-Type", "text/vnd.turbo-stream.html")
	return actionTemplate("append", target).Render(templ.WithChildren(ctx, template), w)
}

// Prepend adds a prepend action to the output stream.
func Prepend(w http.ResponseWriter, target string, template templ.Component) error {
	return PrependWithContext(context.Background(), w, target, template)
}

// PrependWithContext adds a prepend action to the output stream.
func PrependWithContext(ctx context.Context, w http.ResponseWriter, target string, template templ.Component) error {
	w.Header().Set("Content-Type", "text/vnd.turbo-stream.html")
	return actionTemplate("prepend", target).Render(templ.WithChildren(ctx, template), w)
}

// Replace adds a replace action to the output stream.
func Replace(w http.ResponseWriter, target string, template templ.Component) error {
	return ReplaceWithContext(context.Background(), w, target, template)
}

// ReplaceWithContext adds a replace action to the output stream.
func ReplaceWithContext(ctx context.Context, w http.ResponseWriter, target string, template templ.Component) error {
	w.Header().Set("Content-Type", "text/vnd.turbo-stream.html")
	return actionTemplate("replace", target).Render(templ.WithChildren(ctx, template), w)
}

// Update adds an update action to the output stream.
func Update(w http.ResponseWriter, target string, template templ.Component) error {
	return UpdateWithContext(context.Background(), w, target, template)
}

// UpdateWithContext adds an update action to the output stream.
func UpdateWithContext(ctx context.Context, w http.ResponseWriter, target string, template templ.Component) error {
	w.Header().Set("Content-Type", "text/vnd.turbo-stream.html")
	return actionTemplate("update", target).Render(templ.WithChildren(ctx, template), w)
}

// Remove adds a remove action to the output stream.
func Remove(w http.ResponseWriter, target string) error {
	return RemoveWithContext(context.Background(), w, target)
}

// RemoveWithContext adds a remove action to the output stream.
func RemoveWithContext(ctx context.Context, w http.ResponseWriter, target string) error {
	w.Header().Set("Content-Type", "text/vnd.turbo-stream.html")
	return removeTemplate("remove", target).Render(ctx, w)
}

// IsTurboRequest returns true if the incoming request is able to receive a Turbo stream.
// This is determined by checking the request header for "text/vnd.turbo-stream.html"
func IsTurboRequest(r *http.Request) bool {
	return strings.Contains(r.Header.Get("accept"), "text/vnd.turbo-stream.html")
}
