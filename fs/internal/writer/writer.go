package writer

import "io"

// WriteCloser is a type that implements io.WriteCloser
// And calls
type WriteCloser struct {
	closeFn func(string) error
	fp      io.WriteCloser
	buffer  string
}

// New returns a WriteCloser that will call closeFn during Close()
func New(closeFn func(string) error, fp io.WriteCloser, buffer string) *WriteCloser {
	return &WriteCloser{
		closeFn: closeFn,
		fp:      fp,
		buffer:  buffer,
	}
}
func (w *WriteCloser) Write(b []byte) (n int, err error) {
	return w.fp.Write(b)
}

// Close - A Closer which sends an event
func (w *WriteCloser) Close() error {
	if e := w.closeFn(w.buffer); e != nil {
		return e
	}

	return w.fp.Close()
}
