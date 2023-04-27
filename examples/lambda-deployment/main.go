package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/a-h/templ/examples/lambda-deployment/db"
	"github.com/a-h/templ/examples/lambda-deployment/handlers"
	"github.com/a-h/templ/examples/lambda-deployment/session"
	"golang.org/x/exp/slog"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout))
	cs, err := db.NewCountStore(os.Getenv("TABLE_NAME"), os.Getenv("AWS_REGION"))
	if err != nil {
		log.Error("failed to create store", slog.Any("error", err))
		os.Exit(1)
	}
	h := handlers.New(log, cs)

	var secureFlag bool
	if os.Getenv("SECURE_FLAG") == "false" {
		secureFlag = false
	}

	// Add session middleware.
	sh := session.NewMiddleware(h, session.WithSecure(secureFlag))

	server := &http.Server{
		Addr:         "localhost:9000",
		Handler:      sh,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}

	fmt.Printf("Listening on %v\n", server.Addr)
	server.ListenAndServe()
}
