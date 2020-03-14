package fs

import "io"

// WriteCloser is a type that implements io.WriteCloser
type WriteCloser struct {
	c      runner
	fp     io.WriteCloser
	buffer string
}

func (w *WriteCloser) Write(b []byte) (n int, err error) {
	return w.fp.Write(b)
}

// Close - A Closer which sends an event 
func (w *WriteCloser) Close() error {
	w.c.event(w.buffer)
	return w.fp.Close()
}
