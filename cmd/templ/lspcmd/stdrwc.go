package lspcmd

import (
	"errors"
	"io"
	"log/slog"
)

// stdrwc (standard read/write closer) reads from stdin, and writes to stdout.
func newStdRwc(log *slog.Logger, name string, w io.Writer, r io.Reader) stdrwc {
	return stdrwc{
		log:  log,
		name: name,
		w:    w,
		r:    r,
	}
}

type stdrwc struct {
	log  *slog.Logger
	name string
	w    io.Writer
	r    io.Reader
}

func (s stdrwc) Read(p []byte) (int, error) {
	return s.r.Read(p)
}

func (s stdrwc) Write(p []byte) (int, error) {
	return s.w.Write(p)
}

func (s stdrwc) Close() error {
	s.log.Info("rwc: closing", slog.String("name", s.name))
	var errs []error
	if closer, isCloser := s.r.(io.Closer); isCloser {
		if err := closer.Close(); err != nil {
			s.log.Error("rwc: error closing reader", slog.String("name", s.name), slog.Any("error", err))
			errs = append(errs, err)
		}
	}
	if closer, isCloser := s.w.(io.Closer); isCloser {
		if err := closer.Close(); err != nil {
			s.log.Error("rwc: error closing writer", slog.String("name", s.name), slog.Any("error", err))
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
