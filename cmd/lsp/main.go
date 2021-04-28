package lsp

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"

	"github.com/sourcegraph/jsonrpc2"
	"go.uber.org/zap"
)

const (
	exitCodeInterrupt = 2
)

func Run(args []string, stdout io.Writer) error {
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
		case <-signalChan: // first signal, cancel context
			cancel()
		case <-ctx.Done():
		}
		<-signalChan // second signal, hard exit
		os.Exit(exitCodeInterrupt)
	}()
	return run(ctx, args)
}

type rpcLogger struct {
	zapLogger *zap.Logger
}

func (l rpcLogger) Printf(format string, v ...interface{}) {
	l.zapLogger.Info(fmt.Sprintf(format, v...))
}

func run(ctx context.Context, args []string) error {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{
		"/Users/adrian/github.com/a-h/templ/cmd/lsp/log.txt",
	}
	logger, err := cfg.Build()
	if err != nil {
		log.Printf("failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()
	logger.Info("Starting up...")
	handler := NewHandler(logger)

	stream := jsonrpc2.NewBufferedStream(stdrwc{}, jsonrpc2.VSCodeObjectCodec{})
	rpcLogger := jsonrpc2.LogMessages(rpcLogger{zapLogger: logger})
	conn := jsonrpc2.NewConn(ctx, stream, handler, rpcLogger)
	select {
	case <-ctx.Done():
		logger.Info("Signal received")
		conn.Close()
	case <-conn.DisconnectNotify():
		logger.Info("Client disconnected")
	}

	logger.Info("Stopped...")
	return nil
}

type stdrwc struct{}

func (stdrwc) Read(p []byte) (int, error) {
	return os.Stdin.Read(p)
}

func (stdrwc) Write(p []byte) (int, error) {
	return os.Stdout.Write(p)
}

func (stdrwc) Close() error {
	if err := os.Stdin.Close(); err != nil {
		return err
	}
	return os.Stdout.Close()
}
