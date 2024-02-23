package sloghandler

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"sync"

	"github.com/fatih/color"
)

var _ slog.Handler = &Handler{}

type Handler struct {
	h slog.Handler
	m *sync.Mutex
	w io.Writer
}

var levelToIcon = map[slog.Level]string{
	slog.LevelDebug: "(✓)",
	slog.LevelInfo:  "(✓)",
	slog.LevelWarn:  "(!)",
	slog.LevelError: "(✗)",
}
var levelToColor = map[slog.Level]*color.Color{
	slog.LevelDebug: color.New(color.FgCyan),
	slog.LevelInfo:  color.New(color.FgGreen),
	slog.LevelWarn:  color.New(color.FgYellow),
	slog.LevelError: color.New(color.FgRed),
}

func NewHandler(w io.Writer, opts *slog.HandlerOptions) *Handler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &Handler{
		w: w,
		h: slog.NewTextHandler(w, &slog.HandlerOptions{
			Level:     opts.Level,
			AddSource: opts.AddSource,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if opts.ReplaceAttr != nil {
					a = opts.ReplaceAttr(groups, a)
				}
				if a.Key == slog.LevelKey {
					level, ok := levelToIcon[a.Value.Any().(slog.Level)]
					if !ok {
						level = a.Value.Any().(slog.Level).String()
					}
					a.Value = slog.StringValue(level)
					return a
				}
				if a.Key == slog.TimeKey {
					return slog.Attr{}
				}
				return a
			},
		}),
		m: &sync.Mutex{},
	}
}

func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.h.Enabled(ctx, level)
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Handler{h: h.h.WithAttrs(attrs), w: h.w, m: h.m}
}

func (h *Handler) WithGroup(name string) slog.Handler {
	return &Handler{h: h.h.WithGroup(name), w: h.w, m: h.m}
}

var keyValueColor = color.New(color.Faint & color.FgBlack)

func (h *Handler) Handle(ctx context.Context, r slog.Record) (err error) {
	var sb strings.Builder

	sb.WriteString(levelToColor[r.Level].Sprint(levelToIcon[r.Level]))
	sb.WriteString(" ")
	sb.WriteString(r.Message)

	if r.NumAttrs() != 0 {
		sb.WriteString(" [")
		r.Attrs(func(a slog.Attr) bool {
			sb.WriteString(keyValueColor.Sprintf(" %s=%s", a.Key, a.Value.String()))
			return true
		})
		sb.WriteString(" ]")
	}

	sb.WriteString("\n")

	h.m.Lock()
	defer h.m.Unlock()
	_, err = io.WriteString(h.w, sb.String())
	return err
}
