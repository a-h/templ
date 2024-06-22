package runtime

import (
	"io"
	"sync"
)

var bufferPool = sync.Pool{
	New: func() any {
		return new(Buffer)
	},
}

// GetBuffer creates and returns a new buffer if the writer is not already a buffer,
// or returns the existing buffer if it is.
func GetBuffer(w io.Writer) (b *Buffer, existing bool) {
	if w == nil {
		return nil, false
	}
	b, ok := w.(*Buffer)
	if ok {
		return b, true
	}
	b = bufferPool.Get().(*Buffer)
	b.Reset(w)
	return b, false
}

// ReleaseBuffer flushes the buffer and returns it to the pool.
func ReleaseBuffer(w io.Writer) (err error) {
	b, ok := w.(*Buffer)
	if !ok {
		return nil
	}
	err = b.Flush()
	bufferPool.Put(b)
	return err
}
