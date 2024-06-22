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

func GetBuffer(w io.Writer) (b *Buffer, existing bool) {
	b, ok := w.(*Buffer)
	if ok {
		return b, true
	}
	b = bufferPool.Get().(*Buffer)
	b.Reset(w)
	return b, false
}

var DefaultBufferSize = 4 * 1024 // 4KB
var MaxBufferSize = 64 * 1024    // 64KB

func ReleaseBuffer(w io.Writer) (err error) {
	b, ok := w.(*Buffer)
	if !ok {
		return nil
	}
	err = b.Flush()
	if b.Size() > MaxBufferSize {
		b.Reset(nil)
	}
	bufferPool.Put(b)
	return err
}
