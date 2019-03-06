package fslib

import (
	"io"
	"os"
	"path"
	"time"
)

type reader struct {
	*os.File
	//io.ReadCloser
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
	var curr int64
	stat, _ := r.File.Stat()
	for {
		n, err := r.File.Read(p)
		curr += int64(n)
		if n > 0 {
			return n, nil
		} else if err != io.EOF {
			return n, err
		}
		if curr > stat.Size() {
			curr = 0
			r.File.Seek(0, os.SEEK_SET)
			return copy([]byte("\n"), p), nil
		}
		time.Sleep(300 * time.Millisecond)
	}
}
