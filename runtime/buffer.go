package runtime

import (
	"bufio"
	"io"
	"net/http"
)

type Buffer struct {
	Underlying io.Writer
	b          *bufio.Writer
}

func (b *Buffer) Write(p []byte) (n int, err error) {
	return b.b.Write(p)
}

func (b *Buffer) Flush() error {
	if err := b.b.Flush(); err != nil {
		return err
	}
	if f, ok := b.Underlying.(http.Flusher); ok {
		f.Flush()
	}
	return nil
}

func (b *Buffer) Close() error {
	if c, ok := b.Underlying.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

func (b *Buffer) Reset(w io.Writer) {
	if b.b == nil {
		b.b = bufio.NewWriterSize(b, DefaultBufferSize)
	}
	b.Underlying = w
	b.b.Reset(w)
}

func (b *Buffer) Size() int {
	return b.b.Size()
}

func (b *Buffer) WriteString(s string) (n int, err error) {
	return b.b.WriteString(s)
}
