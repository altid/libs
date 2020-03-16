package reader

import (
	"io"
	"os"
	"path"
	"time"
)

// Reader is a simple implementation of tail -f
type Reader struct {
	io.ReadCloser
}

// Since we always have to spin up a control file, what we'll do is tie this all in to the handler
// Then we can just call 'input' the same as always, and remove the 'start' thing from everywhere but the main loop
// Calls to "NewInput" will work just fine, and we can run this all in one thread, managing each tail
// Return an input type even, calling start will just actually add it to the stack
// That way we can assure paths are good and the API doesn't have to actually change
// StartContext might be considered insteresting; we can check the contexts in the loop as well, and use that to clean up instead of something like input.Close()

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
