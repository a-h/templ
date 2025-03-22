package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"os"

	"log/slog"

	"github.com/a-h/templ"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	// Create HTTP routes.
	mux := http.NewServeMux()
	mux.Handle("/", templ.Handler(template()))

	// Wrap the router with CSP middleware to apply the CSP nonce to templ scripts.
	withCSPMiddleware := NewCSPMiddleware(log, mux)

	log.Info("Listening...", slog.String("addr", "127.0.0.1:7001"))
	if err := http.ListenAndServe("127.0.0.1:7001", withCSPMiddleware); err != nil {
		log.Error("failed to start server", slog.Any("error", err))
	}
}

func NewCSPMiddleware(log *slog.Logger, next http.Handler) *CSPMiddleware {
	return &CSPMiddleware{
		Log:  log,
		Next: next,
		Size: 28,
	}
}

type CSPMiddleware struct {
	Log  *slog.Logger
	Next http.Handler
	Size int
}

func (m *CSPMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	nonce, err := m.generateNonce()
	if err != nil {
		m.Log.Error("failed to generate nonce", slog.Any("error", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	ctx := templ.WithNonce(r.Context(), nonce)
	w.Header().Add("Content-Security-Policy", fmt.Sprintf("script-src 'nonce-%s'", nonce))
	m.Next.ServeHTTP(w, r.WithContext(ctx))
}

func (m *CSPMiddleware) generateNonce() (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	ret := make([]byte, m.Size)
	for i := range m.Size {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}
	return string(ret), nil
}
