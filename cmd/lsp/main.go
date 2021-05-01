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
	"go.uber.org/zap/zapcore"
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
	log *zap.Logger
}

func (l rpcLogger) Printf(format string, v ...interface{}) {
	l.log.Info(fmt.Sprintf(format, v...))
}

func run(ctx context.Context, args []string) error {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
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
	handler, err := NewProxy(logger)
	if err != nil {
		log.Printf("failed to create gopls handler: %v\n", err)
		os.Exit(1)
	}
	stream := jsonrpc2.NewBufferedStream(stdrwc{log: logger}, jsonrpc2.VSCodeObjectCodec{})
	//rpcLogger := jsonrpc2.LogMessages(rpcLogger{log: logger})
	conn := jsonrpc2.NewConn(ctx, stream, handler) //, rpcLogger)
	handler.client = conn
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
