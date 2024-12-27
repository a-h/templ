package main

import (
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/a-h/templ"
)

type GzipResponseWriter struct {
	w http.ResponseWriter
}

func (w *GzipResponseWriter) Header() http.Header {
	return w.w.Header()
}

func (w *GzipResponseWriter) Write(b []byte) (int, error) {
	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	defer gzw.Close()

	_, err := gzw.Write(b)
	if err != nil {
		return 0, err
	}
	err = gzw.Close()
	if err != nil {
		return 0, err
	}

	w.w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))

	return w.w.Write(buf.Bytes())
}

func (w *GzipResponseWriter) WriteHeader(statusCode int) {
	w.w.WriteHeader(statusCode)
}

var flagPort = flag.Int("port", 0, "Set the HTTP listen port")
var useGzip = flag.Bool("gzip", false, "Toggle gzip encoding")

func main() {
	flag.Parse()

	if *flagPort == 0 {
		fmt.Println("missing port flag")
		os.Exit(1)
	}

	var count int
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if useGzip != nil && *useGzip {
			w.Header().Set("Content-Encoding", "gzip")
			w = &GzipResponseWriter{w: w}
		}

		count++
		c := Page(count)
		h := templ.Handler(c)
		h.ErrorHandler = func(r *http.Request, err error) http.Handler {
			slog.Error("failed to render template", slog.Any("error", err))
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			})
		}
		h.ServeHTTP(w, r)
	})
	err := http.ListenAndServe(fmt.Sprintf("localhost:%d", *flagPort), nil)
	if err != nil {
		fmt.Printf("Error listening: %v\n", err)
		os.Exit(1)
	}
}
