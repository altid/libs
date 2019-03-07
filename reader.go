package fslib

import (
	"io"
	"os"
	"path"
	"time"
)

type reader struct {
	io.ReadCloser
}

func newReader(name string) (*reader, error) {
	os.MkdirAll(path.Dir(name), 0755)
	f, err := os.OpenFile(name, os.O_CREATE|os.O_RDONLY, 0755)
	if err != nil {
		return &reader{}, err
	}
	if _, err := f.Seek(0, os.SEEK_END); err != nil {
		return &reader{f}, err
	}
	return &reader{f}, err
}

func (r *reader) Read(p []byte) (n int, err error) {
	for {
		n, err := r.ReadCloser.Read(p)
		if n > 0 {
			return n, nil
		} else if err != io.EOF {
			return n, err
		}
		time.Sleep(300 * time.Millisecond)
	}
}
