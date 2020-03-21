package reader

import (
	"io"
	"os"
	"path"
	"time"
)

// Reader is a simple implementation of tail -f
// It has a relatively slow read cycle of 500ms
type Reader struct {
	io.ReadCloser
}

// New returns a new reader, ready to read from
func New(name string) (*Reader, error) {
	os.MkdirAll(path.Dir(name), 0755)
	f, err := os.OpenFile(name, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return &Reader{}, err
	}
	if _, err := f.Seek(0, os.SEEK_END); err != nil {
		return &Reader{f}, err
	}
	return &Reader{f}, err
}

func (r *Reader) Read(p []byte) (n int, err error) {
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
