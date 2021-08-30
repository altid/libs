package reader

import (
	"io"
	"os"
	"path"
	"time"
)

type ReadCloser struct {
	closeFn func(string) error
	fp	io.ReadCloser
	buffer	string
}

func New(closeFn func(string) error, fp io.ReadCloser, buffer string) *ReadCloser {
	return &ReadCloser{
		closeFn: closeFn,
		fp:	 fp,
		buffer:  buffer,
	}
}

func (r *ReadCloser) Read(p []byte) (n int, err error) {
	return r.fp.Read(b)
}

func (r *ReadCloser) Close() error {
	defer r.fp.Close()
	if e := r.closeFn(r.buffer); e != nil {
		return e
	}

	return nil
}

// Poller is a simple implementation of tail -f
// It has a relatively slow read cycle of 500ms
type Poller struct {
	io.ReadCloser
}

// New returns a new reader, ready to read from
func Poll(name string) (*Poller, error) {
	os.MkdirAll(path.Dir(name), 0755)
	f, err := os.OpenFile(name, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return &Poller{}, err
	}
	if _, err := f.Seek(0, os.SEEK_END); err != nil {
		return &Poller{f}, err
	}
	return &Poller{f}, err
}

func (r *Poller) Read(p []byte) (n int, err error) {
	for {
		n, err := r.ReadCloser.Read(p)
		if n > 0 {
			return n, nil
		} else if err != io.EOF {
			return n, err
		}
		time.Sleep(500 * time.Millisecond)
	}
}
