package gintemplrenderer

import (
	"context"
	"net/http"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin/render"
)

var Default = &Renderer{}

func New(ctx context.Context, status int, component templ.Component) *Renderer {
	return &Renderer{
		Ctx:       ctx,
		Status:    status,
		Component: component,
	}
}

type Renderer struct {
	Ctx       context.Context
	Status    int
	Component templ.Component
}

func (t Renderer) Render(w http.ResponseWriter) error {
	t.WriteContentType(w)
	w.WriteHeader(t.Status)
	if t.Component != nil {
		return t.Component.Render(t.Ctx, w)
	}
	return nil
}

func (t Renderer) WriteContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
}

func (t *Renderer) Instance(name string, data any) render.Render {
	templData, ok := data.(templ.Component)
	if !ok {
		return nil
	}
	return &Renderer{
		Ctx:       context.Background(),
		Status:    http.StatusOK,
		Component: templData,
	}
}
