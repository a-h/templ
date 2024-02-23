package main

import (
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/a-h/templ"
)

type GzipResponseWriter struct {
	http.ResponseWriter
}

func (w *GzipResponseWriter) Header() Header {
	return w.Header()
}

func (w *GzipResponseWriter) Write(b []byte) (int, error) {
	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	defer gzw.Close()

	_, err = gzw.Write(b)
	if err != nil {
		return err
	}
	err = gzw.Close()
	if err != nil {
		return err
	}

	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	return w.Write(gzw)
}

func (w *GzipResponseWriter) WriteHeader(statusCode int) (int, error) {
	return w.Write(statusCode)
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
		if useGzip {
			w.Header().Set("Content-Encoding", "gzip")
			w = GzipResponseWriter{w}
		}

		count++
		c := Page(count)
		templ.Handler(c).ServeHTTP(w, r)
	})
	err := http.ListenAndServe(fmt.Sprintf("localhost:%d", *flagPort), nil)
	if err != nil {
		fmt.Printf("Error listening: %v\n", err)
		os.Exit(1)
	}
}
