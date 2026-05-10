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
var origins = []string{
	"http://localhost:8080",
	"http://127.0.0.1:8080",
	"http://localhost:7331",
	"http://127.0.0.1:7331",
}

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
	store := sqlitekv.NewStore(pool)
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

	// Manage CSRF protection.
	csrfProtection := http.NewCrossOriginProtection()
	for _, origin := range origins {
		if err := csrfProtection.AddTrustedOrigin(origin); err != nil {
			log.Error("Failed to add CSRF trusted origin", slog.String("origin", origin), slog.Any("error", err))
			return err
		}
	}

	log.Info("Starting server", slog.String("address", addr))
	return http.ListenAndServe(addr, csrfProtection.Handler(mux))
}
