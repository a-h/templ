package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/a-h/kv/sqlitekv"
	"github.com/a-h/templ/examples/crud/db"
	"github.com/a-h/templ/examples/crud/routes/contacts"
	contactsdelete "github.com/a-h/templ/examples/crud/routes/contacts/delete"
	contactsedit "github.com/a-h/templ/examples/crud/routes/contacts/edit"
	"github.com/a-h/templ/examples/crud/routes/home"
	"zombiezen.com/go/sqlite/sqlitex"
)

var dbURI = "file:data.db?mode=rwc"
var addr = "localhost:8080"

func main() {
	log := slog.Default()
	ctx := context.Background()
	if err := run(ctx, log); err != nil {
		log.Error("Failed to run server", slog.Any("error", err))
		os.Exit(1)
	}
}

func run(ctx context.Context, log *slog.Logger) error {
	pool, err := sqlitex.NewPool(dbURI, sqlitex.PoolOptions{})
	if err != nil {
		log.Error("Failed to open database", slog.Any("error", err))
		return err
	}
	store := sqlitekv.New(pool)
	if err := store.Init(ctx); err != nil {
		log.Error("Failed to initialize store", slog.Any("error", err))
		return err
	}
	db := db.New(store)

	mux := http.NewServeMux()

	homeHandler := home.NewHandler()
	mux.Handle("/", homeHandler)

	ch := contacts.NewHandler(log, db)
	mux.Handle("/contacts", ch)

	ceh := contactsedit.NewHandler(log, db)
	mux.Handle("/contacts/edit", ceh)
	mux.Handle("/contacts/edit/{id}", ceh)

	cdh := contactsdelete.NewHandler(log, db)
	mux.Handle("/contacts/delete/{id}", cdh)

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Info("Starting server", slog.String("address", addr))
	return http.ListenAndServe(addr, mux)
}
