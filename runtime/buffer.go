package runtime

import (
	"bufio"
	"io"
	"net/http"
)

// DefaultBufferSize is the default size of buffers. It is set to 4KB by default, which is the
// same as the default buffer size of bufio.Writer.
var DefaultBufferSize = 4 * 1024 // 4KB

// Buffer is a wrapper around bufio.Writer that enables flushing and closing of
// the underlying writer.
type Buffer struct {
	Underlying io.Writer
	b          *bufio.Writer
}

// Write the contents of p into the buffer.
func (b *Buffer) Write(p []byte) (n int, err error) {
	return b.b.Write(p)
}

// Flush writes any buffered data to the underlying io.Writer and
// calls the Flush method of the underlying http.Flusher if it implements it.
func (b *Buffer) Flush() error {
	if err := b.b.Flush(); err != nil {
		return err
	}
	if f, ok := b.Underlying.(http.Flusher); ok {
		f.Flush()
	}
	return nil
}

// Close closes the buffer and the underlying io.Writer if it implements io.Closer.
func (b *Buffer) Close() error {
	if c, ok := b.Underlying.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

// Reset sets the underlying io.Writer to w and resets the buffer.
func (b *Buffer) Reset(w io.Writer) {
	if b.b == nil {
		b.b = bufio.NewWriterSize(b, DefaultBufferSize)
	}
	b.Underlying = w
	b.b.Reset(w)
}

// Size returns the size of the underlying buffer in bytes.
func (b *Buffer) Size() int {
	return b.b.Size()
}

// WriteString writes the contents of s into the buffer.
func (b *Buffer) WriteString(s string) (n int, err error) {
	return b.b.WriteString(s)
}
