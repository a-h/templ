package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"

	"github.com/a-h/templ"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/basic", func(w http.ResponseWriter, r *http.Request) {
		nonce, err := generateRandomString(28)
		if err != nil {
			// ...
		}
		ctx := templ.WithNonce(r.Context(), nonce)
		w.Header().Add("Content-Security-Policy", fmt.Sprintf("script-src 'nonce-%s'", nonce))
		err = template().Render(ctx, w)
		if err != nil {
			// ...
		}
	})

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := template().Render(r.Context(), w)
		if err != nil {
			// ...
		}
	})
	mux.Handle("/middleware", scriptNonceMiddleware(h))

	http.ListenAndServe("127.0.0.1:7000", mux)
}

func scriptNonceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nonce, err := generateRandomString(28)
		if err != nil {
			// ...
		}
		ctx := templ.WithNonce(r.Context(), nonce)
		w.Header().Add("Content-Security-Policy", fmt.Sprintf("script-src 'nonce-%s'", nonce))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func generateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret), nil
}
