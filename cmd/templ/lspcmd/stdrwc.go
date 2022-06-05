package lspcmd

import (
	"os"

	"go.uber.org/zap"
)

// stdrwc (standard read/write closer) reads from stdin, and writes to stdout.
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
