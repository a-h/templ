package lspcmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"

	"github.com/a-h/protocol"
	"github.com/a-h/templ/cmd/templ/lspcmd/httpdebug"
	"github.com/a-h/templ/cmd/templ/lspcmd/pls"
	"github.com/a-h/templ/cmd/templ/lspcmd/proxy"
	"go.lsp.dev/jsonrpc2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

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

func Run(w io.Writer, args Arguments) error {
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
	return run(ctx, w, args)
}

func run(ctx context.Context, w io.Writer, args Arguments) (err error) {
	log := zap.NewNop()
	if args.Log != "" {
		cfg := zap.NewProductionConfig()
		cfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
		cfg.OutputPaths = []string{
			args.Log,
		}
		log, err = cfg.Build()
		if err != nil {
			_, _ = fmt.Fprintf(w, "failed to create logger: %v\n", err)
			os.Exit(1)
		}
	}
	defer func() {
		_ = log.Sync()
	}()
	log.Info("lsp: starting up...")
	defer func() {
		if r := recover(); r != nil {
			log.Fatal("handled panic", zap.Any("recovered", r))
		}
	}()

	log.Info("lsp: starting gopls...")
	rwc, err := pls.NewGopls(ctx, log, pls.Options{
		Log:      args.GoplsLog,
		RPCTrace: args.GoplsRPCTrace,
	})
	if err != nil {
		log.Error("failed to start gopls", zap.Error(err))
		os.Exit(1)
	}

	cache := proxy.NewSourceMapCache()
	diagnosticCache := proxy.NewDiagnosticCache()

	log.Info("creating client")
	clientProxy, clientInit := proxy.NewClient(log, cache, diagnosticCache)
	_, goplsConn, goplsServer := protocol.NewClient(context.Background(), clientProxy, jsonrpc2.NewStream(rwc), log)
	defer goplsConn.Close()

	log.Info("creating proxy")
	// Create the proxy to sit between.
	serverProxy, serverInit := proxy.NewServer(log, goplsServer, cache, diagnosticCache)

	// Create templ server.
	log.Info("creating templ server")
	templStream := jsonrpc2.NewStream(stdrwc{log: log})
	_, templConn, templClient := protocol.NewServer(context.Background(), serverProxy, templStream, log)
	defer templConn.Close()

	// Allow both the server and the client to initiate outbound requests.
	clientInit(templClient)
	serverInit(templClient)

	// Start the web server if required.
	if args.HTTPDebug != "" {
		log.Info("starting debug http server", zap.String("addr", args.HTTPDebug))
		h := httpdebug.NewHandler(log, serverProxy)
		go func() {
			if err := http.ListenAndServe(args.HTTPDebug, h); err != nil {
				log.Error("web server failed", zap.Error(err))
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
