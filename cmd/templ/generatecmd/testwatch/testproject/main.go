package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/a-h/templ"
)

var flagPort = flag.Int("port", 0, "Set the HTTP listen port")

func main() {
	flag.Parse()

	if *flagPort == 0 {
		fmt.Println("missing port flag")
		os.Exit(1)
	}

	var count int
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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
