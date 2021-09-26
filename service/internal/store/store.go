package store

type WriteCloser struct {

}

// TODO(halfwit) New doesn't need to take a WriteCloser anymore, just a path 
func New(closeFn func(string) error, buffer, filename string) *WriteCloser {
	// We can open our log file on file + document, return handles to that
	// Otherwise just make the file an actual bytes array
}

func (w *WriteCloser) Write(b []byte) (n int, err error) {
	return w.fp.Write(b)
}

func (w *WriteCloser) Close() error {
	defer w.fp.Close()
	if e := w.closeFn(w.buffer); e != nil {
		return e
	}

	return nil
}
