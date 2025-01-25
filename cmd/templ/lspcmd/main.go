package lspcmd

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/a-h/templ/cmd/templ/lspcmd/httpdebug"
	"github.com/a-h/templ/cmd/templ/lspcmd/pls"
	"github.com/a-h/templ/cmd/templ/lspcmd/proxy"
	"github.com/a-h/templ/lsp/jsonrpc2"
	"github.com/a-h/templ/lsp/protocol"

	_ "net/http/pprof"
)

type Arguments struct {
	Log           string
	GoplsLog      string
	GoplsRPCTrace bool
	// PPROF sets whether to start a profiling server on localhost:9999
	PPROF bool
	// HTTPDebug sets the HTTP endpoint to listen on. Leave empty for no web debug.
	HTTPDebug string
}

func Run(stdin io.Reader, stdout, stderr io.Writer, args Arguments) (err error) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	defer func() {
		signal.Stop(signalChan)
		cancel()
	}()
	if args.PPROF {
		go func() {
			_ = http.ListenAndServe("localhost:9999", nil)
		}()
	}
	go func() {
		select {
		case <-signalChan: // First signal, cancel context.
			cancel()
		case <-ctx.Done():
		}
		<-signalChan // Second signal, hard exit.
		os.Exit(2)
	}()
	log := slog.New(slog.NewJSONHandler(io.Discard, nil))
	if args.Log != "" {
		file, err := os.OpenFile(args.Log, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		defer file.Close()

		// Create a new logger with a file writer
		log = slog.New(slog.NewJSONHandler(file, nil))
		log.Debug("Logging to file", slog.String("file", args.Log))
	}
	templStream := jsonrpc2.NewStream(newStdRwc(log, "templStream", stdout, stdin))
	return run(ctx, log, templStream, args)
}

func run(ctx context.Context, log *slog.Logger, templStream jsonrpc2.Stream, args Arguments) (err error) {
	log.Info("lsp: starting up...")
	defer func() {
		if r := recover(); r != nil {
			log.Error("handled panic", slog.Any("recovered", r))
		}
	}()

	log.Info("lsp: starting gopls...")
	rwc, err := pls.NewGopls(ctx, log, pls.Options{
		Log:      args.GoplsLog,
		RPCTrace: args.GoplsRPCTrace,
	})
	if err != nil {
		log.Error("failed to start gopls", slog.Any("error", err))
		os.Exit(1)
	}

	cache := proxy.NewSourceMapCache()
	diagnosticCache := proxy.NewDiagnosticCache()

	log.Info("creating gopls client")
	clientProxy, clientInit := proxy.NewClient(log, cache, diagnosticCache)
	_, goplsConn, goplsServer := protocol.NewClient(ctx, clientProxy, jsonrpc2.NewStream(rwc), log)
	defer goplsConn.Close()

	log.Info("creating proxy")
	// Create the proxy to sit between.
	serverProxy := proxy.NewServer(log, goplsServer, cache, diagnosticCache)

	// Create templ server.
	log.Info("creating templ server")
	_, templConn, templClient := protocol.NewServer(context.Background(), serverProxy, templStream, log)
	defer templConn.Close()

	// Allow both the server and the client to initiate outbound requests.
	clientInit(templClient)

	// Start the web server if required.
	if args.HTTPDebug != "" {
		log.Info("starting debug http server", slog.String("addr", args.HTTPDebug))
		h := httpdebug.NewHandler(log, serverProxy)
		go func() {
			if err := http.ListenAndServe(args.HTTPDebug, h); err != nil {
				log.Error("web server failed", slog.Any("error", err))
			}
		}()
	}

	log.Info("listening")

	select {
	case <-ctx.Done():
		log.Info("context closed")
	case <-templConn.Done():
		log.Info("templConn closed")
	case <-goplsConn.Done():
		log.Info("goplsConn closed")
	}
	log.Info("shutdown complete")
	return
}
