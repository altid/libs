package fs

import (
	"os"
)

// WriteCloser is a type that implements io.WriteCloser
type WriteCloser struct {
	c      runner
	fp     *os.File
	buffer string
}

func (w *WriteCloser) Write(b []byte) (n int, err error) {
	return w.fp.Write(b)
}

func (w *WriteCloser) Close() error {
	w.c.event(w.buffer)
	return w.fp.Close()
}
