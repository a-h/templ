package lspcmd

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/a-h/templ/cmd/lspcmd/pls"
	"github.com/sourcegraph/jsonrpc2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Arguments struct {
	Log           string
	GoplsLog      string
	GoplsRPCTrace bool
}

func Run(args Arguments) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	defer func() {
		signal.Stop(signalChan)
		cancel()
	}()
	go func() {
		select {
		case <-signalChan: // First signal, cancel context.
			cancel()
		case <-ctx.Done():
		}
		<-signalChan // Second signal, hard exit.
		os.Exit(2)
	}()
	return run(ctx, args)
}

func run(ctx context.Context, args Arguments) (err error) {
	logger := zap.NewNop()
	if args.Log != "" {
		cfg := zap.NewProductionConfig()
		cfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
		cfg.OutputPaths = []string{
			args.Log,
		}
		logger, err = cfg.Build()
		if err != nil {
			log.Printf("failed to create logger: %v\n", err)
			os.Exit(1)
		}
	}
	defer logger.Sync()
	logger.Info("Starting up...")

	// Create the proxy.
	proxy := NewProxy(logger)

	// Create the lsp server for the text editor client.
	clientStream := jsonrpc2.NewBufferedStream(stdrwc{log: logger}, jsonrpc2.VSCodeObjectCodec{})
	// If detailed logging is required, it can be enabled with:
	// rpcLogger := jsonrpc2.LogMessages(rpcLogger{log: logger})
	// conn := jsonrpc2.NewConn(ctx, stream, handler, rpcLogger)
	client := jsonrpc2.NewConn(ctx, clientStream, proxy)

	// Start gopls and make a client connection to it.
	gopls, err := pls.NewGopls(logger, proxy.proxyFromGoplsToClient, pls.Options{
		Log:      args.GoplsLog,
		RPCTrace: args.GoplsRPCTrace,
	})
	if err != nil {
		log.Printf("failed to create gopls handler: %v\n", err)
		os.Exit(1)
	}

	// Initialize the proxy.
	proxy.Init(client, gopls)

	// Close the server and gopls client when we're complete.
	select {
	case <-ctx.Done():
		logger.Info("Signal received")
		client.Close()
		gopls.Close()
	case <-client.DisconnectNotify():
		logger.Info("Client disconnected")
	}
	logger.Info("Stopped...")
	return nil
}

type rpcLogger struct {
	log *zap.Logger
}

func (l rpcLogger) Printf(format string, v ...interface{}) {
	l.log.Sugar().Infof(format, v...)
}

type stdrwc struct {
	log *zap.Logger
}

func (s stdrwc) Read(p []byte) (int, error) {
	return os.Stdin.Read(p)
}

func (s stdrwc) Write(p []byte) (int, error) {
	return os.Stdout.Write(p)
}

func (s stdrwc) Close() error {
	s.log.Info("closing connection from LSP to editor")
	if err := os.Stdin.Close(); err != nil {
		s.log.Error("error closing stdin", zap.Error(err))
		return err
	}
	if err := os.Stdout.Close(); err != nil {
		s.log.Error("error closing stdout", zap.Error(err))
		return err
	}
	return nil
}
